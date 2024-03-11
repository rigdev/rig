{{- define "gvList" -}}
## Config
{{ $groupVersions := . -}}
{{- range $groupVersions -}}
{{- template "gvDetails" . -}}
{{- end -}}
{{- end -}}
