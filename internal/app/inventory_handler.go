package app

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

var inventoryTmpl = template.Must(template.New("inventory").Parse(`all:
  children:
    hosts:
      hosts:
{{- range .Hosts}}
        {{.Name}}:
          ansible_host: {{.IP}}
{{- end}}
    vms:
      hosts:
{{- range .VMs}}
        {{.Name}}:
          ansible_host: {{.IP}}
          ansible_ssh_common_args: '-o ProxyJump={{.ParentSSHUser}}@{{.ParentIP}}'
          parent_host: {{.ParentName}}
{{- end}}
`))

type inventoryHandler struct{}

func newInventoryHandler() inventoryHandler {
	return inventoryHandler{}
}

func (h inventoryHandler) execute(c inventoryConfig) error {
	dir := filepath.Join(".generated", "ansible", "inventory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating inventory dir: %v", err)
	}

	f, err := os.Create(filepath.Join(dir, "hosts.yml"))
	if err != nil {
		return fmt.Errorf("creating hosts.yml: %v", err)
	}
	defer func() { _ = f.Close() }()

	if err := inventoryTmpl.Execute(f, c); err != nil {
		return fmt.Errorf("executing inventory template: %v", err)
	}

	return nil
}
