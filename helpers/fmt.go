package helpers

import (
	"strings"
	"text/template"
)

const listTemplate = `
{{- range $k, $v := . }}
- {{ $k }}: {{ $v }}
{{- end }}
`

var parsedListTemplate = template.Must(template.New("list").Parse(listTemplate))

func MapToBulletedList(m map[string]string) string {
	var sb strings.Builder

	if err := parsedListTemplate.Execute(&sb, m); err != nil {
		panic(err)
	}

	return sb.String()
}
