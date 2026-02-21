package main

import (
	"github.com/dannyvelas/conflux"
	"github.com/spf13/cobra"
)

func inventoryCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	inventoryCmd := &cobra.Command{
		Use:   "inventory",
		Short: "Execute ansible inventory commands",
	}

	inventoryCmd.AddCommand(inventoryGenerateCmd(configMux, preflight))

	return inventoryCmd
}
