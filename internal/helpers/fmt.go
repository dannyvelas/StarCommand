package helpers

import (
	"strings"
	"text/template"
)

const sliceListTemplate = `
{{- range . }}
- {{ . }}
{{- end -}}
`

var parsedSliceListTemplate = template.Must(template.New("list").Parse(sliceListTemplate))

func StringSliceToBulletedList(items []string) string {
	var sb strings.Builder

	if err := parsedSliceListTemplate.Execute(&sb, items); err != nil {
		panic(err)
	}

	return sb.String()
}
