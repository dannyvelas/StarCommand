package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func setSSHCmd() *cobra.Command {
	setSSHCmd := &cobra.Command{
		Use:   "config <host-name>",
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			host := args[0]
			fmt.Printf("Generating JSON config for %s...\n", host)
		},
	}

	return setSSHCmd
}
