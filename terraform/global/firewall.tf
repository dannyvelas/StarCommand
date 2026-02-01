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
    private_key = file(var.ssh_private_key_path)
    node {
      name    = var.node
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

# Create security group specifically for managing the proxmox host
resource "proxmox_virtual_environment_cluster_firewall_security_group" "host_mgmt" {
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

# Apply the host management security group to the node itself
resource "proxmox_virtual_environment_firewall_rules" "node_rules" {
  node_name = var.node

  rule {
    security_group = proxmox_virtual_environment_cluster_firewall_security_group.host_mgmt.name
    iface          = "vmbr0"
  }
}

# create generic security group for VMs/LXCs. this will allow us to use port 17031 for ssh
resource "proxmox_virtual_environment_cluster_firewall_security_group" "guest_mgmt" {
  name    = "guest_mgmt"
  comment = "Management access for all containers"

  rule {
    type    = "in"
    action  = "ACCEPT"
    proto   = "tcp"
    dport   = "17031"
    comment = "SSH to Container"
  }
}

# create security group for plex LXC
resource "proxmox_virtual_environment_cluster_firewall_security_group" "plex_lxc" {
  name    = "plex"
  comment = "Plex Media Server Ports"

  rule {
    type    = "in"
    action  = "ACCEPT"
    proto   = "tcp"
    dport   = "32400"
    comment = "Plex Web UI / Media Stream"
  }
}

# security group for wireguard_vm. establishes VPN tunnel
resource "proxmox_virtual_environment_cluster_firewall_security_group" "vpn_outbound" {
  name    = "vpn-handshake"
  comment = "Allow VM to reach VPN Providers"

  rule {
    type    = "out" # Notice this is OUT
    action  = "ACCEPT"
    proto   = "udp"
    dport   = "51820" # Standard WG port, change if your provider uses different
    comment = "Allow encrypted tunnel setup"
  }
}
