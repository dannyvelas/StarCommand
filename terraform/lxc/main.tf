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
      name    = "proxmox"
      address = var.ssh_address
      port    = var.ssh_port
    }
  }
}

data "local_file" "ssh_public_key" {
  filename = var.ssh_public_key
}

resource "proxmox_virtual_environment_download_file" "ubuntu_lxc_template" {
  content_type = "vztmpl"
  datastore_id = "local"
  node_name    = var.node
  url          = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64-root.tar.xz"
  file_name    = "ubuntu-24.04-lxc-template.tar.xz"
}

resource "proxmox_virtual_environment_container" "ubuntu_container" {
  description = "Managed by Terraform"
  tags        = ["terraform", "ubuntu"]
  node_name   = var.node

  cpu {
    cores = 2
  }

  memory {
    dedicated = 4096
  }

  disk {
    datastore_id = "local-lvm"
    size         = 20
  }

  network_interface {
    name   = "eth0"
    bridge = "vmbr0"
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
        gateway = var.router_ip
      }
    }

    user_account {
      keys = [
        trimspace(data.local_file.ssh_public_key.content)
      ]
    }
  }

  operating_system {
    template_file_id = proxmox_virtual_environment_download_file.ubuntu_lxc_template.id
    type             = "ubuntu"
  }

  mount_point {
    # bind mount
    volume = "/mnt/media"
    path   = "/mnt/media"
  }
}
