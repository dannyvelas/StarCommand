package main

import (
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func ansibleCmd(c *models.Config) *cobra.Command {
	ansibleCmd := &cobra.Command{
		Use:   "ansible",
		Short: "Execute ansible commands",
	}

	ansibleCmd.AddCommand(ansiblePlaybookCmd(c)...)

	return ansibleCmd
}
