package helpers

import (
	"strings"
	"text/template"
)

const mapListTemplate = `
{{- range $k, $v := . }}
- {{ $k }}: {{ $v }}
{{- end -}}
`

const sliceListTemplate = `
{{- range . }}
- {{ . }}
{{- end -}}
`

var (
	parsedMapListTemplate   = template.Must(template.New("list").Parse(mapListTemplate))
	parsedSliceListTemplate = template.Must(template.New("list").Parse(sliceListTemplate))
)

func MapToBulletedList(m map[string]string) string {
	var sb strings.Builder

	if err := parsedMapListTemplate.Execute(&sb, m); err != nil {
		panic(err)
	}

	return sb.String()
}

func StringSliceToBulletedList(items []string) string {
	var sb strings.Builder

	if err := parsedSliceListTemplate.Execute(&sb, items); err != nil {
		panic(err)
	}

	return sb.String()
}
