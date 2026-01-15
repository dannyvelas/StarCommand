package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dannyvelas/homelab/internal/helpers"
)

const (
	StatusMissing = "missing"
	StatusLoaded  = "loaded"
)

type validatable interface {
	// Validate receives a diagnostic map where each element corresponds to a key in the config
	// the second return value will be false if at least one key was invalid. otherwise, it will be true
	Validate(map[string]string) bool
}

type fillable interface {
	// FillInKeys takes the keys that are required and uses them to fill out remaining config fields
	FillInKeys() error
}

func validateStruct(v any) (map[string]string, error) {
	diagnostics := make(map[string]string)
	valid := true

	tagToFieldMap, err := helpers.GetTagToFieldMap(v, "labctl", "json")
	if err != nil {
		return nil, fmt.Errorf("error getting tag to field map: %v", err)
	}

	for tag, field := range tagToFieldMap {
		if _, ok := field.Type.Tag.Lookup("required"); !ok {
			continue
		}

		if field.Value.IsZero() {
			diagnostics[tag] = StatusMissing
			valid = false
		} else {
			diagnostics[tag] = StatusLoaded
		}
	}

	if config, ok := v.(validatable); ok {
		valid = valid && config.Validate(diagnostics)
	}

	if !valid {
		return diagnostics, ErrInvalidFields
	}

	return diagnostics, nil
}

func Unmarshal(r Reader, target any) (map[string]string, error) {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("target must be a pointer, got %T", target)
	}

	readResult, err := r.Read()
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error reading: %v", err)
	}
	// if errors.Is(err, ErrInvalidFields) we want to continue
	// because its possible that after helpers.FromMap, the
	// resulting target will have all required fields regardless

	if err := helpers.FromMap(readResult.GetConfigMap(), target); err != nil {
		return nil, fmt.Errorf("error converting map into target: %v", err)
	}

	readDiagnostics := getDiagnostics(readResult)

	val = val.Elem()
	if val.Kind() == reflect.Map {
		return readDiagnostics, nil
	}

	targetDiagnostics, err := validateStruct(target)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error unmarhsalling into config: %v", err)
	}

	mergedDiagnostics := helpers.MergeMaps(readDiagnostics, targetDiagnostics)
	if errors.Is(err, ErrInvalidFields) {
		return mergedDiagnostics, ErrInvalidFields
	}

	if fillableTarget, ok := target.(fillable); ok {
		if err := fillableTarget.FillInKeys(); err != nil {
			return nil, fmt.Errorf("error filling in fields: %v", err)
		}
	}

	return mergedDiagnostics, nil
}

func getDiagnostics(r ReadResult) map[string]string {
	if v, ok := r.(DiagnosticReadResult); ok {
		return v.diagnostics
	}
	return nil
}
