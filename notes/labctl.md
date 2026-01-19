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

| Action | Resource | Host-Alias | Flags             |
|--------|----------|------------|-------------------|
| get    | config   | proxmox    | --for ansible     |
| set    | file     | proxmox    | --for ssh         |
| check  | config   | proxmox    | --for ansible,ssh |

## prompt

I have a Taskfile.yml that has two targets: `setup:proxmox` and `setup:proxmox:check`. The `setup:proxmox` target does three things:
1. get configs from `labctl`
2. run ansible
3. set ~/.ssh/config to have the necessary information so that `ssh proxmox` "just works".
```
    cmds:
      - |
        CONF_JSON=$(./bin/labctl get config proxmox)
        NODE_IP=$(echo $CONF_JSON | jq -r '.node_ip')
        SSH_PORT=$(echo $CONF_JSON | jq -r '.ssh_port')
        if ssh -p "$SSH_PORT" -o ConnectTimeout=3 "admin@$NODE_IP" exit 2>/dev/null; then
          ansible-playbook -i ansible/inventory.ini ansible/setup-proxmox.yml --ask-vault-pass -e "$CONF_JSON"
        else
          ansible-playbook -i ansible/inventory.ini ansible/setup-proxmox.yml -u root --ask-vault-pass -e "$CONF_JSON"
        fi

      - ./bin/labctl set ssh proxmox
```

I want the `setup:proxmox:check` target to basically print a report of found/missing configuration values that are necessary for `setup:proxmox` to succeed. Since I'm using `labctl` as the brain, it felt right for this task to simply have one command which calls `labctl` to do all the logic of checking whether `setup:proxmox:check` will succeed.

I landed on this for the API of the CLI: `labctl check reqs proxmox`. This is good because it's consistent with the other commands (`labctl get config proxmox`, `labctl set ssh proxmox`). But, there is one thing I don't like about it. It feels like division of responsibility is violated a little bit. In this API (`labctl check reqs proxmox`) we are in no way communicating to `labctl` the steps that `setup:proxmox` is doing. We are not communicating that `setup:proxmox` is running ansible, and therefore needs ansible configs to be present, and we are not communicating that `setup:proxmox` is also setting ssh, and therefore needs `ssh` configs to be present. This means that the implementation of `labctl check reqs proxmox` basically needs to know what the `setup:proxmox` target is doing. And, it feels like `labctl`, being a lower-level tool shouldn't necessarily be concerned with this logic. Is there a way to avoid this?
