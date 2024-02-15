{{/*
Expand the name of the chart.
*/}}
{{- define "rig-platform.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "rig-platform.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "rig-platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "rig-platform.labels" -}}
helm.sh/chart: {{ include "rig-platform.chart" . }}
{{ include "rig-platform.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "rig-platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rig-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: rig
{{- end }}

{{/*
Mongodb Common labels
*/}}
{{- define "rig-platform.mongodb.labels" -}}
helm.sh/chart: {{ include "rig-platform.chart" . }}
{{ include "rig-platform.mongodb.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Mongodb Selector labels
*/}}
{{- define "rig-platform.mongodb.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rig-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: mongodb
{{- end }}

{{/*
Postgres Common labels
*/}}
{{- define "rig-platform.postgres.labels" -}}
helm.sh/chart: {{ include "rig-platform.chart" . }}
{{ include "rig-platform.postgres.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Postgres Selector labels
*/}}
{{- define "rig-platform.postgres.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rig-platform.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: postgres
{{- end }}

{{/*
Create the name of the config secret
*/}}
{{- define "rig-platform.secretName" -}}
{{- default (include "rig-platform.fullname" .) .Values.secretName }}
{{- end }}
