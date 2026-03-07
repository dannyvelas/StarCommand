package main

import (
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func inventoryGenerateCmd(c *models.Config) *cobra.Command {
	command := &cobra.Command{
		Use:   "generate",
		Short: "Generate the Ansible inventory file for all hosts, or a single host",
		RunE:  inventoryGenerateCLI(c),
	}

	return command
}

func inventoryGenerateCLI(c *models.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return app.InventoryGenerate(ctx, c)
	}
}
