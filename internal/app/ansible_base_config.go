package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/helpers"
)

type ansibleHostConfig struct {
	Name          string
	IP            string
	SSHUser       string
	SSHPort       int
	SSHPrivateKey string
	Map           map[string]any
}

func newAnsibleBaseConfig(name, ip, sshUser string, sshPort int, sshPrivateKeyPath string) (ansibleHostConfig, *Diagnostics) {
	diagnostics := new(Diagnostics)

	expandedPrivateKey, err := helpers.ExpandPath(sshPrivateKeyPath)
	if err != nil {
		diagnostics.append(Diagnostic{Field: ".ssh.private_key_path", Status: fmt.Sprintf("error expanding path: %v", err)})
	}

	diagnostics.appendChecked(".name", name)
	diagnostics.appendChecked(".ip", ip)
	diagnostics.appendChecked(".ssh.user", sshUser)
	diagnostics.appendChecked(".ssh.port", sshPort)

	return ansibleHostConfig{
		Name:          name,
		IP:            ip,
		SSHUser:       sshUser,
		SSHPort:       sshPort,
		SSHPrivateKey: expandedPrivateKey,
	}, diagnostics
}
