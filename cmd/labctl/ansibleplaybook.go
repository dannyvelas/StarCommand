package main

import (
	"github.com/spf13/cobra"
)

func ansibleCmd() *cobra.Command {
	ansibleCmd := &cobra.Command{
		Use:   "ansible",
		Short: "Display information for a resource",
	}

	ansibleCmd.AddCommand(ansibleRunCmd())

	return ansibleCmd
}
