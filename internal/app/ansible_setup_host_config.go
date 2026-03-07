package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/models"
)

var _ playbookConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	Hosts []ansibleHostConfig

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig(hosts []models.Host) (*ansibleSetupHostConfig, error) {
	setupConfig := new(ansibleSetupHostConfig)

	for _, host := range hosts {
		baseConfig, err := newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error creating base config for %s: %v", host.Name, err)
		}

		baseConfig.Map = map[string]any{
			"incus_storage_pool_name": host.Incus.StoragePoolName,
			"incus_storage_driver":    host.Incus.StoragePoolDriver,
		}

		setupConfig.Hosts = append(setupConfig.Hosts, baseConfig)
	}
	return setupConfig, nil
}

func (c *ansibleSetupHostConfig) validate() map[string]string {
	diagnostics := make(map[string]string)
	setBaseHostDiagnostics(diagnostics, c.Hosts)
	for i, host := range c.Hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)
		setDiagnostic(diagnostics, prefix+".incus.storage_pool_name", host.Map["incus_storage_pool_name"])
		setDiagnostic(diagnostics, prefix+".incus.storage_driver", host.Map["incus_storage_driver"])
	}
	return diagnostics
}

func (c *ansibleSetupHostConfig) hosts() []ansibleHostConfig {
	return c.Hosts
}
