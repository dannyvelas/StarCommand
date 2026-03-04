package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ playbookConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	Hosts []hostConfig

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig(c *config.Config) (*ansibleSetupHostConfig, map[string]string) {
	setupConfig := new(ansibleSetupHostConfig)
	if len(c.Hosts) == 0 {
		return setupConfig, map[string]string{".hosts": statusMissing}
	}

	diagnostics := make(map[string]string)
	for i, host := range c.Hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)

		ansibleBaseConfig, baseDiagnostics := newAnsibleBaseConfig(prefix, host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		mergedDiagnostics := helpers.MergeMaps(diagnostics, baseDiagnostics)

		setDiagnostic(mergedDiagnostics, prefix+".incus.storage_pool_name", host.Incus.StoragePoolName)
		setDiagnostic(mergedDiagnostics, prefix+".incus.storage_pool_driver", host.Incus.StoragePoolDriver)
		setupConfig.Hosts = append(setupConfig.Hosts, setupHostEntry{
			AnsibleBaseConfig: ansibleBaseConfig,
			Incus:             host.Incus,
		})
	}
	return setupConfig, diagnostics
}

func (c *ansibleSetupHostConfig) hosts() []hostConfig {
	return c.Hosts
}

type setupHostEntry struct {
	AnsibleBaseConfig ansibleBaseConfig
	Incus             config.IncusConfig
}

func (e setupHostEntry) name() string {
	return e.AnsibleBaseConfig.Name
}

func (e setupHostEntry) asMap() (map[string]any, error) {
	ansibleBaseMap, err := e.AnsibleBaseConfig.asMap()
	if err != nil {
		return nil, fmt.Errorf("error converting ansible base config to map for %s: %v", e.AnsibleBaseConfig.Name, err)
	}

	return helpers.MergeMaps(ansibleBaseMap, map[string]any{
		"incus_storage_pool_name": e.Incus.StoragePoolName,
		"incus_storage_driver":    e.Incus.StoragePoolDriver,
	}), nil
}
