package app

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/dannyvelas/starcommand/config"
)

type reflectField struct {
	Type  reflect.StructField
	Value reflect.Value
}

const (
	statusMissing = "missing"
	statusLoaded  = "loaded"
)

// configLoader is implemented by all config structs that can be loaded from a
// *config.Config, validated for required fields, and finalized with FillInKeys.
type configLoader interface {
	FillFromConfig(c *config.Config) error
	FillInKeys() error
}

// loadConfig fills cfg from c, builds a diagnostics map for all required fields,
// and runs FillInKeys. If preflight is true, it returns the diagnostics without
// running FillInKeys. Returns an error if any required fields are missing.
func loadConfig(configLoader configLoader, c *config.Config) (map[string]string, error) {
	if err := configLoader.FillFromConfig(c); err != nil {
		return nil, err
	}

	m, err := buildDiagnostics(configLoader)
	if err != nil {
		return nil, fmt.Errorf("internal error building diagnostics: %v", err)
	}

	return m, nil
}

// buildDiagnostics returns a map where each key is the json/yaml tag name of a
// required field, and each value is statusLoaded if the field is non-zero or `statusMissing`
// if it is the zero value. Embedded structs are handled recursively.
func buildDiagnostics(v any) (map[string]string, error) {
	diagnostics := make(map[string]string)

	tagToFieldMap, err := getTagToFieldMap(v, "conflux", "json")
	if err != nil {
		return nil, fmt.Errorf("error getting tag to field map: %v", err)
	}

	for tag, field := range tagToFieldMap {
		if _, ok := field.Type.Tag.Lookup("required"); !ok {
			continue
		}

		if field.Value.IsZero() {
			diagnostics[tag] = statusMissing
		} else {
			diagnostics[tag] = statusLoaded
		}
	}

	return diagnostics, nil
}

// getTagToFieldMap takes a struct and returns a map where each key is
// the value of tag `tagName`. each value is a reflect.Value.
// if `tagName` is not found, it will iterate through `fallbackTags` until it finds a value
func getTagToFieldMap(v any, tagName string, fallbackTags ...string) (map[string]reflectField, error) {
	rv := reflect.ValueOf(v)

	// If a pointer is passed, get the underlying element (the actual struct)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	// If it's not a struct, we can't look up tags
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct as argument")
	}

	tagToFieldMap := make(map[string]reflectField)

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		foundTag := queryForTags(field, tagName, fallbackTags)

		tagToFieldMap[foundTag] = reflectField{field, rv.Field(i)}
	}

	return tagToFieldMap, nil
}

func queryForTags(field reflect.StructField, tagName string, fallbackTags []string) string {
	for i := range len(fallbackTags) + 1 {
		foundTag := field.Tag.Get(tagName)
		if foundTag != "" && foundTag != "-" {
			return strings.Split(foundTag, ",")[0]
		} else if i == len(fallbackTags) {
			break
		}
		tagName = fallbackTags[i]
	}
	return field.Name
}

func hasMissingFields(m map[string]string) bool {
	for _, v := range m {
		if v == statusMissing {
			return true
		}
	}
	return false
}

// diagnosticsToTable takes a diagnostic map and returns it as a pretty-printed formatted table
// This is useful as a user-friendly report of missing and found configuration values
func diagnosticsToTable(data map[string]string) string {
	headerCol1 := "SUBJECT"
	headerCol2 := "STATUS"

	maxKeyLen := len(headerCol1)
	maxValueLen := len(headerCol2)

	for k, v := range data {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
		if len(v) > maxValueLen {
			maxValueLen = len(v)
		}
	}

	type tableContext struct {
		Data       map[string]string
		HeaderCol1 string
		HeaderCol2 string
		KeyWidth   int
		ValueWidth int
		Line       string
	}

	// "| " (2) + maxKeyLen + " | " (3) + maxValueLen + " " (1)
	totalLineLength := maxKeyLen + maxValueLen + 6
	line := strings.Repeat("-", totalLineLength)

	ctx := tableContext{
		Data:       data,
		HeaderCol1: headerCol1,
		HeaderCol2: headerCol2,
		KeyWidth:   maxKeyLen,
		ValueWidth: maxValueLen,
		Line:       line,
	}

	const tableTmpl = `{{ .Line }}
| {{ printf "%-*s" .KeyWidth .HeaderCol1 }} | {{ printf "%-*s" .ValueWidth .HeaderCol2 }} |
{{ .Line }}
{{- range $key, $val := .Data }}
| {{ printf "%-*s" $.KeyWidth $key }} | {{ printf "%-*s" $.ValueWidth $val }} |
{{- end }}
{{ .Line }}
`

	tmpl, err := template.New("table").Parse(tableTmpl)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		panic(err)
	}

	return buf.String()
}
