package main

import (
	"github.com/spf13/cobra"
)

func terraformCmd() *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "Execute terraform commands",
	}

	terraformCmd.AddCommand(terraformApplyCmd())

	return terraformCmd
}
