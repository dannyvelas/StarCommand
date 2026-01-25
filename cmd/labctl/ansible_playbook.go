package main

import (
	"github.com/spf13/cobra"
)

func ansiblePlaybookCmd() *cobra.Command {
	ansibleCmd := &cobra.Command{
		Use: "playbook",
		// TODO: fix
		Short: "Display information for a resource",
	}

	ansibleCmd.AddCommand(ansiblePlaybookRunCmd())

	return ansibleCmd
}
