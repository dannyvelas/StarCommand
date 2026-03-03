package main

import (
	"github.com/dannyvelas/starcommand/config"
	"github.com/spf13/cobra"
)

func inventoryCmd(c *config.Config, preflight bool) *cobra.Command {
	inventoryCmd := &cobra.Command{
		Use:   "inventory",
		Short: "Execute ansible inventory commands",
	}

	inventoryCmd.AddCommand(inventoryGenerateCmd(c, preflight))

	return inventoryCmd
}
