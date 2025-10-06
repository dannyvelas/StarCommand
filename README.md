# Homelab setup

## Prerequisites
* A server
* A computer you can use to ssh into the server
* Terraform installed on that computer
* An ethernet connection between your router and the server

## Instructions
- Flash [Proxmox](https://www.proxmox.com/en/downloads) ISO onto a USB or SSD or disk and then connect that to your server so that you can boot your server with the Proxmox VE OS.
- In the setup, you'll be given an option to name your default proxmox node, pick a name but remember it because you will need it later.
- In setup, on the "Management Network Configuration" page:
  - For management interface, pick the network card that is being used for ethernet.
  - For Hostname (FQDN), put proxmox.lan.
  - For IP address (CIDR), pick an IP address that is not assigned by your router and that you can reserve for your server, let's suppose it's `1.2.3.4`.
  - For gateway, put the IP address of your router.
  - For DNS server, put either your router's IP address or use a public DNS server like `1.1.1.1` (Cloudflare) `8.8.8.8` (Google).
- Verify that from another computer you can `ping 1.2.3.4` and access `https://1.2.3.4:8006`.
- Add an `ssh` key so you have remote access to your server from another computer.
- Run `./setup_proxmox.sh 1.2.3.4 /path/to/your/public/.ssh/key /path/to/your/private/.ssh/key`. This script will:
  - Update package repositories
  - Install `sudo` and `tailscale`
  - Create a `dannyvelasquez` user with full `sudo` permissions and SSH access using `/path/to/your/public/.ssh/key`.
  - Create a `terraform` user with partial `sudo` permissions and SSH access `/path/to/your/public/.ssh/key`.
  - Create a Proxmox `terraform` user with an API token with limited permissions.
  - Harden SSH access so that root and password logins become not permitted. Also, makes `ssh` use a random port instead of `22` to reduce attempted logins.
  - Create a `/mnt/media` directory that will be used for mounting.
  - Verify SSH was set up correctly
- After running the script, it will:
  - Prompt you for a password, enter one and save it in Bitwarden.
  - Show you the API token that was created for the Terraform Proxmox user. Save this in Bitwarden.
  - Show you a port that was randomly chosen port for `ssh`. Suppose it is `1234`. Save it in `~/.ssh/config`:
    ```
    Host proxmox
      Hostname 1.2.3.4
      User dannyvelasquez
      IdentityFile /path/to/your/private/.ssh/key
      Port 1234
    ```
- Create a file in this directory called `terraform.tfvars` it should look like this:
```
endpoint        = "https://1.2.3.4:8006/"
api_token       = "terraform@pve!provider=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ssh_public_key  = "/path/to/your/public/.ssh/key"
ssh_private_key = "/path/to/your/private/.ssh/key"
node            = "whatever-node-name-you-chose-earlier"
```
- Run `terraform apply`. This should create an Ubuntu VM that has a shared mount to the `/mnt/media` directory of its host.
