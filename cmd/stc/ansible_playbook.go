package main

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/dannyvelas/starcommand/internal/config"
	"github.com/spf13/cobra"
)

func ansiblePlaybookCmd(c *config.Config) []*cobra.Command {
	playbooks := []string{"bootstrap-server", "setup-host", "setup-vm"}
	commands := make([]*cobra.Command, 0, len(playbooks))

	for _, playbook := range playbooks {
		command := &cobra.Command{
			Use:   playbook,
			Short: fmt.Sprintf("Run the %s ansible playbook", playbook),
			RunE:  ansiblePlaybookCLI(c, playbook),
		}

		commands = append(commands, command)
	}
	return commands
}

func ansiblePlaybookCLI(c *config.Config, playbook string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return app.AnsibleRun(ctx, c, playbook)
	}
}
