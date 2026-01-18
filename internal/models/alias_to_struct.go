package models

import "fmt"

func AliasToStruct(alias string, targets []string) ([]any, error) {
	result := make([]any, 0, len(targets))
	for _, target := range targets {
		if alias == "proxmox" && target == "ansible" {
			result = append(result, NewAnsibleProxmoxConfig())
		} else if target == "ssh" {
			result = append(result, NewSSHHost(alias))
		} else {
			return nil, fmt.Errorf("unexpected alias(%s) and target(%s) combination", alias, target)
		}
	}
	return result, nil
}
