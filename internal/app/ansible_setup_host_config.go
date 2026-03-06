package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ playbookConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	Hosts []ansibleHostConfig

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig(c *config.Config) (*ansibleSetupHostConfig, error) {
	setupConfig := new(ansibleSetupHostConfig)

	for _, host := range c.Hosts {
		baseConfig, err := newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error creating base config for %s: %v", host.Name, err)
		}

		mergedMaps := helpers.MergeMaps(baseConfig.Map, map[string]any{
			"incus_storage_pool_name": host.Incus.StoragePoolName,
			"incus_storage_driver":    host.Incus.StoragePoolDriver,
		})

		setupConfig.Hosts = append(setupConfig.Hosts, ansibleHostConfig{
			Name: host.Name,
			Map:  mergedMaps,
		})
	}
	return setupConfig, nil
}

func (c *ansibleSetupHostConfig) hosts() []ansibleHostConfig {
	return c.Hosts
}
