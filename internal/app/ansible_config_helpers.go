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

func getAnsibleConfig(playbook string, hosts []config.Host) (playbookConfig, error) {
	switch playbook {
	case "bootstrap-server":
		return newAnsibleBootstrapConfig(hosts)
	case "setup-host":
		return newAnsibleSetupHostConfig(hosts)
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}
