package main

import (
	"github.com/dannyvelas/conflux"
	"github.com/spf13/cobra"
)

func rootCmd(configMux *conflux.ConfigMux) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "iac",
		Short: "Scaffold production infrastructure",
	}

	// get preflight flag
	var preflight bool
	rootCmd.PersistentFlags().BoolVar(&preflight, "preflight", false, "Display config diagnostic table instead of executing")

	rootCmd.AddCommand(inventoryCmd(configMux, preflight))
	rootCmd.AddCommand(ansibleCmd(configMux, preflight))
	rootCmd.AddCommand(sshCmd(configMux, preflight))
	rootCmd.AddCommand(terraformCmd(configMux, preflight))

	return rootCmd
}
