package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type ansibleHandler struct{}

func newAnsibleHandler() ansibleHandler {
	return ansibleHandler{}
}

func (h ansibleHandler) execute(c playbookConfig, playbook string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	if err := h.generateHostVars(c); err != nil {
		return diagnostics, fmt.Errorf("error generating host vars: %v", err)
	}

	if err := h.runAnsiblePlaybook(playbook); err != nil {
		return diagnostics, fmt.Errorf("error running ansible playbook: %v", err)
	}

	return diagnostics, nil
}

func (h ansibleHandler) generateHostVars(c playbookConfig) error {
	for _, host := range c.hosts() {
		hostAsMap, err := host.asMap()
		if err != nil {
			return fmt.Errorf("error getting host config as map for %s: %v", host.name(), err)
		}

		if err := h.writeHostVarsFile(host.name(), hostAsMap); err != nil {
			return fmt.Errorf("error writing host vars file for %s: %v", host.name(), err)
		}
	}

	return nil
}

func (h ansibleHandler) runAnsiblePlaybook(playbook string) error {
	playbookPath := filepath.Join("ansible", "playbooks", playbook+".yml")
	args := []string{
		"-i", filepath.Join(".generated", "ansible", "inventory", "hosts.yml"),
		playbookPath,
	}

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running ansible playbook: %v", err)
	}

	return nil
}

func (h ansibleHandler) writeHostVarsFile(hostname string, vars any) error {
	dir := filepath.Join(".generated", "host_vars", hostname)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating host_vars dir for %s: %v", hostname, err)
	}

	data, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("error marshaling host vars for %s: %v", hostname, err)
	}

	if err := os.WriteFile(filepath.Join(dir, "vars.yml"), data, 0o644); err != nil {
		return fmt.Errorf("error writing host vars file for %s: %v", hostname, err)
	}

	return nil
}
