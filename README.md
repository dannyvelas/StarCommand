# Homelab infra and playbooks

## Prerequisites
* A server which is connected to Ethernet. 
* A computer you can use to ssh into the server.
* [Terraform](https://developer.hashicorp.com/terraform/install) installed on that computer.
* [Ansible](https://formulae.brew.sh/formula/ansible) installed on that computer.
* A [Tailscale](https://login.tailscale.com/start) account.


<details>

<summary>## Install proxmox</summary>
- Flash [Proxmox](https://www.proxmox.com/en/downloads) ISO onto a USB or SSD or disk and then connect that to your server so that you can boot your server with the Proxmox VE OS.
- After accepting the terms and conditions, you can configure your filesystem and how your disk will be provisioned by Proxmox:
  - You probably want `ext4` or `xfs`, unless you know what you're doing.
  - By default, Proxmox will create a few logical volumes inside of `sda3`. The main ones to note are:
    - "root", for your OS and file system. From what I've seen, this takes up around 30% of `hdsize`.
    - "data", for VM disks, which ends up being of size `hdsize - rootsize - swapsize - minfree`.
  - If you have a smaller hard-drive and can't easily add storage, you might want to make these two logical volumes a bit smaller. I made the root logical volume 39.9GiB and the data logical volume 55GiB. This gave me around 135GiB of free space in the `pve` volume group.
- After, you will be asked for an administrator email and password. Create a password, enter it, and store it in Bitwarden. This will be the "root" password. From here on, we will use the special value `<password>` to represent this password.
- In setup, on the "Management Network Configuration" page:
  - For management interface, pick the network card that is being used for ethernet.
  - For Hostname (FQDN), put `proxmox.lan`. The part before the first `.` will become your Proxmox node name. In this case, my node name is `proxmox`. From here on, we will use the special value `<node-name>` to represent your node name.
  - For IP address (CIDR), pick an IP address that is not assigned by your router and that you can reserve for your server. This will be used a lot. From here on, we will use the special value `1.2.3.4` to represent your server's IP address.
  - For gateway, put the IP address of your router. From here on, we will use the special value `10.0.0.1` to represent your router's IP address.
  - For DNS server, put either your router's IP address or use a public DNS server like `1.1.1.1` (Cloudflare) `8.8.8.8` (Google).
- Verify that from another computer you can `ping 1.2.3.4` and access `https://1.2.3.4:8006`.
- Run: `ssh-copy-id -i /path/to/your/public/.ssh/key root@1.2.3.4`. In other words, add an `ssh` key to your Proxmox server and verify afterward that you have remote `ssh` access to your server from your other computer.
- If you're using a laptop as a server, you might want to run this as well so that you can close the lid without it sleeping: `sudo systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target`.

</details>

## Set Ansible variables
- `cp ansible/vars.example.yml /var/homelab.yml`.
- Pick a random port: `echo $RANDOM | jq '. + 1024 | . % 65535'`, this will be used in future steps. From now on, we will use the special value `1234` to represent this randomly generated port.
- Save this port into `/var/homelab.yml`.
- [Generate a Tailscale auth key](https://login.tailscale.com/admin/settings/keys), save it in Bitwarden and put it in `/var/homelab.yml`.
- Update `./ansible/inventory.ini` so that the `proxmox` host has IP address `1.2.3.4`.

## Set up Proxmox with Ansible
- If your public key is anything other than `~/.ssh/id_ed25519.pub`, change it in `./ansible/setup-proxmox.yml`.
- Run `ansible-playbook -i ansible/inventory.ini ansible/setup-proxmox.yml -u root`, this will:
  - Install `sudo`.
  - Create an `admin` user with full `sudo` permissions, that can log-in via SSH with the same key as root.
  - Harden SSH access so that root and password logins become not permitted.
  - Create a `terraform` user with partial `sudo` permissions and SSH access via your public key: `/path/to/your/public/.ssh/key`.
  - Create a Proxmox `terraform` user with an API token with limited permissions.
  - Install `tailscale`.
  - Run `tailscale` and add your server to be a Tailscale node.
  - Make sure that proxmox "local" storage can have items of type "import" and "snippets".
- After running this playbook:
  - It will show you the API token that was created for the Terraform Proxmox user. Save this in Bitwarden.
  - Add the following to your `~/.ssh/config` file, this will be used by the `./ansible/setup-proxmox.yml` playbook:
    ```
    Host proxmox
      Hostname 1.2.3.4
      User admin
      IdentityFile /path/to/your/private/.ssh/key
      Port 1234
    ```
  - Change the proxmox host of `./ansible/inventory.ini` to have these values: `ansible_port=1234 ansible_user=admin`.
  - You should now be able to:
    - Run this playbook as many times as you want (without the `-u root` argument, as that won't work anymore).
    - See your server as a Tailscale node in the [Tailscale machines page](https://login.tailscale.com/admin/machines).

## Create Media logical volume with Ansible
- NOTE: This assumes you have at least 120GiB of space in your `pve` volume group.
- Run `ansible-playbook -i ansible/inventory.ini ansible/add-media-lv.yml`. This will create a new logical volume called "media" in the `pve` volume group of size 120GiB.

## Create a new VM with Terraform
- `cd terraform/vm`.
- Decide on the IP address that you would want for a VM. From now on, we will use the special value `<vm-ip>` to represent your VM's IP address.
- Create a file called `terraform.tfvars`. It should look like this:
```
node            = "<node-name>"
router_ip       = "10.0.0.1"
endpoint        = "https://1.2.3.4:8006/"
api_token       = "terraform@pve!provider=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ssh_address     = "1.2.3.4"
ssh_port        = 1234
ssh_public_key  = "/path/to/your/public/.ssh/key"
ssh_private_key = "/path/to/your/private/.ssh/key"
vm_ip           = "<vm-ip>"
```
- The x's in `api_token` should be replaced with the api token you received in the step before.
- Run `terraform apply`. This should create an Ubuntu VM that can mount to `/mnt/media` on the Proxmox host.
- At this point, you should be able to ssh into the ubuntu VM: `ssh ubuntu@<vm-ip> -i /path/to/your/private/.ssh/key`.

## Create a new LXC container with Terraform
- `cd terraform/lxc`.
- Decide on the IP address that you would want for an LXC container. From now on, we will use the special value `<lxc-ip>` to represent your container's IP address.
- Create a file called `terraform.tfvars`. It should look like this:
```
node           = "proxmox"
router_ip      = "10.0.0.1"
endpoint       = "https://10.0.0.50:8006/"
username       = "root@pam"
password       = "<password>"
ssh_public_key = "/Users/dannyvelasquez/.ssh/id_ed25519.pub"
ip             = "<lxc-ip>"
```
- Unfortunately, Proxmox doesn't support some things in this `main.tf` file without root login, so the authentication here is just root username and password.
- Run `terraform apply`. This should create an Ubuntu LXC container mounted to `/mnt/media` on the Proxmox host.
- At this point, you should be able to ssh into it: `ssh root@<lxc-ip> -i /path/to/your/private/.ssh/key`.

## Install Plex in LXC container
- Update `./ansible/inventory.ini` so that the `plex` host has IP address `<lxc-ip>`.
- Run `ansible-playbook -i ansible/inventory.ini ansible/install-plex.yml -u root`
- After this, you should be able to go to visit `http://<lxc-ip>:32400` and see the Plex welcome screen.
