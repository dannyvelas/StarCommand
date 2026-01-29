package app

import (
	"fmt"

	"github.com/dannyvelas/homelab/internal/models"
)

var _ handler = ansibleProxmoxHandler{}

type ansibleProxmoxHandler struct{}

func newAnsibleProxmoxHandler() ansibleProxmoxHandler {
	return ansibleProxmoxHandler{}
}

func (h ansibleProxmoxHandler) getConfig(_ string) any {
	return models.NewAnsibleProxmoxConfig()
}

func (h ansibleProxmoxHandler) execute(config map[string]string, hostAlias string) (map[string]string, error) {
	fmt.Printf("running ansible on proxmox...\n")
	fmt.Printf("finished running ansible on proxmox...\n")
	return nil, nil
}
