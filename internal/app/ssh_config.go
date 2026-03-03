package app

import (
	"fmt"
	"strconv"

	"github.com/dannyvelas/starcommand/config"
)

type sshConfig struct {
	Alias         string `json:"alias" required:"true"`
	HostName      string `json:"host_name" required:"true"`
	User          string `json:"ssh_user" required:"true"`
	PublicKeyPath string `json:"ssh_public_key_path" required:"true"`
	Port          string `json:"ssh_port" required:"true"`
}

func newSSHHost(hostAlias string) *sshConfig {
	return &sshConfig{Alias: hostAlias}
}

func (c *sshConfig) FillFromConfig(cfg *config.Config) error {
	for _, h := range cfg.Hosts {
		if h.Name == c.Alias {
			c.HostName = h.IP
			c.User = h.SSH.User
			c.Port = portToString(h.SSH.Port)
			c.PublicKeyPath = h.SSH.PublicKeyPath
			return nil
		}
		for _, vm := range h.VMs {
			if vm.Name == c.Alias {
				c.HostName = vm.IP
				c.User = vm.SSH.User
				c.Port = portToString(vm.SSH.Port)
				c.PublicKeyPath = vm.SSH.PublicKeyPath
				return nil
			}
		}
	}
	return fmt.Errorf("host %q %w", c.Alias, errNotFound)
}

func (c *sshConfig) FillInKeys() error { return nil }

func portToString(port int) string {
	if port != 0 {
		return strconv.Itoa(port)
	}
	return "22"
}
