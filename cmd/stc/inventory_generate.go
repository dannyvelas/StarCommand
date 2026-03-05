package main

import (
	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func inventoryGenerateCmd(c *config.Config) *cobra.Command {
	var host string

	command := &cobra.Command{
		Use:   "generate",
		Short: "Generate the Ansible inventory file for all hosts, or a single host",
		RunE:  inventoryGenerateCLI(c, &host),
	}

	command.Flags().StringVar(&host, "host", "", "Limit generation to a single host")

	return command
}

func inventoryGenerateCLI(c *config.Config, host *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return app.InventoryGenerate(ctx, c, host)
	}
}
