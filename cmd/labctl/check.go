package main

import (
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Get diagnostics for a resource",
	}

	checkCmd.AddCommand(checkReqsCmd())

	return checkCmd
}
