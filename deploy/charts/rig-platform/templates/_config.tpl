{{/*
Rig Server config file
*/}}
{{- define "rig-platform.config" -}}
apiVersion: config.rig.dev/v1alpha1
kind: PlatformConfig

{{- with .Values.rig -}}
{{- if or (and .auth.certificateFile .auth.certificateKeyFile) .auth.sso.oidcProviders }}
auth:
  {{- if (and .auth.certificateFile .auth.certificateKeyFile) }}
  certificateFile: {{ .auth.certificateFile | quote }}
  certificateKeyFile: {{ .auth.certificateKeyFile | quote }}
  {{- end }}
  {{- if .auth.sso.oidcProviders }}
  sso:
    oidcProviders:
    {{- range $id, $provider := .auth.sso.oidcProviders }}
      {{ $id | quote }}:
        {{- if $provider.name }}
        name: {{ $provider.name | quote }}
        {{- end }}
        {{- if $provider.icon }}
        icon: {{ $provider.icon | quote }}
        {{- end }}
        issuerURL: {{ $provider.issuerURL | quote }}
        clientID: {{ $provider.clientID | quote }}
        {{- with $provider.allowedDomains }}
        allowedDomains: {{ . | toYaml | nindent 8 }}
        {{- end }}
        {{- with $provider.scopes }}
        scopes: {{ . | toYaml | nindent 8 }}
        {{- end }}
        {{- if .groupsClaim }}
        groupsClaim: {{ .groupsClaim | quote }}
        {{- end }}
        {{- if .disableJITGroups }}
        disableJITGroups: {{ .disableJITGroups }}
        {{- end }}
        {{- with .groupMapping }}
        groupMapping: {{ . | toYaml | nindent 8 }}
        {{- end }}
    {{- end }}
  {{- end }}
{{- end }}

port: {{ $.Values.port }}

{{- if $.Values.ingress.enabled }}
publicURL: {{ printf "https://%s" $.Values.ingress.host | quote }}
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
    database: {{ .client.postgres.database | quote }}
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


environments:
{{- if .environments }}
{{ toYaml .environments | indent 2 }}
{{- else }}
  prod:
    cluster: prod
{{- end }}

clusters:
{{- if .clusters }}
{{ toYaml .clusters | indent 2 }}
{{- else }}
  prod:
    type: k8s
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

{{- if and .cluster .cluster.git }}
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
{{- end }}
auth:
  {{- if not (and .auth.certificateFile .auth.certificateKeyFile) }}
  secret: {{ .auth.secret | quote }}
  {{- end }}
  {{- if .auth.sso.oidcProviders }}
  sso:
    oidcProviders:
    {{- range $name, $provider := .auth.sso.oidcProviders }}
      {{ $name | quote }}:
        clientSecret: {{ $provider.clientSecret | quote }}
    {{- end }}
    {{- end }}
{{- end -}}
{{- end -}}
