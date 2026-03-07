package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/models"
)

type playbookConfig interface {
	hosts() []ansibleHostConfig
	validate() map[string]string
}

type ansibleHostConfig struct {
	Name          string
	IP            string
	SSHUser       string
	SSHPort       int
	SSHPrivateKey string
	Map           map[string]any
}

func getAnsibleConfig(playbook string, hosts []models.Host) (playbookConfig, error) {
	switch playbook {
	case "bootstrap-host":
		return newAnsibleBootstrapConfig(hosts)
	case "setup-host":
		return newAnsibleSetupHostConfig(hosts)
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}
