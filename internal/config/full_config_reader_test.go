package config

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

type testConfig struct {
	SSHPublicKeyPath     string `json:"ssh_public_key_path" required:"true"`
	NodeCIDRAddress      string `json:"node_cidr_address" required:"true"`
	GatewayAddress       string `json:"gateway_address" required:"true"`
	PhysicalNIC          string `json:"physical_nic" required:"true"`
	SSHPort              string `json:"ssh_port" required:"true"`
	AutoUpdateRebootTime string `json:"auto_update_reboot_time" required:"true"`
}

func (c *testConfig) Validate(m map[string]string) bool {
	return true
}

func TestFullConfigReader_Read(t *testing.T) {
	cases := []struct {
		name     string
		fs       fs.FS
		env      []string
		hostName string
		expected testConfig
	}{
		{
			name: "no bitwarden env variables, bitwarden doesn't get used and all values get filled anyway",
			fs: fstest.MapFS{
				"config/all.yml":     {Data: []byte("ssh_port: 17031\nssh_public_key_path: \"~/.ssh/id_ed25519.pub\"\ngateway_address: 10.0.0.1\nphysical_nic: \"enx6c1ff7135975\"\nauto_update_reboot_time: \"05:00\"\n")},
				"config/proxmox.yml": {Data: []byte("node_cidr_address: 10.0.0.50/24\n")},
			},
			env:      []string{},
			hostName: "proxmox",
			expected: testConfig{
				SSHPublicKeyPath:     "~/.ssh/id_ed25519.pub",
				NodeCIDRAddress:      "10.0.0.50/24",
				GatewayAddress:       "10.0.0.1",
				PhysicalNIC:          "enx6c1ff7135975",
				SSHPort:              "17031",
				AutoUpdateRebootTime: "05:00",
			},
		},
		{
			name: "only some bitwarden env variables, bitwarden doesn't get used and all values get filled anyway",
			fs: fstest.MapFS{
				"config/all.yml":     {Data: []byte("ssh_port: 17031\nssh_public_key_path: \"~/.ssh/id_ed25519.pub\"\ngateway_address: 10.0.0.1\nphysical_nic: \"enx6c1ff7135975\"\nauto_update_reboot_time: \"05:00\"\n")},
				"config/proxmox.yml": {Data: []byte("node_cidr_address: 10.0.0.50/24\n")},
			},
			env:      []string{"BWS_ACCESS_TOKEN=123"},
			hostName: "proxmox",
			expected: testConfig{
				SSHPublicKeyPath:     "~/.ssh/id_ed25519.pub",
				NodeCIDRAddress:      "10.0.0.50/24",
				GatewayAddress:       "10.0.0.1",
				PhysicalNIC:          "enx6c1ff7135975",
				SSHPort:              "17031",
				AutoUpdateRebootTime: "05:00",
			},
		},
		{
			name: "<host>.yml file overrides all.yml",
			fs: fstest.MapFS{
				"config/all.yml":     {Data: []byte("ssh_port: 17031\nssh_public_key_path: \"~/.ssh/id_ed25519.pub\"\ngateway_address: 10.0.0.1\nphysical_nic: \"enx6c1ff7135975\"\nauto_update_reboot_time: \"05:00\"\n")},
				"config/proxmox.yml": {Data: []byte("ssh_port: 2222\nnode_cidr_address: 10.0.0.50/24\nssh_public_key_path: \"~/.ssh/other.pub\"\n")},
			},
			env:      []string{},
			hostName: "proxmox",
			expected: testConfig{
				SSHPublicKeyPath:     "~/.ssh/other.pub",
				NodeCIDRAddress:      "10.0.0.50/24",
				GatewayAddress:       "10.0.0.1",
				PhysicalNIC:          "enx6c1ff7135975",
				SSHPort:              "2222",
				AutoUpdateRebootTime: "05:00",
			},
		},
		{
			name: "env value overrides <host>.yml file",
			fs: fstest.MapFS{
				"config/all.yml":     {Data: []byte("ssh_port: 17031\nssh_public_key_path: \"~/.ssh/id_ed25519.pub\"\ngateway_address: 10.0.0.1\nphysical_nic: \"enx6c1ff7135975\"\nauto_update_reboot_time: \"05:00\"\n")},
				"config/proxmox.yml": {Data: []byte("ssh_port: 2222\nnode_cidr_address: 10.0.0.50/24\nssh_public_key_path: \"~/.ssh/other.pub\"\n")},
			},
			env:      []string{"SSH_PORT=9999", "SSH_PUBLIC_KEY_PATH=~/.ssh/env.pub"},
			hostName: "proxmox",
			expected: testConfig{
				SSHPublicKeyPath:     "~/.ssh/env.pub",
				NodeCIDRAddress:      "10.0.0.50/24",
				GatewayAddress:       "10.0.0.1",
				PhysicalNIC:          "enx6c1ff7135975",
				SSHPort:              "9999",
				AutoUpdateRebootTime: "05:00",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewFullConfigReader(tc.hostName, false, WithFilesystem(tc.fs), WithEnviron(tc.env))
			target := testConfig{}
			if _, err := Unmarshal(r, &target); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if target.SSHPublicKeyPath == "" {
				t.Errorf("SSHPublicKeyPath was empty. expected: %s.", "~/.ssh/id_ed25519.pub")
			}
			if target.NodeCIDRAddress == "" {
				t.Errorf("NodeCIDRAddress was empty. expected: %s.", "10.0.0.50/24")
			}
			if target.GatewayAddress == "" {
				t.Errorf("GatewayAddress was empty. expected: %s.", "10.0.0.1")
			}
			if target.PhysicalNIC == "" {
				t.Errorf("PhysicalNIC was empty. expected: %s.", "enx6c1ff7135975")
			}
			if target.SSHPort == "" {
				t.Errorf("SSHPort was empty. expected: %s.", "17031")
			}
			if target.AutoUpdateRebootTime == "" {
				t.Errorf("AutoUpdateRebootTime was empty. expected: %s.", "05:00")
			}
		})
	}
}
