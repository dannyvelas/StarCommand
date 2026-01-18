package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
)

func getConfigCmd() *cobra.Command {
	var targets []string

	getConfigCmd := &cobra.Command{
		Use:   "config <host-alias>",
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),
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

			allConfigs, allDiagnostics := make(map[string]string), make(map[string]string)
			for _, configStruct := range configStructs {
				diagnostics, err := conflux.Unmarshal(configMux, configStruct)
				if errors.Is(err, conflux.ErrInvalidFields) {
					maps.Copy(allDiagnostics, diagnostics)
					continue
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					os.Exit(1)
				}

				if err := mapstructure.Decode(configStruct, &allConfigs); err != nil {
					fmt.Fprintf(os.Stderr, "internal error merging config struct to map: %s\n", err.Error())
					os.Exit(1)
				}
			}

			if len(allDiagnostics) > 0 {
				fmt.Fprintf(os.Stderr, "%v for %s:\n%s\n", conflux.ErrInvalidFields, hostAlias, conflux.DiagnosticsToTable(allDiagnostics))
				return
			}

			bytes, err := json.MarshalIndent(allConfigs, "", "    ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error marshalling to JSON: %s", err.Error())
				os.Exit(1)
			}

			fmt.Println(string(bytes))
		},
	}

	getConfigCmd.Flags().StringSliceVar(&targets, "for", []string{"ansible"}, "Get config for specific integration (e.g. ansible, ssh)")

	return getConfigCmd
}
