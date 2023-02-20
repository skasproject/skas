

{{/*
Expand the name of the chart.
*/}}
{{- define "skas.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "skas.fullname" -}}
{{- if .Values.fullNameOverride }}
{{- .Values.fullNameOverride | trunc 63 | trimSuffix "-" }}
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
{{- define "skas.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "skas.labels" -}}
helm.sh/chart: {{ include "skas.chart" . }}
{{ include "skas.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{/*
Selector labels
*/}}
{{- define "skas.selectorLabels" -}}
app.kubernetes.io/name: {{ include "skas.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the deployment
*/}}
{{- define "skas.deploymentName" -}}
{{- default (printf "%s" (include "skas.fullname" .)) .Values.deploymentName }}
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "skas.serviceAccountName" -}}
{{- default (printf "%s" (include "skas.fullname" .)) .Values.serviceAccountName }}
{{- end }}


{{/*
Create list of config map for reloader.stakater.com
*/}}
{{- define "skas.watchedConfigmap" -}}
{{- if .Values.skMerge.enabled }}
{{- include "skMerge.configmapName" . }},
{{- end }}
{{- if .Values.skStatic.enabled }}
{{- include "skStatic.configmapName" . }},{{ include "skStatic.usersDbName" . }},
{{- end }}
{{- if .Values.skAuth.enabled }}
{{- include "skAuth.configmapName" . }},
{{- end }}
{{- end }}
