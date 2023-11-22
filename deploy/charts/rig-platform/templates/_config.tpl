{{/*
Rig Server config file
*/}}
{{- define "rig-platform.config" -}}
apiVersion: config.rig.dev/v1alpha1
kind: PlatformConfig

{{- with .Values.rig -}}
{{- if and .auth.certificateFile .auth.certificateKeyFile }}
auth:
  certificateFile: {{ .auth.certificateFile | quote }}
  certificateKeyFile: {{ .auth.certificateKeyFile | quote }}
{{- end }}

port: {{ $.Values.port }}

{{- if $.Values.ingress.enabled }}
publicUrl: {{ printf "https://%s" $.Values.ingress.host | quote }}
{{- end }}

telemetryEnabled: {{ .telemetryEnabled }}

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

  {{- if .client.mailjet.apiKey }}
  mailjet:
    apiKey: {{ .client.mailjet.apiKey | quote }}
  {{- end }}

  {{- if .client.smtp.host }}
  smtp:
    host: {{ .client.smtp.host | quote }}
    port: {{ .client.smtp.port }}
    username: {{ .client.smtp.username | quote }}
  {{- end }}

  {{- if .client.operator.baseUrl }}
  operator:
    baseUrl: {{ .client.operator.baseUrl }}
  {{- end }}

cluster:
  type: k8s
  {{- if .cluster.devRegistry.host }}
  devRegistry:
    host: {{ .cluster.devRegistry.host | quote }}
    clusterHost: {{ default .cluster.devRegistry.host .cluster.devRegistry.clusterHost | quote }}
  {{- end }}

  {{- if .cluster.git.url }}
  git:
    {{- with .cluster.git }}
    url: {{ .url | quote }}
    branch: {{ .branch | quote }}
    {{- if .credentials.pathOrefix }}
    pathPrefix: {{ .credentials.pathPrefix | quote }}
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
  {{- if .logging.devMode }}
  devMode: {{ .logging.devMode }}
  {{- end }}

{{- end }}
{{- end -}}

{{/*
Rig Server secret config
*/}}
{{- define "rig-platform.config-secret" -}}
apiVersion: config.rig.dev/v1alpha1
kind: PlatformConfig

{{- with .Values.rig -}}
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
  {{- if .mailjet.secretKey }}
  mailjet:
    secretKey: {{ .mailjet.secretKey | quote }}
  {{- end }}
  {{- if .smtp.password }}
  smtp:
    password: {{ .smtp.password | quote }}
  {{- end }}
{{- end }}

{{- if .repository.secret }}
repository:
  secret: {{ .repository.secret | quote }}
{{- end }}

{{- with .cluster.git.credentials }}
{{- if or .https.password .ssh.privateKey }}
cluster:
  git:
    credentials:
      {{- if .https.password }}
      https:
        password: {{ .https.password | quote }}
      {{- end }}
      {{- if .ssh.privateKey }}
      ssh:
        privateKey: {{ .ssh.privateKey | quote }}
        {{- if .ssh.privateKeyPassword }}
        privateKeyPassword: {{ .ssh.privateKeyPassword | quote }}
        {{- end }}
      {{- end }}
{{- end }}
{{- end }}
auth:
  {{- if not (and .auth.certificateFile .auth.certificateKeyFile) }}
  secret: {{ .auth.secret | quote }}
  {{- end }}

{{- end -}}
{{- end -}}
