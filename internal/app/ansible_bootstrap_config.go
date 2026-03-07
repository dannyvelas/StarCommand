package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/dannyvelas/starcommand/internal/models"
)

var _ ansibleConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	Hosts []ansibleHostConfig

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(hosts []models.Host) (*ansibleBootstrapConfig, *Diagnostics) {
	bootstrapConfig := &ansibleBootstrapConfig{}
	diagnostics := new(Diagnostics)

	for i, host := range hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)

		// query data
		baseConfig, baseDiagnostics := newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		diagnostics.appendWithPrefix(prefix, *baseDiagnostics...)

		pubKeyContent, err := readPublicKey(host.SSH.PublicKeyPath)
		if err != nil {
			diagnostics.append(Diagnostic{Field: prefix + ".ssh.public_key_path", Status: err.Error()})
		} else {
			diagnostics.set(prefix+".ssh.public_key_path", pubKeyContent)
		}

		// set fields
		baseConfig.Map = map[string]any{
			"ssh_public_key":          pubKeyContent,
			"auto_update_reboot_time": autoUpdateRebootTime(host.AutoUpdateRebootTime),
		}

		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, baseConfig)
	}
	return bootstrapConfig, diagnostics
}

func (c *ansibleBootstrapConfig) hosts() []ansibleHostConfig {
	return c.Hosts
}

func readPublicKey(path string) (string, error) {
	if path == "" {
		return "", errNotFound
	}

	expandedPath, err := helpers.ExpandPath(path)
	if err != nil {
		return "", fmt.Errorf("error expanding path: %v", err)
	}

	pubKeyBytes, err := os.ReadFile(expandedPath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return string(pubKeyBytes), nil
}

func autoUpdateRebootTime(t string) string {
	if t == "" {
		return "05:00"
	}
	return t
}
