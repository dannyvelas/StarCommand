package main

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func ansiblePlaybookCmd(c *models.Config) []*cobra.Command {
	playbooks := []string{"bootstrap-host", "setup-host", "setup-vm"}
	commands := make([]*cobra.Command, 0, len(playbooks))

	for _, playbook := range playbooks {
		var hosts []string
		var preflight bool
		command := &cobra.Command{
			Use:   playbook,
			Short: fmt.Sprintf("Run the %s ansible playbook", playbook),
			RunE:  ansiblePlaybookCLI(c, playbook, &hosts, &preflight),
		}
		command.Flags().StringArrayVar(&hosts, "host", nil, "Limit to specific hosts (repeatable)")
		command.Flags().BoolVar(&preflight, "preflight", false, "Show config field statuses without running")

		commands = append(commands, command)
	}
	return commands
}

func ansiblePlaybookCLI(c *models.Config, playbook string, hosts *[]string, preflight *bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		diagnostics, err := app.AnsibleRun(cmd.Context(), c, playbook, *hosts, *preflight)
		if err != nil {
			return err
		}
		if diagnostics != nil {
			fmt.Print(app.DiagnosticsToTable(diagnostics))
		}
		return nil
	}
}
