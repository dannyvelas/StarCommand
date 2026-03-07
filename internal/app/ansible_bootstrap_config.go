package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/dannyvelas/starcommand/internal/models"
)

var _ ansibleConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	hosts       []ansibleHostConfig
	diagnostics *Diagnostics

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(hosts []models.Host) *ansibleBootstrapConfig {
	bootstrapConfig := new(ansibleBootstrapConfig)
	diagnostics := new(Diagnostics)

	for i, host := range hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)

		baseConfig, baseDiagnostics := newAnsibleBaseConfig(host.Name, host.IP, host.SSH.User, host.SSH.Port, host.SSH.PrivateKeyPath)
		diagnostics.appendWithPrefix(prefix, *baseDiagnostics...)

		pubKeyContent, err := readPublicKey(host.SSH.PublicKeyPath)
		if err != nil {
			diagnostics.append(diagnostic{Field: prefix + ".ssh.public_key_path", Status: err.Error()})
		} else {
			diagnostics.append(diagnostic{Field: prefix + ".ssh.public_key_path", Status: statusLoaded})
		}

		baseConfig.Map = map[string]any{
			"ssh_public_key":          pubKeyContent,
			"auto_update_reboot_time": autoUpdateRebootTime(host.AutoUpdateRebootTime),
		}

		bootstrapConfig.hosts = append(bootstrapConfig.hosts, baseConfig)
	}

	bootstrapConfig.diagnostics = diagnostics

	return bootstrapConfig
}

func (c *ansibleBootstrapConfig) getHosts() []ansibleHostConfig {
	return c.hosts
}

func (c *ansibleBootstrapConfig) getDiagnostics() *Diagnostics {
	return c.diagnostics
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
