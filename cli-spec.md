---
title: 'stc cli spec'
date: '2026-03-13'
publish: false
---

## Spec

`stc` is a CLI that:
- Allows you to have one source-of-truth for all your infrastructure configuration, `stc.yml`, so that you don't have to manually create Ansible config files.
- Abstracts Ansible steps, so that you can run simple and intuitive `stc` commands instead of running ansible commands. With ansible commands, you'd have to worry about complicated ansible command syntax and flags like `--limit` or `-e`.

## The stc.yml file

As mentioned above, the `stc` CLI treats `stc.yml` as the source-of-truth of all the information that will be provided to ansible playbooks.

### stc.yml Schema Description

The top-level document contains a single key, `hosts`, whose value is a map of one or more **host** objects. Each key in this map is the host's name (e.g. `host-01`), and each value is a host object. A host object has the following fields:
- `ip` *(string, required)*: The LAN IP address of the physical host (e.g. `192.168.1.10`).
- `ssh` *(object, required)*: SSH connection details for the host. Contains four sub-fields:
  - `user` — the SSH username
  - `port` — the SSH port number (typically `22`)
  - `private_key_path` — path to the private key file (e.g. `~/.ssh/id_ed25519`)
  - `public_key_path` — path to the corresponding public key file (e.g. `~/.ssh/id_ed25519.pub`)
- `auto_update_reboot_time` *(string, required)*: The time at which the host should reboot after unattended security upgrades, expressed as a quoted 24-hour `HH:MM` string (e.g. `"05:00"`).
- `wireguard_endpoint` *(boolean, optional)*: Whether this host is the WireGuard VPN endpoint that receives inbound VPN traffic. Exactly one host across the entire `hosts` map must have this set to `true`; all others must be `false`. If omitted, this defaults to `false`.
- `incus` *(object, required)*: Configuration for the Incus hypervisor running on this host. Contains two sub-fields:
  - `storage_pool_name` — an arbitrary name for the Incus storage pool (e.g. `default`)
  - `storage_pool_driver` — the storage backend driver; must be one of `dir`, `btrfs`, `zfs`, or `lvm`
- `vms` *(map, required)*: A map of one or more **VM** objects representing virtual machines running on this host. Each key is the VM's name (also used as the VM's hostname), and each value is a VM object with the following fields:
  - `ip` *(string, required)*: The IP address assigned to this VM on the OVN overlay network (e.g. `10.0.100.10`). These are distinct from host LAN IPs and exist in a separate address space.
  - `ssh` *(object, required)*: SSH connection details for the VM, with the same four sub-fields as the host's `ssh` block: `user`, `port`, `private_key_path`, and `public_key_path`.
  - `auto_update_reboot_time` *(string, required)*: Same semantics as the host-level field — the quoted 24-hour `HH:MM` time at which the VM reboots after unattended security upgrades.

### stc.yml Example

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

    vms:
      host-02-vm-01:
        ip: 10.0.100.20
        ssh:
          user: admin
          port: 22
          private_key_path: ~/.ssh/id_ed25519
          public_key_path: ~/.ssh/id_ed25519.pub
        auto_update_reboot_time: "05:00"
```

## `stc` has an `inventory generate` command

- `stc` has an `inventory generate` command which takes an optional `--preflight` flag but does not accept `--host` flags. It always generates an inventory for every top-level host and every VM defined in `stc.yml`. There is no way to limit it to a subset.
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
- The required config values for each top-level host are `hosts.<name>.ip`. The required config values for each VM are `hosts.<name>.vms.<vm-name>.ip`, plus the parent host's `hosts.<name>.ssh.user` (needed to construct the `ProxyJump` string). Host and VM names, being map keys, are always present by definition and are not checked.
- If the `--preflight` flag is passed, `stc` will print a diagnostic table showing the status of each required config value for every host and VM, then exit with code 0 without writing any files. There are no secret variables for this command, so all rows will show either `loaded` or `not found`. An example diagnostic table would look something like this (note, this is not based on the example `stc.yml` above, this is just a random example):
  ```
  ---------------------------------------------------
  | CONFIG NAME                         | STATUS      |
  ---------------------------------------------------
  | hosts.host-01.ip                    | loaded      |
  | hosts.host-01.vms.host-01-vm-01.ip  | loaded      |
  | hosts.host-01.ssh.user              | loaded      |
  | hosts.host-02.ip                    | not found   |
  | hosts.host-02.vms.host-02-vm-01.ip  | loaded      |
  | hosts.host-02.ssh.user              | loaded      |
  ---------------------------------------------------
  ```
- If the `--preflight` flag is NOT passed, `stc` will check if any required config values are missing. If any are missing, it will print the same diagnostic table and exit with exit code 1.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## `stc` has `ansible bootstrap-host` command

- `stc` has an `ansible bootstrap-host` command which takes 0 or more `--host <h>` arguments where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`.
- It also takes an optional `--preflight` command.
- This command operates on a collection of hosts. We'll denote this collection as `final_hosts`. `final_hosts` is set by the `--host` arguments that are passed in. If 0 `--host <h>` arguments are passed, `stc` will use every single top-level host name in `stc.yml` as `final_hosts` (VMs are not included in the default).
- Every ansible playbook requires the following 3 fields (at minimum) to be present in every host in `final_hosts`: `.ip`, `.ssh.user`, and `.ssh.port`. The host's name is always present as a map key and is not checked. Let's call these the "base configs".
- The `./ansible/playbooks/bootstrap-host.yml` playbook specifically additionally needs these 3 variables to be set for each host in `final_hosts`: `ssh_port`, `ssh_public_key`, and `auto_update_reboot_time`. Let's call these the "bootstrap configs".
- The `./ansible/playbooks/bootstrap-host.yml` playbook specifically also needs 2 secret variables to be set for each host in `final_hosts`: `admin_email` and `admin_password`. These values will NOT and will never be present in `stc.yml`.  Let's call these "bootstrap secrets".
- In other words, the `bootstrap-host.yml` playbook requires a total of `6n + 2` config values, where `n` is the number of hosts. `6n` because there are 6 stc.yml config values per host. `+2` because of the 2 bootstrap secrets.
- If the `--preflight` flag is passed, `stc` will print a diagnostic table of two columns. There will be one row in this table for each host, for each config value. The header of the first column will be called `CONFIG NAME`. The first column will be the fully qualified path of the config in `stc.yml`. The header of the second column will be `STATUS`. It will be `loaded` if that config value was found or `not found` if that config value was not found. The "bootstrap secret" rows will behave a little bit differently. These don't have a path in `stc.yml` since they will never be in `stc.yml`. Instead of their `CONFIG NAME` column having a fully qualified yaml path, they will have the name of an environmental variable. For `admin_email` the corresponding environmental variable is `STC_ADMIN_EMAIL`. For `admin_password` the corresponding is `STC_ADMIN_PASSWORD`. Since these values will never be in `stc.yml`, they would always come up as `not found`. This wouldn't make much sense. Instead, the code will look for these values in the environment. For `admin_email` it will look for `STC_ADMIN_EMAIL` in the environment. If it finds an entry with a value, it will set the `STATUS` cell for the corresponding row to `loaded`. If it does not find it, it will set the `STATUS` cell for the corresponding row to `will prompt`. After printing, `stc` will exit with a status code of 0. it is done. The sole job of `--preflight` is to print a diagnostic table. An example diagnostic table would look something like this (note, this is not based on the example `stc.yml` above, this is just a random example):
  ```
  -----------------------------------------------
  | CONFIG NAME                       | STATUS      |
  -----------------------------------------------
  | hosts.host-01.ip                  | loaded      |
  | hosts.host-01.ssh.user            | loaded      |
  | hosts.host-01.ssh.port            | loaded      |
  | hosts.host-01.ssh.public_key_path | loaded      |
  | hosts.host-02.ip                  | loaded      |
  | hosts.host-02.ssh.user            | loaded      |
  | hosts.host-02.ssh.port            | not found   |
  | hosts.host-02.ssh.public_key_path | loaded      |
  | STC_ADMIN_EMAIL                   | will prompt |
  | STC_ADMIN_PASSWORD                | loaded      |
  -----------------------------------------------
  ```
- If the `--preflight` command is NOT passed, `stc` will check if any of the "base configs" or "bootstrap configs" are missing. if any are missing, it will print the same exact diagnostic table, which should serve as a user-friendly indication of the required configs that are missing. `stc` should exit with an exit code of 1.
- Otherwise, if we're here, it means `stc` found all 6 "base configs" and "bootstrap configs" per host.
- Next, for any of [`STC_ADMIN_EMAIL`, `STC_ADMIN_PASSWORD`] that are not set in the environment, `stc` will prompt the user to enter the value through stdin in a user-friendly way. In other words:
  - if `STC_ADMIN_EMAIL` is not set in the environment, `stc` will ask the user to enter the value for `admin_email` in a user-friendly way. it will associate the value that the user enters as the admin email, after trimming whitespace. if it IS set in the environment and its, it won't prompt the user. It will associate the value found in the environment as the admin email.
  - if `STC_ADMIN_PASSWORD` is not set in the environment, `stc` will ask the user to enter the value for `admin_password` in a user-friendly way. it will associate the value that the user enters as the admin password, after trimming whitespace. if it IS set in the environment, it won't prompt the user. It will associate the value found in the environment as the admin password.
- Finally, this command will use all 9 values that were found in some combination of stc.yml / prompting / the environment to pass as ansible configs.
- For a given host `my_awesome_host` in `final_hosts`, `stc` should communicate the "base configs" and "bootstrap configs" to ansible by creating the following file: `.generated/ansible/inventory/host_vars/my_awesome_host/vars.yml`, and putting all the "base configs" and "bootstrap configs" inside of that file. `stc` should NOT put the "bootstrap secrets" there. instead, `stc` should create a temporary file, and put the bootstrap secrets in there,
- Finally, `stc` should run `./ansible/playbooks/bootstrap-host.yml` and pass that temporary file as an argument to the ansible playbook.

### edge cases
- If one or more `--host` arguments are provided for names that do not match any top-level host or VM in `stc.yml`, this command should print a user friendly error indicating all of the names that were passed as arguments but were not found in `stc.yml`.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## stc has an `ansible setup-host` command

- `stc` has an `ansible setup-host` command which takes 0 or more `--host <h>` arguments where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`.
- It also takes an optional `--preflight` flag.
- This command operates on a collection of hosts. We'll denote this collection as `final_hosts`. `final_hosts` is set by the `--host` arguments that are passed in. If 0 `--host <h>` arguments are passed, `stc` will use every single top-level host name in `stc.yml` as `final_hosts` (VMs are not included in the default).
- Every ansible playbook requires the following 3 fields (at minimum) to be present in every host in `final_hosts`: `.ip`, `.ssh.user`, and `.ssh.port`. The host's name is always present as a map key and is not checked. Let's call these the "base configs".
- The `./ansible/playbooks/setup-host.yml` playbook does not require any additional non-secret variables beyond the "base configs".
- The `./ansible/playbooks/setup-host.yml` playbook additionally needs 2 secret variables: `smtp_user` and `smtp_password`. These values will NOT and will never be present in `stc.yml`. Let's call these the "setup secrets".
- In other words, the `setup-host.yml` playbook requires a total of `3n + 2` config values, where `n` is the number of hosts.
- If the `--preflight` flag is passed, `stc` will print a diagnostic table of two columns. There will be one row per host per base config, plus one row per setup secret. The header of the first column will be called `CONFIG NAME`. The first column will be the fully qualified path of the config in `stc.yml`. The header of the second column will be `STATUS`. It will be `loaded` if that config value was found, or `not found` if that config value was not found. The "setup secret" rows behave differently: since these values will never be in `stc.yml`, their `CONFIG NAME` will be the name of an environment variable instead of a yaml path. For `smtp_user` the corresponding environment variable is `STC_SMTP_USER`. For `smtp_password` the corresponding is `STC_SMTP_PASSWORD`. The code will look for these values in the environment. If found and non-empty, `STATUS` will be `loaded`. If not found, `STATUS` will be `will prompt`. After printing, `stc` will exit with a status code of 0. An example diagnostic table would look something like this (note, this is not based on the example `stc.yml` above, this is just a random example):
  ```
  ---------------------------------------------
  | CONFIG NAME                 | STATUS      |
  ---------------------------------------------
  | hosts.host-01.ip            | loaded      |
  | hosts.host-01.ssh.user      | loaded      |
  | hosts.host-01.ssh.port      | not found   |
  | STC_SMTP_USER               | will prompt |
  | STC_SMTP_PASSWORD           | loaded      |
  ---------------------------------------------
  ```
- If the `--preflight` flag is NOT passed, `stc` will check if any of the "base configs" are missing. If any are missing, it will print the same exact diagnostic table, which should serve as a user-friendly indication of the required configs that are missing. `stc` should exit with an exit code of 1.
- Otherwise, if we're here, it means `stc` found all 3 "base configs" per host.
- Next, for any of [`STC_SMTP_USER`, `STC_SMTP_PASSWORD`] that are not set in the environment, `stc` will prompt the user to enter the value through stdin in a user-friendly way. In other words:
  - if `STC_SMTP_USER` is not set in the environment, `stc` will ask the user to enter the value for `smtp_user` in a user-friendly way. It will associate the value that the user enters as the SMTP username, after trimming whitespace. If it IS set in the environment, it won't prompt the user.
  - if `STC_SMTP_PASSWORD` is not set in the environment, `stc` will ask the user to enter the value for `smtp_password` in a user-friendly way. It will associate the value that the user enters as the SMTP password, after trimming whitespace. If it IS set in the environment, it won't prompt the user.
- Finally, this command will use all values found in some combination of `stc.yml` / prompting / the environment to pass as ansible configs. The "base configs" are communicated to ansible by creating `.generated/ansible/inventory/host_vars/<host-name>/vars.yml` for each host in `final_hosts`. The "setup secrets" are NOT written there. Instead, `stc` creates a temporary file containing only the setup secrets, and passes it to the ansible playbook as an extra vars file.
- Finally, `stc` should run `./ansible/playbooks/setup-host.yml`.

### edge cases
- If one or more `--host` arguments are provided for names that do not match any top-level host or VM in `stc.yml`, this command should print a user friendly error indicating all of the names that were passed as arguments but were not found in `stc.yml`.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## stc has an `ssh add` command

- `stc` has an `ssh add` command which takes 0 or more `--host <h>` arguments where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`.
- It also takes an optional `--preflight` flag.
- This command operates on a collection of hosts. We'll denote this collection as `final_hosts`. `final_hosts` is set by the `--host` arguments that are passed in. If 0 `--host <h>` arguments are passed, `stc` will use every single top-level host name in `stc.yml` as `final_hosts` (VMs are not included in the default).
- For each host or VM in `final_hosts`, this command reads connection details from `stc.yml` and appends a new `Host` block to `~/.ssh/config` on the local workstation. If `~/.ssh/config` does not exist, it is created with `0600` permissions.
- The required config values for a top-level host are: `hosts.<name>.ip`, `hosts.<name>.ssh.user`, `hosts.<name>.ssh.public_key_path`, and `hosts.<name>.ssh.port`. The host's name, being a map key, is always present and is not checked. The resulting SSH block looks like:
  ```
  Host <name>
    HostName <ip>
    User <ssh.user>
    IdentityFile <ssh.public_key_path>
    Port <ssh.port>
  ```
- For a VM, the same 4 fields are required from the VM's own config (`hosts.<name>.vms.<vm-name>.ip`, etc.), plus the parent host's `hosts.<name>.ssh.user` and `hosts.<name>.ip` in order to construct a `ProxyJump` directive. The resulting SSH block looks like:
  ```
  Host <vm-name>
    HostName <vm.ip>
    User <vm.ssh.user>
    IdentityFile <vm.ssh.public_key_path>
    Port <vm.ssh.port>
    ProxyJump <parent.ssh.user>@<parent.ip>
  ```
- If the `--preflight` flag is passed, `stc` will print a diagnostic table showing the status of each required config value for every host in `final_hosts`, then exit with code 0 without writing anything to `~/.ssh/config`. There are no secret variables for this command, so all rows will show either `loaded` or `not found`. An example for two top-level hosts would look like:
  ```
  -----------------------------------------------
  | CONFIG NAME                       | STATUS      |
  -----------------------------------------------
  | hosts.host-01.ip                  | loaded      |
  | hosts.host-01.ssh.user            | loaded      |
  | hosts.host-01.ssh.public_key_path | loaded      |
  | hosts.host-01.ssh.port            | loaded      |
  | hosts.host-02.ip                  | loaded      |
  | hosts.host-02.ssh.user            | loaded      |
  | hosts.host-02.ssh.public_key_path | not found   |
  | hosts.host-02.ssh.port            | loaded      |
  -----------------------------------------------
  ```
- If the `--preflight` flag is NOT passed, `stc` will check if any of the required config values are missing. If any are missing, it will print the same diagnostic table and exit with exit code 1.
- For each host in `final_hosts`, if `~/.ssh/config` already contains an entry whose `Host` alias matches that host's name, `stc` will skip that host and print a user-friendly message indicating the entry already exists. It will continue processing the remaining hosts and exit with code 0.
- This command has no secret variables and does not prompt for any input.

### edge cases
- If one or more `--host` arguments are provided for names that do not match any top-level host or VM in `stc.yml`, this command should print a user friendly error indicating all of the names that were passed as arguments but were not found in `stc.yml`.
- If there are no hosts in `stc.yml` this command will print a user friendly error indicating this and exit with exit code 1.

## stc has a `setup` command

- `stc` has an `setup` command which takes 0 or more `--host <h>` arguments where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`.
- It also takes an optional `--preflight` flag.
- This command operates on a collection of hosts. We'll denote this collection as `final_hosts`. `final_hosts` is set by the `--host` arguments that are passed in. If 0 `--host <h>` arguments are passed, `stc` will use every single top-level host name in `stc.yml` as `final_hosts` (VMs are not included in the default).
- Suppose that `$hostArgs` is a shell variable that holds 0 or more `--host <h>` arguments, where `<h>` can be either a top-level host name or a VM name as defined in `stc.yml`. Suppose that `$preflight` is a shell variable that either holds the string `--preflight` or the empty string. For every possible value of `$hostArgs`, and for every possible value of `$preflight`, running `stc setup $hostArgs $preflight` would be the same as running the following commands in order:
  - `stc inventory generate`
  - `stc ansible setup-host $hostArgs $preflight`
  - `stc ansible bootstrap-host $hostArgs $preflight`
  - `stc ssh add $hostArgs $preflight`
- However, `stc setup` offers one convenience over manually running those commands individually. If the `--preflight` argument is provided, `stc setup` *merges* the diagnostic tables of all of those commands. In other words, if someone runs `stc setup --preflight`, instead of seeing 1 diagnostic table for `stc inventory generate`, and 1 diagnostic table for `stc ansible setup-host` and 1 diagnostic table for `stc bootstrap-host`, and 1 diagnostic table for `ssh add`, showing a total of 4 diagnostic tables, `stc setup` will show 1 diagnostic table which has the aggregated results of all 4 commands.
- Also, ideally, instead of prompting you once for secret variables necessary for `stc ansible setup-host` and then executing that, and then prompting you again for secret variables necessary for `stc ansible bootstrap-host` and then executing that, `stc setup` will prompt you for all secret variables upfront before executing the ansible playbooks.
