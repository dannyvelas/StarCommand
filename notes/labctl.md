# labctl spec

An internal CLI to configure a homelab

## labctl grammar v1

| Action    | Host-Alias | Flags     |
|-----------|------------|-----------|
| resolve   | proxmox    |           |
| resolve   | proxmox    | --dry-run |
| setup-ssh | proxmox    |           |

## labctl grammar v2

| Action | Resource | Host-Alias |
|--------|----------|------------|
| get    | config   | proxmox    |
| set    | ssh      | proxmox    |
| check  | reqs     | proxmox    |

## labctl grammar v3

| Action | Resource | Host-Alias   | Flags                                   |
|--------|----------|--------------|-----------------------------------------|
| get    | config   | proxmox      | --for run-ansible,ssh,terraform         |
| set    | file     | proxmox      | --for ssh                               |
| check  | config   | proxmox      | --for ansible,ssh,terraform             |
| get    | config   | plex         | --for terraform,run-ansible,set-ansible |
| check  | config   | plex         | --for terraform,run-ansible,set-ansible |
| set    | file     | plex         | --for ansible,ssh                       |
| get    | config   | wireguard-vm | --for terraform                         |
| check  | config   | wireguard-vm | --for terraform                         |
| get    | config   | *            | --for run-ansible,ssh                   |
| check  | config   | *            | --for run-ansible,ssh                   |
| set    | file     | *            | --for ssh

| Action | Resource | Key        |
|--------|----------|------------|
| set    | secret   | secret-key |

## labctl grammar v4 (maybe)

### localized commands
| Sub-command | Domain | Action | Host-Alias   | Flags                       |
|-------------|--------|--------|--------------|-----------------------------|
| host        | config | get    | proxmox      | --for ansible,ssh,terraform |
| host        | file   | set    | proxmox      | --for ssh                   |
| host        | config | check  | proxmox      | --for ansible,ssh,terraform |
| host        | config | get    | plex         | --for terraform,ansible     |
| host        | config | check  | plex         | --for terraform,ansible     |
| host        | file   | set    | plex         | --for ansible               |
| host        | config | get    | wireguard-vm | --for terraform             |
| host        | config | check  | wireguard-vm | --for terraform             |
| host        | config | get    | generic      | --for ansible               |
| host        | config | check  | generic      | --for ansible               |

### global commands
| Domain | Set      | Key        |
|--------|----------|------------|
| secret | set      | secret-key |

## labctl grammar v5 

Goals:
- A command that will allow the terraform project of a given host to get applied
- A command to update the ansible inventory file to include a given host
- A command that will allow the ansible playbook of a given host to run
- A command to update the `~/.ssh/config file` to include connection information to a given host
- One command to check whether the multiple actions for the same host `h` will work (e.g. check if applying the terraform project of `h` will work, check if updating the ansible inventory file to include `h` will work, check if running the ansible playbook for `h` will work, check if updating the ssh config file to include `h` will work)
- One command to set a bitwarden secret

| Sub-command | Action | Host-Alias   | Flags                                   |
|-------------|--------|--------------|-----------------------------------------|
| ansible     | run    | proxmox      |                                         |
| terraform   | run    | proxmox      |                                         |
| ansible     | run    | plex         |                                         |
| ansible     | set    | plex         |                                         |
| terraform   | run    | plex         |                                         |
| terraform   | run    | wireguard-vm |                                         |
| ansible     | run    | *            |                                         |
| ssh         | set    | proxmox      |                                         |
| ssh         | set    | plex         |                                         |
| ssh         | set    | wireguard-vm |                                         |
| ssh         | set    | *            |                                         |


| Sub-command | Action       | Key |
|-------------|--------------|-----|
| secret      | set          | *   |

| Sub-command | Host         | Arguments (minimum one of the following) |
|-------------|--------------|------------------------------------------|
| check       | proxmox      | terraform ansible-set ansible-run ssh    |
| check       | plex         | terraform ansible-set ansible-run ssh    |
| check       | wireguard-vm | terraform ansible-set ansible-run ssh    |
| check       | *            | ansible-set ansible-run ssh              |
