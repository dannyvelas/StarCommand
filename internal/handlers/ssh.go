package handlers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kevinburke/ssh_config"
	"github.com/spf13/afero"
)

var _ Handler = SSHHandler{}

type SSHHandler struct {
	fs      afero.Fs
	homeDir string
}

func NewSSHHandler() SSHHandler {
	homeDir, _ := os.UserHomeDir()

	return SSHHandler{
		fs:      afero.NewOsFs(),
		homeDir: homeDir,
	}
}

func (h SSHHandler) GetConfig(hostAlias string) any {
	return newSSHHost(hostAlias)
}

func (h SSHHandler) Execute(config map[string]string, hostAlias string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	sshFilePath := filepath.Join(h.homeDir, ".ssh", "config")

	alreadyExists, err := h.contentAlreadyExists(sshFilePath, hostAlias)
	if err != nil {
		return nil, fmt.Errorf("error checking if %s already exists in %s file: %v", hostAlias, sshFilePath, err)
	}

	if alreadyExists {
		diagnostics[sshFilePath] = fmt.Sprintf("skipping write: %s host already present", hostAlias)
		return diagnostics, nil
	}

	if err := h.writeFile(config, sshFilePath); err != nil {
		return nil, fmt.Errorf("error writing to %s file: %v", sshFilePath, err)
	}

	return nil, nil
}

func (h SSHHandler) contentAlreadyExists(sshFilePath, hostAlias string) (bool, error) {
	f, err := h.fs.OpenFile(sshFilePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return false, fmt.Errorf("error opening ssh config file: %v", err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return false, fmt.Errorf("error parsing ssh config: %v", err)
	}

	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns {
			if pattern.String() == hostAlias {
				return true, nil
			}
		}
	}

	return false, nil
}

func (h SSHHandler) writeFile(config map[string]string, sshFilePath string) error {
	f, err := h.fs.OpenFile(sshFilePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("error opening ssh config file: %v", err)
	}
	defer f.Close()

	hostBlock := h.buildHostBlock(config)
	if _, err := f.Seek(0, 2); err != nil {
		return fmt.Errorf("error seeking to end of ssh config: %v", err)
	}

	_, err = f.Write(hostBlock)
	return err
}

func (h SSHHandler) buildHostBlock(config map[string]string) []byte {
	const hostTmpl = `
Host {{ index . "alias" }}
  HostName {{ index . "host_name" }}
  User {{ index . "ssh_user" }}
  IdentityFile {{ index . "ssh_public_key_path" }}
  Port {{ index . "ssh_port" }}
`

	tmpl, err := template.New("sshConfig").Parse(hostTmpl)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		panic(err)
	}

	return buf.Bytes()
}
