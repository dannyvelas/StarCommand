# Internals

## How `stc setup` works

`stc setup` applies desired state to all hosts in `stc.yml`. It is safe to run at any time — Ansible's idempotency means already-provisioned hosts are verified quickly and skipped where no changes are needed. New hosts are fully provisioned. Pass one or more `--host` flags to limit execution to a specific subset.

### 1. Generate Ansible inventory
> `stc inventory generate`

An Ansible inventory file is generated from `stc.yml`. It includes all configured hosts and their VMs. VMs are configured with `ProxyJump` pointing to their parent host, so that later Ansible steps can reach them through the host without requiring the VMs to be directly accessible on the network.

### 2. Bootstrap hosts
> `stc ansible bootstrap-host` · [bootstrap-host.yml](../ansible/playbooks/bootstrap-host.yml)

Runs concurrently against all hosts. Hardens the OS: configures UFW, enforces SSH key-only authentication, and enables unattended security updates.

### 3. Set up hosts
> `stc ansible setup-host` · [setup-host.yml](../ansible/playbooks/setup-host.yml)

Runs concurrently against all hosts. Installs and configures host-level services:

- **Incus** — installs the hypervisor, creates a default storage pool and NAT bridge network (`incusbr0`), and opens the Incus API port in UFW
- **WireGuard** — installs and configures a WireGuard VPN server (on the designated VPN host only)
- **Postfix** — configures Postfix as a Gmail SMTP relay so the host can send email alerts
- **Host hardening** — enables IPv4 forwarding, configures NAT port forwarding rules to route inbound traffic to VMs, blocks VM traffic from reaching the home LAN (egress isolation), and optionally installs ddclient for dynamic DNS

---

The following steps run once per host, in order:

### 4. Register host in `~/.ssh/config`
> `stc ssh add <host>`

The host is added to `~/.ssh/config` on your workstation if not already present, so that subsequent SSH and Ansible commands can refer to it by name.

### 5. Join OVN overlay network

OVN provides encrypted east-west networking between hosts.

- **First host:** initializes the OVN central database
  ```
  ovs-vsctl set open . external-ids:ovn-remote=tcp:<host-ip>:6642 \
                          external-ids:ovn-encap-type=geneve \
                          external-ids:ovn-encap-ip=<host-ip>
  ovn-ctl start_northd
  ```
- **Subsequent hosts:** join as a chassis node
  ```
  ovs-vsctl set open . external-ids:ovn-remote=tcp:<first-host-ip>:6642 \
                          external-ids:ovn-encap-type=geneve \
                          external-ids:ovn-encap-ip=<host-ip>
  ```

### 6. Join k3s cluster (server)

k3s runs on each host as a server node.

- **First host:** initializes the cluster
  ```
  curl -sfL https://get.k3s.io | sh -
  ```
- **Subsequent hosts:** join the existing cluster as additional server nodes
  ```
  curl -sfL https://get.k3s.io | K3S_URL=https://<first-host-ip>:6443 \
                                   K3S_TOKEN=<cluster-token> sh -s - server
  ```

### 7. Join Incus cluster

Incus clustering is set up before VMs are created so that Terraform provisions VMs into an already-clustered environment.

- **First host:** initializes the Incus cluster and sets it as the active remote
  ```
  incus cluster enable <host-name>
  incus remote add my-infra <host-ip>
  incus remote switch my-infra
  ```
- **Subsequent hosts:** join the existing Incus cluster
  ```
  incus cluster add <host-name>   # run on an existing cluster member to get a join token
  incus admin init                # run on the new host, providing the join token when prompted
  ```

---

Once all hosts have been registered, VMs are provisioned and configured concurrently across all hosts:

### 8. Create all VMs with Terraform
> `stc terraform apply` · [terraform/main.tf](../terraform/main.tf)

Terraform provisions VMs for all hosts in a single run via Incus. Each VM is on a private NAT subnet and is not directly reachable from the physical network.

### 9. Bootstrap all VMs
> `stc ansible bootstrap-host --host <vm1> --host <vm2> ...` · [bootstrap-host.yml](../ansible/playbooks/bootstrap-host.yml)

Same OS hardening as step 2, now applied concurrently to all newly created VMs across all hosts. `stc setup` passes all VM names automatically.

### 10. Set up all VMs
> `stc ansible setup-vm --host <vm1> --host <vm2> ...` · [setup-vm.yml](../ansible/playbooks/setup-vm.yml)

Installs VM-level services concurrently across all VMs: Docker and storage mount points. `stc setup` passes all VM names automatically.

---

The following steps run once per VM, in order:

### 11. Register VM in `~/.ssh/config`
> `stc ssh add <vm>`

The VM is added to `~/.ssh/config` with a `ProxyJump` directive pointing to its parent host, making it reachable by name from your workstation.

### 12. Join k3s cluster (agent)

The VM joins the k3s cluster as a worker node, making it available for workload scheduling.

```
curl -sfL https://get.k3s.io | K3S_URL=https://<first-host-ip>:6443 \
                                 K3S_TOKEN=<cluster-token> sh -s - agent
```

---

Once all VMs are set up:

### 13. Deploy Traefik

```
kubectl apply -f services/traefik.yml
```

Deploys Traefik as the cluster ingress. Traefik watches for Kubernetes `Ingress` resources and automatically routes traffic to services by subdomain.

### 14. Deploy CoreDNS

```
kubectl apply -f services/coredns.yml
```

Deploys CoreDNS so that `*.infra.example.com` resolves to the cluster's ingress IP. Once your network's DHCP is updated to distribute the cluster as the DNS server, every host on your network resolves service subdomains automatically.

### 15. Print cluster summary

```
incus list
```
