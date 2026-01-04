package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/resolve"
	"github.com/spf13/cobra"
)

// resolveCmd represents the resolve command
var resolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Generate a JSON object of configuration values for a given host",
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		config, err := resolve.ResolveConfig(verbose, host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
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

func init() {
	rootCmd.AddCommand(resolveCmd)
}
