package handlers

import (
	"fmt"

	"github.com/dannyvelas/homelab/internal/models"
)

var _ Handler = AnsibleProxmoxHandler{}

type AnsibleProxmoxHandler struct{}

func NewAnsibleProxmoxHandler() AnsibleProxmoxHandler {
	return AnsibleProxmoxHandler{}
}

func (h AnsibleProxmoxHandler) GetConfig(_ string) any {
	return models.NewAnsibleProxmoxConfig()
}

func (h AnsibleProxmoxHandler) Execute(config map[string]string, hostAlias string) (map[string]string, error) {
	fmt.Printf("running ansible on proxmox...\n")
	fmt.Printf("finished running ansible on proxmox...\n")
	return nil, nil
}
