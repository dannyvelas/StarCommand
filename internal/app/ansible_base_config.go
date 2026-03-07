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

func newAnsibleBaseConfig(name, ip, sshUser string, sshPort int, sshPrivateKeyPath string) (ansibleHostConfig, map[string]string) {
	diagnostics := make(map[string]string)

	// query data
	expandedPrivateKey, err := helpers.ExpandPath(sshPrivateKeyPath)
	if err != nil {
		diagnostics[".ssh.private_key_path"] = fmt.Sprintf("error expanding path: %v", err)
	}

	// set diagnostics
	setDiagnostic(diagnostics, ".name", name)
	setDiagnostic(diagnostics, ".ip", ip)
	setDiagnostic(diagnostics, ".ssh.user", sshUser)
	setDiagnostic(diagnostics, ".ssh.port", sshPort)

	return ansibleHostConfig{
		Name:          name,
		IP:            ip,
		SSHUser:       sshUser,
		SSHPort:       sshPort,
		SSHPrivateKey: expandedPrivateKey,
	}, diagnostics
}
