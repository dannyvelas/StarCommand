terraform {
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "0.83.2"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5.3"
    }
  }
  required_version = "~> 1.13.3"
}

provider "proxmox" {
  endpoint  = var.endpoint
  api_token = var.api_token
  insecure  = true
  ssh {
    agent       = false
    username    = "terraform"
    private_key = file(var.ssh_private_key)
    node {
      name    = var.node
      address = var.ssh_address
      port    = var.ssh_port
    }
  }
}

data "local_file" "ssh_public_key" {
  filename = var.ssh_public_key
}

resource "proxmox_virtual_environment_file" "user_data_cloud_config" {
  content_type = "snippets"
  datastore_id = "local"
  node_name    = var.node

  source_raw {
    data = <<-EOF
    #cloud-config
    hostname: wireguard-vm
    package_update: true
    package_upgrade: true
    users:
      - default
      - name: admin
        primary_group: admin
        groups:
          - sudo
        shell: /bin/bash
        ssh_authorized_keys:
          - ${trimspace(data.local_file.ssh_public_key.content)}
        sudo: ALL=(ALL) NOPASSWD:ALL
    packages:
      - qemu-guest-agent
      - net-tools
      - curl
    # switch port 22 to be 17031
    write_files:
      - path: /etc/systemd/system/ssh.socket.d/listen.conf
        content: |
          [Socket]
          ListenStream=
          ListenStream=0.0.0.0:17031
          ListenStream=[::]:17031
    runcmd:
      # enable qemu-guest-agent
      - systemctl enable qemu-guest-agent
      - systemctl start qemu-guest-agent
      # apply changes to the socket
      - systemctl daemon-reload
      - systemctl stop ssh.socket
      - systemctl start ssh.socket
      - systemctl restart ssh
      # mark done
      - echo "done" > /tmp/cloud-config.done
    EOF

    file_name = "user-data-cloud-config.yaml"
  }
}

resource "proxmox_virtual_environment_vm" "wireguard_vm" {
  name        = "wireguard-vm"
  description = "Managed by Terraform"
  tags        = ["terraform", "ubuntu"]
  node_name   = var.node
  vm_id       = 101

  # enable 'Qemu guest agent' so that proxmox can directly communicate with this VM
  # to get its IP address and display it on the console. also so that it can cleanly and
  # gracefully shut it down without having to just send an ACPI signal
  # for us to enable this though, we have to make sure that we are running the qemu-guest-agent
  # in our cloud-init
  agent {
    enabled = true
  }

  cpu {
    cores = 2
  }

  memory {
    dedicated = 4096
  }

  disk {
    datastore_id = "local-lvm"
    # here, we are telling our VM to scaffold itself with the template we created with the `proxmox_virtual_environment_download_file` resource
    import_from = proxmox_virtual_environment_download_file.ubuntu_cloud_image.id
    interface   = "virtio0"
    iothread    = true
    discard     = "on"
    size        = 20
  }

  network_device {
    bridge   = "vmbr0"
    model    = "virtio"
    firewall = true
  }

  # this initialization block works because:
  # the `proxmox_virtual_environment_download_file` resource created a VM template and
  # stored it in the proxmox "local" storage. this template is configured so that when
  # a VM using this template boots for the first time, there will be a special
  # cloud-init drive in it. this allows us to pass data into the VM like SSH keys, hostname,
  # network config, etc
  initialization {
    datastore_id = "local-lvm"

    ip_config {
      ipv4 {
        address = "${var.ip}/24"
        gateway = var.gateway_address
      }
    }

    # here, we pass in ssh-keys. we also pass in other setup logic that isn't natively
    # supported by the bpg/proxmox provider.
    user_data_file_id = proxmox_virtual_environment_file.user_data_cloud_config.id
  }

  # there are three steps to sharing a host directory with a VM. this is step 2:
  # adding a hardware component to our VM called "media_mount". "media_mount" is
  # a proxmox directory mapping that we created as a resource below.
  # you can think of this as physically plugging-in the aforementioned USB drive to our VM
  virtiofs {
    mapping   = "media_mount"
    cache     = "always"
    direct_io = true
  }
}

# here, we are telling bpg/proxmox to create a VM template that is used by our ubuntu VM.
# we are specifying where that template should be stored.
# we are also specifying the specific cloud image that should be used in this VM template
resource "proxmox_virtual_environment_download_file" "ubuntu_cloud_image" {
  content_type = "import"
  datastore_id = "local"
  node_name    = var.node
  url          = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
  # need to rename the file to *.qcow2 to indicate the actual file format for import
  file_name = "noble-server-cloudimg-amd64.qcow2"
}

# there are three steps to sharing a host directory with a VM. this is step 1:
# create a proxmox directory mapping called "media_mount" which will be hosted at the
# special directory "/mnt/media"
# we can think of this as creating a physical USB drive. when connecting this USB drive
# to a computer, that computer will see a new folder. that new folder will be a symlink
# to a folder on another computer
resource "proxmox_virtual_environment_hardware_mapping_dir" "media_mount" {
  name    = "media_mount"
  comment = "media bind mount"
  map = [
    {
      node = var.node
      path = "/mnt/media"
    },
  ]
}

# firewall killswitch: stops the VM from talking to the internet directly if the internal rules aren't met
resource "proxmox_virtual_environment_firewall_options" "wg_fw_options" {
  node_name = var.node
  vm_id     = proxmox_virtual_environment_vm.wireguard_vm.vm_id

  enabled      = true
  input_policy = "DROP"

  # THE KILLSWITCH: Stop the VM from talking to your LAN/Internet directly if the internal rules aren't met.
  output_policy = "DROP"
}

resource "proxmox_virtual_environment_firewall_rules" "wg_rules" {
  node_name = var.node
  vm_id     = proxmox_virtual_environment_vm.wireguard_vm.vm_id

  # 1. Allow Management (SSH)
  rule {
    security_group = "guest_mgmt"
    iface          = "net0"
  }

  # 2. Allow the Handshake (The only way OUT to the internet)
  rule {
    security_group = "vpn-handshake"
    iface          = "net0"
  }

  # 3. Allow DNS (Usually needed to resolve the VPN provider's endpoint)
  rule {
    type   = "out"
    action = "ACCEPT"
    proto  = "udp"
    dport  = "53"
    iface  = "net0"
  }
}
