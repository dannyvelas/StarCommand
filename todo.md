## finished

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
- [x] make setup:proxmox taskfile task idempotently update the ssh file if needed
- [x] test if you can actually store c.client.Secrets() in a variable in client/bitwarden.go
- [x] maybe rename "resolve" package in go
- [x] make it so that every provider doesn't have to call decode
- [x] make it so that reading from bitwarden is optional. now it is required.
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
- [x] make env reader more testable
- [x] make file reader more testable
- [x] check where `config` interface is being used because maybe it's not needed in those places. in those places, maybe it can be `any` instead. in `validateConfig` we're accepting an `any` anyway..so
  - [x] (dependant on above) maybe change name of `config` interface to be called `validator` or something like this
- [x] make the constructor for fullConfig name the readers it wants to use
- [x] make the constructor for file_reader specify the files it wants to read
- [x] add signature descriptions to every public function
- [x] test if you can use a custom reader
- [x] make sure there are no instances of `SimpleReadResult{}`, or `DiagnosticReadResult{}`
- [x] change file reader name to be FileReaderYAML or something like this
- [x] change `check reqs proxmox` command be `check config proxmox` instead.
- [x] make models.SetSSH return a sentinel error, not a dynamic error
- [x] make app.SetFile return a sentinel error, not a dynamic error
- [x] remove conflux stuff from set_ssh
- [x] make `get config proxmox` accept a flag that looks like this: `--for <target>` where target could be a `ansible` or `ssh`. you can also supply multiple targets like this: `--for ansible --for ssh` or like this: `--for ansible,ssh`. If no flag is passed, `ansible` is assumed.
- [x] make `check config proxmox` accept a flag that looks like this: `--for <target>` where target could be a `ansible` or `ssh`. you can also supply multiple targets like this: `--for ansible --for ssh` or like this: `--for ansible,ssh`. If no flag is passed, `ansible` is assumed.
- [x] add support for other host-aliases other than proxmox. right now `labctl get config` and `labctl check reqs` pretty much will only work for proxmox because it's hardcoded
- [x] see if there's a way to reduce duplication: in `models.AliasToStruct` you are listing all the aliases you support. you are again listing those same aliasas in the cobra configs of `getConfig` and `checkConfig`, in the `ValidArgs` field.
- [x] add test for app.go
- [x] figure out how to share variables
  - vm_id = 100 is both in `terraform/plex_lxc/main.tf` and `ansible/inventory.ini`
  - `proxmox_node_name` is both in terraform variables and `./ansible/group_vars/all/all.yml`
  - port 17031 is both in `./ansible/group_vars/all/all.yml` and `terraform/global/firewall.tf` and `terraform/plex_lxc/main.tf`.
  - this is solved with labctl
- [x] make Ansible playbook send terraform API token directly to Bitwarden Secrets Manager (BWS)
  - this is solved with labctl
- [x] create a "base" terraform LXC module
  - WONT DO
- [x] create a "base" terraform VM module
  - WONT DO

## infra todos
- [x] make sure incus server host has firewall rules like it did before with terraform
- [x] make sure plex has firewall rules like it did with terraform
- [x] migrate README to new labctl
- [x] migrate taskfile to new labctl
- [x] is there a way to mid-playbook switch from "root" to "admin" after ssh_harden runs? if so, do it
  - ehhh it's kind of a pain. better to just split it into two playbooks.
  - so this new to-do item is to switch ssh-hardening to be its own playbook instead of its own role. and, you'll just have to execute both playbooks for the first time
- [x] make it so that the UX allows for two commands: bootstrap everything on a fresh fleet of debian servers that i just got up and running. add a server to my fleet.
- [x] maybe stop making hosts "special". e.g. maybe get rid of the fact that "ansible inventory add incus" and "ansible inventory run incus" have a special behavior that doesn't exist for "ansible inventory add random-name-here".
- [ ] should we make the :check tasks be prerequisits to the regular tasks?
- [ ] test plex works as a k8s application
- [ ] make sure plex VM is mounted as read only
- [ ] fix ssh-restart logic in ssh-harden. it seems to always restart ssh.service even if an LXC uses ssh.socket instead
- [ ] migrate to OVN

## infra next
- [ ] use Netboot.xyz + https://pikvm.org/ + with something like Fedora KickStart

## coding todos
- [x] make `handler.SetFile` more testable
- [x] make it so that my home directory doesn't have to be hardcoded in the tests for `SetFile`
- [x] make it so that ansible configs are read from a file
- [x] add context.Background() which is initialized at main and passed into Execute
- [x] use `runE` in cobra
- [~] (CONFLUX) maybe make bitwarden secrets read things piecemeal, instead of just dumping everything into a map
  - i don't think i can do this
- [~] (CONFLUX) make conflux read configs once instead of every single time that `conflux.Unmarshal` is called. file reads and bitwarden api calls are expensive.
- [x] add support where the user can check the configs that were missing/found that are necessary to run:
  - `stc setup`, or `stc setup --host <host>`
  - `stc inventory [--host <host>]`
  - `stc ansible bootstrap-host`
  - `stc ansible bootstrap-host --vms`
  - `stc ansible setup-host`
  - `stc ansible setup-vm`
  - `stc ssh add <host>`
  - `stc terraform apply`
- [x] see if i can just embed a struct inside of all ansible configs that has {nodeIP, sshPrivateKey, sshUser, sshPort} so i don't have to do as much copy-paste
- [ ] see if i really need `PersistentFlags` in `cmd/labctl/root.go` or if i should use something else
- [ ] make sure no playbook except for "bootstrap" has become: true.
- [ ] stc runs terraform to create VM. still need to test.
- [ ] stc uses ansible to bootstrap+setupVM. still need to test
- [ ] make it so that stc always runs "terraform init"

## terraform-provider-proxmox repo
- [x] make PR to correct the steps necessary to run `make example`
  - create directory `mkdir -p /mnt/bindmounts/shared`
  - use username and password without ssh section at all
- [x] make PR to fix broken link in README for instructions to setup local proxmox 
- [ ] probably will need to make a PR to set up cluster firewall logic in `bpg/terraform-provider-proxmox`.
