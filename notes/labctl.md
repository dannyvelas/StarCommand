## labctl spec

An internal CLI to configure a homelab

Usage:
* labctl [command]

Style 1:
* labctl resolve <host-alias>: Generate a JSON object of configuration values for a given host.
* labctl resolve <host-alias> --dry-run: Generate a table showing all the keys that were found, as well as all the keys that are missing.
* labctl setup-ssh <host-alias>: Update `~/.ssh/config` file to connect to `<host-alias>`.

Sytle 2:
* labctl get config <host-alias>: Generate a JSON object of configuration values for a given host
* labctl set ssh <host-alias>: Update `~/.ssh/config` file to connect to `<host-alias>`.
* labctl check reqs <host-alias>: Check whether configs are set for proxmox

Flags:
* -h, --help: help for labctl
* -v, --verbose: verbose mode
