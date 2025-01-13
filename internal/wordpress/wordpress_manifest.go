package wordpress

import (
	"errors"
	"fmt"

	tpapi "github.com/threeport/threeport/pkg/api/v0"
	kube "github.com/threeport/threeport/pkg/kube/v0"
	util "github.com/threeport/threeport/pkg/util/v0"
	"gorm.io/datatypes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// wordpressYaml returns a YAML manifest for the wordpress workload.
func wordpressYaml(
	definitionName string,
	wordpressReplicas int,
	environment string,
	managedDatabase bool,
	dbStorageGb int,
	dbConnectionSecret string,
) (string, error) {
	var yamlDoc string

	if !managedDatabase {
		var unmanagedDbYamlDoc string

		var serviceAccountMariadb = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-mariadb", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "mariadb",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"environment":                  environment,
					},
				},
				"automountServiceAccountToken": false,
			},
		}
		unmanagedDbYamlDoc, err := kube.AppendObjectToYamlDoc(serviceAccountMariadb, unmanagedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		var secretMariadb = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-mariadb", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "mariadb",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"environment":                  environment,
					},
				},
				"type": "Opaque",
				"data": map[string]interface{}{
					"mariadb-root-password": "WHZOWUhMZ3RFUw==",
					"mariadb-password":      "VHlycG1KVDVPTg==",
				},
			},
		}
		unmanagedDbYamlDoc, err = kube.AppendObjectToYamlDoc(secretMariadb, unmanagedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		var configMapMariadb = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-mariadb", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "mariadb",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"app.kubernetes.io/component":  "primary",
						"environment":                  environment,
					},
				},
				"data": map[string]interface{}{
					"my.cnf": `[mysqld]
skip-name-resolve
explicit_defaults_for_timestamp
basedir=/opt/bitnami/mariadb
plugin_dir=/opt/bitnami/mariadb/plugin
port=3306
socket=/opt/bitnami/mariadb/tmp/mysql.sock
tmpdir=/opt/bitnami/mariadb/tmp
max_allowed_packet=16M
bind-address=*
pid-file=/opt/bitnami/mariadb/tmp/mysqld.pid
log-error=/opt/bitnami/mariadb/logs/mysqld.log
character-set-server=UTF8
collation-server=utf8_general_ci
slow_query_log=0
slow_query_log_file=/opt/bitnami/mariadb/logs/mysqld.log
long_query_time=10.0

[client]
port=3306
socket=/opt/bitnami/mariadb/tmp/mysql.sock
default-character-set=UTF8
plugin_dir=/opt/bitnami/mariadb/plugin

[manager]
port=3306
socket=/opt/bitnami/mariadb/tmp/mysql.sock
pid-file=/opt/bitnami/mariadb/tmp/mysqld.pid`,
				},
			},
		}
		unmanagedDbYamlDoc, err = kube.AppendObjectToYamlDoc(configMapMariadb, unmanagedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		var serviceMariadb = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-mariadb", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "mariadb",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"app.kubernetes.io/component":  "primary",
						"environment":                  environment,
					},
					"annotations": nil,
				},
				"spec": map[string]interface{}{
					"type":            "ClusterIP",
					"sessionAffinity": "None",
					"ports": []interface{}{
						map[string]interface{}{
							"name":       "mysql",
							"port":       3306,
							"protocol":   "TCP",
							"targetPort": "mysql",
							"nodePort":   nil,
						},
					},
					"selector": map[string]interface{}{
						"app.kubernetes.io/name":      "mariadb",
						"app.kubernetes.io/instance":  definitionName,
						"app.kubernetes.io/component": "primary",
					},
				},
			},
		}
		unmanagedDbYamlDoc, err = kube.AppendObjectToYamlDoc(serviceMariadb, unmanagedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		var statefulSetMariadb = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "StatefulSet",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-mariadb", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "mariadb",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"app.kubernetes.io/component":  "primary",
						"environment":                  environment,
					},
				},
				"spec": map[string]interface{}{
					"replicas":             1,
					"revisionHistoryLimit": 10,
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app.kubernetes.io/name":      "mariadb",
							"app.kubernetes.io/instance":  definitionName,
							"app.kubernetes.io/component": "primary",
						},
					},
					"serviceName": fmt.Sprintf("%s-mariadb", definitionName),
					"updateStrategy": map[string]interface{}{
						"type": "RollingUpdate",
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								"checksum/configuration": "abe9c954f29a801817e9c9bae83f5353a24b42f21603fd18da496edd12991d82",
							},
							"labels": map[string]interface{}{
								"app.kubernetes.io/name":       "mariadb",
								"app.kubernetes.io/instance":   definitionName,
								"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
								"app.kubernetes.io/component":  "primary",
								"environment":                  environment,
							},
						},
						"spec": map[string]interface{}{
							"serviceAccountName": fmt.Sprintf("%s-mariadb", definitionName),
							"affinity": map[string]interface{}{
								"podAffinity": nil,
								"podAntiAffinity": map[string]interface{}{
									"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
										map[string]interface{}{
											"podAffinityTerm": map[string]interface{}{
												"labelSelector": map[string]interface{}{
													"matchLabels": map[string]interface{}{
														"app.kubernetes.io/name":      "mariadb",
														"app.kubernetes.io/instance":  definitionName,
														"app.kubernetes.io/component": "primary",
													},
												},
												"topologyKey": "kubernetes.io/hostname",
											},
											"weight": 1,
										},
									},
								},
								"nodeAffinity": nil,
							},
							"securityContext": map[string]interface{}{
								"fsGroup": 1001,
							},
							"containers": []interface{}{
								map[string]interface{}{
									"name":            "mariadb",
									"image":           "docker.io/bitnami/mariadb:10.11.3-debian-11-r0",
									"imagePullPolicy": "IfNotPresent",
									"securityContext": map[string]interface{}{
										"allowPrivilegeEscalation": false,
										"privileged":               false,
										"runAsNonRoot":             true,
										"runAsUser":                1001,
									},
									"env": []interface{}{
										map[string]interface{}{
											"name":  "BITNAMI_DEBUG",
											"value": "false",
										},
										map[string]interface{}{
											"name": "MARIADB_ROOT_PASSWORD",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": fmt.Sprintf("%s-mariadb", definitionName),
													"key":  "mariadb-root-password",
												},
											},
										},
										map[string]interface{}{
											"name":  "MARIADB_USER",
											"value": "bn_wordpress",
										},
										map[string]interface{}{
											"name": "MARIADB_PASSWORD",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": fmt.Sprintf("%s-mariadb", definitionName),
													"key":  "mariadb-password",
												},
											},
										},
										map[string]interface{}{
											"name":  "MARIADB_DATABASE",
											"value": "bitnami_wordpress",
										},
									},
									"ports": []interface{}{
										map[string]interface{}{
											"name":          "mysql",
											"containerPort": 3306,
										},
									},
									"livenessProbe": map[string]interface{}{
										"failureThreshold":    3,
										"initialDelaySeconds": 120,
										"periodSeconds":       10,
										"successThreshold":    1,
										"timeoutSeconds":      1,
										"exec": map[string]interface{}{
											"command": []interface{}{
												"/bin/bash",
												"-ec",
												`password_aux="${MARIADB_ROOT_PASSWORD:-}"
if [[ -f "${MARIADB_ROOT_PASSWORD_FILE:-}" ]]; then
	password_aux=$(cat "$MARIADB_ROOT_PASSWORD_FILE")
fi
mysqladmin status -uroot -p"${password_aux}"
`,
											},
										},
									},
									"readinessProbe": map[string]interface{}{
										"failureThreshold":    3,
										"initialDelaySeconds": 30,
										"periodSeconds":       10,
										"successThreshold":    1,
										"timeoutSeconds":      1,
										"exec": map[string]interface{}{
											"command": []interface{}{
												"/bin/bash",
												"-ec",
												`password_aux="${MARIADB_ROOT_PASSWORD:-}"
if [[ -f "${MARIADB_ROOT_PASSWORD_FILE:-}" ]]; then
	password_aux=$(cat "$MARIADB_ROOT_PASSWORD_FILE")
fi
mysqladmin status -uroot -p"${password_aux}"
`,
											},
										},
									},
									"resources": map[string]interface{}{
										"limits":   map[string]interface{}{},
										"requests": map[string]interface{}{},
									},
									"volumeMounts": []interface{}{
										map[string]interface{}{
											"name":      "data",
											"mountPath": "/bitnami/mariadb",
										},
										map[string]interface{}{
											"name":      "config",
											"mountPath": "/opt/bitnami/mariadb/conf/my.cnf",
											"subPath":   "my.cnf",
										},
									},
								},
							},
							"volumes": []interface{}{
								map[string]interface{}{
									"name": "config",
									"configMap": map[string]interface{}{
										"name": fmt.Sprintf("%s-mariadb", definitionName),
									},
								},
							},
						},
					},
					"volumeClaimTemplates": []interface{}{
						map[string]interface{}{
							"metadata": map[string]interface{}{
								"name": "data",
								"labels": map[string]interface{}{
									"app.kubernetes.io/name":      "mariadb",
									"app.kubernetes.io/instance":  definitionName,
									"app.kubernetes.io/component": "primary",
									"environment":                 environment,
								},
							},
							"spec": map[string]interface{}{
								"accessModes": []interface{}{
									"ReadWriteOnce",
								},
								"resources": map[string]interface{}{
									"requests": map[string]interface{}{
										"storage": fmt.Sprintf("%dGi", dbStorageGb),
									},
								},
							},
						},
					},
				},
			},
		}
		unmanagedDbYamlDoc, err = kube.AppendObjectToYamlDoc(statefulSetMariadb, unmanagedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		var deploymentWordpress = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-wordpress", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "wordpress",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"environment":                  environment,
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app.kubernetes.io/name":     "wordpress",
							"app.kubernetes.io/instance": definitionName,
						},
					},
					"strategy": map[string]interface{}{
						"type": "RollingUpdate",
					},
					"replicas": wordpressReplicas,
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app.kubernetes.io/name":       "wordpress",
								"app.kubernetes.io/instance":   definitionName,
								"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
								"environment":                  environment,
							},
						},
						"spec": map[string]interface{}{
							// yamllint disable rule:indentation
							"hostAliases": []interface{}{
								map[string]interface{}{
									"hostnames": []interface{}{
										"status.localhost",
									},
									"ip": "127.0.0.1",
								},
							},
							// yamllint enable rule:indentation
							"affinity": map[string]interface{}{
								"podAffinity": nil,
								"podAntiAffinity": map[string]interface{}{
									"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
										map[string]interface{}{
											"podAffinityTerm": map[string]interface{}{
												"labelSelector": map[string]interface{}{
													"matchLabels": map[string]interface{}{
														"app.kubernetes.io/name":     "wordpress",
														"app.kubernetes.io/instance": definitionName,
													},
												},
												"topologyKey": "kubernetes.io/hostname",
											},
											"weight": 1,
										},
									},
								},
								"nodeAffinity": nil,
							},
							"securityContext": map[string]interface{}{
								"fsGroup": 1001,
								"seccompProfile": map[string]interface{}{
									"type": "RuntimeDefault",
								},
							},
							"serviceAccountName": "default",
							"containers": []interface{}{
								map[string]interface{}{
									"name":            "wordpress",
									"image":           "docker.io/bitnami/wordpress:6.2.0-debian-11-r22",
									"imagePullPolicy": "IfNotPresent",
									"securityContext": map[string]interface{}{
										"allowPrivilegeEscalation": false,
										"capabilities": map[string]interface{}{
											"drop": []interface{}{
												"ALL",
											},
										},
										"runAsNonRoot": true,
										"runAsUser":    1001,
									},
									"env": []interface{}{
										map[string]interface{}{
											"name":  "BITNAMI_DEBUG",
											"value": "false",
										},
										map[string]interface{}{
											"name":  "ALLOW_EMPTY_PASSWORD",
											"value": "yes",
										},
										map[string]interface{}{
											"name":  "MARIADB_HOST",
											"value": fmt.Sprintf("%s-mariadb", definitionName),
										},
										map[string]interface{}{
											"name":  "MARIADB_PORT_NUMBER",
											"value": "3306",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_DATABASE_NAME",
											"value": "bitnami_wordpress",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_DATABASE_USER",
											"value": "bn_wordpress",
										},
										map[string]interface{}{
											"name": "WORDPRESS_DATABASE_PASSWORD",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": fmt.Sprintf("%s-mariadb", definitionName),
													"key":  "mariadb-password",
												},
											},
										},
										map[string]interface{}{
											"name":  "WORDPRESS_USERNAME",
											"value": "user",
										},
										map[string]interface{}{
											"name": "WORDPRESS_PASSWORD",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": fmt.Sprintf("%s-wordpress", definitionName),
													"key":  "wordpress-password",
												},
											},
										},
										map[string]interface{}{
											"name":  "WORDPRESS_EMAIL",
											"value": "user@example.com",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_FIRST_NAME",
											"value": "FirstName",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_LAST_NAME",
											"value": "LastName",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_HTACCESS_OVERRIDE_NONE",
											"value": "no",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_ENABLE_HTACCESS_PERSISTENCE",
											"value": "no",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_BLOG_NAME",
											"value": "User's Blog!",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_SKIP_BOOTSTRAP",
											"value": "no",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_TABLE_PREFIX",
											"value": "wp_",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_SCHEME",
											"value": "http",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_EXTRA_WP_CONFIG_CONTENT",
											"value": "",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_PLUGINS",
											"value": "none",
										},
										map[string]interface{}{
											"name":  "APACHE_HTTP_PORT_NUMBER",
											"value": "8080",
										},
										map[string]interface{}{
											"name":  "APACHE_HTTPS_PORT_NUMBER",
											"value": "8443",
										},
									},
									"envFrom": nil,
									"ports": []interface{}{
										map[string]interface{}{
											"name":          "http",
											"containerPort": 8080,
										},
										map[string]interface{}{
											"name":          "https",
											"containerPort": 8443,
										},
									},
									"livenessProbe": map[string]interface{}{
										"failureThreshold": 6,
										"httpGet": map[string]interface{}{
											"httpHeaders": []interface{}{},
											"path":        "/wp-admin/install.php",
											"port":        "http",
											"scheme":      "HTTP",
										},
										"initialDelaySeconds": 120,
										"periodSeconds":       10,
										"successThreshold":    1,
										"timeoutSeconds":      5,
									},
									"readinessProbe": map[string]interface{}{
										"failureThreshold": 6,
										"httpGet": map[string]interface{}{
											"httpHeaders": []interface{}{},
											"path":        "/wp-login.php",
											"port":        "http",
											"scheme":      "HTTP",
										},
										"initialDelaySeconds": 30,
										"periodSeconds":       10,
										"successThreshold":    1,
										"timeoutSeconds":      5,
									},
									"resources": map[string]interface{}{
										"limits": map[string]interface{}{},
										"requests": map[string]interface{}{
											"cpu":    "300m",
											"memory": "512Mi",
										},
									},
									"volumeMounts": []interface{}{
										map[string]interface{}{
											"mountPath": "/bitnami/wordpress",
											"name":      "wordpress-data",
											"subPath":   "wordpress",
										},
									},
								},
							},
							"volumes": []interface{}{
								map[string]interface{}{
									"name": "wordpress-data",
									"persistentVolumeClaim": map[string]interface{}{
										"claimName": fmt.Sprintf("%s-wordpress", definitionName),
									},
								},
							},
						},
					},
				},
			},
		}
		unmanagedDbYamlDoc, err = kube.AppendObjectToYamlDoc(deploymentWordpress, unmanagedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		yamlDoc = unmanagedDbYamlDoc
	} else {
		var managedDbYamlDoc string

		var deploymentWordpress = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      fmt.Sprintf("%s-wordpress", definitionName),
					"namespace": "default",
					"labels": map[string]interface{}{
						"app.kubernetes.io/name":       "wordpress",
						"app.kubernetes.io/instance":   definitionName,
						"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
						"environment":                  environment,
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app.kubernetes.io/name":     "wordpress",
							"app.kubernetes.io/instance": definitionName,
						},
					},
					"strategy": map[string]interface{}{
						"type": "RollingUpdate",
					},
					"replicas": wordpressReplicas,
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app.kubernetes.io/name":       "wordpress",
								"app.kubernetes.io/instance":   definitionName,
								"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
								"environment":                  environment,
							},
						},
						"spec": map[string]interface{}{
							// yamllint disable rule:indentation
							"hostAliases": []interface{}{
								map[string]interface{}{
									"hostnames": []interface{}{
										"status.localhost",
									},
									"ip": "127.0.0.1",
								},
							},
							// yamllint enable rule:indentation
							"affinity": map[string]interface{}{
								"podAffinity": nil,
								"podAntiAffinity": map[string]interface{}{
									"preferredDuringSchedulingIgnoredDuringExecution": []interface{}{
										map[string]interface{}{
											"podAffinityTerm": map[string]interface{}{
												"labelSelector": map[string]interface{}{
													"matchLabels": map[string]interface{}{
														"app.kubernetes.io/name":     "wordpress",
														"app.kubernetes.io/instance": definitionName,
													},
												},
												"topologyKey": "kubernetes.io/hostname",
											},
											"weight": 1,
										},
									},
								},
								"nodeAffinity": nil,
							},
							"securityContext": map[string]interface{}{
								"fsGroup": 1001,
								"seccompProfile": map[string]interface{}{
									"type": "RuntimeDefault",
								},
							},
							"serviceAccountName": "default",
							"containers": []interface{}{
								map[string]interface{}{
									"name":            "wordpress",
									"image":           "docker.io/bitnami/wordpress:6.2.0-debian-11-r22",
									"imagePullPolicy": "IfNotPresent",
									"securityContext": map[string]interface{}{
										"allowPrivilegeEscalation": false,
										"capabilities": map[string]interface{}{
											"drop": []interface{}{
												"ALL",
											},
										},
										"runAsNonRoot": true,
										"runAsUser":    1001,
									},
									"env": []interface{}{
										map[string]interface{}{
											"name":  "BITNAMI_DEBUG",
											"value": "false",
										},
										map[string]interface{}{
											"name":  "ALLOW_EMPTY_PASSWORD",
											"value": "yes",
										},
										map[string]interface{}{
											"name": "MARIADB_HOST",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": dbConnectionSecret,
													"key":  "db-endpoint",
												},
											},
										},
										map[string]interface{}{
											"name": "MARIADB_PORT_NUMBER",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": dbConnectionSecret,
													"key":  "db-port",
												},
											},
										},
										map[string]interface{}{
											"name": "WORDPRESS_DATABASE_NAME",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": dbConnectionSecret,
													"key":  "db-name",
												},
											},
										},
										map[string]interface{}{
											"name": "WORDPRESS_DATABASE_USER",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": dbConnectionSecret,
													"key":  "db-user",
												},
											},
										},
										map[string]interface{}{
											"name": "WORDPRESS_DATABASE_PASSWORD",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": dbConnectionSecret,
													"key":  "db-password",
												},
											},
										},
										map[string]interface{}{
											"name":  "WORDPRESS_USERNAME",
											"value": "user",
										},
										map[string]interface{}{
											"name": "WORDPRESS_PASSWORD",
											"valueFrom": map[string]interface{}{
												"secretKeyRef": map[string]interface{}{
													"name": fmt.Sprintf("%s-wordpress", definitionName),
													"key":  "wordpress-password",
												},
											},
										},
										map[string]interface{}{
											"name":  "WORDPRESS_EMAIL",
											"value": "user@example.com",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_FIRST_NAME",
											"value": "FirstName",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_LAST_NAME",
											"value": "LastName",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_HTACCESS_OVERRIDE_NONE",
											"value": "no",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_ENABLE_HTACCESS_PERSISTENCE",
											"value": "no",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_BLOG_NAME",
											"value": "User's Blog!",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_SKIP_BOOTSTRAP",
											"value": "no",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_TABLE_PREFIX",
											"value": "wp_",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_SCHEME",
											"value": "http",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_EXTRA_WP_CONFIG_CONTENT",
											"value": "",
										},
										map[string]interface{}{
											"name":  "WORDPRESS_PLUGINS",
											"value": "none",
										},
										map[string]interface{}{
											"name":  "APACHE_HTTP_PORT_NUMBER",
											"value": "8080",
										},
										map[string]interface{}{
											"name":  "APACHE_HTTPS_PORT_NUMBER",
											"value": "8443",
										},
									},
									"envFrom": nil,
									"ports": []interface{}{
										map[string]interface{}{
											"name":          "http",
											"containerPort": 8080,
										},
										map[string]interface{}{
											"name":          "https",
											"containerPort": 8443,
										},
									},
									"livenessProbe": map[string]interface{}{
										"failureThreshold": 6,
										"httpGet": map[string]interface{}{
											"httpHeaders": []interface{}{},
											"path":        "/wp-admin/install.php",
											"port":        "http",
											"scheme":      "HTTP",
										},
										"initialDelaySeconds": 120,
										"periodSeconds":       10,
										"successThreshold":    1,
										"timeoutSeconds":      5,
									},
									"readinessProbe": map[string]interface{}{
										"failureThreshold": 6,
										"httpGet": map[string]interface{}{
											"httpHeaders": []interface{}{},
											"path":        "/wp-login.php",
											"port":        "http",
											"scheme":      "HTTP",
										},
										"initialDelaySeconds": 30,
										"periodSeconds":       10,
										"successThreshold":    1,
										"timeoutSeconds":      5,
									},
									"resources": map[string]interface{}{
										"limits": map[string]interface{}{},
										"requests": map[string]interface{}{
											"cpu":    "300m",
											"memory": "512Mi",
										},
									},
									"volumeMounts": []interface{}{
										map[string]interface{}{
											"mountPath": "/bitnami/wordpress",
											"name":      "wordpress-data",
											"subPath":   "wordpress",
										},
									},
								},
							},
							"volumes": []interface{}{
								map[string]interface{}{
									"name": "wordpress-data",
									"persistentVolumeClaim": map[string]interface{}{
										"claimName": fmt.Sprintf("%s-wordpress", definitionName),
									},
								},
							},
						},
					},
				},
			},
		}
		managedDbYamlDoc, err := kube.AppendObjectToYamlDoc(deploymentWordpress, managedDbYamlDoc)
		if err != nil {
			return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
		}

		yamlDoc = managedDbYamlDoc
	}

	var secretWordpress = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-wordpress", definitionName),
				"namespace": "default",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       "wordpress",
					"app.kubernetes.io/instance":   definitionName,
					"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
					"environment":                  environment,
				},
			},
			"type": "Opaque",
			"data": map[string]interface{}{
				"wordpress-password": "VkR5MUJhSno5Uw==",
			},
		},
	}
	yamlDoc, err := kube.AppendObjectToYamlDoc(secretWordpress, yamlDoc)
	if err != nil {
		return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
	}

	var serviceWordpress = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      getWordpressServiceName(definitionName),
				"namespace": "default",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       "wordpress",
					"app.kubernetes.io/instance":   definitionName,
					"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
					"environment":                  environment,
				},
			},
			"spec": map[string]interface{}{
				"type":            "ClusterIP",
				"sessionAffinity": "None",
				"ports": []interface{}{
					map[string]interface{}{
						"name":       "http",
						"port":       80,
						"protocol":   "TCP",
						"targetPort": "http",
					},
					map[string]interface{}{
						"name":       "https",
						"port":       443,
						"protocol":   "TCP",
						"targetPort": "https",
					},
				},
				"selector": map[string]interface{}{
					"app.kubernetes.io/name":     "wordpress",
					"app.kubernetes.io/instance": definitionName,
				},
			},
		},
	}
	yamlDoc, err = kube.AppendObjectToYamlDoc(serviceWordpress, yamlDoc)
	if err != nil {
		return yamlDoc, fmt.Errorf("failed to append object to YAML manifest: %w", err)
	}

	return yamlDoc, nil
}

// getWordpressServiceName returns the WordPress deployment's service name.
func getWordpressServiceName(definitionName string) string {
	return fmt.Sprintf("%s-wordpress", definitionName)
}

// getPvcManifest returns the JSON for the persistent volume claim with the
// storage class name set for the infra provider.
func getPvcManifest(
	infraProvider string,
	definitionName string,
	environment string,
) (*datatypes.JSON, error) {
	var storageClassName string
	switch infraProvider {
	case tpapi.KubernetesRuntimeInfraProviderKind:
		storageClassName = "standard"
	case tpapi.KubernetesRuntimeInfraProviderEKS:
		storageClassName = "gp2"
	default:
		return nil, errors.New("unrecognized infra provider")
	}

	var persistentVolumeClaimWordpress = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "PersistentVolumeClaim",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-wordpress", definitionName),
				"namespace": "default",
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       "wordpress",
					"app.kubernetes.io/instance":   definitionName,
					"app.kubernetes.io/managed-by": "wordrpess-threepport-extension",
					"environment":                  environment,
				},
			},
			"spec": map[string]interface{}{
				"storageClassName": storageClassName,
				"accessModes": []interface{}{
					"ReadWriteOnce",
				},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"storage": "10Gi",
					},
				},
			},
		},
	}

	jsonData, err := util.UnstructuredToDatatypesJson(persistentVolumeClaimWordpress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JSON for PVC: %w", err)
	}

	return &jsonData, nil
}
