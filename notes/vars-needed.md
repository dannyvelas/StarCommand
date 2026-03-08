| Variable                | Setup-proxmox (top-level) | ssh-harden role | postfix_relay role | automated_updates role | ufw role |
|-------------------------|---------------------------|-----------------|--------------------|------------------------|----------|
| ssh_public_key          | X                         | X               |                    |                        |          |
| node_cidr_address       | X                         |                 |                    |                        |          |
| gateway_address         | X                         |                 |                    |                        |          |
| physical_nic            | X                         |                 |                    |                        |          |
| admin_password          |                           | X               |                    |                        |          |
| ssh_port                |                           | X               |                    |                        | X        |
| admin_email             |                           |                 |                    | X                      |          |
| auto_update_reboot_time |                           |                 |                    | X                      |          |
| smtp_user               |                           |                 | X                  |                        |          |
| smtp_password           |                           |                 | X                  |                        |          |
