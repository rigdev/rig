{{/*
Expand the name of the chart.
*/}}
{{- define "rig-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "rig-operator.fullname" -}}
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
{{- define "rig-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "rig-operator.labels" -}}
helm.sh/chart: {{ include "rig-operator.chart" . }}
{{ include "rig-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "rig-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rig-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "rig-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "rig-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the config secret to use
*/}}
{{- define "rig-operator.secretName" -}}
{{- default (include "rig-operator.fullname" .) .Values.secretName }}
{{- end }}

{{/*
Create the fullname of apicheck resources
*/}}
{{- define "rig-operator.apicheck.fullname" -}}
{{- include "rig-operator.fullname" . | printf "%s-apicheck" -}}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "rig-operator.apicheck.serviceAccountName" -}}
{{- if .Values.apicheck.serviceAccount.create }}
{{- default (include "rig-operator.apicheck.fullname" .) .Values.apicheck.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.apicheck.serviceAccount.name }}
{{- end }}
{{- end }}
