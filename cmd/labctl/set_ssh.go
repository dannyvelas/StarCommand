package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/host"
	"github.com/dannyvelas/homelab/internal/ssh"
	"github.com/spf13/cobra"
)

func setSSHCmd() *cobra.Command {
	setSSHCmd := &cobra.Command{
		Use:   "ssh <host-name>",
		Short: "Update the `~/.ssh/config` file to connect to a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostName := args[0]

			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(host.FallbackFile, conflux.WithPath(host.GetConfigPath(hostName))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			var sshHost host.SSHHost
			diagnostics, err := conflux.Unmarshal(configMux, &sshHost)
			if errors.Is(err, conflux.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%v for %s:\n%s\n", conflux.ErrInvalidFields, hostName, conflux.DiagnosticsToTable(diagnostics))
				return
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			sshSetter := ssh.NewSSHSetter(hostName, sshHost)
			if err := sshSetter.UpdateConfig(); err != nil && !errors.Is(err, ssh.ErrHostAlreadyExists) {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}
		},
	}

	return setSSHCmd
}
