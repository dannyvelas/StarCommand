package app

import "github.com/dannyvelas/starcommand/internal/models"

type inventoryConfig struct {
	Hosts []inventoryHostConfig
	VMs   []inventoryVMConfig
}

type inventoryHostConfig struct {
	Name string
	IP   string
}

type inventoryVMConfig struct {
	Name          string
	IP            string
	ParentName    string
	ParentIP      string
	ParentSSHUser string
}

func newInventoryConfig(hosts []models.Host) inventoryConfig {
	c := inventoryConfig{}
	for _, host := range hosts {
		c.Hosts = append(c.Hosts, inventoryHostConfig{
			Name: host.Name,
			IP:   host.IP,
		})
		for _, vm := range host.VMs {
			c.VMs = append(c.VMs, inventoryVMConfig{
				Name:          vm.Name,
				IP:            vm.IP,
				ParentName:    host.Name,
				ParentIP:      host.IP,
				ParentSSHUser: host.SSH.User,
			})
		}
	}

	return c
}
