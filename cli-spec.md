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
  setup [--host <h>]...                  Apply desired state to all hosts, or limit to the ones given
  wg add <name>                          Add a WireGuard client (registers peer server-side, generates client config)
  status                                 Show cluster status (hosts, services, VPN, k3s)
  teardown                               Tear down all VMs
  version                                Print version

Low-level commands:
  inventory generate                       Generate the Ansible inventory file for all hosts
  ansible bootstrap-host [--host <h>]...   Run the bootstrap-host playbook against all hosts/VMs, or limit to the ones given
  ansible setup-host [--host <h>]...       Run the setup-host playbook against all hosts, or limit to the ones given
  ansible setup-vm [--host <h>]...         Run the setup-vm playbook against all VMs, or limit to the ones given
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

## stc has a `setup` command

- `stc` has a setup command which takes 0 or more `--host` arguments.
- It also takes an optional `--preflight` argument.
- This `setup` command will read an `stc.yml` file in the root of the directory 
