package main

import (
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func rootCmd(c *models.Config) *cobra.Command {
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
