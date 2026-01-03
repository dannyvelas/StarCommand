# Homelab infra and playbooks

## Prerequisites
* A server which is connected to Ethernet. 
* A computer you can use to ssh into the server.
* [Terraform](https://developer.hashicorp.com/terraform/install) installed on that computer.
* [Ansible](https://formulae.brew.sh/formula/ansible) installed on that computer.
* A [Tailscale](https://login.tailscale.com/start) account.
* Some playbooks will send you an email when your server automatically updates. For this, you'll need an SMTP username and password. If you use Gmail, you can't use your regular password. You'll need to get a 16-character code:
  * Go to your Google Account settings.
  * Search for "App Passwords".
  * Create one called "Ansible Server"
  * Copy the 16-character code.

<details>

<summary><h2>Install Proxmox</h2></summary>

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
  - For Hostname (FQDN), put `pve.lan`. The part before the first `.` will become your Proxmox node name. In this case, my node name is `pve`. From here on, we will use the special value `<node-name>` to represent your node name.
  - For IP address (CIDR), pick an IP address that is not assigned by your router and that you can reserve for your server. This will be used a lot. From here on, we will use the special value `1.2.3.4` to represent your server's IP address.
  - For gateway, put the IP address of your router. From here on, we will use the special value `10.0.0.1` to represent your router's IP address.
  - For DNS server, put either your router's IP address or use a public DNS server like `1.1.1.1` (Cloudflare) `8.8.8.8` (Google).
- Verify that from another computer you can `ping 1.2.3.4` and access `https://1.2.3.4:8006`.
- Run: `ssh-copy-id -i /path/to/your/public/.ssh/key root@1.2.3.4`. In other words, add an `ssh` key to your Proxmox server and verify afterward that you have remote `ssh` access to your server from your other computer.
- If you're using a laptop as a server, you might want to run this as well so that you can close the lid without it sleeping: `sudo systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target`.

</details>

<details>

<summary><h2>Set Ansible variables</h2></summary>

- Pick an admin password for your home server. Save it into Bitwarden as well.
- Run: `ansible-vault create ./ansible/host_vars/proxmox_server/vault.yml`, using a vault password. Save this vault password in Bitwarden.
- In the content of that file put:
  ```
  vault_admin_password: "<admin password for your home server>"
  ```
- Update `./ansible/inventory.ini` so that the `proxmox` host has IP address `1.2.3.4`.
- Run `ansible-vault create ./ansible/group_vars/all/vault.yml`, and add the following:
  ```
  smtp_user: "your-email@example.com"
  smtp_pass: "your 16-character code if gmail, otherwise regular password"
  ```

</details>

<details>

<summary><h2>Set up Proxmox with Ansible</h2></summary>

- If your public key is anything other than `~/.ssh/id_ed25519.pub`, change it in `./ansible/setup-proxmox.yml`.
- Run `ansible-playbook -i ansible/inventory.ini ansible/setup-proxmox.yml -u root --ask-vault-pass`, this will:
  - Configures apt so that it can install non-enterprise packages 
  - Install `sudo`.
  - Create an `admin` user with full `sudo` permissions, that can log-in via SSH with a private key.
  - Harden SSH access so that root and password logins become not permitted.
  - Make SSH happen in port `17031` instead of `22`.
  - Enables auto-updates, notifying you via email every time one happens.
  - Create a `terraform` user with partial `sudo` permissions and SSH access via your public key: `/path/to/your/public/.ssh/key`.
  - Create a Proxmox `terraform` user with an API token with limited permissions.
  - Make sure that Proxmox "local" storage can have items of type "import" and "snippets".
- After running this playbook:
  - It will show you the API token that was created for the Terraform Proxmox user. Save this in Bitwarden.
  - Add the following to your `~/.ssh/config` file, this will be used by the `./ansible/setup-proxmox.yml` playbook:
    ```
    Host proxmox
      Hostname 1.2.3.4
      User admin
      IdentityFile /path/to/your/private/.ssh/key
      Port 17031
    ```
  - You should now be able to run this playbook as many times as you want (without the `-u root` argument, as that won't work anymore).

</details>

<details>

<summary><h2>Create Proxmox Firewall</h2></summary>

- cd `terraform/global`
- Create a file called `terraform.tfvars`. It should look like this:
```
node            = "<node-name>"
endpoint        = "https://1.2.3.4:8006/"
api_token       = "terraform@pve!provider=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ssh_address     = "1.2.3.4"
ssh_port        = 17031
ssh_private_key = "/path/to/your/private/.ssh/key"
```
- Run `terraform init`.
- Run `terraform apply`.
- Change directory back to the root of this repo: `cd ../../`.
- Make sure that `./ansible/group_vars/all/all.yml` has the following variable:
  ```
  proxmox_node_name: "<node-name>"
  ```
- Run ansible playbook to enable firewall: `ansible-playbook -i ansible/inventory.ini ansible/activate-firewall.yml --ask-vault-pass`.
- At this point, only ports `17031` and `8006` will be allowed incoming traffic.

</details>

<details>

<summary><h2>Create a new Plex LXC container</h2></summary>

- `cd terraform/plex_lxc`.
- Decide on the IP address that you would want for your Plex LXC container. From now on, we will use the special value `<plex-lxc-ip>` to represent your Plex container's IP address.
- Create a file called `terraform.tfvars`. It should look like this:
```
node           = "<node-name>"
router_ip      = "10.0.0.1"
endpoint       = "https://1.2.3.4:8006/"
username       = "root@pam"
password       = "<password>"
ssh_public_key = "/path/to/your/public/.ssh/key"
ip             = "<plex-lxc-ip>"
```
- Unfortunately, Proxmox doesn't support some things in this `main.tf` file without root login, so the authentication here is just root username and password.
- Run `terraform apply`. This should create an Ubuntu LXC container mounted to `/mnt/media` on the Proxmox host.
- At this point, you will be locked out of SSH because of some firewall rules. This will be fixed with ansible.
- Get back to the root of the project: `cd ../../`.
- Update `./ansible/inventory.ini` so that it has this:
  ```
  [plex]
  plex_lxc ansible_host=<plex-lxc-ip> vmid=100
  ```
- Run the bootstrap playbook: `ansible-playbook -i ansible/inventory.ini ansible/bootstrap-plex-lxc.yml --ask-vault-pass`.
- Run the setup playbook. In your first run, you'll use root permissions to run the playbook: `ansible-playbook -i ansible/inventory.ini ansible/setup-plex-lxc.yml -e "ansible_user=root" --ask-vault-pass`.
- After this, you should be able to go to visit `http://<plex-lxc-ip>:32400` and see the Plex welcome screen.
- Also, root login with password will be disabled. You'll only be able to login as admin using `/path/to/private/key` at port `17031`.
- You can re-run the setup playbook without root permissions: `ansible-playbook -i ansible/inventory.ini ansible/setup-plex-lxc.yml --ask-vault-pass`.

</details>

<details>

<summary><h2>Create a new VM that is bound to a WireGuard VPN</h2></summary>

- `cd terraform/wireguard-vm`.
- Decide on the IP address that you would want for a VM. From now on, we will use the special value `<vm-ip>` to represent your VM's IP address.
- Create a file called `terraform.tfvars`. It should look like this:
```
node            = "<node-name>"
router_ip       = "10.0.0.1"
endpoint        = "https://1.2.3.4:8006/"
api_token       = "terraform@pve!provider=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
ssh_address     = "1.2.3.4"
ssh_port        = 17031
ssh_public_key  = "/path/to/your/public/.ssh/key"
ssh_private_key = "/path/to/your/private/.ssh/key"
vm_ip           = "<vm-ip>"
```
- The x's in `api_token` should be replaced with the API token that you received when you set up Proxmox with Ansible.
- Run `terraform init`.
- Run `terraform apply`. This should create an Ubuntu VM that can mount to `/mnt/media` on the Proxmox host.
- At this point, you should be able to ssh into the ubuntu VM: `ssh ubuntu@<vm-ip> -i /path/to/your/private/.ssh/key`.

</details>

<details>

<summary><h2>Set up a new server with SSH hardening and automated updates</h2></summary>

The instructions will make it so that only non-root private-key logins are allowed in your server. Also, it will make your server automatically get updates. You will be sent an email when updates happen. These instructions will use "vpn" as the Ansible "host" name and IP address `10.20.30.40`.

- Update `./ansible/inventory.ini` so that under the `remote_vps` group, there is an entry for your new server:
  ```
  [remote_vps]
  vpn ansible_host=10.20.30.40
  ```
- Update `./ansible/host_vars/` so that there is a new directory called `vpn`.
- Run `ansible-vault create ./ansible/host_vars/vpn/vault.yml`, using your vault password.
- In that file put the following, except use an actual password that you'll save to Bitwarden:
  ```
  vault_admin_password: "my super secret password"
  ```
- In your first run, you'll use root permissions and port 22 to run the playbook: `ansible-playbook -i ansible/inventory.ini ansible/setup-server.yml -e "ansible_user=root" -e "ansible_port=22" --ask-vault-pass --ask-pass --limit vpn`.
  - Note: this command makes it so that the `./ansible/setup-server.yml` playbook is only run for your new server (`--limit vpn`). Without this part, the playbook will be run for all hosts under the `remote_vps` group.
- After this, root login with password will be disabled. You'll only be able to login as admin using `/path/to/private/key` at port `17031`.
- You can now re-run this playbook with: `ansible-playbook -i ansible/inventory.ini ansible/setup-server.yml --ask-vault-pass --limit vpn`.

</details>
