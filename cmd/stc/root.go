package main

import (
	"github.com/dannyvelas/starcommand/config"
	"github.com/spf13/cobra"
)

func rootCmd(c *config.Config) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "stc",
		Short: "Scaffold production infrastructure",
	}

	// get preflight flag
	var preflight bool
	rootCmd.PersistentFlags().BoolVar(&preflight, "preflight", false, "Display config diagnostic table instead of executing")

	rootCmd.AddCommand(inventoryCmd(c, preflight))
	rootCmd.AddCommand(ansibleCmd(c, preflight))
	rootCmd.AddCommand(sshCmd(c, preflight))
	rootCmd.AddCommand(terraformCmd(c, preflight))

	return rootCmd
}
