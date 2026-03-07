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

	// query data
	expandedPrivateKey, err := helpers.ExpandPath(sshPrivateKeyPath)
	if err != nil {
		diagnostics.append(Diagnostic{Field: ".ssh.private_key_path", Status: fmt.Sprintf("error expanding path: %v", err)})
	}

	// set diagnostics
	diagnostics.set(".name", name)
	diagnostics.set(".ip", ip)
	diagnostics.set(".ssh.user", sshUser)
	diagnostics.set(".ssh.port", sshPort)

	return ansibleHostConfig{
		Name:          name,
		IP:            ip,
		SSHUser:       sshUser,
		SSHPort:       sshPort,
		SSHPrivateKey: expandedPrivateKey,
	}, diagnostics
}
