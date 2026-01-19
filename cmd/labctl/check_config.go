package main

import (
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/app"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/spf13/cobra"
)

func checkConfigCmd() *cobra.Command {
	var targets []string

	checkConfigCmd := &cobra.Command{
		Use:       "config <host-alias>",
		ValidArgs: app.GetSupportedHostAliases(),
		Short:     "Print a diagnostic report of all the configs that were found/missing for a given resource",
		Args:      cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			a, err := app.New(configMux, hostAlias, targets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			diagnostics, err := a.CheckConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			fmt.Printf("Configs for %s:\n%s\n", hostAlias, app.DiagnosticsToTable(diagnostics))
		},
	}

	checkConfigCmd.Flags().StringSliceVar(&targets, "for", []string{"ansible"}, "Specific integrations to check (e.g. ansible, ssh)")

	return checkConfigCmd
}
