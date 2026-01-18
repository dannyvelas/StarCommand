package main

import (
	"errors"
	"fmt"
	"maps"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/spf13/cobra"
)

func checkConfigCmd() *cobra.Command {
	var targets []string

	checkConfigCmd := &cobra.Command{
		Use:       "config <host-alias>",
		ValidArgs: []string{"proxmox", "plex-lxc", "vm"},
		Short:     "Print a diagnostic report of all the configs that were found/missing for a given resource",
		Args:      cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]

			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			configStructs, err := models.AliasToStruct(hostAlias, targets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			allDiagnostics := make(map[string]string)
			for _, configStruct := range configStructs {
				diagnostics, err := conflux.Unmarshal(configMux, configStruct)
				if errors.Is(err, conflux.ErrInvalidFields) {
					maps.Copy(allDiagnostics, diagnostics)
					continue
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					os.Exit(1)
				}
			}
			if len(allDiagnostics) > 0 {
				fmt.Fprintf(os.Stderr, "%v for %s:\n%s\n", conflux.ErrInvalidFields, hostAlias, conflux.DiagnosticsToTable(allDiagnostics))
				return
			}

			fmt.Printf("Configs for %s:\n%s\n", hostAlias, conflux.DiagnosticsToTable(allDiagnostics))
		},
	}

	checkConfigCmd.Flags().StringSliceVar(&targets, "for", []string{"ansible"}, "Specific integrations to check (e.g. ansible, ssh)")

	return checkConfigCmd
}
