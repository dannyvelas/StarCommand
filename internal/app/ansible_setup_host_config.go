package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ ansiblePlaybookConfig = (*ansibleSetupHostConfig)(nil)

type setupHostEntry struct {
	Name  string
	IP    string
	SSH   config.SSHConfig
	Incus config.IncusConfig
}

type ansibleSetupHostConfig struct {
	Hosts []setupHostEntry `json:"-" required:"true"`

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig() *ansibleSetupHostConfig {
	return &ansibleSetupHostConfig{}
}

func (c *ansibleSetupHostConfig) FillFromConfig(cfg *config.Config) error {
	for _, host := range cfg.Hosts {
		c.Hosts = append(c.Hosts, setupHostEntry{
			Name:  host.Name,
			IP:    host.IP,
			SSH:   host.SSH,
			Incus: host.Incus,
		})
	}
	return nil
}

func (c *ansibleSetupHostConfig) generateHostVars() error {
	for _, host := range c.Hosts {
		if err := generateSetupHostVarsFor(host); err != nil {
			return err
		}
	}
	return nil
}

func generateSetupHostVarsFor(entry setupHostEntry) error {
	ansibleUser, err := determineAnsibleUser(entry.SSH.User, entry.IP, portToString(entry.SSH.Port), entry.SSH.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error determining ansible user for host %s: %v", entry.Name, err)
	}

	expandedPrivateKey, err := helpers.ExpandPath(entry.SSH.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding private key path for host %s: %v", entry.Name, err)
	}

	vars := setupHostHostVars{
		AnsibleHost:          entry.IP,
		AnsiblePort:          portToString(entry.SSH.Port),
		AnsibleSSHPrivateKey: expandedPrivateKey,
		AnsibleUser:          ansibleUser,
		IncusStoragePoolName: entry.Incus.StoragePoolName,
		IncusStorageDriver:   entry.Incus.StoragePoolDriver,
	}

	return writeHostVarsFile(entry.Name, vars)
}
