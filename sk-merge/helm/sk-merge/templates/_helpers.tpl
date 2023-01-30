

{{/*
Expand the name of the chart.
*/}}
{{- define "sk-merge.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "sk-merge.fullname" -}}
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
{{- define "sk-merge.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "sk-merge.labels" -}}
helm.sh/chart: {{ include "sk-merge.chart" . }}
{{ include "sk-merge.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sk-merge.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sk-merge.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "sk-merge.certificateName" -}}
{{- default (printf "%s" (include "sk-merge.fullname" .)) .Values.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "sk-merge.certificateSecretName" -}}
{{- default (printf "%s-cert" (include "sk-merge.fullname" .)) .Values.certificateSecretName }}
{{- end }}

{{/*
Create the name of the deployment
*/}}
{{- define "sk-merge.deploymentName" -}}
{{- default (printf "%s" (include "sk-merge.fullname" .)) .Values.deploymentName }}
{{- end }}

{{/*
Create the name of the configuration configmap
*/}}
{{- define "sk-merge.configName" -}}
{{- default (printf "%s-config" (include "sk-merge.fullname" .)) .Values.configName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "sk-merge.serviceName" -}}
{{- default (printf "%s" (include "sk-merge.fullname" .)) .Values.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "sk-merge.ingressName" -}}
{{- default (printf "%s" (include "sk-merge.fullname" .)) .Values.ingressName }}
{{- end }}

