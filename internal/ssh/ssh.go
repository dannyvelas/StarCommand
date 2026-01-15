package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/dannyvelas/homelab/internal/host"
	"github.com/kevinburke/ssh_config"
)

var ErrHostAlreadyExists = errors.New("host already exists in ssh config file")

type SSHSetter struct {
	hostName string
	sshHost  host.SSHHost
}

func NewSSHSetter(hostName string, sshHost host.SSHHost) SSHSetter {
	return SSHSetter{
		hostName: hostName,
		sshHost:  sshHost,
	}
}

func (s SSHSetter) UpdateConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error: could not find home directory: %v", err)
	}
	path := filepath.Join(home, ".ssh", "config")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("error opening ssh config: %v", err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return fmt.Errorf("error parsing ssh config: %v", err)
	}

	// if host already exists, return
	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns {
			if pattern.String() == s.hostName {
				return ErrHostAlreadyExists
			}
		}
	}

	hostBlock := buildHostBlock(s.sshHost)
	if _, err := f.Seek(0, 2); err != nil {
		return fmt.Errorf("error seeking to end of ssh config: %v", err)
	}

	f.Write(hostBlock)

	return nil
}

func buildHostBlock(host host.SSHHost) []byte {
	const hostTmpl = `
Host {{ .Alias }}
    HostName {{ .HostName }}
    User {{ .User }}
    IdentityFile {{ .IdentityFile }}
		Port {{ .Port }}
`

	tmpl, err := template.New("sshConfig").Parse(hostTmpl)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, host); err != nil {
		panic(err)
	}

	return buf.Bytes()
}
