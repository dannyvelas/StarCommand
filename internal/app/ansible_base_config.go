package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/helpers"
)

func newAnsibleBaseConfig(name, ip, sshUser string, sshPort int, sshPrivateKeyPath string) (ansibleHostConfig, error) {
	expandedPrivateKey, err := helpers.ExpandPath(sshPrivateKeyPath)
	if err != nil {
		return ansibleHostConfig{}, fmt.Errorf("error expanding private key path for %s: %v", name, err)
	}

	return ansibleHostConfig{
		Name:          name,
		IP:            ip,
		SSHUser:       sshUser,
		SSHPort:       sshPort,
		SSHPrivateKey: expandedPrivateKey,
	}, nil
}

func setBaseHostDiagnostics(diagnostics map[string]string, hosts []ansibleHostConfig) {
	for i, host := range hosts {
		prefix := fmt.Sprintf(".hosts[%d]", i)
		setDiagnostic(diagnostics, prefix+".name", host.Name)
		setDiagnostic(diagnostics, prefix+".ip", host.IP)
		setDiagnostic(diagnostics, prefix+".ssh.user", host.SSHUser)
		setDiagnostic(diagnostics, prefix+".ssh.port", host.SSHPort)
		setDiagnostic(diagnostics, prefix+".ssh.private_key_path", host.SSHPrivateKey)
	}
}
