package app

import (
	"github.com/dannyvelas/starcommand/config"
)

var _ playbookConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	Hosts []hostConfig `json:"-" required:"true"`

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig(c *config.Config) *ansibleSetupHostConfig {
	setupConfig := new(ansibleSetupHostConfig)
	for _, host := range c.Hosts {
		setupConfig.Hosts = append(setupConfig.Hosts, setupHostEntry{
			AnsibleBaseConfig: newAnsibleBaseConfig(host.Name, host.IP, host.SSH),
			Incus:             host.Incus,
		})
	}
	return setupConfig
}

func (c *ansibleSetupHostConfig) hosts() []hostConfig {
	return c.Hosts
}

type setupHostEntry struct {
	AnsibleBaseConfig ansibleBaseConfig
	Incus             config.IncusConfig
}

func (e setupHostEntry) ansibleBaseConfig() ansibleBaseConfig {
	return e.AnsibleBaseConfig
}

func (e setupHostEntry) asMap() (map[string]any, error) {
	return map[string]any{
		"incus_storage_pool_name": e.Incus.StoragePoolName,
		"incus_storage_driver":    e.Incus.StoragePoolDriver,
	}, nil
}
