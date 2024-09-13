// generated by 'threeport-sdk gen' but will not be regenerated - intended for modification

package cmd

import cobra "github.com/spf13/cobra"

// GetCmd represents the get command
var GetCmd = &cobra.Command{
	Long:  "Get a Threeport Wordpress object.\n\n\tThe get command does nothing by itself.  Use one of the available subcommands\n\tto get different objects from the system.",
	Short: "Get a Threeport Wordpress object",
	Use:   "get",
}

func init() {
	WordpressCmd.AddCommand(GetCmd)
}
