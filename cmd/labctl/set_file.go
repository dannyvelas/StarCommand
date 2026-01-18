package main

import (
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/internal/app"
	"github.com/spf13/cobra"
)

func setFileCmd() *cobra.Command {
	var targets []string

	setFileCmd := &cobra.Command{
		Use:       "file <host-alias>",
		ValidArgs: []string{"proxmox"},
		Short:     "Update the `~/.ssh/config` file to connect to a given host",
		Args:      cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			a, err := app.New(hostAlias, targets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			diagnostics, err := a.SetFile()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			for _, diagnostic := range diagnostics {
				fmt.Println(diagnostic)
			}

			fmt.Println("SSH config updated successfully!")
		},
	}

	setFileCmd.Flags().StringSliceVar(&targets, "for", []string{"ssh"}, "Write or append to the corresponding file")

	return setFileCmd
}
