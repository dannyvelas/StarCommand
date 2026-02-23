# Internals

## How `iac setup` works

`iac setup` provisions one or more hosts end-to-end. It runs the same sequence of steps for each host, detecting existing cluster state along the way so that subsequent hosts join rather than reinitialize.

### 1. Generate Ansible inventory
> `iac inventory generate`

An Ansible inventory file is generated from `iac.yml`. It includes all configured hosts and their VMs. VMs are configured with `ProxyJump` pointing to their parent host, so that later Ansible steps can reach them through the host without requiring the VMs to be directly accessible on the network.

### 2. Bootstrap hosts
> `iac ansible bootstrap-server`

Runs against all hosts. Hardens the OS: configures UFW, enforces SSH key-only authentication, and enables unattended security updates.

### 3. Set up hosts
> `iac ansible setup-host`

Runs against all hosts. Installs and configures host-level services: Incus (hypervisor) and WireGuard (on the designated VPN host).

---

The following steps run once per host, in order:

### 4. Register host in `~/.ssh/config`
> `iac ssh add <host>`

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

### 8. Create VMs with Terraform
> `iac terraform apply`

Terraform provisions the VMs for this host via Incus. Each VM is on a private NAT subnet and is not directly reachable from the physical network.

### 9. Bootstrap VMs
> `iac ansible bootstrap-server --vms`

Same OS hardening as step 2, now applied to the newly created VMs.

### 10. Set up VMs
> `iac ansible setup-vm`

Runs against all VMs for this host. Installs VM-level services: Docker, Traefik (reverse proxy), and storage mount points.

### 11. Register VMs in `~/.ssh/config`
> `iac ssh add <vm>`

Each VM is added to `~/.ssh/config` with a `ProxyJump` directive pointing to its parent host, making it reachable by name from your workstation.

### 12. Join k3s cluster (agent)

Each VM joins the k3s cluster as a worker node, making it available for workload scheduling.

```
curl -sfL https://get.k3s.io | K3S_URL=https://<first-host-ip>:6443 \
                                 K3S_TOKEN=<cluster-token> sh -s - agent
```

---

Once all hosts are processed, `iac setup` prints a summary of the Incus cluster with `incus list`.
