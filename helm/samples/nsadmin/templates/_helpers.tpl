
{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nsadmin.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nsadmin.labels" -}}
helm.sh/chart: {{ include "nsadmin.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{/*
Create the adminName
*/}}
{{- define "nsadmin.adminName" -}}
{{- default (printf "%s-admin" .Release.Namespace ) .Values.adminNameOverride }}
{{- end }}

