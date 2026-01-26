package app

//import (
//	"errors"
//	"fmt"
//	"maps"
//
//	"github.com/dannyvelas/conflux"
//	"github.com/dannyvelas/homelab/internal/models"
//	"github.com/go-viper/mapstructure/v2"
//	"github.com/spf13/afero"
//)
//
//type ProxmoxHandler struct {
//	configMux *conflux.ConfigMux
//	fs        afero.Fs
//	homeDir   string
//
//	hostAlias string
//	targets   []string
//}
//
//func NewProxmoxHandler(configMux *conflux.ConfigMux, targets []string) *ProxmoxHandler {
//	return &ProxmoxHandler{
//		configMux: configMux,
//		fs:        afero.NewOsFs(),
//		hostAlias: "proxmox",
//		targets:   targets,
//	}
//}
//
//func (h *ProxmoxHandler) GetConfig() (map[string]string, map[string]string, error) {
//	allConfigs, allDiagnostics := make(map[string]string), make(map[string]string)
//	for _, target := range h.targets {
//		targetStruct, err := h.targetToStruct(target)
//		if err != nil {
//			return nil, nil, fmt.Errorf("error: %w: unexpected alias(%s) and target(%s) combination", ErrInvalidArgs, h.hostAlias, target)
//		}
//
//		diagnostics, err := conflux.Unmarshal(h.configMux, targetStruct)
//		if errors.Is(err, conflux.ErrInvalidFields) {
//			maps.Copy(allDiagnostics, diagnostics)
//			continue
//		} else if err != nil {
//			return nil, nil, fmt.Errorf("internal error unmarshalling: %v", err)
//		}
//
//		config := &mapstructure.DecoderConfig{TagName: "json", Result: &allConfigs}
//		decoder, err := mapstructure.NewDecoder(config)
//		if err != nil {
//			return nil, nil, fmt.Errorf("internal error creating new decoder: %v", err)
//		}
//
//		if err := decoder.Decode(targetStruct); err != nil {
//			return nil, nil, fmt.Errorf("internal error merging config struct to map: %v", err)
//		}
//	}
//
//	return allConfigs, allDiagnostics, nil
//}
//
//func (h *ProxmoxHandler) CheckConfig() (map[string]string, error) {
//	allDiagnostics := make(map[string]string)
//	for _, target := range h.targets {
//		targetStruct, err := h.targetToStruct(target)
//		if err != nil {
//			return nil, fmt.Errorf("error: %w: unexpected alias(%s) and target(%s) combination", ErrInvalidArgs, h.hostAlias, target)
//		}
//
//		diagnostics, err := conflux.Unmarshal(h.configMux, targetStruct)
//		if err != nil {
//			return nil, fmt.Errorf("internal error unmarshalling: %v", err)
//		}
//
//		maps.Copy(allDiagnostics, diagnostics)
//	}
//
//	return allDiagnostics, nil
//}
//
//func (h *ProxmoxHandler) SetFile() ([]string, error) {
//	writableFiles, nonWritableTargets := make([]WritableFile, 0), make([]string, 0)
//	for _, target := range h.targets {
//		targetStruct, err := h.targetToStruct(target)
//		if err != nil {
//			return nil, fmt.Errorf("error: %w: unexpected alias(%s) and target(%s) combination", ErrInvalidArgs, h.hostAlias, target)
//		}
//
//		if writableFile, ok := targetStruct.(WritableFile); !ok {
//			nonWritableTargets = append(nonWritableTargets, target)
//		} else {
//			writableFiles = append(writableFiles, writableFile)
//		}
//	}
//
//	if len(nonWritableTargets) > 0 {
//		return nil, fmt.Errorf("error: %w: the following targets do not support writing to a file: %v", ErrInvalidArgs, nonWritableTargets)
//	}
//
//	diagnostics := make([]string, 0)
//	for _, writableFile := range writableFiles {
//		alreadyExists, err := writableFile.ContentAlreadyExists(h.fs, h.homeDir)
//		if err != nil {
//			return nil, fmt.Errorf("error checking if %s already exists in %s file: %v", writableFile.Resource(), writableFile.Name(), err)
//		}
//
//		if alreadyExists {
//			diagnostics = append(diagnostics, fmt.Sprintf("skipping write to %s because %s already exists in that file", writableFile.Name(), writableFile.Resource()))
//			continue
//		}
//
//		if err := writableFile.SetFile(h.fs, h.homeDir); err != nil {
//			return nil, fmt.Errorf("error writing to %s file: %v", writableFile.Name(), err)
//		}
//	}
//
//	return diagnostics, nil
//}
//
//func (h *ProxmoxHandler) useFS(fs afero.Fs) {
//	h.fs = fs
//}
//
//func (h *ProxmoxHandler) useHomeDir(homeDir string) {
//	h.homeDir = homeDir
//}
//
//func (h *ProxmoxHandler) getHomeDir() string {
//	return h.homeDir
//}
//
//func (h *ProxmoxHandler) targetToStruct(target string) (any, error) {
//	switch target {
//	case "ansible":
//		return models.NewAnsibleProxmoxConfig(), nil
//	default:
//		return fallbackTargetToStruct(h.hostAlias, target)
//	}
//}
