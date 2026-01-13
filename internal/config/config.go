package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dannyvelas/homelab/internal/helpers"
)

const (
	statusMissing = "missing"
	statusLoaded  = "loaded"
)

type config interface {
	// Validate receives a map of validation results where each element corresponds to a key in the config
	// the second return value will be false if at least one key was invalid. otherwise, it will be true
	Validate(map[string]string) bool
}

type fillableConfig interface {
	// FillInKeys takes the keys that are required and uses them to fill out remaining config fields
	FillInKeys() error
}

func validateConfig(v any) (map[string]string, error) {
	results := make(map[string]string)
	valid := true

	tagToFieldMap, err := helpers.GetTagToFieldMap(v, "bw", "json")
	if err != nil {
		return nil, fmt.Errorf("error getting tag to field map: %v", err)
	}

	for tag, field := range tagToFieldMap {
		if _, ok := field.Type.Tag.Lookup("required"); !ok {
			continue
		}

		if field.Value.IsZero() {
			results[tag] = statusMissing
			valid = false
		} else {
			results[tag] = statusLoaded
		}
	}

	if config, ok := v.(config); ok {
		valid = valid && config.Validate(results)
	}

	if !valid {
		return results, ErrInvalidFields
	}

	return results, nil
}

func UnmarshalInto(r unvalidatedReader, target any) error {
	m, err := r.ReadUnvalidated()
	if errors.Is(err, ErrInvalidFields) {
		return err
	} else if err != nil {
		return fmt.Errorf("error reading: %v", err)
	}

	if err := decode(m, target); err != nil {
		return fmt.Errorf("error converting map into target: %v", err)
	}

	return nil
}

func decode(src, dest any) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("error marshalling map: %v", err)
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("error unmarshalling map into target: %v", err)
	}
	return nil
}
