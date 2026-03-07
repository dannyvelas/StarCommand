package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/dannyvelas/starcommand/internal/models"
)

var _ playbookConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	Hosts []ansibleHostConfig

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(hosts []models.Host) (*ansibleBootstrapConfig, error) {
	bootstrapConfig := new(ansibleBootstrapConfig)

	for _, host := range hosts {
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

		baseConfig.Map = map[string]any{
			"ssh_public_key":          string(pubKeyBytes),
			"auto_update_reboot_time": host.AutoUpdateRebootTime,
		}

		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, baseConfig)
	}
	return bootstrapConfig, nil
}

func (c *ansibleBootstrapConfig) validate() map[string]string {
	diagnostics := make(map[string]string)
	setBaseHostDiagnostics(diagnostics, c.Hosts)
	for i, host := range c.Hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)
		setDiagnostic(diagnostics, prefix+".ssh.public_key_path", host.Map["ssh_public_key"])
	}
	return diagnostics
}

func (c *ansibleBootstrapConfig) hosts() []ansibleHostConfig {
	return c.Hosts
}
