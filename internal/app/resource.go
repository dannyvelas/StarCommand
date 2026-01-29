package app

type Resource string

const (
	AnsiblePlaybookResource  Resource = "ansiblePlaybook"
	AnsibleInventoryResource Resource = "ansibleInventory"
	TerraformResource        Resource = "terraformResource"
	SSHResource              Resource = "ssh"
)
