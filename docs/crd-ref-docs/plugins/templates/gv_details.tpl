{{- define "gvDetails" -}}
{{- $gv := . -}}
{{ range $gv.SortedTypes }}
{{ template "type" . }}
{{ end }}
{{- end -}}
