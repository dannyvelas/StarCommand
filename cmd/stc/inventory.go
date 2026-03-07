package main

import (
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func inventoryCmd(c *models.Config) *cobra.Command {
	inventoryCmd := &cobra.Command{
		Use:   "inventory",
		Short: "Execute ansible inventory commands",
	}

	inventoryCmd.AddCommand(inventoryGenerateCmd(c))

	return inventoryCmd
}
