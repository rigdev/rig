{{- define "type_members" -}}
{{- $field := . -}}
{{ markdownRenderFieldDoc $field.Doc }}
{{- end -}}
