package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// ansiblePlaybookConfig is implemented by all ansible config types.
// It extends configLoader with a method for generating per-host host_vars files.
type playbookConfig interface {
	configLoader
	generateHostVars() error
}

type bootstrapHostVars struct {
	AnsibleHost          string `yaml:"ansible_host"`
	AnsiblePort          string `yaml:"ansible_port"`
	AnsibleSSHPrivateKey string `yaml:"ansible_ssh_private_key_file"`
	AnsibleUser          string `yaml:"ansible_user"`
	SSHPublicKey         string `yaml:"ssh_public_key"`
	AutoUpdateRebootTime string `yaml:"auto_update_reboot_time"`
}

type setupHostHostVars struct {
	AnsibleHost          string `yaml:"ansible_host"`
	AnsiblePort          string `yaml:"ansible_port"`
	AnsibleSSHPrivateKey string `yaml:"ansible_ssh_private_key_file"`
	AnsibleUser          string `yaml:"ansible_user"`
	IncusStoragePoolName string `yaml:"incus_storage_pool_name"`
	IncusStorageDriver   string `yaml:"incus_storage_driver"`
}

type setupVMHostVars struct {
	AnsibleHost          string `yaml:"ansible_host"`
	AnsiblePort          string `yaml:"ansible_port"`
	AnsibleSSHPrivateKey string `yaml:"ansible_ssh_private_key_file"`
	AnsibleUser          string `yaml:"ansible_user"`
}

func writeHostVarsFile(hostname string, vars any) error {
	dir := filepath.Join(".generated", "host_vars", hostname)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating host_vars dir for %s: %v", hostname, err)
	}

	data, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("error marshaling host vars for %s: %v", hostname, err)
	}

	if err := os.WriteFile(filepath.Join(dir, "vars.yml"), data, 0o644); err != nil {
		return fmt.Errorf("error writing host vars file for %s: %v", hostname, err)
	}

	return nil
}
