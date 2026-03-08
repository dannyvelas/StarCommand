variable "incus_remote" {
  description = "The name of the Incus remote to target (as configured in your local client)"
  type        = string
  default     = "local"
}

variable "ssh_public_key_path" {
  description = "Path to the SSH public key to inject into instances"
  type        = string
  default     = "~/.ssh/id_ed25519.pub"
}

variable "network_bridge" {
  description = "The name of the network bridge on the host"
  type        = string
  default     = "incusbr0"
}

variable "storage_pool_name" {
  description = "Name of the Incus storage pool"
  type        = string
  default     = "default"
}

variable "storage_pool_driver" {
  description = "Storage driver for the Incus storage pool (e.g. dir, btrfs, zfs)"
  type        = string
  default     = "dir"
}
