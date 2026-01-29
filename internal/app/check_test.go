package app

import (
	"errors"
	"testing"

	"github.com/dannyvelas/conflux"
)

func TestCheckError(t *testing.T) {
	configMux := conflux.NewConfigMux()

	tests := []struct {
		name        string
		targetArgs  []string
		expectedErr error
	}{
		{
			name:        "SSH unsupported action",
			targetArgs:  []string{"ssh:run"},
			expectedErr: errUnsupportedCombination,
		},
		{
			name:        "SSH trailing word",
			targetArgs:  []string{"ssh:run:here"},
			expectedErr: errExpectedEOF,
		},
		{
			name:        "SSH unrecognized action",
			targetArgs:  []string{"ssh:ssh"},
			expectedErr: errUnrecognizedAction,
		},
		{
			name:        "Ansible unrecognized subcommand",
			targetArgs:  []string{"ansible:hi"},
			expectedErr: errUnrecognizedAnsibleSubCommand,
		},
		{
			name:        "Multiple: SSH unsupported action and ansible unrecognized subcommand",
			targetArgs:  []string{"ssh:run", "ansible:hi"},
			expectedErr: errUnrecognizedAnsibleSubCommand,
		},
		{
			name:        "Multiple: SSH unsupported action and ansible action instead of subcommand",
			targetArgs:  []string{"ssh:run", "ansible:run"},
			expectedErr: errUnrecognizedAnsibleSubCommand,
		},
		{
			name:        "Multiple: SSH unsupported action and ansible missing action",
			targetArgs:  []string{"ssh:run", "ansible:inventory"},
			expectedErr: errExpectingAction,
		},
		{
			name:        "Multiple: SSH unsupported action and ansible unsupported action",
			targetArgs:  []string{"ssh:run", "ansible:inventory:apply"},
			expectedErr: errUnsupportedCombination,
		},
		{
			name:        "Multiple: SSH unsupported action and ansible unrecognized action",
			targetArgs:  []string{"ssh:run", "ansible:inventory:asdf"},
			expectedErr: errUnrecognizedAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Check(configMux, "proxmox", tt.targetArgs)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("ToTargets() error was %v. wanted: %v", err, tt.expectedErr)
				return
			}
		})
	}
}
