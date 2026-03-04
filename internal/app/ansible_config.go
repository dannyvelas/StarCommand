package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
)

type playbookConfig interface {
	generateHostVars() error
}

func getAnsibleConfig(c *config.Config, playbook string) (playbookConfig, error) {
	switch playbook {
	case "bootstrap-server":
		return newAnsibleBootstrapConfig(c), nil
	case "setup-host":
		return newAnsibleSetupHostConfig(c), nil
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}
