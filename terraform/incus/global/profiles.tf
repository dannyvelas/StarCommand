terraform {
  required_providers {
    incus = {
      source  = "lxc/incus"
      version = ">= 1.0.0"
    }
  }
  required_version = "~> 1.13.3"
}

provider "incus" {
  remote {
    name = var.incus_remote
  }
}

locals {
  ssh_key = trimspace(file(var.ssh_public_key_path))
}

resource "incus_storage_pool" "default" {
  name   = var.storage_pool_name
  driver = var.storage_pool_driver

  config = {
    source = "/var/lib/incus/storage-pools/${var.storage_pool_name}"
  }
}

resource "incus_network" "bridge" {
  name = var.network_bridge

  config = {
    "ipv4.address" = "10.0.100.1/24"
    "ipv4.nat"     = "true"
    "ipv6.address" = "none"
  }
}

resource "incus_profile" "basic" {
  name        = "basic"
  description = "Basic networking and root disk configuration"

  device {
    name = "eth0"
    type = "nic"
    properties = {
      network = incus_network.bridge.name
      name    = "eth0"
    }
  }

  device {
    name = "root"
    type = "disk"
    properties = {
      path = "/"
      pool = incus_storage_pool.default.name
    }
  }
}

resource "incus_profile" "management" {
  name        = "management"
  description = "Management access (SSH keys)"

  config = {
    "user.user-data" = <<-EOT
      #cloud-config
      ssh_authorized_keys:
        - ${local.ssh_key}
      packages:
        - curl
        - git
        - vim
        - htop
    EOT
  }
}
