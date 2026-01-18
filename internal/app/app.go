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

type TargetStruct struct {
	Target string
	Struct any
}

type WritableFile interface {
	SetFile() error
}

type App struct {
	hostAlias     string
	configMux     *conflux.ConfigMux
	targetStructs []TargetStruct
}

func New(hostAlias string, targets []string) (App, error) {
	configMux := conflux.NewConfigMux(
		conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
		conflux.WithEnvReader(),
		conflux.WithBitwardenSecretReader(),
	)

	targetStructs, err := aliasAndTargetsToStructs(hostAlias, targets)
	if err != nil {
		return App{}, fmt.Errorf("error: %w: no supported combination for hostAlias(%s) and targets(%v)", ErrInvalidArgs, hostAlias, targets)
	}

	return App{
		hostAlias:     hostAlias,
		configMux:     configMux,
		targetStructs: targetStructs,
	}, nil
}

func (a App) GetConfig() (map[string]string, map[string]string, error) {
	allConfigs, allDiagnostics := make(map[string]string), make(map[string]string)
	for _, targetStruct := range a.targetStructs {
		diagnostics, err := conflux.Unmarshal(a.configMux, targetStruct.Struct)
		if errors.Is(err, conflux.ErrInvalidFields) {
			maps.Copy(allDiagnostics, diagnostics)
			continue
		} else if err != nil {
			return nil, nil, fmt.Errorf("error unmarshalling: %v", err)
		}

		if err := mapstructure.Decode(targetStruct.Struct, &allConfigs); err != nil {
			return nil, nil, fmt.Errorf("error merging config struct to map: %v", err)
		}
	}

	return allConfigs, allDiagnostics, nil
}

func (a App) CheckConfig() (map[string]string, error) {
	allDiagnostics := make(map[string]string)
	for _, targetStruct := range a.targetStructs {
		diagnostics, err := conflux.Unmarshal(a.configMux, targetStruct.Struct)
		if err != nil {
			return nil, fmt.Errorf("internal error unmarshalling: %v", err)
		}
		maps.Copy(allDiagnostics, diagnostics)
	}

	return allDiagnostics, nil
}

func (a App) SetFile() ([]string, error) {
	writableFiles, nonWritableTargets := make([]WritableFile, 0), make([]string, 0)
	for _, configStruct := range a.targetStructs {
		if writableFile, ok := configStruct.Struct.(WritableFile); !ok {
			nonWritableTargets = append(nonWritableTargets, configStruct.Target)
		} else {
			writableFiles = append(writableFiles, writableFile)
		}
	}

	if len(nonWritableTargets) > 0 {
		return nil, fmt.Errorf("error: %w: the following targets do not support writing to a file: %v", ErrInvalidArgs, nonWritableTargets)
	}

	diagnostics := make([]string, 0)
	for _, writableFile := range writableFiles {
		var errAlreadyExists *models.ErrAlreadyExists
		if err := writableFile.SetFile(); errors.As(err, &errAlreadyExists) {
			diagnostics = append(diagnostics, "skipping write to %s because %s already exists in that file", errAlreadyExists.Name, errAlreadyExists.Resource)
		} else if err != nil {
			return nil, fmt.Errorf("error writing to file: %v", err)
		}
	}

	return diagnostics, nil
}

func aliasAndTargetsToStructs(alias string, targets []string) ([]TargetStruct, error) {
	result := make([]TargetStruct, 0, len(targets))
	for _, target := range targets {
		if alias == "proxmox" && target == "ansible" {
			result = append(result, TargetStruct{target, models.NewAnsibleProxmoxConfig()})
		} else if target == "ssh" {
			result = append(result, TargetStruct{target, models.NewSSHHost(alias)})
		} else {
			return nil, fmt.Errorf("unexpected alias(%s) and target(%s) combination", alias, target)
		}
	}
	return result, nil
}
