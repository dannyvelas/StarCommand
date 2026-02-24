package main

import (
	"github.com/dannyvelas/conflux"
	"github.com/spf13/cobra"
)

func terraformCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "Execute terraform commands",
	}

	terraformCmd.AddCommand(terraformApplyCmd(configMux, preflight))

	return terraformCmd
}
