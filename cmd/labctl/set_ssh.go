package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/dannyvelas/homelab/internal/ssh"
	"github.com/spf13/cobra"
)

func setSSHCmd() *cobra.Command {
	setSSHCmd := &cobra.Command{
		Use:   "ssh <host-alias>",
		Short: "Update the `~/.ssh/config` file to connect to a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			sshSetter, err := ssh.NewSSHSetter(hostAlias)
			if errors.Is(err, ssh.ErrHostAlreadyExists) {
				fmt.Println("Host already exists in ssh config file, skipping...")
				return
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			sshHost := models.NewSSHHost(hostAlias)
			diagnostics, err := conflux.Unmarshal(configMux, sshHost)
			if errors.Is(err, conflux.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%v for setting ssh for %s host:\n%s\n", conflux.ErrInvalidFields, hostAlias, conflux.DiagnosticsToTable(diagnostics))
				return
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			if err := sshSetter.UpdateConfig(sshHost); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			fmt.Println("SSH config updated successfully!")
		},
	}

	return setSSHCmd
}
