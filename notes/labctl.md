## labctl spec

An internal CLI to configure a homelab

Usage:
* labctl [command]

Style 1:
* labctl resolve <host-name>: Generate a JSON object of configuration values for a given host.
* labctl resolve <host-name> --dry-run: Generate a table showing all the keys that were found, as well as all the keys that are missing.
* labctl setup-ssh <host-name>: Update `~/.ssh/config` file to connect to `<host-name>`.

Sytle 2:
* labctl get config <host-name>: Generate a JSON object of configuration values for a given host
* labctl check config <host-name>: Generate a table showing all the keys that were found, as well as all the keys that are missing.
* labctl set ssh <host-name>: Update `~/.ssh/config` file to connect to `<host-name>`.
* labctl check ssh proxmox

Flags:
* -h, --help: help for labctl
* -v, --verbose: verbose mode
