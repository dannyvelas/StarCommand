package main

import (
	"os"

	"github.com/dannyvelas/homelab/internal/env"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "labctl",
	Short: "An internal CLI to configure a homelab",
}

// execute adds all child commands to the root command and sets flags appropriately.
// this is called by main.main(). It only needs to happen once to the rootCmd.
func execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initialize(env env.Env) {
	var verbose bool
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.AddCommand(newResolveCmd(env, verbose))
}
