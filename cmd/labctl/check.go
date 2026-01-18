package main

import (
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check whether some resource is in a valid state",
	}

	checkCmd.AddCommand(checkConfigCmd())

	return checkCmd
}
