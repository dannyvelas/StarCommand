package helpers

import (
	"fmt"
	"reflect"
)

type ReflectField struct {
	Type  reflect.StructField
	Value reflect.Value
}

// GetTagToFieldMap takes a struct and returns a map where each key is
// the value of tag `tagName`. each value is a reflect.Value.
// if `tagName` is not found, it will iterate through `fallbackTags` until it finds a value
func GetTagToFieldMap(v any, tagName string, fallbackTags ...string) (map[string]ReflectField, error) {
	rv := reflect.ValueOf(v)

	// If a pointer is passed, get the underlying element (the actual struct)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	// If it's not a struct, we can't look up tags
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct as argument")
	}

	tagToFieldMap := make(map[string]ReflectField)

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		foundTag := queryForTags(field, tagName, fallbackTags)
		if foundTag == "" {
			return nil, fmt.Errorf("field %s is missing a tag", field.Name)
		}

		tagToFieldMap[foundTag] = ReflectField{field, rv.Field(i)}
	}

	return tagToFieldMap, nil
}

func queryForTags(field reflect.StructField, tagName string, fallbackTags []string) string {
	for i := range len(fallbackTags) + 1 {
		foundTag := field.Tag.Get(tagName)
		if foundTag != "" {
			return foundTag
		} else if i == len(fallbackTags) {
			return ""
		}
		tagName = fallbackTags[i]
	}
	return ""
}
