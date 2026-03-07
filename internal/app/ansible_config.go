package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/models"
)

type ansibleConfig interface {
	getHosts() []ansibleHostConfig
	getDiagnostics() *Diagnostics
}

func getAnsibleConfig(playbook string, hosts []models.Host) (ansibleConfig, error) {
	switch playbook {
	case "bootstrap-host":
		return newAnsibleBootstrapConfig(hosts), nil
	case "setup-host":
		return newAnsibleSetupHostConfig(hosts), nil
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}
