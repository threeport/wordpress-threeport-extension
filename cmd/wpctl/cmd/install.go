package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	cli "github.com/threeport/threeport/pkg/cli/v0"
	client "github.com/threeport/threeport/pkg/client/v0"
	config "github.com/threeport/threeport/pkg/config/v0"
	kube "github.com/threeport/threeport/pkg/kube/v0"
	installer "github.com/threeport/wordpress-threeport-extension/pkg/installer/v0"
	"gopkg.in/yaml.v2"
)

// installCmd represents install command
var installCmd = &cobra.Command{
	Use:          "install",
	Example:      "wpctl install",
	Short:        "Install the wordpress controller to an existing Threeport control plane",
	Long:         `Install the wordpress controller to an existing Threeport control plane.`,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// get threeport config
		configData, err := ioutil.ReadFile("/Users/lander2k2/.config/threeport/config.yaml")
		if err != nil {
			log.Fatalf("Error reading YAML file: %v", err)
		}

		var threeportConfig config.ThreeportConfig
		err = yaml.UnmarshalStrict(configData, &threeportConfig)
		if err != nil {
			log.Fatalf("Error unmarshalling YAML: %v", err)
		}

		// get Threeport API client and endpoint
		apiClient, err := threeportConfig.GetHTTPClient(threeportConfig.CurrentControlPlane)
		if err != nil {
			cli.Error("failed to get Threeport API client", err)
			os.Exit(1)
		}

		apiEndpoint, err := threeportConfig.GetThreeportAPIEndpoint(threeportConfig.CurrentControlPlane)
		if err != nil {
			cli.Error("failed to get Threeport API endpoint", err)
			os.Exit(1)
		}

		// get Kubernetes runtime instance for control plane
		queryString := "ThreeportControlPlaneHost=true"
		kubernetesRuntimeInstances, err := client.GetKubernetesRuntimeInstancesByQueryString(
			apiClient,
			apiEndpoint,
			queryString,
		)
		if err != nil {
			cli.Error("failed to get kubernetes runtime instances", err)
			os.Exit(1)
		}
		if len(*kubernetesRuntimeInstances) != 1 {
			cli.Error(fmt.Sprintf("found %d k8s runtime instances", len(*kubernetesRuntimeInstances)), err)
			os.Exit(1)
		}
		kubeRuntimes := *kubernetesRuntimeInstances

		// get encryption key
		encryptionKey, err := threeportConfig.GetThreeportEncryptionKey(threeportConfig.CurrentControlPlane)
		if err != nil {
			cli.Error("failed to get Threeport API encryption key", err)
			os.Exit(1)
		}

		// get Kubernetes client
		dynamicInterface, restMapper, err := kube.GetClient(
			&kubeRuntimes[0],
			false,
			apiClient,
			apiEndpoint,
			encryptionKey,
		)
		if err != nil {
			cli.Error("failed to get Kube client", err)
			os.Exit(1)
		}

		// create installer
		installer := installer.NewInstaller(dynamicInterface, restMapper)

		// install wordpress controller extension
		if err := installer.InstallWordpressExtension(); err != nil {
			cli.Error("failed to install wordpress extension", err)
			os.Exit(1)
		}

		cli.Complete("wordpress controller extension installed")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
