package app

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/go-viper/mapstructure/v2"
)

type rule struct {
	Name           string
	Match          func(resource Resource, action Action, hostAlias string) bool
	GetConfig      func() any
	ExecuteCommand func(configs map[string]string) error
}

var registry = []rule{
	{
		Name: "ansible run proxmox",
		Match: func(resource Resource, action Action, host string) bool {
			return resource == AnsiblePlaybookResource && action == RunAction && host == "proxmox"
		},
		GetConfig: func() any {
			return models.NewAnsibleProxmoxConfig()
		},
	},
	{
		Name: "Default Ansible",
		Match: func(resource Resource, action Action, host string) bool {
			return resource == AnsiblePlaybookResource && action == RunAction
		},
		GetConfig: func() any {
			// ... Logic
			return map[string]string{}
		},
	},
}

func execute(configMux *conflux.ConfigMux, resource Resource, action Action, hostAlias string, dryRun bool) (map[string]string, error) {
	rule, err := findRule(resource, action, hostAlias)
	if err != nil {
		return nil, err
	}

	configStruct := rule.GetConfig()
	diagnostics, err := conflux.Unmarshal(configMux, configStruct)
	if errors.Is(err, conflux.ErrInvalidFields) {
		return diagnostics, fmt.Errorf("error getting configs for running %s on %s host:\n%s", resource, hostAlias, DiagnosticsToTable(diagnostics))
	} else if err != nil {
		return nil, fmt.Errorf("internal error unmarshalling configs to struct for %s on %s: %v", resource, hostAlias, err)
	}

	if dryRun {
		return diagnostics, nil
	}

	_, err = configAsMap(configStruct)
	if err != nil {
		return nil, fmt.Errorf("internal error converting config to map: %v", err)
	}

	fmt.Println("running ansible...")
	fmt.Println("ran ansible...")

	return map[string]string{}, nil
}

func findRule(resource Resource, action Action, hostAlias string) (rule, error) {
	for _, rule := range registry {
		if rule.Match(resource, action, hostAlias) {
			return rule, nil
		}
	}

	return rule{}, fmt.Errorf("error: unsupported combination of resource(%s), action(%s), and hostAlias(%s)", resource, action, hostAlias)
}

func configAsMap(config any) (map[string]string, error) {
	configMap := make(map[string]string)
	decoderConfig := &mapstructure.DecoderConfig{TagName: "json", Result: &configMap}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, fmt.Errorf("internal error creating new decoder: %v", err)
	}

	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("internal error merging config struct to map: %v", err)
	}

	return configMap, nil
}
