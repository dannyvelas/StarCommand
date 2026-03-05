package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ playbookConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	Hosts []hostConfig

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(c *config.Config) *ansibleBootstrapConfig {
	bootstrapConfig := new(ansibleBootstrapConfig)
	if len(c.Hosts) == 0 {
		return bootstrapConfig
	}

	for _, host := range c.Hosts {
		if host.AutoUpdateRebootTime == "" {
			host.AutoUpdateRebootTime = "05:00"
		}

		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, bootstrapHostEntry{
			AnsibleBaseConfig:    newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath),
			SSHPublicKeyPath:     host.SSH.PublicKeyPath,
			AutoUpdateRebootTime: host.AutoUpdateRebootTime,
		})
	}
	return bootstrapConfig
}

func (c *ansibleBootstrapConfig) hosts() []hostConfig {
	return c.Hosts
}

type bootstrapHostEntry struct {
	AnsibleBaseConfig    ansibleBaseConfig
	SSHPublicKeyPath     string
	AutoUpdateRebootTime string
}

func (e bootstrapHostEntry) name() string {
	return e.AnsibleBaseConfig.Name
}

func (e bootstrapHostEntry) asMap() (map[string]any, error) {
	ansibleBaseMap, err := e.AnsibleBaseConfig.asMap()
	if err != nil {
		return nil, fmt.Errorf("error converting ansible base config to map for %s: %v", e.AnsibleBaseConfig.Name, err)
	}

	expandedPublicKey, err := helpers.ExpandPath(e.SSHPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error expanding public key path for %s: %v", e.AnsibleBaseConfig.Name, err)
	}

	pubKeyBytes, err := os.ReadFile(expandedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("error reading public key for %s: %v", e.AnsibleBaseConfig.Name, err)
	}

	return helpers.MergeMaps(ansibleBaseMap, map[string]any{
		"ssh_public_key":          string(pubKeyBytes),
		"auto_update_reboot_time": e.AutoUpdateRebootTime,
	}), nil
}
