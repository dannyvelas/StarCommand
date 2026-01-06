package config

import (
	"bytes"
	"strings"
	"text/template"
)

// fmtTable takes a map and returns a formatted string table.
func fmtTable(data map[string]string) string {
	// 1. Calculate the maximum length of the keys
	maxKeyLen := 3 // Minimum width to fit the "KEY" header
	for k := range data {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	// 2. Create a data structure to pass both the map and the width
	type tableContext struct {
		Data  map[string]string
		Width int
		Line  string
	}

	// Create a horizontal line based on the dynamic width
	// (maxKeyLen + 3 for padding/borders + "STATUS" length)
	line := strings.Repeat("-", maxKeyLen+20)

	ctx := tableContext{
		Data:  data,
		Width: maxKeyLen,
		Line:  line,
	}

	// 3. Update the template to use the dynamic Width
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
	err = tmpl.Execute(&buf, ctx)
	if err != nil {
		panic(err)
	}

	return buf.String()
}
