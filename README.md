# Bare-Metal Infrastructure Automation

Infrastructure as Code for self-hosted environments. A single Go CLI (`iac`) reads your configuration and orchestrates Terraform and Ansible to provision one or more servers with a hypervisor, a WireGuard VPN, a Traefik reverse proxy, OVN networking, and a k3s cluster. Engineers deploy any containerized service with `kubectl` — no application-specific code in the platform itself.

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

**Two-layer isolation**: Trusted infrastructure (WireGuard, KVM, OVN) runs on the host. Application workloads run inside VMs behind NAT. A container escape lands in the VM kernel, not the host.

### Security layers

1. **OS hardening** — UFW firewall, SSH key-only auth, unattended security updates
2. **NAT networking** — VMs on private subnets, isolated from the physical network
3. **Egress blocking** — VMs cannot initiate connections to other network hosts
4. **VM kernel isolation** — Container breakouts are contained by the VM boundary
5. **OVN overlay** — Encrypted east-west traffic between hosts, isolated from the physical network
6. **Reverse proxy** — Single ingress point with TLS, security headers, rate limiting
7. **Read-only containers** — Immutable root filesystem, writable only for data volumes

### Monitoring and alerting

- **Grafana dashboard** — real-time metrics for all deployed services (resource usage, uptime, response times)
- **Alerting** — automatic notifications when any service goes down or degrades (email, Slack, PagerDuty)

### Go links

Internal short URLs for quick access to services and dashboards:

| Link | Destination |
|------|-------------|
| `go/grafana` | Monitoring dashboard |
| `go/alerts` | Alert configuration |
| `go/vpn` | VPN client setup guide |

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
- A C toolchain (e.g., `xcode-select --install` on macOS)
- SSH client with key-based auth configured to all servers

### Network preparation

Forward UDP port **51820** and TCP port **443** on your gateway to the server that will act as the WireGuard endpoint and ingress point.

### Remote access (optional)

If you want to access services from outside the network perimeter without VPN, set up a DDNS hostname before provisioning:

1. Create an account with a DDNS provider (e.g., DuckDNS, No-IP)
2. Register a hostname (e.g., `infra.example.com`)
3. Note your login credentials — you'll add them to `iac.yml`

During provisioning, the system automatically installs ddclient to keep the hostname pointed at your public IP. If you skip this step, everything still works on your local network and over VPN.

## Setup

### 1. Clone and build

```bash
git clone <repo-url>
cd <repo>
make
```

### 2. Configure

Copy the environment file and fill in your Bitwarden credentials:

```bash
cp .env.example .env
```

The `.env` file (see `.env.example` for required values) is used by `iac` to authenticate to Bitwarden Secrets Manager. At runtime, `iac` merges configuration from two sources, with BWS taking priority:

| Priority | Source | Notes |
|----------|--------|-------|
| 1 (highest) | Bitwarden Secrets Manager | Secrets and sensitive values |
| 2 | `iac.yml` | Infrastructure configuration |

Populate both:

```bash
bws secret create <KEY> <VALUE> <PROJECT_ID>  # store sensitive values in BWS
vim ./iac.yml                                  # fill in infrastructure config
```

`iac` generates all tool-specific configs (Ansible inventory, Terraform tfvars) into `.generated/`. You never edit those files directly.

## Usage

### Bootstrap infrastructure

Provision each server — this hardens the OS, installs WireGuard (on the designated VPN host), installs a hypervisor (Incus), creates a workload VM with Traefik, configures OVN overlay networking, and joins k3s:

```bash
iac setup                     # setup all hosts specified in your config
iac setup --host <your-host>  # setup only one host. If a cluster already exists, join this host to that cluster. Otherwise, initialize a new cluster
```

### Add VPN clients

```bash
iac wg add alice-laptop
iac wg add alice-phone
```

Client configs are saved to `.generated/vpn-clients/`. Import them into the WireGuard app on each device, then delete the `.conf` files from your workstation — they contain the client's private key and preshared key. The `.generated/` directory is gitignored and the files are created with `0600` permissions, but they should be treated as sensitive and not kept around longer than needed.

### Set up local DNS

Deploy CoreDNS on the cluster so that `*.infra.example.com` resolves to the ingress host's IP. This is what allows hosts on your network to reach services by subdomain (e.g., `grafana.infra.example.com`).

```bash
kubectl apply -f services/coredns.yml
```

Then update your network's DHCP configuration to distribute the cluster as the DNS server. This is a one-time change — after this, every host on the network resolves service subdomains automatically.

### Deploy services

This platform is application-agnostic. After infrastructure is set up, deploy any service using standard Kubernetes manifests and kubectl. See `services/` for full working examples.

k3s handles scheduling and restarts. Traefik picks up new services automatically via Ingress resources and routes by subdomain.

```bash
kubectl apply -f services/grafana.yml
kubectl apply -f services/golinks.yml
```

### Verify

```bash
iac status # cluster overview: hosts, services, VPN, k3s
```

### Tear down

```bash
iac teardown # destroy all VMs via Terraform
```

## CLI reference

```
iac <command> [options]

Commands:
  setup [--host <host>]                  Setup one or more physical hosts (hardening, hypervisor, VPN, Reverse Proxy, VM, OVN, k3s)
  wg add <name>                          Add a WireGuard client (registers peer server-side, generates client config)
  status                                 Show cluster status (hosts, services, VPN, k3s)
  teardown                               Tear down all VMs
  version                                Print version

Low-level commands:
  inventory <host>                       Generate the Ansible inventory file for a host and its VMs
  ansible bootstrap-server <host>        Run the bootstrap-server playbook against a single host
  ansible bootstrap-server <host> --vms  Run the bootstrap-server playbook against a host's VMs
  ansible setup-host <host>              Run the setup-host playbook against a single host
  ansible setup-vm <host>                Run the setup-vm playbook against a host's VMs
  ssh add <host>                         Add a host and its VMs to ~/.ssh/config
  terraform apply                        Apply the Terraform project
```

`iac` wraps these low-level commands because they require config resolution — secret fetching, inventory generation, and var merging — that would otherwise need to be done manually.

## Project structure

```
iac.yml                      # infrastructure configuration
services/                    # service manifests (one per app)
cmd/                         # Go CLI source
  iac/                       # entrypoint and command routing
ansible/
  playbooks/                 # ansible playbooks
  roles/                     # ansible roles
terraform/                   # incus VM lifecycle
.generated/                  # auto-generated configs (gitignored)
```

## Scaling to new hardware

When you add or migrate servers:

1. Update `iac.yml` with the new host IPs and storage paths
2. Run `iac setup --host <new-host>` for each new host
3. k3s automatically joins the new node to the cluster
4. OVN extends the overlay network to the new host
5. Deploy services with `kubectl` — no changes to manifests needed, k3s schedules across the cluster

## Tech stack

| Component    | Tool                  | Why                                          |
|--------------|-----------------------|----------------------------------------------|
| CLI          | Go                    | Single binary, no runtime dependencies       |
| Provisioning | Ansible               | Agentless, SSH-based, idempotent             |
| VM lifecycle | Terraform + Incus     | Declarative, reproducible                    |
| Scheduling   | k3s                   | Kubernetes-native, rolling updates, auto-restart |
| Networking   | OVN                   | Overlay networking, encrypted east-west traffic |
| Firewall     | UFW                   | Simple, readable firewall rules              |
| VPN          | WireGuard             | Kernel-level, no third-party trust           |
| Proxy        | Traefik               | Auto-discovery, TLS, subdomain routing       |
| Monitoring   | Grafana + Prometheus  | Dashboards, alerting, service health         |
| Go links     | golinks               | Internal short URLs for quick service access |
| OS           | Debian 12             | Stable, security updates, KVM support        |
