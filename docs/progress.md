## Progress

- [ ] Finish creating example `iac.yml` file structure, maybe switching to cue
- [ ] Finish writing logic to read config so that ansible playbooks have the data they need
- [ ] Explore if there's a way to reduce duplication in `../internal/app/app.go`
- [ ] Test that bootstrap-server.yml playbook works via `iac ansible bootstrap-server`
- [ ] Test that setup-host.yml playbook works via `iac ansible setup-host`
- [ ] Test that setup-vm.yml playbook works via `iac ansible setup-vm`
- [ ] Test that terraform project works via `iac terraform apply` 
- [ ] Test that terraform project works via `iac terraform destroy` 
- [ ] Add `fail2ban`
- [ ] Add `host_harden` role back to `setup-host.yml` and test that it works
- [ ] Add `wireguard` role back to `setup-host.yml` and test that it works
- [ ] Add K3s support
- [ ] add support for `iac setup` running `kubectl apply -f services/traefik.yml`
- [ ] add support for `iac setup` running `kubectl apply -f services/coredns.yml`
- [ ] Add OVN support
- [ ] Add support for `status` subcommand
