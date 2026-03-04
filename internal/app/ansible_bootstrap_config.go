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

func newAnsibleBootstrapConfig(c *config.Config) (*ansibleBootstrapConfig, map[string]string) {
	bootstrapConfig := new(ansibleBootstrapConfig)
	diagnostics := make(map[string]string)

	if len(c.Hosts) == 0 {
		diagnostics[".hosts"] = statusMissing
		return bootstrapConfig, diagnostics
	}

	for i, host := range c.Hosts {
		buildStructDiagnostics(host, fmt.Sprintf(".hosts[%d]", i), diagnostics)
		if host.AutoUpdateRebootTime == "" {
			host.AutoUpdateRebootTime = "05:00"
		}

		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, bootstrapHostEntry{
			AnsibleBaseConfig:    newAnsibleBaseConfig(host.Name, host.IP, host.SSH),
			AutoUpdateRebootTime: host.AutoUpdateRebootTime,
		})
	}
	return bootstrapConfig, diagnostics
}

func (c *ansibleBootstrapConfig) hosts() []hostConfig {
	return c.Hosts
}

type bootstrapHostEntry struct {
	AnsibleBaseConfig    ansibleBaseConfig
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

	return helpers.MergeMaps(ansibleBaseMap, map[string]any{
		"ssh_public_key":          string(pubKeyBytes),
		"auto_update_reboot_time": autoUpdateRebootTime,
	}), nil
}
