package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ playbookConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	Hosts []hostConfig `json:"-" required:"true"`

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(c *config.Config) *ansibleBootstrapConfig {
	bootstrapConfig := new(ansibleBootstrapConfig)
	for _, host := range c.Hosts {
		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, bootstrapHostEntry{
			AnsibleBaseConfig:    newAnsibleBaseConfig(host.Name, host.IP, host.SSH),
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
	AutoUpdateRebootTime string
}

func (e bootstrapHostEntry) ansibleBaseConfig() ansibleBaseConfig {
	return e.AnsibleBaseConfig
}

func (e bootstrapHostEntry) asMap() (map[string]any, error) {
	expandedPublicKey, err := helpers.ExpandPath(e.AnsibleBaseConfig.SSH.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error expanding public key path for %s: %v", e.AnsibleBaseConfig.Name, err)
	}

	pubKeyBytes, err := os.ReadFile(expandedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("error reading public key for %s: %v", e.AnsibleBaseConfig.Name, err)
	}

	autoUpdateRebootTime := e.AutoUpdateRebootTime
	if autoUpdateRebootTime == "" {
		autoUpdateRebootTime = "05:00"
	}

	return map[string]any{
		"ssh_public_key":          string(pubKeyBytes),
		"auto_update_reboot_time": autoUpdateRebootTime,
	}, nil
}
