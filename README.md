# Homelab infra and playbooks

## Prerequisites
* A server (Ubuntu/Debian) connected to Ethernet.
* A computer you can use to ssh into the server (Controller).
* [Terraform](https://developer.hashicorp.com/terraform/install) installed.
* [Ansible](https://formulae.brew.sh/formula/ansible) installed.
* [Go](https://go.dev/) installed.
* [Task](https://taskfile.dev/) installed.
* A C toolchain on that computer. You can install this on macOS with `xcode-select --install`.
* A [Tailscale](https://login.tailscale.com/start) account.
* [Bitwarden Secrets Manager](https://bitwarden.com/help/secrets-manager-quick-start/).

## Deployment Reference
| Component     | Command                  | Action performed                                                    |
|---------------|--------------------------|---------------------------------------------------------------------|
| Incus Server  | task setup:incus         | Bootstraps host (Ansible) & creates global profiles (Terraform).    |
| Plex          | task setup:plex          | Deploys container (Terraform) & configures mounts/service (Ansible) |
| WireGuard     | task setup:wireguard_vm  | Deploys VM (Terraform) & exposes UDP+SSH ports.                     |
| Remote server | task setup:remote_server | Configures a newly-rented service with hardened security (Ansible)  |

## Configs
* Each playbook requires a unique set of configs.
* You can identify the configs needed by running `task <playbookname>:check`.
* Add your configs in any combination of: `./config/all.yml`, `./config/incus.yml`, the environment (or `.env`), and a Bitwarden secrets vault.
