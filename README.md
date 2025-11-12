# Homelab setup

## Prerequisites
* A server which is connected to Ethernet. 
* A computer you can use to ssh into the server.
* [Terraform](https://developer.hashicorp.com/terraform/install) installed on that computer.
* [Ansible](https://formulae.brew.sh/formula/ansible) installed on that computer.
* A [Tailscale](https://login.tailscale.com/start) account.

## Instructions

### Set up Proxmox manually
- Flash [Proxmox](https://www.proxmox.com/en/downloads) ISO onto a USB or SSD or disk and then connect that to your server so that you can boot your server with the Proxmox VE OS.
- After accepting the terms and conditions, you can configure your filesystem and the amount of space on your drive that will be used by Proxmox:
  - You probably want `ext4` or `xfs`, unless you know what you're doing.
  - You can specify how much space on your hard-drive you want Proxmox to use using the `hdsize` field. By default it will use the whole thing. It will create two main logical volumes:
    - "root", for your OS and file system. From what I've seen this takes up around 30% of the space you gave it.
    - "data", for VM disks. This seems to take up whatever remaining space there is from the space you gave it.
  - If you don't want Proxmox to use all the remaining space on your drive for VM disks, then you should use a value for the `hdsize` field which is smaller than the total capacity of your drive. This is especially true if you have a smaller hard-drive and can't easily add storage. You can put the available space to better use, like for storing media content.
- In setup, on the "Management Network Configuration" page:
  - For management interface, pick the network card that is being used for ethernet.
  - For Hostname (FQDN), put proxmox.lan.
  - For IP address (CIDR), pick an IP address that is not assigned by your router and that you can reserve for your server. This will be used a lot. From here on, we will use the special value `1.2.3.4` to represent your server's IP address.
  - For gateway, put the IP address of your router. From here on, we will use the special value `10.0.0.1` to represent your router's IP address.
  - For DNS server, put either your router's IP address or use a public DNS server like `1.1.1.1` (Cloudflare) `8.8.8.8` (Google).
- Verify that from another computer you can `ping 1.2.3.4` and access `https://1.2.3.4:8006`.
- Add an `ssh` key to your Proxmox server and verify afterward that you have remote `ssh` access to your server from your other computer.

## Ansible
### Variables set up
- `cp ansible/vars.example.yml /var/homelab.yml`.
- Pick a random port: `echo $RANDOM | jq '. + 1024 | . % 65535'`, this will be used in future steps. From now on, we will use the special value `1234` to represent this randomly generated port.
- Save this port into `/var/homelab.yml`.
- [Generate a Tailscale auth key](https://login.tailscale.com/admin/settings/keys), save it in Bitwarden and put it in `/var/homelab.yml`.
- Update `./ansible/inventory.ini` so that the `proxmox` host has IP address `1.2.3.4`.

### Run playbook
- If your public key is anything other than `~/.ssh/id_ed25519.pub`, change it in `./ansible/setup-proxmox.yml`.
- Add the following to your `~/.ssh/config` file, this will be used by the `./ansible/setup-proxmox.yml` playbook:
  ```
  Host proxmox
    Hostname 1.2.3.4
    User root
    IdentityFile /path/to/your/private/.ssh/key
    Port 22
  ```
- Run `ansible-playbook -i ansible/inventory.ini ansible/setup-proxmox.yml -u root`, this will:
  - Install `sudo`.
  - Create an `admin` user with full `sudo` permissions, that can log-in via SSH with the same key as root.
  - Harden SSH access so that root and password logins become not permitted.
  - Create a `terraform` user with partial `sudo` permissions and SSH access `/path/to/your/public/.ssh/key`.
  - Create a Proxmox `terraform` user with an API token with limited permissions.
  - Install `tailscale`.
  - Run `tailscale` and add your server to be a Tailscale node.
  - Create a `/mnt/media` directory that will be used for mounting.
- After running this playbook:
  - It will show you the API token that was created for the Terraform Proxmox user. Save this in Bitwarden.
  - ssh logins with the `root` user or port 22 will no longer work, so update the `User` in `~/.ssh/config` to be `admin` instead of `root`. Also update the `Port` to be the port from before.
  - You should be able to run it as many times as you want, except as admin (`-u admin`) and not as root as we did above.
  - You should be able to go to the [Tailscale machines page](https://login.tailscale.com/admin/machines) and see your server there as a Tailscale node.

## Terraform
- Decide on the IP address that you would want for a new Plex VM. From now on, we will use the special value `<plex-vm-ip>` to represent your plex VM's IP address.
- Create a file in this directory called `terraform.tfvars`. It should look like this:
```
node            = "proxmox"
router_ip       = "10.0.0.1"
endpoint        = "https://1.2.3.4:8006/"
api_token       = "terraform@pve!provider=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ssh_address     = "1.2.3.4"
ssh_port        = 1234
ssh_public_key  = "/path/to/your/public/.ssh/key"
ssh_private_key = "/path/to/your/private/.ssh/key"
plex_vm_ip      = "<plex-vm-ip>"
```
- If you chose a hostname other than "proxmox" in your management configuration page, when you were manually setting up Proxmox in the first step, set the "node" key to that value.
- The x's in `api_token` should be replaced with the api token you received in the step before.
- Run `terraform apply`. This should create an Ubuntu VM on IP that can mount to `/mnt/media` on the Proxmox host.
- At this point, you should be able to ssh into the ubuntu VM: `ssh ubuntu@<plex-vm-ip> -i /path/to/your/private/.ssh/key`.

## Ansible for Plex VM
- Update `./ansible/inventory.ini` so that the `plex` host has IP address `<plex-vm-ip>`.
- Run `ansible-playbook -i ansible/inventory.ini ansible/setup-plex-vm.yml -u ubuntu`
- After this, you should be able to go to visit `http://<plex-vm-ip>:32400` and see the Plex welcome screen.
