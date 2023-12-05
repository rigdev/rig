{{- define "gvList" -}}
{{- $groupVersions := . -}}
---
custom_edit_url: null
---

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

<hr class="solid" />


:::info generated from source code
This page is generated based on go source code. If you have suggestions for
improvements for this page, please open an issue at
[github.com/rigdev/rig](https://github.com/rigdev/rig/issues/new), or a pull
request with changes to [the go source
files](https://github.com/rigdev/rig/tree/main/pkg/api).
:::

{{- end -}}
