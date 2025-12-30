terraform {
  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "0.90.0"
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


# Enable Firewall at the Datacenter level
resource "proxmox_virtual_environment_cluster_firewall" "cluster" {
  enabled      = true
  input_policy = "DROP"
}

# Define the Security Group (Reusable for any VM/LXC)
resource "proxmox_virtual_environment_cluster_firewall_security_group" "mgmt" {
  name    = "management"
  comment = "Essential Admin Access"

  rule {
    type    = "in"
    action  = "ACCEPT"
    proto   = "tcp"
    dport   = "8006"
    comment = "Proxmox UI"
  }

  rule {
    type    = "in"
    action  = "ACCEPT"
    proto   = "tcp"
    dport   = "17031"
    comment = "Custom SSH Port"
  }
}

# Apply the group to the Node (Host) itself
resource "proxmox_virtual_environment_firewall_rules" "node_rules" {
  node_name = var.node

  rule {
    security_group = proxmox_virtual_environment_cluster_firewall_security_group.mgmt.name
    iface          = "vmbr0"
  }
}
