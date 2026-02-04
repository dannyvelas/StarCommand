package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "labctl",
	Short: "An internal CLI to configure a homelab",
}

// execute adds all child commands to the root command and sets flags appropriately.
// this is called by main.main(). It only needs to happen once to the rootCmd.
func execute(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func initialize() {
	rootCmd.AddCommand(ansibleCmd())
	rootCmd.AddCommand(sshCmd())
	rootCmd.AddCommand(terraformCmd())
	rootCmd.AddCommand(checkCmd())
}
