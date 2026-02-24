package main

import (
	"github.com/dannyvelas/conflux"
	"github.com/spf13/cobra"
)

func ansibleCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	ansibleCmd := &cobra.Command{
		Use:   "ansible",
		Short: "Execute ansible commands",
	}

	ansibleCmd.AddCommand(ansiblePlaybookCmd(configMux, preflight)...)

	return ansibleCmd
}
