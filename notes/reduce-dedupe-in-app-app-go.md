
| Function       | arg1     |
|----------------|----------|
| AnsibleRun     | playbook |
| SSHAdd         | <host>   |
| TerraformApply |          |
| InventoryAdd   | <host>[] |
| Setup          | <host>[] |


| Resource  | verb             | arg2     |
|-----------|------------------|----------|
| Ansible   | bootstrap-server | <host>[] |
| Ansible   | setup-host       | <host>[] |
| Ansible   | setup-remote     | <host>[] |
| Ansible   | setup-vm         | <host>[] |
| SSH       | add              | <host>   |
| Terraform | apply            |          |
| Inventory | add              |          |
| Setup     |                  | <host>[] |
