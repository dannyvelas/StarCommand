package app

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/go-viper/mapstructure/v2"
)

type rule struct {
	Name    string
	Match   func(resource resource, action action, hostAlias string) bool
	handler handler
}

var registry = []rule{
	{
		Name: "ansible run proxmox",
		Match: func(resource resource, action action, hostAlias string) bool {
			return resource == ansiblePlaybookResource && action == runAction && hostAlias == "proxmox"
		},
		handler: newAnsibleProxmoxHandler(),
	},
	{
		Name: "ssh add <any-host-alias>",
		Match: func(resource resource, action action, hostAlias string) bool {
			return resource == sshResource && action == addAction
		},
		handler: newSSHHandler(),
	},
}

func execute(configMux *conflux.ConfigMux, resource resource, action action, hostAlias string, dryRun bool) (map[string]string, error) {
	rule, err := findRule(resource, action, hostAlias)
	if err != nil {
		return nil, err
	}

	configStruct := rule.handler.getConfig(hostAlias)
	diagnostics, err := conflux.Unmarshal(configMux, configStruct)
	if errors.Is(err, conflux.ErrInvalidFields) {
		return diagnostics, fmt.Errorf("error getting config for running %s on %s host:\n%s", resource, hostAlias, conflux.DiagnosticsToTable(diagnostics))
	} else if err != nil {
		return nil, fmt.Errorf("internal error unmarshalling config to struct for %s on %s: %v", resource, hostAlias, err)
	}

	if dryRun {
		return diagnostics, nil
	}

	configMap, err := configAsMap(configStruct)
	if err != nil {
		return nil, fmt.Errorf("internal error converting config to map: %v", err)
	}

	handlerDiagnostics, err := rule.handler.execute(configMap, hostAlias)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func findRule(resource resource, action action, hostAlias string) (rule, error) {
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
