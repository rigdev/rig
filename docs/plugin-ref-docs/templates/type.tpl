{{- define "type" -}}
{{- $type := . -}}
{{- if markdownShouldRenderType $type -}}
{{- if ne $type.Name "Config" }}
### {{ $type.Name }}
{{- end }}
{{ if $type.IsAlias }}_Underlying type:_ _{{ markdownRenderTypeLink $type.UnderlyingType  }}_{{ end }}
{{ $type.Doc }}

{{ if $type.Members -}}
| Field | Description |
| --- | --- |
{{ range $type.Members -}}
| `{{ .Name  }}` _{{ markdownRenderType .Type }}_ | {{ template "type_members" . }} |
{{ end -}}

{{ end -}}

{{- end -}}
{{- end -}}
