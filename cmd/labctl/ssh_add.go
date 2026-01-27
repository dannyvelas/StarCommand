package main

import (
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/app"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/spf13/cobra"
)

func sshAddCmd() *cobra.Command {
	sshAddCmd := &cobra.Command{
		Use:   "add <host-alias>",
		Short: "Update the `~/.ssh/config` file to connect to a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			diagnostics, err := app.SSHAdd(configMux, hostAlias)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			if len(diagnostics) > 0 {
				fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
			}
		},
	}

	return sshAddCmd
}
