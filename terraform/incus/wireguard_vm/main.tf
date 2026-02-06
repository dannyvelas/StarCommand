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

resource "incus_profile" "wireguard" {
  name = "wireguard"

  # Expose UDP 51820 for VPN Handshake
  device {
    name = "vpn_handshake"
    type = "proxy"
    properties = {
      listen  = "udp:0.0.0.0:51820"
      connect = "udp:127.0.0.1:51820"
    }
  }

  # Expose port 17031 via proxy device (Host Port -> Container Port 17031)
  device {
    name = "ssh_custom"
    type = "proxy"
    properties = {
      listen  = "tcp:0.0.0.0:17032"
      connect = "tcp:127.0.0.1:17031"
    }
  }
}

resource "incus_instance" "wireguard_vm" {
  name      = "wireguard-vm"
  image     = "images:ubuntu/24.04"
  type      = "virtual-machine"
  ephemeral = false

  profiles = ["basic", "management", incus_profile.wireguard.name]

  config = {
    "limits.cpu"    = "2"
    "limits.memory" = "4GiB"
    # Ensure guest agent channel is available
    "user.user-data" = <<-EOT
      #cloud-config
      packages:
        - qemu-guest-agent
        - net-tools
        - curl
      runcmd:
        - systemctl enable --now qemu-guest-agent
    EOT
  }

  # Override eth0 from 'basic' profile to add static IP
  device {
    name = "eth0"
    type = "nic"
    properties = {
      network        = "incusbr0"
      "ipv4.address" = var.ip
    }
  }

  # Virtio-FS mount for Media (replicating original config)
  device {
    name = "media_mount"
    type = "disk"
    properties = {
      source = "/mnt/media"
      path   = "/mnt/media"
    }
  }
}
