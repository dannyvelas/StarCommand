package app

import (
	"errors"
	"fmt"
	"maps"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/handlers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/go-viper/mapstructure/v2"
)

type WritableFile interface {
	SetFile() error
}

type App struct {
	hostAlias string
	targets   []string
	configMux *conflux.ConfigMux
	handler   handlers.Handler
}

var handlerMap = map[string]handlers.Handler{
	"proxmox": handlers.NewProxmoxHandler(),
}

func New(configMux *conflux.ConfigMux, hostAlias string, targets []string) (App, error) {
	handler, ok := handlerMap[hostAlias]
	if !ok {
		return App{}, fmt.Errorf("error: %w: unsupported host(%s)", ErrInvalidArgs, hostAlias)
	}

	return App{
		hostAlias: hostAlias,
		targets:   targets,
		configMux: configMux,
		handler:   handler,
	}, nil
}

func (a App) GetConfig() (map[string]string, map[string]string, error) {
	allConfigs, allDiagnostics := make(map[string]string), make(map[string]string)
	for _, target := range a.targets {
		targetStruct, err := a.handler.TargetToStruct(target)
		if err != nil {
			return nil, nil, fmt.Errorf("error: %w: unexpected alias(%s) and target(%s) combination", ErrInvalidArgs, a.hostAlias, target)
		}

		diagnostics, err := conflux.Unmarshal(a.configMux, targetStruct)
		if errors.Is(err, conflux.ErrInvalidFields) {
			maps.Copy(allDiagnostics, diagnostics)
			continue
		} else if err != nil {
			return nil, nil, fmt.Errorf("internal error unmarshalling: %v", err)
		}

		config := &mapstructure.DecoderConfig{TagName: "json", Result: &allConfigs}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return nil, nil, fmt.Errorf("internal error creating new decoder: %v", err)
		}

		if err := decoder.Decode(targetStruct); err != nil {
			return nil, nil, fmt.Errorf("internal error merging config struct to map: %v", err)
		}
	}

	return allConfigs, allDiagnostics, nil
}

func (a App) CheckConfig() (map[string]string, error) {
	allDiagnostics := make(map[string]string)
	for _, target := range a.targets {
		targetStruct, err := a.handler.TargetToStruct(target)
		if err != nil {
			return nil, fmt.Errorf("error: %w: unexpected alias(%s) and target(%s) combination", ErrInvalidArgs, a.hostAlias, target)
		}

		diagnostics, err := conflux.Unmarshal(a.configMux, targetStruct)
		if err != nil {
			return nil, fmt.Errorf("internal error unmarshalling: %v", err)
		}

		maps.Copy(allDiagnostics, diagnostics)
	}

	return allDiagnostics, nil
}

func (a App) SetFile() ([]string, error) {
	writableFiles, nonWritableTargets := make([]WritableFile, 0), make([]string, 0)
	for _, target := range a.targets {
		targetStruct, err := a.handler.TargetToStruct(target)
		if err != nil {
			return nil, fmt.Errorf("error: %w: unexpected alias(%s) and target(%s) combination", ErrInvalidArgs, a.hostAlias, target)
		}

		if writableFile, ok := targetStruct.(WritableFile); !ok {
			nonWritableTargets = append(nonWritableTargets, target)
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
			diagnostics = append(diagnostics, fmt.Sprintf("skipping write to %s because %s already exists in that file", errAlreadyExists.Name, errAlreadyExists.Resource))
		} else if err != nil {
			return nil, fmt.Errorf("error writing to file: %v", err)
		}
	}

	return diagnostics, nil
}
