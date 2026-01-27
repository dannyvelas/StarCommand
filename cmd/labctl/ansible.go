package main

import (
	"github.com/spf13/cobra"
)

func ansibleCmd() *cobra.Command {
	ansibleCmd := &cobra.Command{
		Use:   "ansible",
		Short: "Execute ansible commands",
	}

	ansibleCmd.AddCommand(ansiblePlaybookCmd())

	return ansibleCmd
}
