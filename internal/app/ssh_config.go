package app

import (
	"github.com/dannyvelas/starcommand/internal/models"
)

type sshConfig struct {
	Alias         string `json:"alias"`
	HostName      string `json:"host_name"`
	User          string `json:"ssh_user"`
	PublicKeyPath string `json:"ssh_public_key_path"`
	Port          int    `json:"ssh_port"`
}

func newSSHConfig(c *models.Config, host models.Host) (*sshConfig, error) {
	return &sshConfig{
		Alias:         host.Name,
		HostName:      host.IP,
		User:          host.SSH.User,
		PublicKeyPath: host.SSH.PublicKeyPath,
		Port:          host.SSH.Port,
	}, nil
}
