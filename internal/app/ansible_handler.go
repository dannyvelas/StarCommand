package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ansibleHandler struct{}

func newAnsibleHandler() ansibleHandler {
	return ansibleHandler{}
}

func (h ansibleHandler) execute(c playbookConfig, playbook string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	if err := c.generateHostVars(); err != nil {
		return diagnostics, fmt.Errorf("error generating host vars: %v", err)
	}

	if err := h.runAnsiblePlaybook(playbook); err != nil {
		return diagnostics, fmt.Errorf("error running ansible playbook: %v", err)
	}

	return diagnostics, nil
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
