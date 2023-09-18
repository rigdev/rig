{{/*
Rig Server config file
*/}}
{{- define "rig.config" -}}
auth:
  jwt:
    secret: shhhdonotshare
repository:
  secret:
    mongodb:
      key: thisisasecret
client:
  {{- with .Values.rig.client.mongo }}
  mongo:
    {{- if $.Values.mongodb.enabled }}
    host: "{{ include "rig.fullname" $ }}-mongodb:27017"
    {{- else }}
    host: {{ .host | quote }}
    {{- end }}
    user: {{ .user | quote }}
    password: {{ .password | quote }}
  {{- end }}
  {{- with .Values.rig.client.minio }}
  minio:
    endpoint: {{ .endpoint | quote }}
    secure: {{ .secure }}
    access_key_id: {{ .access_key_id | quote }}
    secret_access_key: {{ .secret_access_key | quote }}
  {{- end }}
cluster:
  type: k8s
  {{- with .Values.rig.cluster.dev_registry }}
  dev_registry:
    enabled: {{ .enabled }}
    host: {{ .host }}
    cluster_host: {{ .cluster_host }}
  {{- end }}
{{- with .Values.rig.email }}
email:
  type: {{ .type | quote }}
{{- end }}
{{- with .Values.rig.registry }}
registry:
  enabled: {{ .enabled }}
  port: {{ .port }}
  log_level: {{ .log_level }}
{{- end }}
{{- end -}}
