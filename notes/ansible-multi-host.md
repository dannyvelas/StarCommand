l# Multi-Host Ansible Support

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

### Why the SSH check is lazy

The SSH check happens at playbook-run time, not during inventory generation. This is by
design: host SSH state changes dynamically. Before `bootstrap-server` runs, only root
can SSH. After it runs, only admin can. Running the check eagerly (e.g. at `stc
inventory generate` time) would bake in a stale result.

### What is currently broken
  
when `stc ansible setup-host` runs, the code should be smart enough to run the `../ansible/playbooks/setup-host.yml` playbook for all hosts in *config.Config. same for other playbooks like `../ansible/playbooks/bootstrap-server.yml` and `../ansible/playbooks/setup-vm.yml`.
* suppose there are two hosts in *config.Config, "h" and "k" where "h" needs the playbook to run as "root" and "k" needs the playbook to run as "admin." in this case, the code should execute the playbook as root for host "h" and as admin for host "k".
* regardless of required "root"/"admin" permissions for playbooks for the hosts in *config.Config, the playbook should execute in parallel for all of the hosts in *config.Config. I expect this should be possible because ansible's default functionality is to support running the same playbook on multiple target hosts concurrently.
* suppose that there are two hosts "h" and "k" in *config.Config. suppose both of these hosts have different config values for ip, ssh_port, ssh_private_key_path, ssh_user, etc. our code should make sure that when playbook "p" runs, the configs for host "h" will be used for "p" when "p" is running on "h", and the configs for host "k" will be used for "p" when "p" is running on host "k".

you might wonder why why need `ip`, `ssh_port`, `ssh_private_key_path`, etc for each host. this is because before running a playbook "p" we want to do an ssh check to determine whether we should run "p" as admin or root for a given host

right now it's implementing it wrong. right now, the code only targets one host. it just picks the first host as the target which was arbitrary and wrong.

### How i think it should be fixed

The standard Ansible way to assign specific values for specific hosts is to create a directory named `host_vars`. Inside that directory, you create a file named after each host. Ansible will automatically detect these files and apply the variables to the correct host during execution.

I think we should change the code so that when `stc ansible setup-host` runs, we generate this `host_vars` logic and put it in a specific top-level directory (which is gitignored) in this repo, `.generated`.

that way, when someone runs `stc ansible setup-host`, they can see the vars that were used by peering into the `.generated` directory.

We should probably refactor the ansible config structs so that they can hold the information for more than just one host. they should be able to hold the information for multiple hosts. that way, when that config struct is passed to `ansibleHandler.execute` (e.g. `ansibleHandler.execute(playbookConfig)` on line 48 of `../internal/app/app.go`), `ansibleHandler.execute` can read the config struct and generate the `host_vars` files in the `.generated` directory for each host.
