package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/kevinburke/ssh_config"
	"github.com/spf13/afero"
)

type sshHandler struct {
	fs      afero.Fs
	homeDir string
}

func newSSHHandler(homeDir string) sshHandler {
	return sshHandler{
		fs:      afero.NewOsFs(),
		homeDir: homeDir,
	}
}

func (h sshHandler) execute(sshConfig *sshConfig, hostAlias string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	sshFilePath := filepath.Join(h.homeDir, ".ssh", "config")

	alreadyExists, err := h.contentAlreadyExists(sshFilePath, hostAlias)
	if err != nil {
		return diagnostics, fmt.Errorf("error checking if %s already exists in %s file: %v", hostAlias, sshFilePath, err)
	}

	if alreadyExists {
		diagnostics["Write to "+sshFilePath] = fmt.Sprintf("skipping: %s host already present", hostAlias)
		return diagnostics, nil
	}

	if err := h.writeFile(sshConfig, sshFilePath); err != nil {
		return diagnostics, fmt.Errorf("error writing to %s file: %v", sshFilePath, err)
	}

	return diagnostics, nil
}

func (h sshHandler) contentAlreadyExists(sshFilePath, hostAlias string) (bool, error) {
	f, err := h.fs.OpenFile(sshFilePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return false, fmt.Errorf("error opening ssh config file: %v", err)
	}
	defer func() { _ = f.Close() }()

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

func (h sshHandler) writeFile(config *sshConfig, sshFilePath string) error {
	f, err := h.fs.OpenFile(sshFilePath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("error opening ssh config file: %v", err)
	}
	defer func() { _ = f.Close() }()

	hostBlock := h.buildHostBlock(config)
	if _, err := f.Seek(0, 2); err != nil {
		return fmt.Errorf("error seeking to end of ssh config: %v", err)
	}

	_, err = f.Write(hostBlock)
	return err
}

func (h sshHandler) buildHostBlock(config *sshConfig) []byte {
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
	if err := tmpl.Execute(&buf, config); err != nil {
		panic(err)
	}

	return buf.Bytes()
}
