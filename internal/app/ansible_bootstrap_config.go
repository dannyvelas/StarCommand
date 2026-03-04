package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/goccy/go-yaml"
)

type bootstrapHostEntry struct {
	Name                 string
	IP                   string
	SSH                  config.SSHConfig
	AutoUpdateRebootTime string
}

type bootstrapHostVars struct {
	AnsibleHost          string `yaml:"ansible_host"`
	AnsiblePort          int    `yaml:"ansible_port"`
	AnsibleSSHPrivateKey string `yaml:"ansible_ssh_private_key_file"`
	AnsibleUser          string `yaml:"ansible_user"`
	SSHPublicKey         string `yaml:"ssh_public_key"`
	AutoUpdateRebootTime string `yaml:"auto_update_reboot_time"`
}

type ansibleBootstrapConfig struct {
	Hosts []bootstrapHostEntry `json:"-" required:"true"`

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(c *config.Config) *ansibleBootstrapConfig {
	bootstrapConfig := new(ansibleBootstrapConfig)
	for _, host := range c.Hosts {
		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, bootstrapHostEntry{
			Name:                 host.Name,
			IP:                   host.IP,
			SSH:                  host.SSH,
			AutoUpdateRebootTime: host.AutoUpdateRebootTime,
		})
	}
	return bootstrapConfig
}

func (c *ansibleBootstrapConfig) generateHostVars() error {
	for _, host := range c.Hosts {
		ansibleUser, err := determineAnsibleUser(host.SSH.User, host.IP, host.SSH.Port, host.SSH.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error determining ansible user for %s: %v", host.Name, err)
		}

		expandedPrivateKey, err := helpers.ExpandPath(host.SSH.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error expanding private key path for %s: %v", host.Name, err)
		}

		expandedPublicKey, err := helpers.ExpandPath(host.SSH.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("error expanding public key path for %s: %v", host.Name, err)
		}

		pubKeyBytes, err := os.ReadFile(expandedPublicKey)
		if err != nil {
			return fmt.Errorf("error reading public key for %s: %v", host.Name, err)
		}

		autoUpdateRebootTime := host.AutoUpdateRebootTime
		if autoUpdateRebootTime == "" {
			autoUpdateRebootTime = "05:00"
		}

		vars := bootstrapHostVars{
			AnsibleHost:          host.IP,
			AnsiblePort:          host.SSH.Port,
			AnsibleSSHPrivateKey: expandedPrivateKey,
			AnsibleUser:          ansibleUser,
			SSHPublicKey:         string(pubKeyBytes),
			AutoUpdateRebootTime: autoUpdateRebootTime,
		}

		dir := filepath.Join(".generated", "host_vars", host.Name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("error creating host_vars dir for %s: %v", host.Name, err)
		}

		data, err := yaml.Marshal(vars)
		if err != nil {
			return fmt.Errorf("error marshaling host vars for %s: %v", host.Name, err)
		}

		if err := os.WriteFile(filepath.Join(dir, "vars.yml"), data, 0o644); err != nil {
			return fmt.Errorf("error writing host vars file for %s: %v", host.Name, err)
		}
	}
	return nil
}
