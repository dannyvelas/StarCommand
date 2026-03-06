
| Function       | arg1     |
|----------------|----------|
| AnsibleRun     | playbook |
| SSHAdd         | <host>   |
| TerraformApply |          |
| InventoryAdd   | <host>[] |
| Setup          | <host>[] |


| Resource  | verb             | arg2     |
|-----------|------------------|----------|
| Ansible   | bootstrap-server |          |
| Ansible   | setup-host       |          |
| Ansible   | setup-remote     |          |
| Ansible   | setup-vm         |          |
| SSH       | add              | <host>   |
| Terraform | apply            |          |
| Inventory | add              | <host>[] |
| Setup     |                  | <host>[] |
