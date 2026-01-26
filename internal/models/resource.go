package models

import "fmt"

type Resource string

const (
	AnsiblePlaybookResource Resource = "ansiblePlaybook"
	SSHResource             Resource = "ssh"
)

func StringToResource(s string) (Resource, error) {
	r := Resource(s)
	switch r {
	case AnsiblePlaybookResource, SSHResource:
		return r, nil
	default:
		return "", fmt.Errorf("invalid resource: %v", s)
	}
}
