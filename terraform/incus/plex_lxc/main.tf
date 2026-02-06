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

resource "incus_profile" "plex" {
  name = "plex"

  # Expose port 32400 via proxy device (Host Port -> Container Port)
  device {
    name = "plex_web"
    type = "proxy"
    properties = {
      listen  = "tcp:0.0.0.0:32400"
      connect = "tcp:127.0.0.1:32400"
    }
  }

  # Expose port 17031 via proxy device (Host Port -> Container Port 17031)
  device {
    name = "ssh_custom"
    type = "proxy"
    properties = {
      listen  = "tcp:0.0.0.0:17031"
      connect = "tcp:127.0.0.1:17031"
    }
  }
}

resource "incus_instance" "plex_lxc" {
  name      = "plex-lxc"
  image     = "images:ubuntu/24.04"
  type      = "container"
  ephemeral = false

  profiles = ["basic", "management", incus_profile.plex.name]

  # Override eth0 from 'basic' profile to add static IP
  device {
    name = "eth0"
    type = "nic"
    properties = {
      network        = "incusbr0"
      "ipv4.address" = var.ip
    }
  }

  # Bind Mount: Media
  device {
    name = "media_mount"
    type = "disk"
    properties = {
      source = "/mnt/media"
      path   = "/mnt/media"
    }
  }

  # Bind Mount: Plex Configuration
  device {
    name = "plex_config"
    type = "disk"
    properties = {
      source = "/mnt/media/plex-config"
      path   = "/var/lib/plexmediaserver/Library/Application Support/Plex Media Server"
    }
  }

  # Ensure Cloud Init SSH key is present (inherited from management, 
  # but we define ssh_public_key_path variable to facilitate if we need customization)
}
