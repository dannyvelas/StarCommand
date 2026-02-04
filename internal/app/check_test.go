package app

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/dannyvelas/conflux"
	"github.com/google/go-cmp/cmp"
)

const testYAML = `admin_email: "admin@example.com"
auto_update_reboot_time: "05:00"
gateway_address: 10.0.0.1
node_cidr_address: 10.0.0.50/24
physical_nic: "enx6c1ff7135975"
proxmox_admin_password: "not-a-password"
smtp_password: "not-a-password"
smtp_user: "admin"
ssh_port: 17031
ssh_public_key_path: "~/.ssh/id_ed25519.pub"
ssh_user: "admin"
`

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
			_, err := Check(context.Background(), configMux, "proxmox", tt.targetArgs)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("ToTargets() error was %v. wanted: %v", err, tt.expectedErr)
				return
			}
		})
	}
}

func TestCheckConfig(t *testing.T) {
	cases := []struct {
		name                string
		hostAlias           string
		targets             []string
		expectedDiagnostics map[string]string
	}{
		{
			name:      "ansible as target",
			hostAlias: "proxmox",
			targets:   []string{"ansible:playbook:run"},
			expectedDiagnostics: map[string]string{
				"admin_email":             "loaded",
				"auto_update_reboot_time": "loaded",
				"gateway_address":         "loaded",
				"node_cidr_address":       "loaded",
				"physical_nic":            "loaded",
				"proxmox_admin_password":  "loaded",
				"smtp_password":           "loaded",
				"smtp_user":               "loaded",
				"ssh_port":                "loaded",
				"ssh_public_key_path":     "loaded",
			},
		},
		{
			name:      "ssh as target",
			hostAlias: "proxmox",
			targets:   []string{"ssh:add"},
			expectedDiagnostics: map[string]string{
				"node_cidr_address":   "loaded",
				"ssh_port":            "loaded",
				"ssh_public_key_path": "loaded",
				"ssh_user":            "loaded",
			},
		},
		{
			name:      "ansible and ssh as targets",
			hostAlias: "proxmox",
			targets:   []string{"ansible:playbook:run", "ssh:add"},
			expectedDiagnostics: map[string]string{
				"admin_email":             "loaded",
				"auto_update_reboot_time": "loaded",
				"gateway_address":         "loaded",
				"node_cidr_address":       "loaded",
				"physical_nic":            "loaded",
				"proxmox_admin_password":  "loaded",
				"smtp_password":           "loaded",
				"smtp_user":               "loaded",
				"ssh_port":                "loaded",
				"ssh_public_key_path":     "loaded",
				"ssh_user":                "loaded",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := fstest.MapFS{"config/all.yml": {Data: []byte(testYAML)}}

			configMux := conflux.NewConfigMux(conflux.WithYAMLFileReader("config/all.yml", conflux.WithFileSystem(mockFS)))

			gotDiagnostics, err := Check(context.Background(), configMux, tc.hostAlias, tc.targets)
			if err != nil {
				t.Fatalf("unexpected error getting config: %v", err)
			}

			if diff := cmp.Diff(tc.expectedDiagnostics, gotDiagnostics); diff != "" {
				t.Errorf("diagnostics mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
