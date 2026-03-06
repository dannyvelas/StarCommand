package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/config"
)

type playbookConfig interface {
	hosts() []ansibleHostConfig
}

type ansibleHostConfig struct {
	Name string
	Map  map[string]any
}

func getAnsibleConfig(c *config.Config, playbook string) (playbookConfig, error) {
	switch playbook {
	case "bootstrap-server":
		return newAnsibleBootstrapConfig(c)
	case "setup-host":
		return newAnsibleSetupHostConfig(c)
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}
