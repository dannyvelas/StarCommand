package config

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

var errInvalidFields = errors.New("")

// newErrInvalidFields takes a map and returns a formatted error
func newErrInvalidFields(data map[string]string) error {
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
	err = tmpl.Execute(&buf, ctx)
	if err != nil {
		panic(err)
	}

	return fmt.Errorf("%w%s", errInvalidFields, buf.String())
}
