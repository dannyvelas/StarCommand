package app

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/go-viper/mapstructure/v2"
)

type rule struct {
	Name    string
	Match   func(resource Resource, action Action, hostAlias string) bool
	Execute func(configMux *conflux.ConfigMux, hostAlias string, dryRun bool) (map[string]string, error)
}

var registry = []rule{
	{
		Name: "ansible run proxmox",
		Match: func(resource Resource, action Action, host string) bool {
			return resource == AnsiblePlaybookResource && action == RunAction && host == "proxmox"
		},
		Execute: func(configMux *conflux.ConfigMux, host string, dryRun bool) (map[string]string, error) {
			targetStruct := models.NewAnsibleProxmoxConfig()
			diagnostics, err := conflux.Unmarshal(configMux, targetStruct)
			if errors.Is(err, conflux.ErrInvalidFields) {
				return diagnostics, fmt.Errorf("error getting configs for running ansible playbook on proxmox host:\n%s", DiagnosticsToTable(diagnostics))
			} else if err != nil {
				return nil, fmt.Errorf("internal error unmarshalling configs to struct for ansible playbook on proxmox host: %v", err)
			}

			if dryRun {
				return diagnostics, nil
			}

			proxmoxAnsibleConfigMap := make(map[string]string)
			config := &mapstructure.DecoderConfig{TagName: "json", Result: &proxmoxAnsibleConfigMap}
			decoder, err := mapstructure.NewDecoder(config)
			if err != nil {
				return nil, fmt.Errorf("internal error creating new decoder: %v", err)
			}

			if err := decoder.Decode(targetStruct); err != nil {
				return nil, fmt.Errorf("internal error merging config struct to map: %v", err)
			}

			fmt.Println("running ansible...")
			fmt.Println("ran ansible...")

			return map[string]string{}, nil
		},
	},
	{
		Name: "Default Ansible",
		Match: func(resource Resource, action Action, host string) bool {
			return resource == AnsiblePlaybookResource && action == RunAction
		},
		Execute: func(configMux *conflux.ConfigMux, host string, dryRun bool) (map[string]string, error) {
			// ... Logic
			return map[string]string{}, nil
		},
	},
}

func execute(configMux *conflux.ConfigMux, resource Resource, action Action, hostAlias string, dryRun bool) (map[string]string, error) {
	for _, rule := range registry {
		if rule.Match(resource, action, hostAlias) {
			return rule.Execute(configMux, hostAlias, dryRun)
		}
	}

	return nil, fmt.Errorf("error: unsupported combination of resource(%s), action(%s), and hostAlias(%s)", resource, action, hostAlias)
}
