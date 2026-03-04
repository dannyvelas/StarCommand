package app

import (
	"fmt"
	"slices"

	"github.com/dannyvelas/starcommand/config"
)

type sshConfig struct {
	Alias         string `json:"alias" required:"true"`
	HostName      string `json:"host_name" required:"true"`
	User          string `json:"ssh_user" required:"true"`
	PublicKeyPath string `json:"ssh_public_key_path" required:"true"`
	Port          int    `json:"ssh_port" required:"true"`
}

func newSSHConfig(c *config.Config, hostAlias string) (*sshConfig, map[string]string, error) {
	i := slices.IndexFunc(c.Hosts, func(h config.Host) bool { return h.Name == hostAlias })
	if i < 0 {
		return nil, nil, fmt.Errorf("host alias %s not found in config", hostAlias)
	}
	configHost := c.Hosts[i]

	diagnostics := make(map[string]string)
	buildStructDiagnostics(configHost, fmt.Sprintf(".hosts[%d]", i), diagnostics)

	return &sshConfig{
		Alias:         configHost.Name,
		HostName:      configHost.IP,
		User:          configHost.SSH.User,
		PublicKeyPath: configHost.SSH.PublicKeyPath,
		Port:          configHost.SSH.Port,
	}, diagnostics, nil
}
