---
title: 'stc cli spec'
date: '2026-03-13'
publish: false
---

## Spec

`stc` is a CLI that:
- Allows you to have one source-of-truth for all your infrastructure configuration, `stc.yml`, so that you don't have to manually copy-paste values between Ansible and Terraform config files.
- Abstracts Ansible and Terraform steps, so that you can run a single command to provision your entire infrastructure without having to worry about the order of operations or the specific Ansible/Terraform commands to run for each tool

## The stc.yml file

As mentioned above, the `stc` CLI treats `stc.yml` as the source-of-truth of all the information that will be provided to terraform and ansible.

### stc.yml Schema Description

The top-level document contains a single key, `hosts`, whose value is a list of one or more **host** objects. Each `host` represents a physical machine and has the following fields:
- `name` *(string, required)*: A human-readable label for the host (e.g. `host-01`).
- `ip` *(string, required)*: The LAN IP address of the physical host (e.g. `192.168.1.10`).
- `ssh` *(object, required)*: SSH connection details for the host. Contains four sub-fields:
  - `user` — the SSH username
  - `port` — the SSH port number (typically `22`)
  - `private_key_path` — path to the private key file (e.g. `~/.ssh/id_ed25519`)
  - `public_key_path` — path to the corresponding public key file (e.g. `~/.ssh/id_ed25519.pub`)
- `auto_update_reboot_time` *(string, required)*: The time at which the host should reboot after unattended security upgrades, expressed as a quoted 24-hour `HH:MM` string (e.g. `"05:00"`).
- `wireguard_endpoint` *(boolean, optional)*: Whether this host is the WireGuard VPN endpoint that receives inbound VPN traffic. Exactly one host across the entire `hosts` list must have this set to `true`; all others must be `false`. If omitted, this defaults to `false`.
- `incus` *(object, required)*: Configuration for the Incus hypervisor running on this host. Contains two sub-fields:
  - `storage_pool_name` — an arbitrary name for the Incus storage pool (e.g. `default`)
  - `storage_pool_driver` — the storage backend driver; must be one of `dir`, `btrfs`, `zfs`, or `lvm`
- `vms` *(list, required)*: A list of one or more **VM** objects representing virtual machines running on this host. Each entry in a host's `vms` list has the following fields:
  - `name` *(string, required)*: A label for the VM. This value is also used as the VM's hostname.
  - `ip` *(string, required)*: The IP address assigned to this VM on the OVN overlay network (e.g. `10.0.100.10`). These are distinct from host LAN IPs and exist in a separate address space.
  - `ssh` *(object, required)*: SSH connection details for the VM, with the same four sub-fields as the host's `ssh` block: `user`, `port`, `private_key_path`, and `public_key_path`.
  - `auto_update_reboot_time` *(string, required)*: Same semantics as the host-level field — the quoted 24-hour `HH:MM` time at which the VM reboots after unattended security upgrades.

### stc.yml Example

```yaml
hosts:
  - name: host-01
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
      - name: host-01-vm-01
        ip: 10.0.100.10
        ssh:
          user: admin
          port: 22
          private_key_path: ~/.ssh/id_ed25519
          public_key_path: ~/.ssh/id_ed25519.pub
        auto_update_reboot_time: "05:00"

  - name: host-02
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

    vms:
      - name: host-02-vm-01
        ip: 10.0.100.20
        ssh:
          user: admin
          port: 22
          private_key_path: ~/.ssh/id_ed25519
          public_key_path: ~/.ssh/id_ed25519.pub
        auto_update_reboot_time: "05:00"
```

## Overview of commands that `stc` should support:

```
stc <command> [options]

Commands:
  inventory generate                       Generate the Ansible inventory file for all hosts
  ansible bootstrap-host [--host <h>]...   Run the bootstrap-host playbook against all hosts/VMs, or limit to the ones given
  ansible setup-host [--host <h>]...       Run the setup-host playbook against all hosts, or limit to the ones given
  ssh add <host>                           Add a host to ~/.ssh/config
  terraform apply                          Apply the Terraform project

<h> is the host name as defined in stc.yml.
```

## `stc` has an `inventory generate` command

- `stc` has an `inventory generate` command which does NOT have any expect any flags.
- This command will read information in the `stc.yml` file and use it to generate an ansible inventory file in yaml format called `.generated/ansible/inventory/hosts.yml`.
- If the `stc.yml` file looks like the example above, running `stc inventory generate` should create the following content in `.generated/ansible/inventory/hosts.yml`:
  ```yaml
  all:
    children:
      hosts:
        hosts:
          host-01:
            ansible_host: 192.168.1.10
          host-02:
            ansible_host: 192.168.1.11
      vms:
        hosts:
          host-01-vm-01:
            ansible_host: 10.0.100.10
            ansible_ssh_common_args: '-o ProxyJump=admin@192.168.1.10'
            parent_host: host-01
          host-02-vm-01:
            ansible_host: 10.0.100.20
            ansible_ssh_common_args: '-o ProxyJump=admin@192.168.1.11'
            parent_host: host-02
    ```
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## `stc` has `ansible bootstrap-host` command

- `stc` has an `ansible bootstrap-host` command which takes 0 or more `--host <h>` arguments where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`.
- It also takes an optional `--preflight` command.
- This command operates on a collection of hosts. We'll denote this collection as `hosts`. `hosts` is set by the `--host` arguments that are passed in. If 0 `--host <h>` arguments are passed, `stc` will use every single top-level host name in `stc.yml` as `hosts` (VMs are not included in the default).
- Every ansible playbook requires the following 4 fields (at minimum) to be present in every host in `hosts`: `.name`, `.ip`, `.ssh.user`, and `.ssh.port`. Let's call these the "base configs".
- The `./ansible/playbooks/bootstrap-host.yml` playbook specifically additionally needs these 3 variables to be set for each host in `hosts`: `ssh_port`, `ssh_public_key`, and `auto_update_reboot_time`. Let's call these the "bootstrap configs".
- The `./ansible/playbooks/bootstrap-host.yml` playbook specifically also needs 2 secret variables to be set for each host in `hosts`: `admin_email` and `admin_password`. These values will NOT and will never be present in `stc.yml`.  Let's call these "bootstrap secrets".
- In other words, the `bootstrap-host.yml` playbook requires a total of 9 config values.
- If the `--preflight` flag is passed, `stc` will print a diagnostic table of two columns. There will be one row in this table for each host, for each config value. in other words, if there are `n` hosts, there will be `7n + 2` rows. `7n` because there are `7` config values per host. `+2` because of the 2 bootstrap secrets.
  - The header of the first column will be called `CONFIG NAME`. The first column will be the fully qualified path of the config in `stc.yml`. The header of the second column will `STATUS`. It will be `loaded` if that config value was found or `not found` if that config value was not found. The "bootstrap secret" rows will behave a little bit differently. These don't have a path in `stc.yml` since they will never be in `stc.yml`. Instead of their `CONFIG NAME` column having a fully qualified yaml path, they will have the name of an environmental variable. For `admin_email` the corresponding environmental variable is `STC_ADMIN_EMAIL`. For `admin_password` the corresponding is `STC_ADMIN_PASSWORD`. Since these values will never be in `stc.yml`, they would always come up as `not found`. This wouldn't make much sense. Instead, the code will look for these values in the environment. For `admin_email` it will look for `STC_ADMIN_EMAIL` in the environment. If it finds an entry with a value, it will set the `STATUS` cell for the corresponding row to `loaded`. If it does not find it, it will set the `STATUS` cell for the corresponding row to `will prompt`. After printing, `stc` will exit with a status code of 0. it is done. The sole job of `--preflight` is to print a diagnostic table. An example diagnostic table would look something like this (note, this is not based on the example `stc.yml` above, this is just a random example):
  ```
  ---------------------------------------------
  | CONFIG NAME                 | STATUS      |
  ---------------------------------------------
  | host-01.name                | loaded      |
  | host-01.ip                  | loaded      |
  | host-01.ssh.user            | loaded      |
  | host-01.ssh.port            | loaded      |
  | host-01.ssh.public_key_path | loaded      |
  | host-02.name                | loaded      |
  | host-02.ip                  | loaded      |
  | host-02.ssh.user            | loaded      |
  | host-02.ssh.port            | not found   |
  | host-02.ssh.public_key_path | loaded      |
  | STC_ADMIN_EMAIL             | will prompt |
  | STC_ADMIN_PASSWORD          | loaded      |
  ---------------------------------------------
  ```
- If the `--preflight` command is NOT passed, `stc` will check if any of the "base configs" or "bootstrap configs" are missing. if any are missing, it will print the same exact diagnostic table, which should serve as a user-friendly indication of the required configs that are missing. `stc` should exit with an exit code of 1.
- Otherwise, if we're here, it means `stc` found all 7 "base configs" and "bootstrap configs"
- Next, for any of [`STC_ADMIN_EMAIL`, `STC_ADMIN_PASSWORD`] that are not set in the environment, `stc` will prompt the user to enter the value through stdin in a user-friendly way. In other words:
  - if `STC_ADMIN_EMAIL` is not set in the environment, `stc` will ask the user to enter the value for `admin_email` in a user-friendly way. it will associate the value that the user enters as the admin email, after trimming whitespace. if it IS set in the environment and its, it won't prompt the user. It will associate the value found in the environment as the admin email.
  - if `STC_ADMIN_PASSWORD` is not set in the environment, `stc` will ask the user to enter the value for `admin_password` in a user-friendly way. it will associate the value that the user enters as the admin password, after trimming whitespace. if it IS set in the environment, it won't prompt the user. It will associate the value found in the environment as the admin password.
- Finally, this command will use all 9 values that were found in some combination of stc.yml / prompting / the environment to pass as ansible configs.
- For a given host `my_awesome_host` in `hosts`, `stc` should communicate the "base configs" and "bootstrap configs" to ansible by creating the following file: `.generated/ansible/inventory/host_vars/my_awesome_host/vars.yml`, and putting all the "base configs" and "bootstrap configs" inside of that file. `stc` should NOT put the "bootstrap secrets" there. instead, `stc` should create a temporary file, and put the bootstrap secrets in there,
- Finally, `stc` should run `./ansible/playbooks/bootstrap-host.yml` and pass that temporary file as an argument to the ansible playbook.

### edge cases
- If one or more `--host` arguments are provided for names that do not match any top-level host or VM in `stc.yml`, this command should print a user friendly error indicating all of the names that were passed as arguments but were not found in `stc.yml`.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## stc has an `ansible setup-host` command

- `stc` has an `ansible setup-host` command which takes 0 or more `--host <h>` arguments where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`.
- It also takes an optional `--preflight` flag.
- This command operates on a collection of hosts. We'll denote this collection as `hosts`. `hosts` is set by the `--host` arguments that are passed in. If 0 `--host <h>` arguments are passed, `stc` will use every single top-level host name in `stc.yml` as `hosts` (VMs are not included in the default).
- Every ansible playbook requires the following 4 fields (at minimum) to be present in every host in `hosts`: `.name`, `.ip`, `.ssh.user`, and `.ssh.port`. Let's call these the "base configs".
- The `./ansible/playbooks/setup-host.yml` playbook does not require any additional non-secret variables beyond the "base configs".
- The `./ansible/playbooks/setup-host.yml` playbook additionally needs 2 secret variables: `smtp_user` and `smtp_password`. These values will NOT and will never be present in `stc.yml`. Let's call these the "setup secrets".
- In other words, the `setup-host.yml` playbook requires a total of `4n + 2` config values, where `n` is the number of hosts.
- If the `--preflight` flag is passed, `stc` will print a diagnostic table of two columns. There will be one row per host per base config, plus one row per setup secret. The header of the first column will be called `CONFIG NAME`. The first column will be the fully qualified path of the config in `stc.yml`. The header of the second column will be `STATUS`. It will be `loaded` if that config value was found, or `not found` if that config value was not found. The "setup secret" rows behave differently: since these values will never be in `stc.yml`, their `CONFIG NAME` will be the name of an environment variable instead of a yaml path. For `smtp_user` the corresponding environment variable is `STC_SMTP_USER`. For `smtp_password` the corresponding is `STC_SMTP_PASSWORD`. The code will look for these values in the environment. If found and non-empty, `STATUS` will be `loaded`. If not found, `STATUS` will be `will prompt`. After printing, `stc` will exit with a status code of 0. An example diagnostic table would look something like this (note, this is not based on the example `stc.yml` above, this is just a random example):
  ```
  ---------------------------------------------
  | CONFIG NAME                 | STATUS      |
  ---------------------------------------------
  | host-01.name                | loaded      |
  | host-01.ip                  | loaded      |
  | host-01.ssh.user            | loaded      |
  | host-01.ssh.port            | not found   |
  | STC_SMTP_USER               | will prompt |
  | STC_SMTP_PASSWORD           | loaded      |
  ---------------------------------------------
  ```
- If the `--preflight` flag is NOT passed, `stc` will check if any of the "base configs" are missing. If any are missing, it will print the same exact diagnostic table, which should serve as a user-friendly indication of the required configs that are missing. `stc` should exit with an exit code of 1.
- Otherwise, if we're here, it means `stc` found all 4 "base configs" per host.
- Next, for any of [`STC_SMTP_USER`, `STC_SMTP_PASSWORD`] that are not set in the environment, `stc` will prompt the user to enter the value through stdin in a user-friendly way. In other words:
  - if `STC_SMTP_USER` is not set in the environment, `stc` will ask the user to enter the value for `smtp_user` in a user-friendly way. It will associate the value that the user enters as the SMTP username, after trimming whitespace. If it IS set in the environment, it won't prompt the user.
  - if `STC_SMTP_PASSWORD` is not set in the environment, `stc` will ask the user to enter the value for `smtp_password` in a user-friendly way. It will associate the value that the user enters as the SMTP password, after trimming whitespace. If it IS set in the environment, it won't prompt the user.
- Finally, this command will use all values found in some combination of `stc.yml` / prompting / the environment to pass as ansible configs. The "base configs" are communicated to ansible by creating `.generated/ansible/inventory/host_vars/<host-name>/vars.yml` for each host in `hosts`. The "setup secrets" are NOT written there. Instead, `stc` creates a temporary file containing only the setup secrets, and passes it to the ansible playbook as an extra vars file.
- Finally, `stc` should run `./ansible/playbooks/setup-host.yml`.

### edge cases
- If one or more `--host` arguments are provided for names that do not match any top-level host or VM in `stc.yml`, this command should print a user friendly error indicating all of the names that were passed as arguments but were not found in `stc.yml`.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.


## stc has an `ssh add` command

- `stc` has an `ssh add` command which takes exactly 1 positional argument `<host>`, where `<host>` is either a top-level host name or a VM name as defined in `stc.yml`.
- This command reads connection details for the given host or VM from `stc.yml` and appends a new `Host` block to `~/.ssh/config` on the local workstation. If `~/.ssh/config` does not exist, it is created with `0600` permissions.
- The required config values for a top-level host are: `.name`, `.ip`, `.ssh.user`, `.ssh.public_key_path`, and `.ssh.port`. The resulting block looks like:
  ```
  Host <name>
    HostName <ip>
    User <ssh.user>
    IdentityFile <ssh.public_key_path>
    Port <ssh.port>
  ```
- For a VM, the same 5 fields are required from the VM's own config, plus the parent host's `.ip` and `.ssh.user` in order to construct a `ProxyJump` directive. The resulting block looks like:
  ```
  Host <vm.name>
    HostName <vm.ip>
    User <vm.ssh.user>
    IdentityFile <vm.ssh.public_key_path>
    Port <vm.ssh.port>
    ProxyJump <parent.ssh.user>@<parent.ip>
  ```
- If `~/.ssh/config` already contains an entry whose `Host` alias matches the given name, `stc` will skip the write and print a user-friendly message indicating the entry already exists. It will exit with code 0.
- This command has no secret variables and does not prompt for any input.

### edge cases
- If the given `<host>` argument does not match any top-level host or VM in `stc.yml`, this command should print a user friendly error and exit with exit code 1.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## Constitution

1. The code should promote a world-class user experience.
2. The code should be very well-structured for the problem at hand. Think carefully about the right code patterns that would be perfect for this problem. The code shouldn't be messy or have duplication. Duplication is bad because a developer might update one part of the code and forget to update another part of the code. In forgetting to update the other part of the code a bug could get introduced. The could should be written in a way that will allow the code to grow and scale if many more `ansible` subcommands are added.
