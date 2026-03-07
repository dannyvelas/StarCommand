## Progress

- [ ] Finish creating example `stc.yml` file structure
- [ ] Finish writing logic to read config so that ansible playbooks have the data they need
- [ ] Test if sensitive value reading works
- [ ] Explore if there's a way to reduce duplication in `../internal/app/app.go`
- [ ] Test that bootstrap-host.yml playbook works via `stc ansible bootstrap-host`
- [ ] Test that setup-host.yml playbook works via `stc ansible setup-host`
- [ ] Test that setup-vm.yml playbook works via `stc ansible setup-vm`
- [ ] Test that terraform project works via `stc terraform apply` 
- [ ] Test that terraform project works via `stc terraform destroy` 
- [ ] Add `fail2ban`
- [ ] Add `host_harden` role back to `setup-host.yml` and test that it works
- [ ] Add `wireguard` role back to `setup-host.yml` and test that it works
- [ ] Add K3s support
- [ ] add support for `stc setup` running `kubectl apply -f services/traefik.yml`
- [ ] add support for `stc setup` running `kubectl apply -f services/coredns.yml`
- [ ] Add OVN support
- [ ] Add VLAN segmentation (A management VLAN for SSH access to hosts, a separate VLAN for OVN overlay traffic between hosts, A VLAN isolating your server network from your outside network)
- [ ] Add support for `status` subcommand
