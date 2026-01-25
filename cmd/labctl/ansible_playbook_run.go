package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/handlers"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/spf13/cobra"
)

func ansiblePlaybookRunCmd() *cobra.Command {
	var targets []string

	ansibleRunCmd := &cobra.Command{
		Use: "run <host-alias>",
		// TODO: fix
		ValidArgs: nil,
		// TODO: fix
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			handler, err := handlers.New(configMux, hostAlias, targets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			configs, diagnostics, err := handler.GetConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			} else if len(diagnostics) > 0 {
				fmt.Fprintf(os.Stderr, "invalid or missing configs for %s:\n%s\n", hostAlias, handlers.DiagnosticsToTable(diagnostics))
				return
			}

			bytes, err := json.MarshalIndent(configs, "", "    ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "internal error marshalling configs to JSON: %s\n", err.Error())
				os.Exit(1)
			}

			fmt.Println(string(bytes))
		},
	}

	ansibleRunCmd.Flags().StringSliceVar(&targets, "for", []string{"ansible"}, "Get config for specific integration (e.g. ansible, ssh)")

	return ansibleRunCmd
}
