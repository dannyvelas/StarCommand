package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/internal/config"
	"github.com/dannyvelas/homelab/internal/env"
	"github.com/spf13/cobra"
)

// resolveCmd represents the resolve command
func newResolveCmd(env env.Env, verbose bool) *cobra.Command {
	var showRequirements bool

	resolveCmd := &cobra.Command{
		Use:   "resolve <host-name>",
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			host := args[0]
			if showRequirements {
				requiredKeys, err := config.GetRequiredKeys(host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error getting required keys: %s", err.Error())
					os.Exit(1)
				}

				fmt.Printf("Required keys:%s", requiredKeys)
				return
			}

			config, err := config.Resolve(env, verbose, host)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			bytes, err := json.MarshalIndent(config, "", "    ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error marshalling to JSON: %s", err.Error())
				os.Exit(1)
			}

			fmt.Println(string(bytes))
		},
	}

	resolveCmd.Flags().BoolVar(&showRequirements, "requirements", false, "Check for missing configuration requirements")

	return resolveCmd
}
