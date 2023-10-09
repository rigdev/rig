{{/*
Render configfile
*/}}
{{- define "rig-operator.config" -}}
apiVersion: config.rig.dev/v1alpha1
kind: OperatorConfig
{{ .Values.config | toYaml }}
{{- end }}
