package config

import (
	"bytes"
	"html/template"
	"strings"
)

// DiagnosticsToTable takes a diagnostic map and returns it as a pretty-printed formatted table
// This is useful as a user-friendly report of missing and found configuration values
func DiagnosticsToTable(data map[string]string) string {
	// calculate the maximum length of the keys
	maxKeyLen := 3 // Minimum width to fit the "KEY" header
	for k := range data {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	// create a data structure to pass both the map and the width
	type tableContext struct {
		Data  map[string]string
		Width int
		Line  string
	}

	// create a horizontal line based on the dynamic width
	line := strings.Repeat("-", maxKeyLen+20)

	ctx := tableContext{
		Data:  data,
		Width: maxKeyLen,
		Line:  line,
	}

	// make template use the dynamic Width
	// We use printf with a dynamic precision: %-*s
	// The '*' tells printf to get the width from the next argument.
	const tableTmpl = `{{ .Line }}
| {{ printf "%-*s" .Width "KEY" }} | STATUS
{{ .Line }}
{{- range $key, $val := .Data }}
| {{ printf "%-*s" $.Width $key }} | {{ $val }}
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
