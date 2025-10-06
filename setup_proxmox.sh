#!/bin/bash

# Check if required arguments are provided
if [ $# -ne 2 ]; then
  echo "Usage: $0 <ip-addr> <path_to_public_key>"
  echo "Example: $0 192.168.1.100 ~/.ssh/id_ed25519.pub"
  exit 1
fi

IP=$1
PUBLIC_KEY_PATH=$2

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
    set  -e

    # Install sudo
    apt update && apt install -y sudo

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
    echo 'terraform ALL=(root) NOPASSWD: /sbin/pvesm' > /etc/sudoers.d/terraform
    echo 'terraform ALL=(root) NOPASSWD: /sbin/qm' >> /etc/sudoers.d/terraform
    echo 'terraform ALL=(root) NOPASSWD: /usr/bin/tee /var/lib/vz/*' >> /etc/sudoers.d/terraform
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

    echo 'Setup completed successfully!'
'"
