package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Generate a JSON object of configuration values for a given host",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		fmt.Printf("verbose mode: %t.\nhost: %s.\n", verbose, host)
	},
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}
