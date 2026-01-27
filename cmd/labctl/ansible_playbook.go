package main

import (
	"github.com/spf13/cobra"
)

func ansiblePlaybookCmd() *cobra.Command {
	ansibleCmd := &cobra.Command{
		Use:   "playbook",
		Short: "Execute ansible playbook commands",
	}

	ansibleCmd.AddCommand(ansiblePlaybookRunCmd())

	return ansibleCmd
}
