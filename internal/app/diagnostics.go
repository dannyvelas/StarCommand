package app

import (
	"bytes"
	"reflect"
	"strings"
	"text/template"
)

const (
	statusLoaded     = "loaded"
	statusWillPrompt = "will prompt"
)

type diagnostic struct {
	Field  string
	Status string
}

type Diagnostics []diagnostic

var tableTmpl = template.Must(template.New("table").Parse(`{{ .Line }}
| {{ printf "%-*s" .KeyWidth .HeaderCol1 }} | {{ printf "%-*s" .ValueWidth .HeaderCol2 }} |
{{ .Line }}
{{- range .Data }}
| {{ printf "%-*s" $.KeyWidth .Field }} | {{ printf "%-*s" $.ValueWidth .Status }} |
{{- end }}
{{ .Line }}
`))

// ToTable returns a pretty-printed diagnostic table
// This is useful as a user-friendly report of missing and found configuration values
func (d *Diagnostics) ToTable() string {
	headerCol1 := "SUBJECT"
	headerCol2 := "STATUS"

	maxKeyLen := len(headerCol1)
	maxValueLen := len(headerCol2)

	for _, d := range *d {
		if len(d.Field) > maxKeyLen {
			maxKeyLen = len(d.Field)
		}
		if len(d.Status) > maxValueLen {
			maxValueLen = len(d.Status)
		}
	}

	type tableContext struct {
		Data       []diagnostic
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
		Data:       *d,
		HeaderCol1: headerCol1,
		HeaderCol2: headerCol2,
		KeyWidth:   maxKeyLen,
		ValueWidth: maxValueLen,
		Line:       line,
	}

	var buf bytes.Buffer
	_ = tableTmpl.Execute(&buf, ctx)

	return buf.String()
}

func (d *Diagnostics) hasErrors() bool {
	for _, d := range *d {
		if d.Status != statusLoaded {
			return true
		}
	}
	return false
}

func (d *Diagnostics) appendChecked(field string, val any) {
	if reflect.ValueOf(val).IsZero() {
		*d = append(*d, diagnostic{Field: field, Status: errNotFound.Error()})
	} else {
		*d = append(*d, diagnostic{Field: field, Status: statusLoaded})
	}
}

func (d *Diagnostics) appendWithPrefix(prefix string, src ...diagnostic) {
	for _, v := range src {
		*d = append(*d, diagnostic{Field: prefix + v.Field, Status: v.Status})
	}
}

func (d *Diagnostics) append(src ...diagnostic) {
	*d = append(*d, Diagnostics(src)...)
}

func (d *Diagnostics) Len() int {
	return len(*d)
}
