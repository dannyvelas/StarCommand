package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/dannyvelas/homelab/internal/config"
	"github.com/kevinburke/ssh_config"
)

var ErrHostAlreadyExists = errors.New("host already exists in ssh config file")

type SSHSetter struct {
	hostName     string
	configReader config.Reader
}

func NewSSHSetter(hostName string, configReader config.Reader) SSHSetter {
	return SSHSetter{
		hostName:     hostName,
		configReader: configReader,
	}
}

func (s SSHSetter) UpdateConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error: could not find home directory: %v")
	}
	path := filepath.Join(home, ".ssh", "config")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("error opening ssh config: %v")
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return fmt.Errorf("error parsing ssh config: %v")
	}

	// if host already exists, return
	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns {
			if pattern.String() == s.hostName {
				return ErrHostAlreadyExists
			}
		}
	}

	if _, err := config.UnmarshalInto(s.configReader, &sshHost); err != nil {
	}

	hostBlock := buildHostBlock(hostName)
	if _, err := f.Seek(0, 2); err == nil {
		f.WriteString(hostBlock)
	}
}

func buildHostBlock(host SSHHost) string {
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

	return buf.String()
}
