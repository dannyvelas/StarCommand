# MVP Backlog

## MVP Scope

The MVP delivers a fully working `stc setup` that provisions a set of Debian hosts
end-to-end from a single command. It includes every command `stc setup` depends on, plus
`--preflight` on all commands.

**In scope:** `inventory generate`, `ansible bootstrap-host`, `ansible setup-host`,
`ssh add`, `setup`, `--preflight` on all of the above.

**Out of scope (post-MVP):** `ansible setup-vm`, `terraform apply`, `wg add`, `status`,
`teardown`.

---

## Notes for all tasks

**YAML library:** Use `github.com/goccy/go-yaml` for all YAML parsing and marshaling.
`gopkg.in/yaml.v3` was archived in 2024 and should not be used.

**Go templates:** All commands that generate text files or formatted output (inventory YAML,
SSH config blocks, diagnostic table) must use `text/template` from the Go standard library.
This makes the output format easy to read and modify — you can open the template and
immediately see what the generated file will look like.

**Script tests:** Command-level behavior must be covered by txtar script tests using the
`github.com/rogpeppe/go-internal/testscript` package. These tests compile the `stc` binary,
run it inside a temporary directory populated with the files defined in the test, and assert
on stdout, stderr, exit code, and generated file contents. Test files live in
`testdata/scripts/`. Each `.txtar` file covers one scenario. The format is:

```
# Description of the test
exec stc <command> [flags]
stdout 'expected pattern'
exists .generated/some/file

-- stc.yml --
<file contents>
```

Verbs: `exec` (expect exit 0), `! exec` (expect non-zero), `stdout`, `stderr`, `exists`,
`grep <pattern> <file>`, `env KEY=VALUE`.

**Unit tests:** Logic that does not involve running the binary (parsing, rendering, resolving)
must be covered by standard Go unit tests (`*_test.go` files).

---

## Milestone 1 — Project Foundation

The project builds. `stc.yml` is loaded and parsed.

---

### Task 1 — Initialize Go project with Cobra CLI framework

**Story points:** 2

**Description**

Create a new Go project for the `stc` CLI. Use the
[Cobra](https://github.com/spf13/cobra) library for the CLI framework — it handles
subcommands, flags, and `--help` generation automatically.

**Project structure:**
```
cmd/stc/        # CLI entrypoint: main.go, root command, subcommands
internal/       # Business logic (no CLI concerns)
testdata/scripts/  # txtar script tests
Makefile
go.mod
```

**Root command requirements:**
- Binary name: `stc`
- A `version` subcommand that prints `stc v0.1.0` to stdout and exits 0
- The root command's long description (shown by `stc --help`) must be informative. It should
  give a high-level sense of what `stc` does without being exhaustive — the README and `docs/`
  cover the full details. Something like:

  ```
  stc is a CLI for provisioning and managing Debian servers.

  It treats stc.yml as the single source of truth for all infrastructure
  configuration. Run 'stc setup' to provision all hosts end-to-end: it
  generates an Ansible inventory, hardens the OS, installs host services
  (Incus, WireGuard, Postfix), provisions VMs with Terraform, and registers
  every host and VM in ~/.ssh/config — in the right order, automatically.

  Run any subcommand with --preflight to preview its required config values
  without executing anything.
  ```

  The exact wording may vary, but the description must cover: what `stc` is,
  `stc.yml` as the config source-of-truth, a brief summary of what `stc setup`
  orchestrates, and `--preflight`.

**Suppress usage on runtime errors.** By default Cobra prints the full help text whenever
any error is returned — this is noisy for runtime failures like "host not found". It should
only appear for flag/grammar errors (e.g. unknown flag). Fix this with a
`PersistentPreRunE` on the root command that sets `cmd.SilenceUsage = true`. This function
runs after flags are parsed, so usage still appears for flag errors but is hidden for errors
that happen during execution.

**Makefile targets:** `all` (default, runs `build`), `build` (compiles to `./stc`),
`clean` (removes `./stc`).

**Testscript setup:** Add `github.com/rogpeppe/go-internal/testscript` as a dependency.
Create `cmd/stc/script_test.go` containing a `TestScript` function that:
1. Builds the `stc` binary
2. Calls `testscript.Run` pointed at `testdata/scripts/`

This test runner is used by all future txtar tests throughout the project.

**Required Tests**

*Unit test — `TestVersionOutput`:*
- Run `./stc version`
- Assert stdout contains `stc v0.1.0`
- Assert exit code is 0

*Txtar test — `testdata/scripts/version.txtar`:*
```
# version subcommand prints version string
exec stc version
stdout 'stc v0.1.0'

-- stc.yml --
hosts: {}
```

The runtime-error-no-usage behavior must be implemented in task 1 (via `PersistentPreRunE`),
but the txtar test for it should be added in the first task that introduces a command with a
runtime error path (e.g. task 3 — `inventory generate` can produce a runtime error when
`stc.yml` is missing). At that point, add:

```
# testdata/scripts/runtime-error-no-usage.txtar
# Runtime errors do not print the usage block
! exec stc inventory generate
! stdout 'Usage:'
stderr '.'
```

(No `-- stc.yml --` section so the file is absent, which triggers a runtime error.)

**Definition of Done**
- `make` produces `./stc`
- `./stc --help` lists subcommands
- `./stc version` prints `stc v0.1.0`
- `go test ./...` passes (including the testscript runner)
- A runtime error does not print the usage block

---

### Task 2 — Define Config structs and stc.yml parser

**Story points:** 2

**Description**

Define the Go structs that represent `stc.yml` and implement a function to parse it.

**Schema:** `hosts` is a `map[string]Host` where each key is the host's name (e.g.
`"host-01"`). `vms` inside each host is a `map[string]VM` where each key is the VM's name.
Names are not repeated inside the object — they are the map key.

**Structs** (suggested location: `internal/models/`):

```go
type Config struct {
    Hosts map[string]Host `yaml:"hosts"`
}

type Host struct {
    IP                   string      `yaml:"ip"`
    SSH                  SSHConfig   `yaml:"ssh"`
    AutoUpdateRebootTime string      `yaml:"auto_update_reboot_time"`
    WireguardEndpoint    bool        `yaml:"wireguard_endpoint"`
    Incus                IncusConfig `yaml:"incus"`
    VMs                  map[string]VM `yaml:"vms"`
}

type VM struct {
    IP                   string    `yaml:"ip"`
    SSH                  SSHConfig `yaml:"ssh"`
    AutoUpdateRebootTime string    `yaml:"auto_update_reboot_time"`
}

type SSHConfig struct {
    User           string `yaml:"user"`
    Port           int    `yaml:"port"`
    PrivateKeyPath string `yaml:"private_key_path"`
    PublicKeyPath  string `yaml:"public_key_path"`
}

type IncusConfig struct {
    StoragePoolName   string `yaml:"storage_pool_name"`
    StoragePoolDriver string `yaml:"storage_pool_driver"`
}
```

**Parser function:** `LoadConfig(path string) (*Config, error)` using
`github.com/goccy/go-yaml`.

**Required Tests**

*Unit test — `TestLoadConfig_ValidTwoHostConfig`:*

Input `stc.yml`:
```yaml
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: true
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms:
      host-01-vm-01:
        ip: 10.0.100.10
        ssh:
          user: admin
          port: 22
          private_key_path: ~/.ssh/id_ed25519
          public_key_path: ~/.ssh/id_ed25519.pub
        auto_update_reboot_time: "05:00"
  host-02:
    ip: 192.168.1.11
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

Expected: `config.Hosts["host-01"].IP == "192.168.1.10"`,
`config.Hosts["host-01"].SSH.User == "admin"`,
`config.Hosts["host-01"].VMs["host-01-vm-01"].IP == "10.0.100.10"`,
`config.Hosts["host-02"].WireguardEndpoint == false`, no error.

*Unit test — `TestLoadConfig_MissingFile`:*
- Input: path that does not exist
- Expected: non-nil error, error message contains the path

*Unit test — `TestLoadConfig_InvalidYAML`:*
- Input: `hosts: [invalid: {`
- Expected: non-nil error

**Definition of Done**
- All three unit tests pass
- `LoadConfig` returns a correctly populated `*Config` for the valid input above

---

### Task 3 — Wire config loading into all commands

**Story points:** 1

**Description**

Load `stc.yml` from the current working directory before dispatching to any subcommand
that needs config. If loading fails, print a user-friendly error and exit 1.

The `version` subcommand must NOT require a `stc.yml` — it must work without one.

One clean approach: load config in `PersistentPreRunE` on the root command, but skip
loading for commands that declare they don't need it (e.g. check a per-command flag or
use a dedicated `PersistentPreRunE` override on the `version` subcommand).

Error format: `error: could not load stc.yml: <reason>`

**Required Tests**

*Txtar test — `testdata/scripts/missing-stc-yml.txtar`:*
```
# Missing stc.yml produces a friendly error and exits 1
! exec stc inventory generate
stderr 'stc.yml'
```
(No `-- stc.yml --` section, so the file does not exist in the test directory.)

*Txtar test — `testdata/scripts/version-no-stc-yml.txtar`:*
```
# version works without stc.yml
exec stc version
stdout 'stc v0.1.0'
```
(No `-- stc.yml --` section.)

**Definition of Done**
- `./stc inventory generate` with no `stc.yml` prints an error mentioning `stc.yml` and exits 1
- `./stc version` with no `stc.yml` exits 0
- All txtar tests pass

---

## Milestone 2 — Diagnostic Infrastructure

The `Diagnostics` type and ASCII table renderer are ready. `--preflight` is wired globally.

---

### Task 4 — Implement Diagnostic type and ASCII table renderer

**Story points:** 2

**Description**

Implement the shared `Diagnostics` type and ASCII table renderer used by every `--preflight`
output in the CLI.

**Types and methods:**
```go
type Diagnostic struct {
    Field  string
    Status string
}

type Diagnostics []Diagnostic
```

Status constants: `"loaded"`, `"not found"`, `"will prompt"`.

**Methods:**
- `(d *Diagnostics) Append(diags ...Diagnostic)` — appends entries
- `(d Diagnostics) HasErrors() bool` — returns true if any entry has status `"not found"`
- `(d Diagnostics) ToTable() string` — renders a bordered ASCII table

**`ToTable()` requirements:**
- Use `text/template` from the Go standard library to render the table
- Column widths auto-size to the widest entry in each column (including the header row)
- Header row: `CONFIG NAME` | `STATUS`
- Border line: `-` characters
- Pipe characters `|` separate columns with one space of padding on each side

Example output (column widths adapt to content):
```
---------------------------------------------
| CONFIG NAME                 | STATUS      |
---------------------------------------------
| hosts.host-01.ip            | loaded      |
| hosts.host-01.ssh.user      | not found   |
| STC_ADMIN_EMAIL             | will prompt |
---------------------------------------------
```

The template should define the per-row structure. Column widths are pre-computed and passed
as template data, not computed inside the template itself.

**Required Tests**

*Unit test — `TestDiagnosticsToTable_ColumnWidths`:*
- Input: `Diagnostics{ {"hosts.host-01.ip", "loaded"}, {"hosts.host-01.ssh.user", "not found"} }`
- Expected: table where the `CONFIG NAME` column is wide enough to fit
  `"hosts.host-01.ssh.user"` (the longer of the two) with padding, and the `STATUS` column
  is wide enough to fit `"not found"`

*Unit test — `TestDiagnosticsToTable_ExactOutput`:*
- Input: `Diagnostics{ {"hosts.host-01.ip", "loaded"} }`
- Expected: exact string match against a hardcoded expected table string (include borders,
  pipes, and spacing)

*Unit test — `TestDiagnosticsHasErrors_AllLoaded`:*
- Input: all `"loaded"` statuses
- Expected: `false`

*Unit test — `TestDiagnosticsHasErrors_OneNotFound`:*
- Input: mix of `"loaded"` and one `"not found"`
- Expected: `true`

**Definition of Done**
- All four unit tests pass
- Column widths adapt dynamically — no truncation, no excess whitespace

---

### Task 5 — Implement global `--preflight` flag

**Story points:** 1

**Description**

Add a `--preflight` boolean flag as a persistent flag on the root Cobra command so it is
automatically available to every subcommand.

```go
var preflight bool
rootCmd.PersistentFlags().BoolVar(&preflight, "preflight", false,
    "Print a diagnostic table without executing the command")
```

The `preflight` variable must be accessible to all command handler functions.

**Required Tests**

*Txtar test — `testdata/scripts/preflight-flag-accepted.txtar`:*
```
# --preflight flag is accepted by subcommands without error
exec stc inventory generate --preflight
! stderr 'unknown flag'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

**Definition of Done**
- `--preflight` is accepted by `inventory generate`, `ansible bootstrap-host`,
  `ansible setup-host`, `ssh add`, and `setup` without an "unknown flag" error
- `./stc --help` shows `--preflight` in the flags section

---

## Milestone 3 — `stc inventory generate`

The first fully working end-to-end command.

---

### Task 6 — Implement Ansible inventory YAML generation

**Story points:** 2

**Description**

Given a parsed `Config`, render the Ansible inventory YAML and write it to
`.generated/ansible/inventory/hosts.yml`, creating parent directories as needed.

**Use `text/template` to render the inventory.** Define the template so that the
structure of the output file is immediately legible from reading the template source.

The output structure must be:
```yaml
all:
  children:
    hosts:
      hosts:
        <host-name>:
          ansible_host: <host-ip>
        ...
    vms:
      hosts:
        <vm-name>:
          ansible_host: <vm-ip>
          ansible_ssh_common_args: '-o ProxyJump=<parent-ssh-user>@<parent-ip>'
          parent_host: <parent-name>
        ...
```

The `ProxyJump` value is `<parent.SSH.User>@<parent.IP>`. Host and VM names must be
output in sorted order for deterministic output (Go maps iterate randomly).

Write the file with `0644` permissions.

**Required Tests**

*Unit test — `TestGenerateInventory_TwoHostsTwoVMs`:*

Input:
```go
Config{Hosts: map[string]Host{
    "host-01": {IP: "192.168.1.10", SSH: SSHConfig{User: "admin"}, VMs: map[string]VM{
        "host-01-vm-01": {IP: "10.0.100.10"},
    }},
    "host-02": {IP: "192.168.1.11", SSH: SSHConfig{User: "admin"}, VMs: map[string]VM{
        "host-02-vm-01": {IP: "10.0.100.20"},
    }},
}}
```

Expected output contains (in order):
- `host-01:` with `ansible_host: 192.168.1.10`
- `host-02:` with `ansible_host: 192.168.1.11`
- `host-01-vm-01:` with `ansible_host: 10.0.100.10` and
  `ansible_ssh_common_args: '-o ProxyJump=admin@192.168.1.10'` and `parent_host: host-01`
- `host-02-vm-01:` with `ansible_host: 10.0.100.20` and
  `ansible_ssh_common_args: '-o ProxyJump=admin@192.168.1.11'` and `parent_host: host-02`

*Unit test — `TestGenerateInventory_SortedOutput`:*
- Input: config with hosts `"zebra"` and `"alpha"`
- Expected: `"alpha"` appears before `"zebra"` in the output

**Definition of Done**
- Both unit tests pass
- Generated file matches the required structure exactly

---

### Task 7 — Implement `stc inventory generate` command

**Story points:** 2

**Description**

Wire up the `inventory generate` subcommand. This command accepts `--preflight` but does
NOT accept `--host` flags — it always operates on the full config.

**Fields to validate** (for diagnostics):
- Per top-level host: `hosts.<name>.ip`
- Per VM: `hosts.<parent-name>.vms.<vm-name>.ip` and `hosts.<parent-name>.ssh.user`

Names are always present (they are map keys) and are not checked.

**Behavior:**
- Empty `hosts` map → print `error: no hosts defined in stc.yml`, exit 1
- `--preflight` → build diagnostics, print table via `ToTable()`, exit 0, write no files
- Non-preflight, any field missing → print diagnostic table, exit 1
- Non-preflight, all fields present → write `.generated/ansible/inventory/hosts.yml`, print
  a success message, exit 0

**Required Tests**

*Txtar test — `testdata/scripts/inventory-generate.txtar`:*
```
# Happy path: inventory generate writes the correct file
exec stc inventory generate
stdout 'Inventory written'
exists .generated/ansible/inventory/hosts.yml
grep 'ansible_host: 192.168.1.10' .generated/ansible/inventory/hosts.yml
grep 'ansible_host: 10.0.100.10' .generated/ansible/inventory/hosts.yml
grep 'ProxyJump=admin@192.168.1.10' .generated/ansible/inventory/hosts.yml
grep 'parent_host: host-01' .generated/ansible/inventory/hosts.yml

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: true
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms:
      host-01-vm-01:
        ip: 10.0.100.10
        ssh:
          user: admin
          port: 22
          private_key_path: ~/.ssh/id_ed25519
          public_key_path: ~/.ssh/id_ed25519.pub
        auto_update_reboot_time: "05:00"
```

*Txtar test — `testdata/scripts/inventory-generate-zero-hosts.txtar`:*
```
# Zero hosts: exits 1 with friendly error
! exec stc inventory generate
stderr 'no hosts'

-- stc.yml --
hosts: {}
```

*Txtar test — `testdata/scripts/inventory-generate-missing-ip.txtar`:*
```
# Missing ip: prints diagnostic table and exits 1 without writing file
! exec stc inventory generate
stdout 'not found'
! exists .generated/ansible/inventory/hosts.yml

-- stc.yml --
hosts:
  host-01:
    ip: ""
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/inventory-generate-preflight.txtar`:*
```
# --preflight prints diagnostic table, exits 0, writes no files
exec stc inventory generate --preflight
stdout 'CONFIG NAME'
stdout 'STATUS'
stdout 'loaded'
! exists .generated/ansible/inventory/hosts.yml

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

**Definition of Done**
- All four txtar tests pass
- `go test ./...` passes

---

## Milestone 4 — Ansible Execution Infrastructure

Shared machinery used by both `ansible bootstrap-host` and `ansible setup-host`.

---

### Task 8 — Implement host and VM resolution from stc.yml

**Story points:** 2

**Description**

Implement the shared logic that resolves `--host` flag values into concrete records from
`stc.yml`. All `ansible` subcommands and `ssh add` use this.

**VM names are globally unique across the entire `stc.yml`** — they must be, because
`--host host-01-vm-01` would be ambiguous otherwise. The resolver must search both
top-level hosts and all VM maps.

**Functions to implement:**

```go
// ResolveTargetNames returns cliHosts if non-empty, otherwise all top-level host
// names from config (VMs are not included in the default).
func ResolveTargetNames(config *Config, cliHosts []string) []string

// ResolveTargets looks up each name in config. It searches top-level hosts first,
// then all VM maps across all hosts. Returns an error listing any names not found.
func ResolveTargets(config *Config, names ...string) ([]Target, error)
```

```go
type Target struct {
    Name                 string
    IP                   string
    SSH                  SSHConfig
    AutoUpdateRebootTime string
    IsVM                 bool
    ParentName           string // empty for top-level hosts
    ParentIP             string
    ParentSSHUser        string
}
```

**Required Tests**

*Unit test — `TestResolveTargetNames_NoCLIHosts`:*
- Config: two top-level hosts (`"host-01"`, `"host-02"`), each with one VM
- Input `cliHosts`: empty
- Expected: `["host-01", "host-02"]` (VMs not included)

*Unit test — `TestResolveTargetNames_WithCLIHosts`:*
- Config: two hosts with VMs
- Input `cliHosts`: `["host-01"]`
- Expected: `["host-01"]`

*Unit test — `TestResolveTargets_TopLevelHost`:*
- Config: `host-01` with `IP: "192.168.1.10"`
- Input names: `["host-01"]`
- Expected: one `Target` with `Name="host-01"`, `IP="192.168.1.10"`, `IsVM=false`

*Unit test — `TestResolveTargets_VM`:*
- Config: `host-01` containing VM `host-01-vm-01` with `IP: "10.0.100.10"`;
  `host-01` has `IP: "192.168.1.10"`, `SSH.User: "admin"`
- Input names: `["host-01-vm-01"]`
- Expected: one `Target` with `Name="host-01-vm-01"`, `IP="10.0.100.10"`, `IsVM=true`,
  `ParentName="host-01"`, `ParentIP="192.168.1.10"`, `ParentSSHUser="admin"`

*Unit test — `TestResolveTargets_UnknownName`:*
- Input names: `["host-01", "nonexistent", "also-missing"]`
- Expected: error message contains `"nonexistent"` and `"also-missing"` but not `"host-01"`

**Definition of Done**
- All five unit tests pass

---

### Task 9 — Implement sensitive field resolver

**Story points:** 2

**Description**

Implement the shared mechanism for resolving secret values that are never stored in
`stc.yml`. For each secret field, look it up in the environment first; if not set, prompt
interactively.

Use Go reflection (`reflect` standard library) to find struct fields tagged
`sensitive:"true"` and `json:"<key>"`. This avoids a hard-coded field list — any struct
that follows the tagging convention automatically gets the right behavior.

**Env var naming:** `"STC_" + strings.ToUpper(jsonKey)`. For a field tagged
`json:"admin_email"`, the env var is `STC_ADMIN_EMAIL`. Lookup is case-insensitive.

**Functions to implement:**

```go
// PromptSensitiveFields fills all fields tagged sensitive:"true" on the struct
// that v points to. For each field, checks the env var first; if absent, reads
// from r with a prompt written to w.
func PromptSensitiveFields(v any, r io.Reader, w io.Writer) error

// AppendSensitiveDiagnostics appends one Diagnostic per sensitive field of v.
// Field is the env var name (e.g. "STC_ADMIN_EMAIL").
// Status is "loaded" if the env var is set and non-empty, "will prompt" otherwise.
func AppendSensitiveDiagnostics(d *Diagnostics, v any) error
```

The argument `v` must be a pointer to a struct. Return an error if it is not.

**Required Tests**

*Unit test — `TestPromptSensitiveFields_FromEnv`:*

Input struct:
```go
type testConfig struct {
    Email string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
}
```
Environment: `STC_ADMIN_EMAIL=test@example.com`
Input reader: empty (should not be read)
Expected: `Email == "test@example.com"`, no error

*Unit test — `TestPromptSensitiveFields_FromPrompt`:*
- Same struct, no env var set
- Input reader: `strings.NewReader("test@example.com\n")`
- Expected: `Email == "test@example.com"`, no error

*Unit test — `TestPromptSensitiveFields_EmptyInput`:*
- Same struct, no env var set
- Input reader: `strings.NewReader("\n")` (just a newline)
- Expected: non-nil error

*Unit test — `TestAppendSensitiveDiagnostics_Mixed`:*
- Same struct
- Environment: `STC_ADMIN_EMAIL` is not set
- Expected: one `Diagnostic` with `Field="STC_ADMIN_EMAIL"`, `Status="will prompt"`

*Unit test — `TestAppendSensitiveDiagnostics_Loaded`:*
- Same struct
- Environment: `STC_ADMIN_EMAIL=x`
- Expected: one `Diagnostic` with `Field="STC_ADMIN_EMAIL"`, `Status="loaded"`

**Definition of Done**
- All five unit tests pass

---

### Task 10 — Implement ansible-playbook executor

**Story points:** 2

**Description**

Implement the shared logic for running an Ansible playbook against a set of resolved
targets. Both `ansible bootstrap-host` and `ansible setup-host` use this.

**Steps:**
1. For each target, write non-sensitive vars to
   `.generated/ansible/inventory/host_vars/<name>/vars.yml` using `github.com/goccy/go-yaml`
   to marshal a `map[string]any`.
2. If there are sensitive vars, write them to a temp file via
   `os.CreateTemp("", "stc-vars-*.yml")`. Use `defer os.Remove(tmpFile.Name())` so the
   file is always cleaned up, even if ansible fails.
3. Build the `ansible-playbook` invocation:
   ```
   ansible-playbook \
     -i .generated/ansible/inventory/hosts.yml \
     --limit <comma-separated target names> \
     [-e @<tmpfile>]   # only when sensitive vars are present \
     ansible/playbooks/<playbook>.yml
   ```
4. Stream stdout and stderr from the subprocess directly to the terminal.

**Required Tests**

*Unit test — `TestBuildPlaybookArgs_WithSensitiveVars`:*
- Input: playbook `"bootstrap-host"`, hosts `["host-01", "host-02"]`,
  extraVarsFile `"/tmp/stc-vars-123.yml"`
- Expected args slice:
  `["-i", ".generated/ansible/inventory/hosts.yml", "--limit", "host-01,host-02",
  "-e", "@/tmp/stc-vars-123.yml", "ansible/playbooks/bootstrap-host.yml"]`

*Unit test — `TestBuildPlaybookArgs_NoSensitiveVars`:*
- Input: same but `extraVarsFile` is `""`
- Expected: `-e` and the file path are absent from the args slice

*Unit test — `TestCreateTempVarsFile_CreatesAndCleansUp`:*
- Input: `map[string]any{"admin_email": "x@example.com", "admin_password": "secret"}`
- Expected: temp file exists after creation, contains both keys as YAML, file is deleted
  after the cleanup function returned by the constructor is called

**Definition of Done**
- All three unit tests pass
- `host_vars` files are written before `ansible-playbook` is invoked
- Temp file does not persist after the executor returns (verify in a real run with
  `ls /tmp/stc-vars-*`)

---

## Milestone 5 — `stc ansible bootstrap-host` and `stc ansible setup-host`

---

### Task 11 — Implement `stc ansible bootstrap-host` command

**Story points:** 3

**Description**

Wire up the `ansible bootstrap-host` command. It runs `ansible/playbooks/bootstrap-host.yml`
against a set of hosts after validating config and collecting secrets.

**Flags:** 0 or more `--host <h>` (defaults to all top-level hosts), optional `--preflight`.

**Config fields to validate per host** (for diagnostics and pre-run checks):

| CONFIG NAME | Description |
|---|---|
| `hosts.<name>.ip` | Host LAN IP |
| `hosts.<name>.ssh.user` | SSH username |
| `hosts.<name>.ssh.port` | SSH port |
| `hosts.<name>.ssh.public_key_path` | Path to public key (read at runtime to get contents) |
| `hosts.<name>.auto_update_reboot_time` | HH:MM reboot time for unattended updates |

**Secrets** (looked up from env / prompted; never from stc.yml):

| Env var | Ansible var name | Description |
|---|---|---|
| `STC_ADMIN_EMAIL` | `admin_email` | Email for admin notifications |
| `STC_ADMIN_PASSWORD` | `admin_password` | Admin account password |

**`host_vars/<name>/vars.yml` must contain** (non-sensitive only):
- `ansible_host`: host IP
- `ansible_port`: SSH port
- `ansible_user`: determined at runtime (attempt configured SSH user; fall back to `"root"`)
- `ssh_port`: SSH port (used by the playbook to configure sshd)
- `ssh_public_key`: full file contents of `ssh.public_key_path`
- `auto_update_reboot_time`: from stc.yml

Secrets go in the temp vars file, not in `host_vars`.

**Behavior:**
- `--preflight` → print diagnostic table (per-host fields + both secrets), exit 0
- Any per-host field missing (non-preflight) → print diagnostic table, exit 1
- All fields present → prompt for unset secrets, run playbook
- Unknown `--host` value → print error listing the unknown names, exit 1

**Required Tests**

*Txtar test — `testdata/scripts/bootstrap-host-preflight.txtar`:*
```
# --preflight prints diagnostic table with all required fields and exits 0
exec stc ansible bootstrap-host --preflight
stdout 'CONFIG NAME'
stdout 'STATUS'
stdout 'hosts.host-01.ip'
stdout 'hosts.host-01.ssh.user'
stdout 'hosts.host-01.ssh.port'
stdout 'hosts.host-01.ssh.public_key_path'
stdout 'hosts.host-01.auto_update_reboot_time'
stdout 'STC_ADMIN_EMAIL'
stdout 'STC_ADMIN_PASSWORD'
stdout 'will prompt'
! exists .generated/ansible/inventory/host_vars/host-01/vars.yml

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: true
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/bootstrap-host-preflight-env-loaded.txtar`:*
```
# When STC_ADMIN_EMAIL is set in env, it shows as "loaded" in the table
env STC_ADMIN_EMAIL=admin@example.com
exec stc ansible bootstrap-host --preflight
stdout 'STC_ADMIN_EMAIL'
stdout 'loaded'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/bootstrap-host-unknown-host.txtar`:*
```
# Unknown --host value exits 1 with a clear error
! exec stc ansible bootstrap-host --host nonexistent
stderr 'nonexistent'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

**Definition of Done**
- All three txtar tests pass
- `go test ./...` passes

---

### Task 12 — Implement `stc ansible setup-host` command

**Story points:** 2

**Description**

Wire up the `ansible setup-host` command. It is structurally identical to `bootstrap-host`
(same flags, same flow) but uses a different playbook, fewer per-host config requirements,
and different secrets.

**Flags:** 0 or more `--host <h>` (defaults to all top-level hosts), optional `--preflight`.

**Config fields to validate per host:**

| CONFIG NAME | Description |
|---|---|
| `hosts.<name>.ip` | Host LAN IP |
| `hosts.<name>.ssh.user` | SSH username |
| `hosts.<name>.ssh.port` | SSH port |

**Secrets:**

| Env var | Ansible var name | Description |
|---|---|---|
| `STC_SMTP_USER` | `smtp_user` | SMTP username for Postfix relay |
| `STC_SMTP_PASSWORD` | `smtp_password` | SMTP password |

**`host_vars/<name>/vars.yml` must contain:**
- `ansible_host`, `ansible_port`, `ansible_user` (same as `bootstrap-host`)

Secrets go in the temp vars file.

**Playbook:** `ansible/playbooks/setup-host.yml`

**Required Tests**

*Txtar test — `testdata/scripts/setup-host-preflight.txtar`:*
```
# --preflight prints the correct fields for setup-host
exec stc ansible setup-host --preflight
stdout 'hosts.host-01.ip'
stdout 'hosts.host-01.ssh.user'
stdout 'hosts.host-01.ssh.port'
stdout 'STC_SMTP_USER'
stdout 'STC_SMTP_PASSWORD'
! stdout 'auto_update_reboot_time'
! stdout 'STC_ADMIN_EMAIL'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/setup-host-smtp-env-loaded.txtar`:*
```
# STC_SMTP_USER set in env shows as loaded
env STC_SMTP_USER=smtp@example.com
exec stc ansible setup-host --preflight
stdout 'STC_SMTP_USER'
stdout 'loaded'
stdout 'STC_SMTP_PASSWORD'
stdout 'will prompt'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

**Definition of Done**
- Both txtar tests pass
- `go test ./...` passes

---

## Milestone 6 — `stc ssh add`

---

### Task 13 — Implement SSH config file writer

**Story points:** 2

**Description**

Implement the logic for appending `Host` blocks to `~/.ssh/config`.

**Use `text/template` to render each `Host` block.** Define two templates — one for
top-level hosts and one for VMs — so that the output format is immediately legible from
reading the template source.

Top-level host template:
```
Host {{.Name}}
  HostName {{.IP}}
  User {{.User}}
  IdentityFile {{.PublicKeyPath}}
  Port {{.Port}}
```

VM template (same as above plus):
```
  ProxyJump {{.ParentSSHUser}}@{{.ParentIP}}
```

**Implementation requirements:**
- Open `~/.ssh/config` with `O_RDWR|O_CREATE` and `0600` permissions
- Parse the existing file to check if a `Host <name>` alias already exists. Use
  `github.com/kevinburke/ssh_config` for parsing.
- If the alias already exists: skip, print `<name>: already present in ~/.ssh/config, skipping`
- If not: seek to end of file and append the rendered block
- Process all targets; do not stop on a skip

**Required Tests**

*Txtar test — `testdata/scripts/ssh-add-host.txtar`:*
```
# ssh add appends a Host block for a top-level host
exec stc ssh add --host host-01
stdout 'Added "host-01"'
grep 'Host host-01' $HOME/.ssh/config
grep 'HostName 192.168.1.10' $HOME/.ssh/config
grep 'User admin' $HOME/.ssh/config
grep 'Port 22' $HOME/.ssh/config
! grep 'ProxyJump' $HOME/.ssh/config

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/ssh-add-idempotent.txtar`:*
```
# Running ssh add twice for the same host: second run skips gracefully
exec stc ssh add --host host-01
stdout 'Added'

exec stc ssh add --host host-01
stdout 'skipping'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/ssh-add-vm.txtar`:*
```
# ssh add for a VM includes a ProxyJump directive
exec stc ssh add --host host-01-vm-01
stdout 'Added "host-01-vm-01"'
grep 'Host host-01-vm-01' $HOME/.ssh/config
grep 'HostName 10.0.100.10' $HOME/.ssh/config
grep 'ProxyJump admin@192.168.1.10' $HOME/.ssh/config

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms:
      host-01-vm-01:
        ip: 10.0.100.10
        ssh:
          user: admin
          port: 22
          private_key_path: ~/.ssh/id_ed25519
          public_key_path: ~/.ssh/id_ed25519.pub
        auto_update_reboot_time: "05:00"
```

**Definition of Done**
- All three txtar tests pass
- `go test ./...` passes

---

### Task 14 — Implement `stc ssh add` command

**Story points:** 2

**Description**

Wire up the `ssh add` command using the SSH config writer from Task 13 and the host
resolver from Task 8.

**Flags:** 0 or more `--host <h>` (defaults to all top-level hosts; VMs not included in
default), optional `--preflight`.

**Required config fields per top-level host** (for diagnostics):
- `hosts.<name>.ip`
- `hosts.<name>.ssh.user`
- `hosts.<name>.ssh.public_key_path`
- `hosts.<name>.ssh.port`

**Required config fields per VM** (in addition to the VM's own 4 fields):
- `hosts.<parent-name>.ip` (for ProxyJump)
- `hosts.<parent-name>.ssh.user` (for ProxyJump)

No secret variables. No interactive prompting.

**Behavior:**
- `--preflight` → print diagnostic table, exit 0, do NOT touch `~/.ssh/config`
- Any required field missing (non-preflight) → print diagnostic table, exit 1
- All fields present → write entries (skipping any already present), exit 0
- Unknown `--host` → print error listing unknown names, exit 1

**Required Tests**

*Txtar test — `testdata/scripts/ssh-add-preflight.txtar`:*
```
# --preflight prints required fields and exits without modifying ~/.ssh/config
exec stc ssh add --preflight
stdout 'CONFIG NAME'
stdout 'hosts.host-01.ip'
stdout 'hosts.host-01.ssh.user'
stdout 'hosts.host-01.ssh.public_key_path'
stdout 'hosts.host-01.ssh.port'
! exists $HOME/.ssh/config

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/ssh-add-missing-field.txtar`:*
```
# Missing public_key_path: diagnostic table shown, exits 1
! exec stc ssh add
stdout 'not found'
stdout 'ssh.public_key_path'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ""
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

**Definition of Done**
- All txtar tests pass (including the three from Task 13)
- `go test ./...` passes

---

## Milestone 7 — `stc setup`

---

### Task 15 — Implement `stc setup` command with merged preflight

**Story points:** 3

**Description**

Wire up the `setup` command, which orchestrates full provisioning end-to-end. This is the
primary command most users run.

**Flags:** 0 or more `--host <h>` (defaults to all top-level hosts), optional `--preflight`.

**Execution order:**
1. `inventory generate`
2. `ansible setup-host`
3. `ansible bootstrap-host`
4. `ssh add`

**Preflight behavior (key complexity):**
When `--preflight` is passed, collect diagnostics from all four sub-operations and merge
them into a **single** table. Print one combined table and exit 0. Do NOT execute any side
effects (no files written, no subprocesses launched).

**Upfront secret collection (second key complexity):**
When NOT in preflight mode, collect ALL secrets before starting execution. Prompt for
`STC_SMTP_USER`, `STC_SMTP_PASSWORD`, `STC_ADMIN_EMAIL`, `STC_ADMIN_PASSWORD` at the
start — the user should not be interrupted once the playbooks are running.

**Required Tests**

*Txtar test — `testdata/scripts/setup-preflight-merged-table.txtar`:*
```
# setup --preflight shows one merged table covering all four commands
exec stc setup --preflight
stdout 'CONFIG NAME'

# inventory generate fields
stdout 'hosts.host-01.ip'

# bootstrap-host fields
stdout 'hosts.host-01.ssh.public_key_path'
stdout 'hosts.host-01.auto_update_reboot_time'
stdout 'STC_ADMIN_EMAIL'
stdout 'STC_ADMIN_PASSWORD'

# setup-host secrets
stdout 'STC_SMTP_USER'
stdout 'STC_SMTP_PASSWORD'

# ssh add fields
stdout 'hosts.host-01.ssh.user'

# No files written
! exists .generated/ansible/inventory/hosts.yml
! exists $HOME/.ssh/config

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: true
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/setup-preflight-no-duplicate-rows.txtar`:*
```
# Fields shared between multiple subcommands (e.g. hosts.host-01.ip) appear only once
exec stc setup --preflight
stdout 'hosts.host-01.ip'

-- stc.yml --
hosts:
  host-01:
    ip: 192.168.1.10
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

*Txtar test — `testdata/scripts/setup-missing-field-exits-before-execution.txtar`:*
```
# A missing required field causes setup to exit before generating any files
! exec stc setup
stdout 'not found'
! exists .generated/ansible/inventory/hosts.yml

-- stc.yml --
hosts:
  host-01:
    ip: ""
    ssh:
      user: admin
      port: 22
      private_key_path: ~/.ssh/id_ed25519
      public_key_path: ~/.ssh/id_ed25519.pub
    auto_update_reboot_time: "05:00"
    wireguard_endpoint: false
    incus:
      storage_pool_name: default
      storage_pool_driver: dir
    vms: {}
```

**Definition of Done**
- All three txtar tests pass
- `./stc setup --preflight` shows a single merged table (not four separate tables)
- No prompts appear mid-execution when running `./stc setup`
- `go test ./...` passes
