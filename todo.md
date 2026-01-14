- [x] learn about what stuff won't destruct nicely on my server with `terraform destroy` and i would have to manually destruct
  - as of right now, everything would destruct nicely. the only thing not supported natively by bpg/proxmox is cloud-init configurations to mount `media_mount` to `/mnt/media` and enable `qemu-guest-agent`. But that's okay. When we run `terraform destroy` that will wipe the VM entirely so it doesn't matter.
- [x] switch to make terraform authenticate via token not via root
- [x] switch over to setup proxmox with ansible
- [x] make sections collapsible in README
- [x] add media to /mnt/media directory
- [x] make LXC plex config path (`/var/lib/plexmediaserver/Library/Application Support/Plex Media Server`) get mounted to host directory (e.g. `/mnt/media/plex-config` or something like this)
- [x] make it so that when "harden-ssh" tasks get run from "setup-proxmox" they use "homelab_admin_password" and when those tasks get run from "setup-server" for "host A", they use the "hostA_admin_password", and when those tasks get run from "setup-server" for "host B", they use the "hostB_admin_password".
- [x] maybe merge /var/homelab.yml and secrets?
- [x] document your use of secrets now
- [x] add ufw protections to proxmox
- [x] see if there's a better way to structure repository (roles, but not needed for now)
- [x] make it so that proxmox also gets server updates
- [x] add postfix support so that VPS server updates actually go to your email 
- [x] test that firewall actually works and that plex is still working
- [x] rename LXC to plexLXC in readme and otherwise
- [x] see if there are any changes that need to be made to plexLXC for firewall
- [x] rename VM to be called wireguard VM
- [x] switch VM to be on port 17031 instead of 22
- [x] add firewall rules to VM
- [x] test VM firewall
  - output:
    ```
    PORT      STATE SERVICE
    3128/tcp  open  squid-http
    8006/tcp  open  wpl-analytics
    17031/tcp open  unknown
    ```
- [x] fix the fact that `labctl resolve --help` doesn't tell you about `<host-name>`
- [x] actually make the "ssh_public_key" variable passed to ansible be the actual public key, not the file path
- [ ] create a "base" terraform LXC module
- [ ] create a "base" terraform VM module
- [ ] add jump-host LXC (re-adding tailscale stuff to README for it)
- [ ] add jump-host LXC to readme
- [ ] see if there are any changes that need to be made to jumpLXC for firewall
- [ ] figure out how to share variables
  - vm_id = 100 is both in `terraform/plex_lxc/main.tf` and `ansible/inventory.ini`
  - `proxmox_node_name` is both in terraform variables and `./ansible/group_vars/all/all.yml`
  - port 17031 is both in `./ansible/group_vars/all/all.yml` and `terraform/global/firewall.tf` and `terraform/plex_lxc/main.tf`.
- [ ] fix ssh-restart logic in ssh-harden. it seems to always restart ssh.service even if an LXC uses ssh.socket instead
- [ ] move configure apt stuff from ansible to terraform
- [ ] it seems like sometimes "terraform destroy" on the "global" terraform project doesn't actually clean the `/etc/pve/firewall/cluster.fw` settings. check if this is consistent and why this is happening. also, fix it
- [ ] see how we can convert the README to a program
- [ ] use Netboot.xyz + https://pikvm.org/ + proxmox answers file to remotely shutdown/reboot and re-install proxmox
- [ ] make Ansible playbook send terraform API token directly to Bitwarden Secrets Manager (BWS)
- [ ] figure out a way to make it so that plex data (about watch history, users with access to my plex) is stored somewhere externally so that if I nuke Proxmox, it doesn't get lost.
- [ ] make setup:proxmox taskfile task idempotently update the ssh file if needed
- [ ] migrate all variables to "./configs" dir, effectively deleting all ansible and terraform config files
- [ ] test if you can actually store c.client.Secrets() in a variable in client/bitwarden.go
- [x] maybe rename "resolve" package in go
- [x] make it so that every provider doesn't have to call decode
- [ ] make it so that reading from bitwarden is optional. now it is required.
- [x] add test so that if `validateConfig` runs for something that doesn't implement `config`, it can return `true`. and if it runs for something that does implement `config`, it can return `false`
  - NOT NECESSARY anymore
- [x] rename "results" to diagnosticMap
- [x] rename the name that the receivers of `*reader` structs use to refer to "self". right now it's "p" but that kinda doesn't make sense
- [x] remove "unvalidated" from everything. we can just call it readResult or something
- [x] probably make `ErrInvalidFields` not public anymore
  - WONT DO: it's needed for when ssh calls `config.UnmarshalInto`
- [x] use a different tag name than "bw". people won't necessarily use bitwarden. use something like "config" instead.
- [x] maybe see if we can make "Validate" a little more functional - instead of making it mutate its input argument
  - eh, Validate implementations don't have any reason to to be forced to create a completely new map. they're going to alter the original map anyway. so, whatever
- [ ] maybe make env reader more testable
- [ ] maybe make file reader more testable

## terraform-provider-proxmox repo
- [x] make PR to correct the steps necessary to run `make example`
  - create directory `mkdir -p /mnt/bindmounts/shared`
  - use username and password without ssh section at all
- [x] make PR to fix broken link in README for instructions to setup local proxmox 
