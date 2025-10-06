#!/bin/bash

# Check if required arguments are provided
if [ $# -ne 3 ]; then
  echo "Usage: $0 <ip-addr> <path-to-public_key> <path-to-private-key>"
  echo "Example: $0 192.168.1.100 ~/.ssh/id_ed25519.pub ~/.ssh/id_ed25519"
  exit 1
fi

IP=$1
PUBLIC_KEY_PATH=$2
PRIVATE_KEY_PATH=$2

# Check if public key file exists
if [ ! -f "$PUBLIC_KEY_PATH" ]; then
  echo "Error: Public key file not found at $PUBLIC_KEY_PATH"
  exit 1
fi

# Read public key
PUBLIC_KEY=$(cat "$PUBLIC_KEY_PATH")
if [ -z "$PUBLIC_KEY" ]; then
  echo "Error: Public key file is empty"
  exit 1
fi

# Prompt for password
echo -n "Enter password for user 'dannyvelasquez': "
read -s PASSWORD
echo

# Generate a random port between 1024 and 65535
PORT=$((RANDOM % 64512 + 1024))

# SSH commands to be executed on the remote server
ssh root@"$IP" bash -c "'
    # stop if there is an error
    set -e
    
    # Add Tailscales GPG key and repository
    mkdir -p --mode=0755 /usr/share/keyrings
    curl -fsSL https://pkgs.tailscale.com/stable/debian/trixie.noarmor.gpg | tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
    curl -fsSL https://pkgs.tailscale.com/stable/debian/trixie.tailscale-keyring.list | tee /etc/apt/sources.list.d/tailscale.list

    # Install sudo and tailscale
    apt update && apt upgrade && apt install -y sudo
    apt-get install tailscale

    # Create dannyvelasquez user
    useradd -G sudo -m dannyvelasquez -s /bin/bash
    echo \"dannyvelasquez:$PASSWORD\" | chpasswd

    # Setup SSH for dannyvelasquez
    mkdir -p /home/dannyvelasquez/.ssh
    chmod 700 /home/dannyvelasquez/.ssh
    chown dannyvelasquez:dannyvelasquez /home/dannyvelasquez/.ssh
    echo \"$PUBLIC_KEY\" > /home/dannyvelasquez/.ssh/authorized_keys
    chmod 600 /home/dannyvelasquez/.ssh/authorized_keys
    chown dannyvelasquez:dannyvelasquez /home/dannyvelasquez/.ssh/authorized_keys

    # Create terraform user and setup Proxmox access
    pveum user add terraform@pve
    pveum role add Terraform -privs \"Realm.AllocateUser, VM.PowerMgmt, VM.GuestAgent.Unrestricted, Sys.Console, Sys.Audit, Sys.AccessNetwork, VM.Config.Cloudinit, VM.Replicate, Pool.Allocate, SDN.Audit, Realm.Allocate, SDN.Use, Mapping.Modify, VM.Config.Memory, VM.GuestAgent.FileSystemMgmt, VM.Allocate, SDN.Allocate, VM.Console, VM.Clone, VM.Backup, Datastore.AllocateTemplate, VM.Snapshot, VM.Config.Network, Sys.Incoming, Sys.Modify, VM.Snapshot.Rollback, VM.Config.Disk, Datastore.Allocate, VM.Config.CPU, VM.Config.CDROM, Group.Allocate, Datastore.Audit, VM.Migrate, VM.GuestAgent.FileWrite, Mapping.Use, Datastore.AllocateSpace, Sys.Syslog, VM.Config.Options, Pool.Audit, User.Modify, VM.Config.HWType, VM.Audit, Sys.PowerMgmt, VM.GuestAgent.Audit, Mapping.Audit, VM.GuestAgent.FileRead, Permissions.Modify\"
    pveum aclmod / -user terraform@pve -role Terraform

    # Create API token (this will output to console for manual saving)
    echo \"Creating Terraform API token - SAVE THIS TOKEN:\"
    pveum user token add terraform@pve provider --privsep=0

    # Create local terraform user
    useradd -m terraform

    # Setup sudo permissions for terraform
    echo \"terraform ALL=(root) NOPASSWD: /sbin/pvesm\" > /etc/sudoers.d/terraform
    echo \"terraform ALL=(root) NOPASSWD: /sbin/qm\" >> /etc/sudoers.d/terraform
    echo \"terraform ALL=(root) NOPASSWD: /usr/bin/tee /var/lib/vz/*\" >> /etc/sudoers.d/terraform
    chmod 440 /etc/sudoers.d/terraform

    # Setup SSH for terraform user
    mkdir -p /home/terraform/.ssh
    chmod 700 /home/terraform/.ssh
    chown terraform:terraform /home/terraform/.ssh
    echo \"$PUBLIC_KEY\" > /home/terraform/.ssh/authorized_keys
    chmod 600 /home/terraform/.ssh/authorized_keys
    chown terraform:terraform /home/terraform/.ssh/authorized_keys

    # Harden SSH login
    cat > /etc/ssh/sshd_config.d/10-hardening.conf <<EOF
# SSH Hardening Configuration
Port $PORT
PasswordAuthentication no
PermitRootLogin no
EOF

    # Verify config
    sshd -t || {
        echo \"Invalid sshd_config, removing changes...\"
        rm /etc/ssh/sshd_config.d/10-hardening.conf
        exit 1
    }

    systemctl restart sshd
    echo \"sshd_config updated. New SSH port: $PORT\"

    # Create media directory
    mkdir -p /mnt/media

    # Start Tailscale!
    sudo tailscale up

    echo \"Setup completed successfully!\"
'"

echo "Verifying SSH access for dannyvelasquez user..."
if ssh -o BatchMode=yes -o StrictHostKeyChecking=no -i "$PRIVATE_KEY_PATH" -p "$PORT" "dannyvelasquez@$IP" "echo 'SSH access successful for dannyvelasquez'"; then
  echo "dannyvelasquez SSH access verified"
else
  echo "Error: Unable to verify SSH access for dannyvelasquez"
fi

echo "Verifying SSH access and privileges for terraform user..."
if ssh -o BatchMode=yes -o StrictHostKeyChecking=no -i "$PRIVATE_KEY_PATH" -p "$PORT" "terraform@$HOST" "sudo pvesm apiinfo"; then
  echo "terraform SSH access and privileges verified"
else
  echo "Error: Unable to verify SSH access or privileges for terraform user"
fi
