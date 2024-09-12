{{/*
Rig Server config file
*/}}
{{- define "rig-platform.config" -}}
apiVersion: config.rig.dev/v1alpha1
kind: PlatformConfig

{{- with .Values.rig -}}
{{- if or (and .auth.certificateFile .auth.certificateKeyFile) .auth.sso.oidcProviders .auth.allowRegister .auth.requireVerification .auth.sendWelcomeEmail }}
auth:
  allowRegister: {{ .auth.allowRegister }}
  requireVerification: {{ .auth.requireVerification }}
  sendWelcomeEmail: {{ .auth.sendWelcomeEmail }}
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

  {{- if .client.mailjets }}
  mailjets:
    {{- range $id, $provider := .client.mailjets }}
      {{ $id | quote }}:
        apiKey: {{ $provider.apiKey | quote }}
    {{- end }}
  {{- end }}

  {{- if .client.smtps }}
  smtps:
    {{- range $id, $provider := .client.smtps }}
      {{ $id | quote }}:
        host: {{ $provider.host | quote }}
        port: {{ $provider.port }}
        username: {{ $provider.username | quote }}
    {{- end}}
  {{- end }}

  {{- if .client.operator.baseUrl }}
  operator:
    baseUrl: {{ .client.operator.baseUrl }}
  {{- end }}

  {{- if .client.git }}
  git:
    author: {{ .client.git.author | toYaml | nindent 6 }}
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

{{- if .dockerRegistries }}
dockerRegistries:
{{- range $host, $registry := .dockerRegistries }}
  {{ $host | quote }}:
    username: {{ $registry.username | quote }}
    script: {{ $registry.script | quote}}
    expire: {{ $registry.expire | quote}}
{{- end }}
{{- end}}

{{- if .email.id }}
email:
  id: {{ .email.id | quote }}
  from: {{ .email.from | quote }}
{{- end }}
logging:
  level: {{ .logging.level | quote }}
  {{- if .logging.devMode }}
  devMode: {{ .logging.devMode }}
  {{- end }}

{{- with .capsuleExtensions }}
{{ toYaml . }}
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

  {{- if .mailjets }}
  mailjets:
    {{- range $id, $provider := .mailjets }}
      {{ $id | quote }}:
        secretKey: {{ $provider.secretKey | quote }}
    {{- end }}
  {{- end }}

  {{- if .smtps }}
  smtps:
    {{- range $id, $provider := .smtps }}
      {{ $id | quote }}:
        password: {{ $provider.password | quote }}
    {{- end}}
  {{- end }}

  {{- if .slack }}
  slack:
  {{- range $workspace, $provider := .slack }}
    {{ $workspace | quote }}:
      token: {{ $provider.token | quote }}
  {{- end }}
  {{- end }}

  {{- with .git }}
  {{- with .auths }}
  git:
    auths:
    {{- range . }}
      - {{ . | toYaml | indent 8 | trim }}
    {{- end }}
  {{- end }}
  {{- end }}
{{- end }}

{{- if .repository.secret }}
repository:
  secret: {{ .repository.secret | quote }}
{{- end }}

{{- if .dockerRegistries }}
dockerRegistries:
{{- range $host, $registry := .dockerRegistries }}
  {{ $host | quote }}:
    password: {{ $registry.password | quote}}
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

