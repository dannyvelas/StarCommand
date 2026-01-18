package app

import (
	"errors"
	"fmt"
	"maps"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/go-viper/mapstructure/v2"
)

type App struct {
	configMux     *conflux.ConfigMux
	configStructs []any
}

func New(hostAlias string, targets []string) (App, error) {
	configMux := conflux.NewConfigMux(
		conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
		conflux.WithEnvReader(),
		conflux.WithBitwardenSecretReader(),
	)

	configStructs, err := aliasAndTargetsToStructs(hostAlias, targets)
	if err != nil {
		return App{}, fmt.Errorf("error: %w: no supported combination for hostAlias(%s) and targets(%v)", ErrInvalidArgs, hostAlias, targets)
	}

	return App{
		configMux:     configMux,
		configStructs: configStructs,
	}, nil
}

func (a App) GetConfig() (map[string]string, map[string]string, error) {
	allConfigs, allDiagnostics := make(map[string]string), make(map[string]string)
	for _, configStruct := range a.configStructs {
		diagnostics, err := conflux.Unmarshal(a.configMux, configStruct)
		if errors.Is(err, conflux.ErrInvalidFields) {
			maps.Copy(allDiagnostics, diagnostics)
			continue
		} else if err != nil {
			return nil, nil, fmt.Errorf("error unmarshalling: %v", err)
		}

		if err := mapstructure.Decode(configStruct, &allConfigs); err != nil {
			return nil, nil, fmt.Errorf("error merging config struct to map: %v", err)
		}
	}

	return allConfigs, allDiagnostics, nil
}

func (a App) CheckConfig() (map[string]string, error) {
	allDiagnostics := make(map[string]string)
	for _, configStruct := range a.configStructs {
		diagnostics, err := conflux.Unmarshal(a.configMux, configStruct)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling: %v", err)
		}
		maps.Copy(allDiagnostics, diagnostics)
	}

	return allDiagnostics, nil
}

func aliasAndTargetsToStructs(alias string, targets []string) ([]any, error) {
	result := make([]any, 0, len(targets))
	for _, target := range targets {
		if alias == "proxmox" && target == "ansible" {
			result = append(result, models.NewAnsibleProxmoxConfig())
		} else if target == "ssh" {
			result = append(result, models.NewSSHHost(alias))
		} else {
			return nil, fmt.Errorf("unexpected alias(%s) and target(%s) combination", alias, target)
		}
	}
	return result, nil
}
