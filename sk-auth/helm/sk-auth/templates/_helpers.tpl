

{{/*
Expand the name of the chart.
*/}}
{{- define "sk-auth.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Define the namespace hosting users and groupbinding definition
*/}}
{{- define "sk-auth.tokenNamespace" -}}
{{- default .Release.Namespace  .Values.tokenConfig.namespace }}
{{- end }}


{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "sk-auth.fullname" -}}
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
{{- define "sk-auth.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "sk-auth.labels" -}}
helm.sh/chart: {{ include "sk-auth.chart" . }}
{{ include "sk-auth.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sk-auth.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sk-auth.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "sk-auth.certificateName" -}}
{{- default (printf "%s" (include "sk-auth.fullname" .)) .Values.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "sk-auth.certificateSecretName" -}}
{{- default (printf "%s-cert" (include "sk-auth.fullname" .)) .Values.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "sk-auth.serviceAccountName" -}}
{{- default (include "sk-auth.fullname" .) .Values.serviceAccountName }}
{{- end }}

{{/*
Create the name of the deployment
*/}}
{{- define "sk-auth.deploymentName" -}}
{{- default (printf "%s" (include "sk-auth.fullname" .)) .Values.deploymentName }}
{{- end }}

{{/*
Create the name of the configuration configmap
*/}}
{{- define "sk-auth.configName" -}}
{{- default (printf "%s-config" (include "sk-auth.fullname" .)) .Values.configName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "sk-auth.serviceName" -}}
{{- default (printf "%s" (include "sk-auth.fullname" .)) .Values.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "sk-auth.ingressName" -}}
{{- default (printf "%s" (include "sk-auth.fullname" .)) .Values.ingressName }}
{{- end }}

{{/*
Create the name of the cluster role for the server to access token namespaces when not same as deployment namespace
*/}}
{{- define "sk-auth.clusterRoleName" -}}
{{- default (printf "skas:%s-%s" .Release.Namespace (include "sk-auth.fullname" .)) .Values.clusterRoleName }}
{{- end }}

{{/*
Create the name of the Role for the server to access token when in same namespace
*/}}
{{- define "sk-auth.roleName" -}}
{{- default (printf "%s" (include "sk-auth.fullname" .)) .Values.roleName }}
{{- end }}

{{/*
Create the name of the Role for a manager to access tokens
*/}}
{{- define "sk-auth.editorRoleName" -}}
{{- default (printf "%s-editor" (include "sk-auth.fullname" .)) .Values.editorRoleName }}
{{- end }}

{{/*
Compute the serverUrl for the kubeconfig.User.AuthServerUrl (Automatic client configuration)
*/}}
{{- define "sk-auth.kubeconfig.authServerUrl" -}}
{{- if .Values.ingress.enabled }}
{{- default (printf "https://%s" .Values.ingress.host) .Values.kubeconfig.user.authServerUrl }}
{{- else }}
{{- .Values.kubeconfig.user.authServerUrl }}
{{- end }}
{{- end }}
