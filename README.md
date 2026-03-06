# Star Command

> **Work in progress.** Not everything described here is fully implemented yet. See [docs/progress.md](docs/progress.md) for a detailed breakdown of what's done and what's still being built.

Fork this repo to get a fully declarative, version-controlled infrastructure for one or more Debian servers.

This repo also comes with a CLI which:
- Allows you to have one source-of-truth for all your infrastructure configuration, so that you don't have to manually copy-paste values between Ansible and Terraform config files
- Abstracts Ansible and Terraform steps, so that you can run a single command to provision your entire infrastructure without having to worry about the order of operations or the specific Ansible/Terraform commands to run for each tool

## Architecture

```
Private Network (192.168.1.0/24)
  |
  +-- Gateway
  |     \-- UDP 51820 forwarded -> Host 01
  |         TCP 443 forwarded -> Host 01
  |
  +-- Host 01 (192.168.1.10)
  |     +-- WireGuard (wg0, VPN endpoint)
  |     +-- UFW (NAT, port forwarding, egress blocking)
  |     +-- OVN (overlay networking between hosts)
  |     +-- k3s server (workload scheduler)
  |     \-- VM (192.168.122.50) <- private subnet, not on physical network
  |           +-- k3s agent
  |           +-- Traefik (reverse proxy, subdomain routing)
  |           \-- workload containers (read-only root)
  |
  +-- Host 02 (192.168.1.11)
  |     +-- UFW (NAT, port forwarding, egress blocking)
  |     +-- OVN (overlay networking between hosts)
  |     +-- k3s server
  |     \-- VM (192.168.123.50) <- private subnet, not on physical network
  |           +-- k3s agent
  |           +-- Traefik (reverse proxy, subdomain routing)
  |           \-- workload containers (read-only root)
  |
  +-- ...additional hosts...
```

**Two-layer isolation**: Trusted infrastructure (WireGuard, Incus, OVN) runs on the host. Application workloads run inside VMs behind NAT. A container escape lands in the VM kernel, not the host.

### Security layers

1. **OS hardening** — UFW firewall, SSH key-only auth, unattended security updates
2. **NAT networking** — VMs on private subnets, isolated from the physical network
3. **Egress blocking** — VMs cannot initiate connections to other network hosts
4. **VM kernel isolation** — Container breakouts are contained by the VM boundary
5. **OVN overlay** — Encrypted east-west traffic between hosts, isolated from the physical network
6. **Reverse proxy** — Single ingress point with TLS, security headers, rate limiting
7. **Read-only containers** — Immutable root filesystem, writable only for data volumes


## Prerequisites

### Hardware

- One or more Debian 12 servers, each with:
  - 8 GB+ RAM
  - Network access

### Software (on your workstation)

- [Go](https://go.dev/) 1.21+ (to build the CLI)
- [Terraform](https://www.terraform.io/) (latest stable)
- [Ansible](https://docs.ansible.com/) (latest stable)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (for deploying and managing services)
- [WireGuard client](https://www.wireguard.com/install/) (for VPN access)
- SSH client with key-based auth configured to all servers

### Network preparation

Forward UDP port **51820** and TCP port **443** on your gateway to the server that will act as the WireGuard endpoint and ingress point.

### Remote access (optional)

If you want to access services from outside the network perimeter without VPN, set up a DDNS hostname before provisioning:

1. Create an account with a DDNS provider (e.g., DuckDNS, No-IP)
2. Register a hostname (e.g., `infra.example.com`)
3. Note your login credentials — you'll add them to `stc.yml`

During provisioning, the system automatically installs ddclient to keep the hostname pointed at your public IP. If you skip this step, everything still works on your local network and over VPN.

## Setup

### 1. Fork and clone

Fork this repo on GitHub, then clone your fork:

```bash
git clone https://github.com/ <your-username >/starcommand
cd starcommand
make
```

Your fork is your infrastructure repo — commit your config, playbook changes, and service manifests to it.

### 2. Configure

Copy the example config and fill in your values:

```bash
cp stc.example.yml stc.yml
vim stc.yml
```

Each field is explained inline in [`stc.example.yml`](stc.example.yml). Non-sensitive values live in `stc.yml`. Sensitive values (e.g. `admin_password`) are never stored by `stc` — it will prompt for them interactively at runtime when needed. For automation (e.g. CI), you can supply them as environment variables instead:

```bash
export STC_ADMIN_PASSWORD=...
```

`stc` generates all tool-specific configs (Ansible inventory, Terraform vars) into `.generated/`. You never edit those files directly.

## Usage

### Provision hosts

Apply the desired state to all hosts in `stc.yml`. Ansible's idempotency means this is safe to run at any time — already-provisioned hosts are verified quickly, new hosts are fully provisioned:

```bash
stc setup                                        # apply desired state to all hosts
stc setup --host <host>                          # limit to one host
stc setup --host <host1> --host <host2>          # limit to a subset of hosts
```

After the first `stc setup` completes, update your network's DHCP configuration to distribute the cluster as the DNS server. This is a one-time manual step — after this, every host on your network resolves service subdomains automatically.

For a deeper look at how `stc setup` works internally, see [docs/internals.md](docs/internals.md).

### Add VPN clients

```bash
stc wg add alice-laptop
stc wg add alice-phone
```

Client configs are saved to `.generated/vpn-clients/`. Import them into the WireGuard app on each device, then delete the `.conf` files from your workstation — they contain the client's private key and preshared key. The `.generated/` directory is gitignored and the files are created with `0600` permissions, but they should be treated as sensitive and not kept around longer than needed.

### Deploy services

This platform is application-agnostic. After infrastructure is set up, deploy any service using standard Kubernetes manifests and kubectl. See `services/` for full working examples.

k3s handles scheduling and restarts. Traefik picks up new services automatically via Ingress resources and routes by subdomain.

```bash
kubectl apply -f services/grafana.yml
kubectl apply -f services/golinks.yml
```

### Verify

```bash
stc status # cluster overview: hosts, services, VPN, k3s
```

### Tear down

```bash
stc teardown # destroy all VMs via Terraform
```

## CLI reference

```
stc <command> [options]

Commands:
  setup [--host <h>]...                  Apply desired state to all hosts, or limit to the ones given
  wg add <name>                          Add a WireGuard client (registers peer server-side, generates client config)
  status                                 Show cluster status (hosts, services, VPN, k3s)
  teardown                               Tear down all VMs
  version                                Print version

Low-level commands:
  inventory generate                       Generate the Ansible inventory file for all hosts
  ansible bootstrap-server [--host <h>]... Run the bootstrap-server playbook against all hosts/VMs, or limit to the ones given
  ansible setup-host [--host <h>]...       Run the setup-host playbook against all hosts, or limit to the ones given
  ansible setup-vm [--host <h>]...         Run the setup-vm playbook against all VMs, or limit to the ones given
  ssh add <host>                           Add a host to ~/.ssh/config
  terraform apply                          Apply the Terraform project
```

`stc` wraps these low-level commands because they require config resolution — secret fetching, inventory generation, and var merging — that would otherwise need to be done manually.

## Project structure

```
stc.yml                      # infrastructure configuration (your values)
stc.example.yml              # example config with field descriptions
services/                    # service manifests (one per app)
ansible/
  playbooks/                 # ansible playbooks
  roles/                     # ansible roles
terraform/                   # incus VM lifecycle
cmd/                         # Go CLI source
.generated/                  # auto-generated configs (gitignored)
```

## Pulling upstream changes

To get improvements from this repo into your fork:

```bash
git fetch upstream
git merge upstream/main
```

If you've modified playbooks or Terraform files, standard git conflict resolution applies.

## Scaling to new hardware

When you add or migrate servers:

1. Update `stc.yml` with the new host IPs and storage paths
2. Run `stc setup` — already-provisioned hosts are skipped, new ones are fully provisioned
3. k3s automatically joins the new node to the cluster
4. OVN extends the overlay network to the new host
5. Deploy services with `kubectl` — no changes to manifests needed, k3s schedules across the cluster

## Why Star Command?

Debian names its releases after characters from Toy Story — Woody, Buzz, Jessie, Buster, Bookworm. Star Command is where Buzz Lightyear comes from. It seemed like the right name for a command-line tool built for Debian.

## Tech stack

| Component    | Tool                 | Why                                               |
|--------------|----------------------|---------------------------------------------------|
| CLI          | Go                   | Single binary, orchestrates Ansible and Terraform |
| Provisioning | Ansible              | Agentless, SSH-based, idempotent                  |
| VM lifecycle | Terraform + Incus    | Declarative, reproducible                         |
| Scheduling   | k3s                  | Kubernetes-native, rolling updates, auto-restart  |
| Networking   | OVN                  | Overlay networking, encrypted east-west traffic   |
| Firewall     | UFW                  | Simple, readable firewall rules                   |
| VPN          | WireGuard            | Kernel-level, no third-party trust                |
| Proxy        | Traefik              | Auto-discovery, TLS, subdomain routing            |
| Monitoring   | Grafana + Prometheus | Dashboards, alerting, service health              |
| Go links     | golinks              | Internal short URLs for quick service access      |
| OS           | Debian 12            | Stable, security updates, KVM support             |
