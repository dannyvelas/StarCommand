# Homelab setup

## Prerequisites
* A server which is connected to Ethernet. 
* A computer you can use to ssh into the server.
* [Terraform](https://developer.hashicorp.com/terraform/install) installed on that computer.
* [Ansible](https://formulae.brew.sh/formula/ansible) installed on that computer.
* A [Tailscale](https://login.tailscale.com/start) account.

## Instructions

### Set up proxmox manually
- Flash [Proxmox](https://www.proxmox.com/en/downloads) ISO onto a USB or SSD or disk and then connect that to your server so that you can boot your server with the Proxmox VE OS.
- In the setup, you'll be given an option to name your default proxmox node, pick a name but remember it because you will need it later.
- In setup, on the "Management Network Configuration" page:
  - For management interface, pick the network card that is being used for ethernet.
  - For Hostname (FQDN), put proxmox.lan.
  - For IP address (CIDR), pick an IP address that is not assigned by your router and that you can reserve for your server, let's suppose it's `1.2.3.4`.
  - For gateway, put the IP address of your router.
  - For DNS server, put either your router's IP address or use a public DNS server like `1.1.1.1` (Cloudflare) `8.8.8.8` (Google).
- Verify that from another computer you can `ping 1.2.3.4` and access `https://1.2.3.4:8006`.
- Add an `ssh` key to your proxmox server and verify afterward that you have remote `ssh` access to your server from your other computer.
- Pick a random port: `echo $RANDOM | jq '. + 1024 | . % 65535'`, this will be used in future steps. Let's suppose it's `1234`.
- Add the following to your `~/.ssh/config` file, and put the port that you picked in the previous step:
  ```
  Host proxmox
    Hostname 1.2.3.4
    User dannyvelasquez
    IdentityFile /path/to/your/private/.ssh/key
    Port 1234
  ```

## Ansible
- `cp ansible/vars.example.yml ansible/vars.yml`
- [Generate a Tailscale auth key](https://login.tailscale.com/admin/settings/keys), save it in Bitwarden and put it in `./ansible/vars.yml`.
- Save the port from before into `./ansible/vars.yml`.
- Run `ansible-playbook -i ansible/inventory.ini ansible/proxmox_setup.yml`, this will:
  - Install `sudo` and `tailscale`
  - Create a `dannyvelasquez` user with full `sudo` permissions and SSH access using `/path/to/your/public/.ssh/key`.
  - Create a `terraform` user with partial `sudo` permissions and SSH access `/path/to/your/public/.ssh/key`.
  - Create a Proxmox `terraform` user with an API token with limited permissions.
  - Harden SSH access so that root and password logins become not permitted. Also, makes `ssh` use a random port instead of `22` to reduce attempted logins.
  - Create a `/mnt/media` directory that will be used for mounting.
- After running the playbook, it will show you the API token that was created for the Terraform Proxmox user. Save this in Bitwarden.

## Terraform
- Create a file in this directory called `terraform.tfvars` it should look like this:
```
endpoint        = "https://1.2.3.4:8006/"
api_token       = "terraform@pve!provider=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ssh_public_key  = "/path/to/your/public/.ssh/key"
ssh_private_key = "/path/to/your/private/.ssh/key"
node            = "whatever-node-name-you-chose-earlier"
```
- Run `terraform apply`. This should create an Ubuntu VM that has a shared mount to the `/mnt/media` directory of its host.
