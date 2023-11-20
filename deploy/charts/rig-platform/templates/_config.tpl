{{/*
Rig Server config file
*/}}
{{- define "rig-platform.config" -}}
{{- with .Values.rig -}}
{{- if and .auth.jwt.certificate_file .auth.jwt.certificate_key_file }}
auth:
  jwt:
    certificate_file: {{ .auth.jwt.certificate_file | quote }}
    certificate_key_file: {{ .auth.jwt.certificate_key_file | quote }}
{{- end }}
port: {{ $.Values.port }}
{{- if $.Values.ingress.enabled }}
public_url: {{ printf "https://%s" $.Values.ingress.host | quote }}
{{- end }}
telemetry:
  enabled: {{ .telemetry.enabled }}
client:
  {{- if or .client.postgres.host $.Values.postgres.enabled }}
  postgres:
    {{- if $.Values.postgres.enabled }}
    host: "{{ include "rig-platform.fullname" $ }}-postgres:5432"
    insecure: true
    {{- else }}
    host: {{ .client.postgres.host | quote }}
    insecure: {{ .client.postgres.insecure }}
    {{- end }}
    user: {{ .client.postgres.user | quote }}
  {{- end }}

  {{- if or .client.mongo.host $.Values.mongodb.enabled }}
  mongo:
    {{- if $.Values.mongodb.enabled }}
    host: "{{ include "rig-platform.fullname" $ }}-mongodb:27017"
    {{- else }}
    host: {{ .client.mongo.host | quote }}
    {{- end }}
    user: {{ .client.mongo.user | quote }}
  {{- end }}

  {{- if .client.minio.host }}
  minio:
    host: {{ .client.minio.host | quote }}
    endpoint: {{ .client.minio.endpoint | quote }}
    secure: {{ .client.minio.secure }}
  {{- end }}

  {{- if .client.mailjet.api_key }}
  mailjet:
    api_key: {{ .client.mailjet.api_key | quote }}
  {{- end }}

  {{- if .client.smtp.host }}
  smtp:
    host: {{ .client.smtp.host | quote }}
    port: {{ .client.smtp.port }}
    username: {{ .client.smtp.username | quote }}
  {{- end }}

  {{- if .client.operator.base_url }}
  operator:
    base_url: {{ .client.operator.base_url }}
  {{- end }}
repository:
  capsule: {{ include "rig-platform.repository" $ | nindent 4 }}
  service_account: {{ include "rig-platform.repository" $ | nindent 4 }}
  group: {{ include "rig-platform.repository" $ | nindent 4 }}
  project: {{ include "rig-platform.repository" $ | nindent 4 }}
  cluster_config: {{ include "rig-platform.repository" $ | nindent 4 }}
  session: {{ include "rig-platform.repository" $ | nindent 4 }}
  user: {{ include "rig-platform.repository" $ | nindent 4 }}
  verification_code: {{ include "rig-platform.repository" $ | nindent 4 }}
  secret: {{ include "rig-platform.repository" $ | nindent 4 }}
cluster:
  type: k8s
  {{- if .cluster.dev_registry.host }}
  dev_registry:
    host: {{ .cluster.dev_registry.host | quote }}
    cluster_host: {{ default .cluster.dev_registry.host .cluster.dev_registry.cluster_host | quote }}
  {{- end }}
  {{- if .cluster.git.url }}
  git:
    {{- with .cluster.git }}
    url: {{ .url | quote }}
    branch: {{ .branch | quote }}
    {{- if .credentials.path_prefix }}
    path_prefix: {{ .credentials.path_prefix | quote }}
    {{- end }}
    {{- if .credentials.https.username }}
    credentials:
      https:
        username: {{ .credentials.https.username | quote }}
    {{- end }}
    {{- end }}
  {{- end }}

{{- if .email.type }}
email:
  type: {{ .email.type | quote }}
  from: {{ .email.from | quote }}
{{- end }}
logging:
  level: {{ .logging.level | quote }}
  {{- if .logging.dev_mode }}
  dev_mode: {{ .logging.dev_mode }}
  {{- end }}

{{- end }}
{{- end -}}

{{/*
Rig platform repository
*/}}
{{- define "rig-platform.repository" -}}
{{- if or .Values.rig.client.mongo.host .Values.mongodb.enabled -}}
store: "mongodb"
{{- else -}}
store: "postgres"
{{- end -}}
{{- end -}}

{{/*
Rig Server secret config
*/}}
{{- define "rig-platform.config-secret" -}}
{{- with .Values.rig -}}

auth:

  {{- with .auth.jwt }}
  jwt:
    {{- if not (and .certificate_file .certificate_key_file) }}
    secret: {{ .secret | quote }}
    {{- end }}
  {{- end }}

{{- with .client }}
client:
  {{- if .postgres.password }}
  postgres:
    password: {{ .postgres.password | quote }}
  {{- end }}
  {{- if .mongo.password }}
  mongo:
    password: {{ .mongo.password | quote }}
  {{- end }}
  {{- if .minio.secret_access_key }}
  minio:
    secret_access_key: {{ .minio.secret_access_key | quote }}
  {{- end }}
  {{- if .mailjet.secret_key }}
  mailjet:
    secret_key: {{ .mailjet.secret_key | quote }}
  {{- end }}
  {{- if .smtp.password }}
  smtp:
    password: {{ .smtp.password | quote }}
  {{- end }}
{{- end }}

{{- with .repository.secret }}
{{- if .key }}
repository:
  secret:
    {{- if $.Values.rig.client.mongo.host }}
    mongodb:
      key: {{ .key | quote }}
    {{- else }}
    postgres:
      key: {{ .key | quote }}
    {{- end }}
{{- end }}
{{- end }}

{{- with .cluster.git.credentials }}
{{- if or .https.password .ssh.private_key }}
cluster:
  git:
    credentials:
      {{- if .https.password }}
      https:
        password: {{ .https.password | quote }}
      {{- end }}
      {{- if .ssh.private_key }}
      ssh:
        private_key: {{ .ssh.private_key | quote }}
        {{- if .ssh.private_key_password }}
        private_key_password: {{ .ssh.private_key_password | quote }}
        {{- end }}
      {{- end }}
{{- end }}
{{- end }}

{{- end -}}
{{- end -}}
