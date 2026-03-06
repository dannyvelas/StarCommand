package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/internal/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ playbookConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	Hosts []ansibleHostConfig

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(c *config.Config) (*ansibleBootstrapConfig, error) {
	bootstrapConfig := new(ansibleBootstrapConfig)

	for _, host := range c.Hosts {
		baseConfig, err := newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error creating base config for %s: %v", host.Name, err)
		}

		expandedPublicKey, err := helpers.ExpandPath(host.SSH.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error expanding public key path for %s: %v", host.Name, err)
		}

		pubKeyBytes, err := os.ReadFile(expandedPublicKey)
		if err != nil {
			return nil, fmt.Errorf("error reading public key for %s: %v", host.Name, err)
		}

		if host.AutoUpdateRebootTime == "" {
			host.AutoUpdateRebootTime = "05:00"
		}

		mergedMaps := helpers.MergeMaps(baseConfig.Map, map[string]any{
			"ssh_public_key":          string(pubKeyBytes),
			"auto_update_reboot_time": host.AutoUpdateRebootTime,
		})

		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, ansibleHostConfig{
			Name: host.Name,
			Map:  mergedMaps,
		})
	}
	return bootstrapConfig, nil
}

func (c *ansibleBootstrapConfig) hosts() []ansibleHostConfig {
	return c.Hosts
}
