package main

import (
	"github.com/dannyvelas/starcommand/internal/config"
	"github.com/spf13/cobra"
)

func rootCmd(c *config.Config) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "stc",
		Short: "Scaffold production infrastructure",
	}

	rootCmd.AddCommand(inventoryCmd(c))
	rootCmd.AddCommand(ansibleCmd(c))
	rootCmd.AddCommand(sshCmd(c))
	rootCmd.AddCommand(terraformCmd(c))

	return rootCmd
}
