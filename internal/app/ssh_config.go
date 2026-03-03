package app

import "github.com/dannyvelas/starcommand/config"

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

func (c *sshConfig) FillFromConfig(_ *config.Config) error { return nil }
func (c *sshConfig) FillInKeys() error                     { return nil }
