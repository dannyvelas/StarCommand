package app

import (
	"testing"
	"testing/fstest"

	"github.com/dannyvelas/conflux"
	"github.com/google/go-cmp/cmp"
)

func TestGetConfig(t *testing.T) {
	cases := []struct {
		name                string
		hostAlias           string
		targets             []string
		expectedConfig      map[string]string
		expectedDiagnostics map[string]string
	}{
		{
			name:      "ansible as target",
			hostAlias: "proxmox",
			targets:   []string{"ansible"},
			expectedConfig: map[string]string{
				"ssh_port":                "17031",
				"ssh_public_key_path":     "~/.ssh/id_ed25519.pub",
				"gateway_address":         "10.0.0.1",
				"physical_nic":            "enx6c1ff7135975",
				"auto_update_reboot_time": "05:00",
				"admin_email":             "admin@example.com",
				"admin_password":          "not-a-password",
				"smtp_user":               "admin",
				"smtp_password":           "not-a-password",
			},
			expectedDiagnostics: map[string]string{},
		},
		{
			name:      "ssh as target",
			hostAlias: "proxmox",
			targets:   []string{"ssh"},
			expectedConfig: map[string]string{
				"alias":               "proxmox",
				"host_name":           "10.0.0.50",
				"ssh_user":            "admin",
				"ssh_public_key_path": "~/.ssh/id_ed25519.pub",
				"ssh_port":            "17031",
				"node_cidr_address":   "10.0.0.50/24",
			},
			expectedDiagnostics: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := fstest.MapFS{
				"config/all.yml":     {Data: []byte("ssh_user: \"admin\"\nssh_port: 17031\nssh_public_key_path: \"~/.ssh/id_ed25519.pub\"\ngateway_address: 10.0.0.1\nphysical_nic: \"enx6c1ff7135975\"\nauto_update_reboot_time: \"05:00\"\n")},
				"config/proxmox.yml": {Data: []byte("node_cidr_address: 10.0.0.50/24\n")},
			}
			mockEnv := []string{"ADMIN_EMAIL=admin@example.com", "PROXMOX_ADMIN_PASSWORD=not-a-password", "SMTP_USER=admin", "SMTP_PASSWORD=not-a-password"}
			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader("config/all.yml", conflux.WithPath("config/proxmox.yml"), conflux.WithFileSystem(mockFS)),
				conflux.WithEnvReader(conflux.WithEnviron(mockEnv)),
			)

			a, err := New(configMux, tc.hostAlias, tc.targets)
			if err != nil {
				t.Fatalf("unexpected error initializing app: %v", err)
			}

			gotConfig, gotDiagnostics, err := a.GetConfig()
			if err != nil {
				t.Fatalf("unexpected error getting config: %v", err)
			}

			for key, value := range tc.expectedConfig {
				found, ok := gotConfig[key]
				if !ok {
					t.Errorf("error: missing entry %s=%s in config", key, value)
					continue
				}

				if found != value {
					t.Errorf("error: for key %s, expected %s but found=%s.", key, value, found)
				}
			}

			if diff := cmp.Diff(tc.expectedDiagnostics, gotDiagnostics); diff != "" {
				t.Errorf("diagnostics mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
