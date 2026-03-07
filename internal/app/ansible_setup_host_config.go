package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/models"
)

var _ ansibleConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	Hosts []ansibleHostConfig

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig(hosts []models.Host) (*ansibleSetupHostConfig, *Diagnostics) {
	setupConfig := new(ansibleSetupHostConfig)
	diagnostics := new(Diagnostics)

	for i, host := range hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)

		// query data
		baseConfig, baseDiagnostics := newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		diagnostics.appendWithPrefix(prefix, *baseDiagnostics...)

		// set diagnostics
		diagnostics.set(prefix+".incus.storage_pool_name", host.Incus.StoragePoolName)
		diagnostics.set(prefix+".incus.storage_driver", host.Incus.StoragePoolDriver)

		baseConfig.Map = map[string]any{
			"incus_storage_pool_name": host.Incus.StoragePoolName,
			"incus_storage_driver":    host.Incus.StoragePoolDriver,
		}

		setupConfig.Hosts = append(setupConfig.Hosts, baseConfig)
	}
	return setupConfig, diagnostics
}

func (c *ansibleSetupHostConfig) hosts() []ansibleHostConfig {
	return c.Hosts
}
