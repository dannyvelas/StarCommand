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

### host commands

| Action | Resource | Host-Alias   | Flags                       |
|--------|----------|--------------|-----------------------------|
| get    | config   | proxmox      | --for ansible,ssh,terraform |
| set    | file     | proxmox      | --for ssh                   |
| check  | config   | proxmox      | --for ansible,ssh,terraform |
| get    | config   | plex         | --for terraform,ansible     |
| check  | config   | plex         | --for terraform,ansible     |
| set    | file     | plex         | --for ansible               |
| get    | config   | wireguard-vm | --for terraform             |
| check  | config   | wireguard-vm | --for terraform             |
| get    | config   | generic      | --for ansible               |
| check  | config   | generic      | --for ansible               |

| Action | Resource | Key        |
|--------|----------|------------|
| set    | secret   | secret-key |

## labctl grammar v4 ?

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
