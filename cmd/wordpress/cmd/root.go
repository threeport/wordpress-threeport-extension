// generated by 'threeport-sdk gen' but will not be regenerated - intended for modification

package cmd

import (
	cobra "github.com/spf13/cobra"
	cli "github.com/threeport/threeport/pkg/cli/v0"
)

var CliArgs = &cli.GenesisControlPlaneCLIArgs{}

// WordpressCmd represents the wordpress command which is the root command for
// the wordpress plugin.
var WordpressCmd = &cobra.Command{
	Long:  "Manage the Wordpress Threeport extension",
	Short: "Manage the Wordpress Threeport extension",
	Use:   "wordpress",
}
