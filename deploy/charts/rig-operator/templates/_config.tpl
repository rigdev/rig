{{/*
Render configfile
*/}}
{{- define "rig-operator.config" -}}
apiVersion: config.rig.dev/v1alpha1
kind: OperatorConfig
webhooksEnabled: {{ .Values.config.webhooksEnabled }}
devModeEnabled: {{ .Values.config.devModeEnabled }}
leaderElectionEnabled: {{ .Values.config.leaderElectionEnabled }}
{{- if .Values.config.certManager.clusterIssuer }}
certManager:
  clusterIssuer: {{ .Values.config.certManager.clusterIssuer | quote }}
  createCertificateResources: {{ .Values.config.certManager.createCertificateResources }}
{{- end }}
{{- end }}
