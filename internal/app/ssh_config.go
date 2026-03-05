package app

import (
	"fmt"
	"slices"

	"github.com/dannyvelas/starcommand/config"
)

type sshConfig struct {
	Alias         string `json:"alias"`
	HostName      string `json:"host_name"`
	User          string `json:"ssh_user"`
	PublicKeyPath string `json:"ssh_public_key_path"`
	Port          int    `json:"ssh_port"`
}

func newSSHConfig(c *config.Config, hostAlias string) (*sshConfig, error) {
	i := slices.IndexFunc(c.Hosts, func(h config.Host) bool { return h.Name == hostAlias })
	if i < 0 {
		return nil, fmt.Errorf("host alias %s not found in config", hostAlias)
	}
	configHost := c.Hosts[i]

	return &sshConfig{
		Alias:         configHost.Name,
		HostName:      configHost.IP,
		User:          configHost.SSH.User,
		PublicKeyPath: configHost.SSH.PublicKeyPath,
		Port:          configHost.SSH.Port,
	}, nil
}
