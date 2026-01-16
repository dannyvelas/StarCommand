package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/dannyvelas/homelab/internal/models"
	"github.com/kevinburke/ssh_config"
)

var ErrHostAlreadyExists = errors.New("host already exists in ssh config file")

type SSHSetter struct {
	hostAlias string
}

func NewSSHSetter(hostAlias string) (SSHSetter, error) {
	f, err := openHomeSSHFile()
	if err != nil {
		return SSHSetter{}, fmt.Errorf("error opening ssh config file: %v", err)
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return SSHSetter{}, fmt.Errorf("error parsing ssh config: %v", err)
	}

	// if host already exists, return
	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns {
			if pattern.String() == hostAlias {
				return SSHSetter{}, ErrHostAlreadyExists
			}
		}
	}

	return SSHSetter{hostAlias: hostAlias}, nil
}

func (s SSHSetter) UpdateConfig(host *models.SSHHost) error {
	f, err := openHomeSSHFile()
	if err != nil {
		return fmt.Errorf("error opening ssh config file: %v", err)
	}

	hostBlock := buildHostBlock(host)
	if _, err := f.Seek(0, 2); err != nil {
		return fmt.Errorf("error seeking to end of ssh config: %v", err)
	}

	f.Write(hostBlock)

	return nil
}

func openHomeSSHFile() (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error: could not find home directory: %v", err)
	}
	path := filepath.Join(home, ".ssh", "config")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("error opening ssh config: %v", err)
	}

	return f, nil
}

func buildHostBlock(host *models.SSHHost) []byte {
	const hostTmpl = `
Host {{ .Alias }}
  HostName {{ .HostName }}
  User {{ .User }}
  IdentityFile {{ .PublicKeyPath }}
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
