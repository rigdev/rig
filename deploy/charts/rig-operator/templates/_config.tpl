{{/*
Render configfile
*/}}
{{- define "rig-operator.config" -}}
apiVersion: config.rig.dev/v1alpha1
kind: OperatorConfig

webhooksEnabled: {{ .Values.config.webhooksEnabled }}
devModeEnabled: {{ .Values.config.devModeEnabled }}
leaderElectionEnabled: {{ .Values.config.leaderElectionEnabled }}
{{- end }}
