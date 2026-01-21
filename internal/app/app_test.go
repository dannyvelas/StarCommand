package app

import (
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

func TestGetConfig(t *testing.T) {
	cases := []struct {
		name           string
		hostAlias      string
		targets        []string
		expectedConfig map[string]string
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
		},
		{
			name:      "ansible and ssh as targets",
			hostAlias: "proxmox",
			targets:   []string{"ansible", "ssh"},
			expectedConfig: map[string]string{
				"admin_email":             "admin@example.com",
				"admin_password":          "not-a-password",
				"alias":                   "proxmox",
				"auto_update_reboot_time": "05:00",
				"gateway_address":         "10.0.0.1",
				"host_name":               "10.0.0.50",
				"node_cidr_address":       "10.0.0.50/24",
				"node_ip":                 "10.0.0.50",
				"physical_nic":            "enx6c1ff7135975",
				"smtp_password":           "not-a-password",
				"smtp_user":               "admin",
				"ssh_port":                "17031",
				"ssh_public_key_path":     "~/.ssh/id_ed25519.pub",
				"ssh_user":                "admin",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := fstest.MapFS{"config/all.yml": {Data: []byte(testYAML)}}

			configMux := conflux.NewConfigMux(conflux.WithYAMLFileReader("config/all.yml", conflux.WithFileSystem(mockFS)))

			a, err := New(configMux, tc.hostAlias, tc.targets)
			if err != nil {
				t.Fatalf("unexpected error initializing app: %v", err)
			}

			gotConfig, _, err := a.GetConfig()
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
			targets:   []string{"ansible"},
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
			targets:   []string{"ssh"},
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
			targets:   []string{"ansible", "ssh"},
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

			a, err := New(configMux, tc.hostAlias, tc.targets)
			if err != nil {
				t.Fatalf("unexpected error initializing app: %v", err)
			}

			gotDiagnostics, err := a.CheckConfig()
			if err != nil {
				t.Fatalf("unexpected error getting config: %v", err)
			}

			if diff := cmp.Diff(tc.expectedDiagnostics, gotDiagnostics); diff != "" {
				t.Errorf("diagnostics mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
