package main

import (
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func terraformCmd(c *models.Config) *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:   "terraform",
		Short: "Execute terraform commands",
	}

	terraformCmd.AddCommand(terraformApplyCmd(c))

	return terraformCmd
}
