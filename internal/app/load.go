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

func hasErrors(m map[string]string) bool {
	for _, v := range m {
		if v != statusLoaded {
			return true
		}
	}
	return false
}

// setDiagnostic records whether val is the zero value for its type.
func setDiagnostic(diagnostics map[string]string, key string, val any) {
	if reflect.ValueOf(val).IsZero() {
		diagnostics[key] = statusMissing
	} else {
		diagnostics[key] = statusLoaded
	}
}

// DiagnosticsToTable takes a diagnostic map and returns it as a pretty-printed formatted table
// This is useful as a user-friendly report of missing and found configuration values
func DiagnosticsToTable(data map[string]string) string {
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

func mapsCopyWithPrefix(dest, src map[string]string, prefix string) {
	for k, v := range src {
		dest[prefix+k] = v
	}
}
