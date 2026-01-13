package helpers

import (
	"fmt"
	"reflect"
	"strings"
)

type ReflectField struct {
	Type  reflect.StructField
	Value reflect.Value
}

// FromMap takes a map[string]string and writes it to "dst"
// "dst" could either be a map[string]string or a struct with only string fields
// if "dst" is a map[string]string, then entries in "src" are copied to "dst"
func FromMap(src map[string]string, dst any) error {
	val := reflect.ValueOf(dst)

	// We must have a pointer to a struct, or a pointer to a map to be able to set values
	if val.Kind() != reflect.Pointer || (val.Elem().Kind() != reflect.Struct && val.Elem().Kind() != reflect.Map) {
		return fmt.Errorf("target must be a pointer to a map[string]string or a pointer to a struct, got %T", dst)
	}

	// Dereference the pointer
	dstVal := val.Elem()

	// if destination is map
	if dstVal.Kind() == reflect.Map {
		for k, v := range src {
			dstVal.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
		}
		return nil
	}

	// otherwise, as struct
	structType := dstVal.Type()

	// create a normalized/lowercase version of the source map
	// this is important for matching to struct tags
	normalizedSrc := make(map[string]string, len(src))
	for k, v := range src {
		normalizedSrc[strings.ToLower(k)] = v
	}

	// iterate through all fields of the struct
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := dstVal.Field(i)

		// we can only set the field if it is capitalized (exported) and of type string
		if !fieldVal.CanSet() || field.Type.Kind() != reflect.String {
			continue
		}

		// get tag value
		configTag := queryForTags(field, "bw", []string{"json"})
		if configTag == "" {
			continue
		}

		// If the tag exists as a key in our source map, set the field
		if val, exists := normalizedSrc[strings.ToLower(configTag)]; exists {
			fieldVal.SetString(val)
		}
	}

	return nil
}

// ToMap takes a struct as an argument and converts it to a map[string]string
func ToMap(src any) (map[string]string, error) {
	val := reflect.ValueOf(src)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("source must be a struct, got %T", src)
	}

	dest := make(map[string]string)
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fVal := val.Field(i)

		// Only process exported string fields
		if !fVal.CanInterface() || f.Type.Kind() != reflect.String {
			continue
		}

		tag := queryForTags(f, "bw", []string{"json"})
		if tag != "" {
			dest[tag] = fVal.String()
		}
	}
	return dest, nil
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
