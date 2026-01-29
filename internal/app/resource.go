package app

type resource string

const (
	AnsiblePlaybookResource  resource = "ansiblePlaybook"
	AnsibleInventoryResource resource = "ansibleInventory"
	TerraformResource        resource = "terraformResource"
	SSHResource              resource = "ssh"
)
