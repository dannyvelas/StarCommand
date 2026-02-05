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
  endpoint = var.endpoint
  username = var.username
  password = var.password
  insecure = true
}

data "local_file" "ssh_public_key" {
  filename = var.ssh_public_key
}

resource "proxmox_virtual_environment_download_file" "ubuntu_lxc_template" {
  content_type = "vztmpl"
  datastore_id = "local"
  node_name    = var.node
  url          = "http://download.proxmox.com/images/system/ubuntu-24.04-standard_24.04-2_amd64.tar.zst"
  file_name    = "ubuntu-24.04-standard_24.04-2_amd64.tar.zst"
}

resource "proxmox_virtual_environment_container" "plex_lxc" {
  description  = "Managed by Terraform"
  tags         = ["terraform", "ubuntu"]
  node_name    = var.node
  vm_id        = 100
  unprivileged = true

  cpu {
    cores = 2
  }

  memory {
    dedicated = 1024
  }

  disk {
    datastore_id = "local-lvm"
    size         = 20
  }

  network_interface {
    name     = "eth0"
    bridge   = "vmbr0"
    firewall = true
  }

  # this initialization block works because:
  # the `proxmox_virtual_environment_download_file` resource created a VM template and
  # stored it in the proxmox "local" storage. this template is configured so that when
  # a VM using this template boots for the first time, there will be a special
  # cloud-init drive in it. this allows us to pass data into the VM like SSH keys, hostname,
  # network config, etc
  initialization {
    hostname = "terraform-provider-proxmox-ubuntu-container"

    ip_config {
      ipv4 {
        address = "${var.ip}/24"
        gateway = var.gateway_address
      }
    }

    user_account {
      keys = [
        trimspace(data.local_file.ssh_public_key.content)
      ]
    }

    dns {
      servers = ["1.1.1.1"]
    }
  }

  operating_system {
    template_file_id = proxmox_virtual_environment_download_file.ubuntu_lxc_template.id
    type             = "ubuntu"
  }

  mount_point {
    # bind mount for media
    volume = "/mnt/media"
    path   = "/mnt/media"
  }

  mount_point {
    # bind mount for plex metadata
    volume = "/mnt/media/plex-config"
    path   = "/var/lib/plexmediaserver/Library/Application Support/Plex Media Server"
  }
}

# enable firewall options at container level which is default deny
resource "proxmox_virtual_environment_firewall_options" "plex_fw_options" {
  node_name    = var.node
  container_id = proxmox_virtual_environment_container.plex_lxc.vm_id

  enabled       = true
  input_policy  = "DROP"   # Everything not allowed is blocked
  output_policy = "ACCEPT" # Allow the container to reach out for updates
}

# enable firewall rule which associates the "plex" security group with this plex_lxc
# the "plex" security group should have been created prior (it is defined in terraform/global)
resource "proxmox_virtual_environment_firewall_rules" "plex_rules" {
  node_name    = var.node
  container_id = proxmox_virtual_environment_container.plex_lxc.vm_id

  rule {
    security_group = "plex"
    iface          = "net0"
  }

  rule {
    security_group = "guest_mgmt"
    iface          = "net0"
  }
}
