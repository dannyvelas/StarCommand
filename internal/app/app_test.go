package app

import (
	"bytes"
	"html/template"
	"testing"
	"testing/fstest"

	"github.com/dannyvelas/conflux"
	"github.com/google/go-cmp/cmp"
)

func TestGetConfig(t *testing.T) {
	const (
		adminEmail           = "admin@example.com"
		adminPassword        = "not-a-password"
		alias                = "proxmox"
		autoUpdateRebootTime = "05:00"
		gatewayAddress       = "10.0.0.1"
		hostName             = "10.0.0.50"
		nodeCIDRAddress      = "10.0.0.50/24"
		nodeIP               = "10.0.0.50"
		physicalNIC          = "enx6c1ff7135975"
		smtpPassword         = "not-a-password"
		smtpUser             = "admin"
		sshPort              = "17031"
		sshPublicKeyPath     = "~/.ssh/id_ed25519.pub"
		sshUser              = "admin"
	)

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
				"admin_email":             adminEmail,
				"admin_password":          adminPassword,
				"auto_update_reboot_time": autoUpdateRebootTime,
				"gateway_address":         gatewayAddress,
				"physical_nic":            physicalNIC,
				"smtp_password":           smtpPassword,
				"smtp_user":               smtpUser,
				"ssh_port":                sshPort,
				"ssh_public_key_path":     sshPublicKeyPath,
			},
			expectedDiagnostics: map[string]string{},
		},
		{
			name:      "ssh as target",
			hostAlias: "proxmox",
			targets:   []string{"ssh"},
			expectedConfig: map[string]string{
				"alias":               alias,
				"host_name":           hostName,
				"ssh_user":            sshUser,
				"ssh_public_key_path": sshPublicKeyPath,
				"ssh_port":            sshPort,
				"node_cidr_address":   nodeCIDRAddress,
			},
			expectedDiagnostics: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const configTemplate = `admin_email: "{{.AdminEmail}}"
auto_update_reboot_time: "{{.AutoUpdateRebootTime}}"
gateway_address: {{.GatewayAddress}}
node_cidr_address: {{.NodeCIDRAddress}}
physical_nic: "{{.PhysicalNIC}}"
proxmox_admin_password: "{{.ProxmoxAdminPassword}}"
smtp_password: "{{.SMTPPassword}}"
smtp_user: "{{.SMTPUser}}"
ssh_port: {{.SSHPort}}
ssh_public_key_path: "{{.SSHPublicKeyPath}}"
ssh_user: "{{.SSHUser}}"
`

			// 2. Prepare the data (using a struct or a map)
			data := struct {
				AdminEmail           string
				AutoUpdateRebootTime string
				GatewayAddress       string
				NodeCIDRAddress      string
				PhysicalNIC          string
				ProxmoxAdminPassword string
				SMTPPassword         string
				SMTPUser             string
				SSHPort              string
				SSHPublicKeyPath     string
				SSHUser              string
			}{
				AdminEmail:           adminEmail,
				AutoUpdateRebootTime: autoUpdateRebootTime,
				GatewayAddress:       gatewayAddress,
				NodeCIDRAddress:      nodeCIDRAddress,
				PhysicalNIC:          physicalNIC,
				ProxmoxAdminPassword: adminPassword,
				SMTPPassword:         smtpPassword,
				SMTPUser:             smtpUser,
				SSHPort:              sshPort,
				SSHPublicKeyPath:     sshPublicKeyPath,
				SSHUser:              sshUser,
			}

			// 3. Render the template
			var tpl bytes.Buffer
			tmpl := template.Must(template.New("config").Parse(configTemplate))
			if err := tmpl.Execute(&tpl, data); err != nil {
				t.Fatalf("failed to render template: %v", err)
			}

			// 4. Use it in MapFS
			mockFS := fstest.MapFS{"config/all.yml": {Data: tpl.Bytes()}}

			configMux := conflux.NewConfigMux(conflux.WithYAMLFileReader("config/all.yml", conflux.WithFileSystem(mockFS)))

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
