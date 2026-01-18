package models

import (
	"bytes"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kevinburke/ssh_config"
)

type SSHHost struct {
	Alias           string `json:"alias"`
	HostName        string `json:"hostName"`
	User            string `json:"ssh_user" required:"true"`
	PublicKeyPath   string `json:"ssh_public_key_path" required:"true"`
	Port            string `json:"ssh_port" required:"true"`
	NodeCIDRAddress string `json:"node_cidr_address" required:"true"`
}

func NewSSHHost(hostAlias string) *SSHHost {
	return &SSHHost{
		Alias: hostAlias,
	}
}

func (s *SSHHost) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(s.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", s.NodeCIDRAddress, err)
	}
	s.HostName = prefix.Addr().String()

	return nil
}

func (s *SSHHost) SetFile() error {
	f, err := openHomeSSHFile()
	if err != nil {
		return fmt.Errorf("error opening ssh config file: %v", err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return fmt.Errorf("error parsing ssh config: %v", err)
	}

	// if host already exists, return
	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns {
			if pattern.String() == s.Alias {
				return NewErrAlreadyExists("ssh", "host")
			}
		}
	}

	hostBlock := s.buildHostBlock()
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

func (s *SSHHost) buildHostBlock() []byte {
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
	if err := tmpl.Execute(&buf, s); err != nil {
		panic(err)
	}

	return buf.Bytes()
}
