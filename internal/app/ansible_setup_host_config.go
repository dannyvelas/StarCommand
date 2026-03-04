package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/goccy/go-yaml"
)

type setupHostEntry struct {
	Name  string
	IP    string
	SSH   config.SSHConfig
	Incus config.IncusConfig
}

type setupHostHostVars struct {
	AnsibleHost          string `yaml:"ansible_host"`
	AnsiblePort          string `yaml:"ansible_port"`
	AnsibleSSHPrivateKey string `yaml:"ansible_ssh_private_key_file"`
	AnsibleUser          string `yaml:"ansible_user"`
	IncusStoragePoolName string `yaml:"incus_storage_pool_name"`
	IncusStorageDriver   string `yaml:"incus_storage_driver"`
}

type ansibleSetupHostConfig struct {
	Hosts []setupHostEntry `json:"-" required:"true"`

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig(c *config.Config) *ansibleSetupHostConfig {
	setupConfig := new(ansibleSetupHostConfig)
	for _, host := range c.Hosts {
		setupConfig.Hosts = append(setupConfig.Hosts, setupHostEntry{
			Name:  host.Name,
			IP:    host.IP,
			SSH:   host.SSH,
			Incus: host.Incus,
		})
	}
	return setupConfig
}

func (c *ansibleSetupHostConfig) generateHostVars() error {
	for _, host := range c.Hosts {
		ansibleUser, err := determineAnsibleUser(host.SSH.User, host.IP, portToString(host.SSH.Port), host.SSH.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error determining ansible user for host %s: %v", host.Name, err)
		}

		expandedPrivateKey, err := helpers.ExpandPath(host.SSH.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error expanding private key path for host %s: %v", host.Name, err)
		}

		vars := setupHostHostVars{
			AnsibleHost:          host.IP,
			AnsiblePort:          portToString(host.SSH.Port),
			AnsibleSSHPrivateKey: expandedPrivateKey,
			AnsibleUser:          ansibleUser,
			IncusStoragePoolName: host.Incus.StoragePoolName,
			IncusStorageDriver:   host.Incus.StoragePoolDriver,
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
