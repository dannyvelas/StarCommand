package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ playbookConfig = (*ansibleBootstrapConfig)(nil)

type bootstrapHostEntry struct {
	Name                 string
	IP                   string
	SSH                  config.SSHConfig
	AutoUpdateRebootTime string
}

type ansibleBootstrapConfig struct {
	Hosts []bootstrapHostEntry `json:"-" required:"true"`

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig() *ansibleBootstrapConfig {
	return &ansibleBootstrapConfig{}
}

func (c *ansibleBootstrapConfig) FillFromConfig(cfg *config.Config) error {
	for _, host := range cfg.Hosts {
		c.Hosts = append(c.Hosts, bootstrapHostEntry{
			Name:                 host.Name,
			IP:                   host.IP,
			SSH:                  host.SSH,
			AutoUpdateRebootTime: host.AutoUpdateRebootTime,
		})
	}
	return nil
}

func (c *ansibleBootstrapConfig) generateHostVars() error {
	for _, host := range c.Hosts {
		if err := generateBootstrapVarsFor(host); err != nil {
			return err
		}
	}
	return nil
}

func generateBootstrapVarsFor(entry bootstrapHostEntry) error {
	ansibleUser, err := determineAnsibleUser(entry.SSH.User, entry.IP, portToString(entry.SSH.Port), entry.SSH.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error determining ansible user for %s: %v", entry.Name, err)
	}

	expandedPrivateKey, err := helpers.ExpandPath(entry.SSH.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding private key path for %s: %v", entry.Name, err)
	}

	expandedPublicKey, err := helpers.ExpandPath(entry.SSH.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding public key path for %s: %v", entry.Name, err)
	}

	pubKeyBytes, err := os.ReadFile(expandedPublicKey)
	if err != nil {
		return fmt.Errorf("error reading public key for %s: %v", entry.Name, err)
	}

	autoUpdateRebootTime := entry.AutoUpdateRebootTime
	if autoUpdateRebootTime == "" {
		autoUpdateRebootTime = "05:00"
	}

	vars := bootstrapHostVars{
		AnsibleHost:          entry.IP,
		AnsiblePort:          portToString(entry.SSH.Port),
		AnsibleSSHPrivateKey: expandedPrivateKey,
		AnsibleUser:          ansibleUser,
		SSHPublicKey:         string(pubKeyBytes),
		AutoUpdateRebootTime: autoUpdateRebootTime,
	}

	return writeHostVarsFile(entry.Name, vars)
}
