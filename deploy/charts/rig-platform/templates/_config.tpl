{{/*
Rig Server config file
*/}}
{{- define "rig-platform.config" -}}
auth:
  jwt:
    secret: shhhdonotshare
repository:
  secret:
    mongodb:
      key: thisisasecret
client:
  {{- if or .Values.rig.client.postgres.host .Values.postgres.enabled }}
  postgres:
    {{- if .Values.postgres.enabled }}
    host: "{{ include "rig-platform.fullname" . }}-postgres:5432"
    insecure: true
    {{- else }}
    host: {{ .Values.rig.client.postgres.host | quote }}
    insecure: {{ .Values.rig.client.postgres.insecure }}
    {{- end }}
    user: {{ .Values.rig.client.postgres.user | quote }}
    password: {{ .Values.rig.client.postgres.password | quote }}
  {{- end }}
  {{- if or .Values.rig.client.mongo.host .Values.mongodb.enabled }}
  mongo:
    {{- if .Values.mongodb.enabled }}
    host: "{{ include "rig-platform.fullname" . }}-mongodb:27017"
    {{- else }}
    host: {{ .Values.rig.client.mongo.host | quote }}
    {{- end }}
    user: {{ .Values.rig.client.mongo.user | quote }}
    password: {{ .Values.rig.client.mongo.password | quote }}
  {{- end }}
  {{- with .Values.rig.client.minio }}
  minio:
    endpoint: {{ .endpoint | quote }}
    secure: {{ .secure }}
    access_key_id: {{ .access_key_id | quote }}
    secret_access_key: {{ .secret_access_key | quote }}
  {{- end }}
  {{- if .Values.rig.client.operator.base_url }}
  operator:
    base_url: {{ .Values.rig.client.operator.base_url }}
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
