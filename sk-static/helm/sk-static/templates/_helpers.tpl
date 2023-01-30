

{{/*
Expand the name of the chart.
*/}}
{{- define "sk-static.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "sk-static.fullname" -}}
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
{{- define "sk-static.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "sk-static.labels" -}}
helm.sh/chart: {{ include "sk-static.chart" . }}
{{ include "sk-static.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sk-static.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sk-static.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "sk-static.certificateName" -}}
{{- default (printf "%s" (include "sk-static.fullname" .)) .Values.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "sk-static.certificateSecretName" -}}
{{- default (printf "%s-cert" (include "sk-static.fullname" .)) .Values.certificateSecretName }}
{{- end }}

{{/*
Create the name of the deployment
*/}}
{{- define "sk-static.deploymentName" -}}
{{- default (printf "%s" (include "sk-static.fullname" .)) .Values.deploymentName }}
{{- end }}

{{/*
Create the name of the configuration configmap
*/}}
{{- define "sk-static.configName" -}}
{{- default (printf "%s-config" (include "sk-static.fullname" .)) .Values.configName }}
{{- end }}

{{/*
Create the name of the usersDb configmap
*/}}
{{- define "sk-static.usersDbName" -}}
{{- default (printf "%s-users" (include "sk-static.fullname" .)) .Values.usersDbName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "sk-static.serviceName" -}}
{{- default (printf "%s" (include "sk-static.fullname" .)) .Values.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "sk-static.ingressName" -}}
{{- default (printf "%s" (include "sk-static.fullname" .)) .Values.ingressName }}
{{- end }}

