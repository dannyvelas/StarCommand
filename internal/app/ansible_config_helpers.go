package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/models"
)

type playbookConfig interface {
	hosts() []ansibleHostConfig
}

func getAnsibleConfig(playbook string, hosts []models.Host) (playbookConfig, map[string]string, error) {
	switch playbook {
	case "bootstrap-host":
		c, d := newAnsibleBootstrapConfig(hosts)
		return c, d, nil
	case "setup-host":
		c, d := newAnsibleSetupHostConfig(hosts)
		return c, d, nil
	}

	return nil, nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}
