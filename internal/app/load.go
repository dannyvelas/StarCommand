package app

import (
	"bytes"
	"reflect"
	"strings"
	"text/template"
)

const (
	statusMissing = "missing"
	statusLoaded  = "loaded"
)

// buildStructDiagnostics walks v (a struct) using yaml tags for naming and records whether
// required fields are zero or not into diagnostics. Nested structs are recursed into automatically.
// Only fields tagged with `required:"true"` are validated; nested struct fields are always recursed into.
func buildStructDiagnostics(v any, prefix string, diagnostics map[string]string) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldVal := rv.Field(i)

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		fieldPrefix := prefix + "." + strings.Split(yamlTag, ",")[0]

		if fieldVal.Kind() == reflect.Struct {
			buildStructDiagnostics(fieldVal.Interface(), fieldPrefix, diagnostics)
			continue
		}

		if field.Tag.Get("required") != "true" {
			continue
		}

		if fieldVal.IsZero() {
			diagnostics[fieldPrefix] = statusMissing
		} else {
			diagnostics[fieldPrefix] = statusLoaded
		}
	}
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
