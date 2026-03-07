// Package models has the global structs used throughout the application
package models

type Config struct {
	Hosts []Host `yaml:"hosts"`
}

type Host struct {
	Name                 string      `yaml:"name"`
	IP                   string      `yaml:"ip"`
	SSH                  SSHConfig   `yaml:"ssh"`
	AutoUpdateRebootTime string      `yaml:"auto_update_reboot_time"`
	WireGuardEndpoint    bool        `yaml:"wireguard_endpoint"`
	Incus                IncusConfig `yaml:"incus"`
	VMs                  []VM        `yaml:"vms"`
}

type VM struct {
	Name                 string    `yaml:"name"`
	IP                   string    `yaml:"ip"`
	SSH                  SSHConfig `yaml:"ssh"`
	AutoUpdateRebootTime string    `yaml:"auto_update_reboot_time"`
}

type SSHConfig struct {
	User           string `yaml:"user"`
	Port           int    `yaml:"port"`
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
}

type IncusConfig struct {
	StoragePoolName   string `yaml:"storage_pool_name"`
	StoragePoolDriver string `yaml:"storage_pool_driver"`
}
