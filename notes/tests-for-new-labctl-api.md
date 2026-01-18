## tests for new labctl api

| Command                                             | Expected                                                                              | Result |
|-----------------------------------------------------|---------------------------------------------------------------------------------------|--------|
| labctl get config proxmox                           | get config for ansible                                                                | Passed |
| labctl get config proxmox --for ansible             | get config for ansible                                                                | Passed |
| labctl get config proxmox --for ssh                 | get config for ssh                                                                    | Passed |
| labctl get config proxmox --for ansible,ssh         | get config for ansible and ssh                                                        | Passed |
| labctl get config proxmox --for ssh,ansible         | get config for ansible and ssh                                                        | Passed |
| labctl get config proxmox --for ansible --for ssh   | get config for ansible and ssh                                                        | Passed |
| labctl get config proxmox --for ssh --for ansible   | get config for ansible and ssh                                                        | Passed |
| labctl check config proxmox                         | check config for ansible                                                              | Passed |
| labctl check config proxmox --for ansible           | check config for ansible                                                              | Passed |
| labctl check config proxmox --for ssh               | check config for ssh                                                                  | Passed |
| labctl check config proxmox --for ansible,ssh       | check config for ansible and ssh                                                      | Passed |
| labctl check config proxmox --for ssh,ansible       | check config for ansible and ssh                                                      | Passed |
| labctl check config proxmox --for ansible --for ssh | check config for ansible and ssh                                                      | Passed |
| labctl check config proxmox --for ssh --for ansible | check config for ansible and ssh                                                      | Passed |
| labctl set file proxmox                             | update ~/.ssh/config file with proxmox info                                           | Passed |
| labctl set file proxmox --for ssh                   | update ~/.ssh/config file with proxmox info                                           | Passed |
| labctl set file proxmox --for ansible               | error: invalid args: the following targets do not support writing to a file [ansible] | Passed |
