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
	diagnosticMap := make(map[string]string)
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
			diagnosticMap[tag] = StatusMissing
			valid = false
		} else {
			diagnosticMap[tag] = StatusLoaded
		}
	}

	if config, ok := v.(validatable); ok {
		valid = valid && config.Validate(diagnosticMap)
	}

	if !valid {
		return diagnosticMap, ErrInvalidFields
	}

	return diagnosticMap, nil
}

func Unmarshal(r Reader, target any) (map[string]string, error) {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("target must be a pointer, got %T", target)
	}

	readResult, err := r.read()
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error reading: %v", err)
	}
	// if errors.Is(err, ErrInvalidFields) we want to continue
	// because its possible that after helpers.FromMap, the
	// resulting target will have all required fields regardless

	if err := helpers.FromMap(readResult.getConfigMap(), target); err != nil {
		return nil, fmt.Errorf("error converting map into target: %v", err)
	}

	readDiagnosticMap := getDiagnosticMap(readResult)

	val = val.Elem()
	if val.Kind() == reflect.Map {
		return readDiagnosticMap, nil
	}

	targetDiagnosticMap, err := validateStruct(target)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error unmarhsalling into config: %v", err)
	}

	mergedDiagnostics := helpers.MergeMaps(readDiagnosticMap, targetDiagnosticMap)
	if errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("%w:\n%s", ErrInvalidFields, diagnosticMapToTable(mergedDiagnostics))
	}

	if fillableTarget, ok := target.(fillable); ok {
		if err := fillableTarget.FillInKeys(); err != nil {
			return nil, fmt.Errorf("error filling in fields: %v", err)
		}
	}

	return mergedDiagnostics, nil
}

func getDiagnosticMap(r readResult) map[string]string {
	if v, ok := r.(diagnosticReadResult); ok {
		return v.diagnosticMap
	}
	return nil
}
