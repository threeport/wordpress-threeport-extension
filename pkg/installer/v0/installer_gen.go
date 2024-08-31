// generated by 'threeport-sdk gen' - do not edit

package v0

import (
	"context"
	"fmt"
	tp_database "github.com/threeport/threeport/pkg/api-server/v0/database"
	kube "github.com/threeport/threeport/pkg/kube/v0"
	database "github.com/threeport/wordpress-threeport-extension/pkg/api-server/v0/database"
	errors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	dynamic "k8s.io/client-go/dynamic"
)

const (
	DbInitFilename            = "db.sql"
	DbInitLocation            = "/etc/threeport/db-create"
	defaultNamespace          = "threeport-wordpress"
	defaultThreeportNamespace = "threeport-control-plane"
	natsLabelSelector         = "app.kubernetes.io/name=nats"
)

// Installer contains the values needed for an extension installation.
type Installer struct {
	// dynamice interface client for Kubernetes API
	KubeClient dynamic.Interface

	// Kubernetes API REST mapper
	KubeRestMapper *meta.RESTMapper

	// The Kubernetes namespace to install the extension components in.
	ExtensionNamespace string

	// The Kubernetes namespace the Threeport control plane is installed in.
	ThreeportNamespace string
}

// NewInstaller returns a wordpress extension installer with default values.
func NewInstaller(
	kubeClient dynamic.Interface,
	restMapper *meta.RESTMapper,
) *Installer {
	defaultInstaller := Installer{
		ExtensionNamespace: defaultNamespace,
		KubeClient:         kubeClient,
		KubeRestMapper:     restMapper,
		ThreeportNamespace: defaultThreeportNamespace,
	}

	return &defaultInstaller
}

// InstallWordpressExtension installs the controller and API for the wordpress extension.
func (i *Installer) InstallWordpressExtension() error {
	// get NATS service name from cluster
	gvr := schema.GroupVersionResource{
		Group:    "",
		Resource: "services",
		Version:  "v1",
	}
	services, err := i.KubeClient.Resource(gvr).Namespace(i.ThreeportNamespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: natsLabelSelector},
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve NATS service name: %w", err)
	}
	if len(services.Items) != 1 {
		return fmt.Errorf(
			"expected one NATS service with label '%s' but found %d",
			natsLabelSelector,
			len(services.Items),
		)
	}
	natsServiceName := services.Items[0].GetName()

	// create namespace
	var namespace = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": i.ExtensionNamespace,
			},
		},
	}

	if _, err := kube.CreateOrUpdateResource(
		namespace,
		i.KubeClient,
		*i.KubeRestMapper,
	); err != nil {
		return fmt.Errorf("failed to create/update wordpress extension namespace: %w", err)
	}

	// copy secrets into extension namespace
	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"db-root-cert",
		i.ThreeportNamespace,
		i.ExtensionNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"db-threeport-cert",
		i.ThreeportNamespace,
		i.ExtensionNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"encryption-key",
		i.ThreeportNamespace,
		i.ExtensionNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	if err := copySecret(
		i.KubeClient,
		*i.KubeRestMapper,
		"controller-config",
		i.ThreeportNamespace,
		i.ExtensionNamespace,
	); err != nil {
		return fmt.Errorf("failed to copy secret: %w", err)
	}

	// create secret for database connection
	var apiSecret = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      "db-config",
				"namespace": i.ExtensionNamespace,
			},
			"stringData": map[string]interface{}{
				"env": fmt.Sprintf(
					"DB_HOST=%s.%s.svc.cluster.local\nDB_USER=%s\nDB_NAME=%s\nDB_PORT=%s\nDB_SSL_MODE=%s\nNATS_HOST=%s.%s.svc.cluster.local\nNATS_PORT=4222\n",
					tp_database.ThreeportDatabaseHost,
					i.ThreeportNamespace,
					tp_database.ThreeportDatabaseUser,
					database.ThreeportWordpressDatabaseName,
					tp_database.ThreeportDatabasePort,
					tp_database.ThreeportDatabaseSslMode,
					natsServiceName,
					i.ThreeportNamespace,
				),
			},
		},
	}
	if _, err := kube.CreateOrUpdateResource(apiSecret, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update API server secret for DB connection: %w", err)
	}

	// create configmap used to initialize API database
	var dbCreateConfig = &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"data": map[string]interface{}{
			"db.sql": "CREATE USER IF NOT EXISTS threeport;\nCREATE DATABASE IF NOT EXISTS threeport_wordpress_api encoding='utf-8';\nGRANT ALL ON DATABASE threeport_wordpress_api TO threeport;",
		},
		"kind": "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "db-create",
			"namespace": i.ExtensionNamespace,
		},
	}}

	if _, err := kube.CreateOrUpdateResource(dbCreateConfig, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wordpress DB initialization configmap: %w", err)
	}

	// install wordpress API server
	var wordpressApiDeploy = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "threeport-wordpress-api-server",
				"namespace": i.ExtensionNamespace,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name": "threeport-wordpress-api-server",
					},
				},
				"strategy": map[string]interface{}{
					"rollingUpdate": map[string]interface{}{
						"maxSurge":       "25%",
						"maxUnavailable": "25%",
					},
					"type": "RollingUpdate",
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": nil,
						"labels": map[string]interface{}{
							"app.kubernetes.io/name": "threeport-wordpress-api-server",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"args": []interface{}{
									"-auto-migrate=true",
									"-auth-enabled=false",
								},
								"command": []interface{}{
									"/rest-api",
								},
								"envFrom": []interface{}{
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "encryption-key",
										},
									},
								},
								"image":           "localhost:5001/threeport-wordpress-rest-api:dev",
								"imagePullPolicy": "IfNotPresent",
								"name":            "api-server",
								"ports": []interface{}{
									map[string]interface{}{
										"containerPort": 1323,
										"name":          "api",
										"protocol":      "TCP",
									},
								},
								"readinessProbe": map[string]interface{}{
									"failureThreshold": 1,
									"httpGet": map[string]interface{}{
										"path":   "/readyz",
										"port":   8081,
										"scheme": "HTTP",
									},
									"initialDelaySeconds": 1,
									"periodSeconds":       2,
									"successThreshold":    1,
									"timeoutSeconds":      1,
								},
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/etc/threeport/",
										"name":      "db-config",
									},
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-certs",
										"name":      "db-threeport-cert",
									},
								},
							},
						},
						"initContainers": []interface{}{
							map[string]interface{}{
								"command": []interface{}{
									"bash",
									"-c",
									fmt.Sprintf("cockroach sql --certs-dir=/etc/threeport/db-certs --host crdb.%s.svc.cluster.local --port 26257 -f /etc/threeport/db-create/db.sql", i.ThreeportNamespace),
								},
								"image":           "cockroachdb/cockroach:v23.1.14",
								"imagePullPolicy": "IfNotPresent",
								"name":            "db-init",
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-create",
										"name":      "db-create",
									},
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-certs",
										"name":      "db-root-cert",
									},
								},
							},
							map[string]interface{}{
								"args": []interface{}{
									"-env-file=/etc/threeport/env",
									"up",
								},
								"command": []interface{}{
									"/database-migrator",
								},
								"image":           "localhost:5001/threeport-wordpress-database-migrator:dev",
								"imagePullPolicy": "IfNotPresent",
								"name":            "database-migrator",
								"volumeMounts": []interface{}{
									map[string]interface{}{
										"mountPath": "/etc/threeport/",
										"name":      "db-config",
									},
									map[string]interface{}{
										"mountPath": "/etc/threeport/db-certs",
										"name":      "db-threeport-cert",
									},
								},
							},
						},
						"restartPolicy":                 "Always",
						"terminationGracePeriodSeconds": 30,
						"volumes": []interface{}{
							map[string]interface{}{
								"name": "db-root-cert",
								"secret": map[string]interface{}{
									"defaultMode": 420,
									"secretName":  "db-root-cert",
								},
							},
							map[string]interface{}{
								"name": "db-threeport-cert",
								"secret": map[string]interface{}{
									"defaultMode": 420,
									"secretName":  "db-threeport-cert",
								},
							},
							map[string]interface{}{
								"name": "db-config",
								"secret": map[string]interface{}{
									"defaultMode": 420,
									"secretName":  "db-config",
								},
							},
							map[string]interface{}{
								"configMap": map[string]interface{}{
									"defaultMode": 420,
									"name":        "db-create",
								},
								"name": "db-create",
							},
						},
					},
				},
			},
		},
	}

	if _, err := kube.CreateOrUpdateResource(wordpressApiDeploy, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wordpress API deployment: %w", err)
	}

	// install wordpress controller
	var wordpressControllerDeploy = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "threeport-wordpress-controller",
				"namespace": i.ExtensionNamespace,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app.kubernetes.io/name": "threeport-wordpress-controller",
					},
				},
				"strategy": map[string]interface{}{
					"rollingUpdate": map[string]interface{}{
						"maxSurge":       "25%",
						"maxUnavailable": "25%",
					},
					"type": "RollingUpdate",
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app.kubernetes.io/name": "threeport-wordpress-controller",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"args": []interface{}{
									"-auth-enabled=false",
								},
								"command": []interface{}{
									"/wordpress-controller",
								},
								"envFrom": []interface{}{
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "controller-config",
										},
									},
									map[string]interface{}{
										"secretRef": map[string]interface{}{
											"name": "encryption-key",
										},
									},
								},
								"image":           "localhost:5001/threeport-wordpress-controller:dev",
								"imagePullPolicy": "IfNotPresent",
								"name":            "wordpress-controller",
								"readinessProbe": map[string]interface{}{
									"failureThreshold": 1,
									"httpGet": map[string]interface{}{
										"path":   "/readyz",
										"port":   8081,
										"scheme": "HTTP",
									},
									"initialDelaySeconds": 1,
									"periodSeconds":       2,
									"successThreshold":    1,
									"timeoutSeconds":      1,
								},
							},
						},
						"restartPolicy":                 "Always",
						"terminationGracePeriodSeconds": 30,
					},
				},
			},
		},
	}

	if _, err := kube.CreateOrUpdateResource(wordpressControllerDeploy, i.KubeClient, *i.KubeRestMapper); err != nil {
		return fmt.Errorf("failed to create/update wordpress controller deployment: %w", err)
	}

	return nil
}

// copySecret copies a secret from one namespace to another.  The function
// returns without error if the secret already exists in the target namespace.
func copySecret(
	dynamicClient dynamic.Interface,
	restMapper meta.RESTMapper,
	secretName string,
	sourceNamespace string,
	targetNamespace string,
) error {
	secretGVR := schema.GroupVersionResource{
		Group:    "",
		Resource: "secrets",
		Version:  "v1",
	}
	secretGK := schema.GroupKind{
		Group: "",
		Kind:  "Secret",
	}

	mapping, err := restMapper.RESTMapping(secretGK, secretGVR.Version)
	if err != nil {
		return fmt.Errorf("failed to get RESTMapping for Secret resource: %w", err)
	}

	targetSecretResource := dynamicClient.Resource(mapping.Resource).Namespace(targetNamespace)
	_, err = targetSecretResource.Get(context.TODO(), secretName, metav1.GetOptions{})
	if err == nil {
		// secret already exists, return nil
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf(
			"failed to check if Secret '%s' exists in namespace '%s': %w",
			secretName,
			targetNamespace,
			err,
		)
	}

	secretResource := dynamicClient.Resource(mapping.Resource).Namespace(sourceNamespace)
	secret, err := secretResource.Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf(
			"failed to get Secret '%s' from namespace '%s': %w",
			secretName,
			sourceNamespace,
			err,
		)
	}

	secret.SetNamespace(targetNamespace)
	secret.SetResourceVersion("")
	secret.SetUID("")
	secret.SetSelfLink("")
	secret.SetCreationTimestamp(metav1.Time{})
	secret.SetManagedFields(nil)

	_, err = targetSecretResource.Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create/update Secret in namespace '%s': %w", targetNamespace, err)
	}

	return nil
}
