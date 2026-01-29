package app

type resource string

const (
	ansiblePlaybookResource  resource = "ansiblePlaybook"
	ansibleInventoryResource resource = "ansibleInventory"
	terraformResource        resource = "terraformResource"
	sshResource              resource = "ssh"
)
