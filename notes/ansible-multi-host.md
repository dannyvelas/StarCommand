# Multi-Host Ansible Support

## Context

### What this project is

`stc` is a Go CLI for managing production server infrastructure. Users fork the repo, fill in
`stc.yml` with their server details, and run `stc` commands to provision servers and
VMs. The CLI abstracts Ansible and Terraform so users don't have to manually manage
config files, ordering of operations, or tool-specific flags.

### User configuration (`stc.yml` → `config.Config`)

Users describe their infrastructure in `stc.yml`. It maps to `config.Config`
(`config/models.go`):

```
Config
  Hosts []Host
    Host
      Name                  string   // used as CLI arg and Ansible inventory hostname
      IP                    string
      SSH
        User                string
        Port                int
        PrivateKeyPath      string
        PublicKeyPath       string
      AutoUpdateRebootTime  string   // e.g. "05:00"
      WireGuardEndpoint     bool
      Incus
        StoragePoolName     string
        StoragePoolDriver   string
      VMs []VM
        VM
          Name              string
          IP                string
          SSH               SSHConfig
          AutoUpdateRebootTime string
```

All fields are per-host.

Sensitive values (`admin_password`, `smtp_password`, etc.) are never stored in
`stc.yml`. The CLI either prompts for them interactively or reads them from `STC_*`
env vars at runtime.

### Ansible inventory (`stc inventory generate`)

`stc inventory generate` produces a minimal YAML inventory at `ansible/inventory.yml`.
The current implementation is a **stub** (`return nil, nil` in `internal/app/app.go`).
The intended output format is:

```yaml
all:
  children:
    metal:
      hosts:
        host-01:
          ansible_host: 192.168.1.10
        host-02:
          ansible_host: 192.168.1.11
    vms:
      hosts:
        host-01-vm:
          ansible_host: 192.168.122.49
          parent_host: host-01
        host-02-vm:
          ansible_host: 192.168.123.49
          parent_host: host-02
```

The inventory is **intentionally minimal** — it only contains hostnames and IPs. It
does not include connection params (`ansible_port`, `ansible_ssh_private_key_file`) or
playbook-specific vars.

### Ansible playbooks

Three playbooks live in `ansible/playbooks/`:

| Playbook           | `hosts:` target | What it does                                                           |
|--------------------|-----------------|------------------------------------------------------------------------|
| `bootstrap-server` | `all`           | OS hardening, creates `admin` user, sets SSH port, disables root login |
| `setup-host`       | `metal`         | Installs Incus, configures postfix relay                               |
| `setup-vm`         | `vms`           | Installs Docker, sets up storage                                       |

**Critical detail for `bootstrap-server`**: before this playbook runs, only `root` can
SSH into a fresh Debian server. The playbook creates the `admin` user and disables root
login. After it runs, only `admin` can SSH.

### The `configLoader` pattern (`internal/app/load.go`)

All config structs that need to be populated from `stc.yml` implement:

```go
import "github.com/dannyvelas/starcommand/config"

type configLoader interface {
	FillFromConfig(c *config.Config) error
	FillInKeys() error
}
```

`loadConfig(cfg configLoader, c *config.Config, name string)` orchestrates:
1. `FillFromConfig(c)` — populate fields from `stc.yml`
2. `buildDiagnostics(cfg)` — reflect over struct fields with `required:"true"` tag; any
   zero-valued required field is marked as `"missing"`. Returns an error with a table if
   any are missing.
3. `FillInKeys()` — derive injected fields (e.g. expand `~` in paths, read a public key
   file into a string field)

### Current `ansibleRun` flow

Entry point: `AnsibleRun(ctx, c *config.Config, playbook string, preflight bool)` in
`internal/app/app.go`.

1. `ansibleHandler.getConfig(playbook)` returns an `ansibleConfig` (one of three
   structs below)
2. `loadConfig(playbookConfig, c, playbook)` — calls `FillFromConfig`, validates
   required fields, calls `FillInKeys`. If `preflight` is true, returns diagnostics
   without running.
3. `promptSensitiveFields(playbookConfig, stdin, stdout)` — prompts interactively for
   fields tagged `sensitive:"true"`, or reads from `STC_<FIELD_NAME>` env vars.
4. `ansibleHandler.execute(playbookConfig)` → `runAnsiblePlaybook(playbookConfig)`:
   - Does an **SSH connectivity check** by dialing `NodeIP:SSHPort` as `SSHUser` with
     `SSHPrivateKeyPath`
   - Encodes the config struct to a temp JSON file (`/tmp/labctl-vars-*.json`)
   - Runs `ansible-playbook -i ansible/inventory.ini ansible/setup-proxmox.yml -e @<tmpfile>`
   - If the SSH check returned `errConnectingSSH`, appends `-u root` to the command

### Why the SSH check is lazy

The SSH check happens at playbook-run time, not during inventory generation. This is by
design: host SSH state changes dynamically. Before `bootstrap-server` runs, only root
can SSH. After it runs, only admin can. Running the check eagerly (e.g. at `stc
inventory generate` time) would bake in a stale result.

### Ansible config structs

All three ansible config structs embed `ansibleBaseConfig`
(`internal/app/ansible_base_config.go`):

```go
type ansibleBaseConfig struct {
	NodeIP            string `json:"node_ip" required:"true"`
	SSHUser           string `json:"ssh_user" required:"true"`
	SSHPort           string `json:"ssh_port" required:"true"`
	SSHPrivateKeyPath string `json:"ssh_private_key_path" required:"true"`

	// Injected by fillInBaseKeys()
	AnsibleUser string `json:"ansible_user"`
	AnsiblePort string `json:"ansible_port"`
}
```

Additional fields per config:

**`ansibleBootstrapConfig`** (`ansible_bootstrap_config.go`):
- Required: `SSHPublicKeyPath`, `AutoUpdateRebootTime` (default `"05:00"`)
- Injected: `SSHPublicKey` (file contents of `SSHPublicKeyPath`, read in `FillInKeys`)
- Sensitive: `AdminEmail`, `AdminPassword`

**`ansibleSetupHostConfig`** (`ansible_setup_host_config.go`):
- Required: `IncusStoragePoolName`, `IncusStorageDriver`
- Sensitive: `SMTPUser`, `SMTPPassword`

**`ansibleSetupVMConfig`** (`ansible_setup_vm_config.go`):
- No additional fields

### What is currently broken

- All `FillFromConfig` methods are stubs that return `nil`. Required fields in
  `ansibleBaseConfig` (and the playbook-specific structs) are never populated, so
  `loadConfig` always returns a "missing fields" error. No ansible command currently
  works end-to-end.
- `ansible_handler.go` hardcodes `ansible/setup-proxmox.yml` for all playbooks.
  Should use the actual playbook path per command (e.g.
  `ansible/playbooks/bootstrap-server.yml`).
- The inventory path is hardcoded as `ansible/inventory.ini`. Should be
  `ansible/inventory.yml`.
- The SSH check and `-u root` flag are single-host only. There is no multi-host
  support.

---

## Feature Requirements

### R1 — Fix hardcoded paths in `ansible_handler.go`

`runAnsiblePlaybook` hardcodes `ansible/setup-proxmox.yml` and `ansible/inventory.ini`.
These must be corrected:
- Inventory: `ansible/inventory.yml`
- Playbook: passed in per-command, e.g. `ansible/playbooks/bootstrap-server.yml`

The playbook path mapping belongs in `getConfig` or `execute`, whichever is cleaner.

### R2 — Implement `FillFromConfig` for all ansible config structs

Each struct's `FillFromConfig` must populate required fields from `*config.Config`.

The open design question is **where per-host vars come from** when there are multiple
hosts. Two options:

**Option A (simpler)**: Treat shared fields as globally uniform.
In practice, all hosts in a  use the same SSH key, port, and user. `FillFromConfig`
reads these from `cfg.Hosts[0]` as a representative value. The single global extra vars
JSON applies uniformly to all hosts.

**Option B (architecturally cleaner)**: `stc inventory generate` writes all per-host
vars into the inventory YAML (`ansible_port`, `ansible_ssh_private_key_file`,
`auto_update_reboot_time`, `incus_storage_pool_name`, etc.). `FillFromConfig` for
ansible configs only populates the truly global vars (sensitive fields, `ssh_public_key`
content). This requires implementing `stc inventory generate` first.

Given the current state (inventory generation is a stub), Option A unblocks everything
now. Option B is a follow-up once inventory generation is implemented.

Regardless of option, the field mapping is:

| Struct field | Source in `config.Config` |
|---|---|
| `NodeIP` | `Hosts[0].IP` |
| `SSHUser` | `Hosts[0].SSH.User` |
| `SSHPort` | `Hosts[0].SSH.Port` (int → string; default `"22"` if zero) |
| `SSHPrivateKeyPath` | `Hosts[0].SSH.PrivateKeyPath` |
| `SSHPublicKeyPath` | `Hosts[0].SSH.PublicKeyPath` |
| `AutoUpdateRebootTime` | `Hosts[0].AutoUpdateRebootTime` (keep constructor default `"05:00"` if empty) |
| `IncusStoragePoolName` | `Hosts[0].Incus.StoragePoolName` |
| `IncusStorageDriver` | `Hosts[0].Incus.StoragePoolDriver` |

### R3 — Multi-host SSH check

`runAnsiblePlaybook` currently dials a single host. It must dial all relevant hosts
from `*config.Config` concurrently and record the result (admin SSH succeeded or
failed) per host.

For `bootstrap-server` and `setup-host`: iterate `cfg.Hosts`.
For `setup-vm`: iterate all VMs across `cfg.Hosts[*].VMs`.

`runAnsiblePlaybook` needs access to `*config.Config` (it currently only receives
`ansibleConfig`). Update the signature accordingly.

### R4 — Per-host `ansible_user` without modifying the inventory

The SSH check result determines whether each host needs `ansible_user=root` or the
configured user. Ansible has no per-host flag mechanism outside of inventory or
`host_vars/` files. The approach that avoids touching the inventory:

Partition hosts into two groups based on SSH check results, then run `ansible-playbook`
**twice in parallel**:
- `ansible-playbook -i ansible/inventory.yml <playbook> --limit host-01,host-03 -u root -e @<tmpfile>` (hosts where admin SSH failed)
- `ansible-playbook -i ansible/inventory.yml <playbook> --limit host-02 -e @<tmpfile>` (hosts where admin SSH succeeded)

Skip whichever group is empty. Wait for both processes to finish. Surface errors from
either.

The `--limit` value is a comma-separated list of host names (matching the names in
`stc.yml` / inventory).

### Summary of files to change

| File | Change |
|---|---|
| `internal/app/ansible_handler.go` | Fix hardcoded paths; update SSH check to iterate all hosts; run playbook twice in parallel with `--limit` |
| `internal/app/ansible_base_config.go` | Add `fillBaseFromHost` / `fillBaseFromVM` helpers |
| `internal/app/ansible_bootstrap_config.go` | Implement `FillFromConfig` |
| `internal/app/ansible_setup_host_config.go` | Implement `FillFromConfig` |
| `internal/app/ansible_setup_vm_config.go` | Implement `FillFromConfig` |
