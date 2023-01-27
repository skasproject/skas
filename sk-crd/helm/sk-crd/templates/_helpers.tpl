

{{/*
Expand the name of the chart.
*/}}
{{- define "sk-crd.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Define the namespace hosting users and groupbinding definition
*/}}
{{- define "sk-crd.userdbNamespace" -}}
{{- default .Release.Namespace  .Values.userdbNamespace }}
{{- end }}


{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "sk-crd.fullname" -}}
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
{{- define "sk-crd.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "sk-crd.labels" -}}
helm.sh/chart: {{ include "sk-crd.chart" . }}
{{ include "sk-crd.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "sk-crd.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sk-crd.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "sk-crd.certificateName" -}}
{{- default (printf "%s" (include "sk-crd.fullname" .)) .Values.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "sk-crd.certificateSecretName" -}}
{{- default (printf "%s-cert" (include "sk-crd.fullname" .)) .Values.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "sk-crd.serviceAccountName" -}}
{{- default (include "sk-crd.fullname" .) .Values.serviceAccountName }}
{{- end }}

{{/*
Create the name of the deployment
*/}}
{{- define "sk-crd.deploymentName" -}}
{{- default (printf "%s" (include "sk-crd.fullname" .)) .Values.deploymentName }}
{{- end }}

{{/*
Create the name of the configuration configmap
*/}}
{{- define "sk-crd.configName" -}}
{{- default (printf "%s-config" (include "sk-crd.fullname" .)) .Values.configName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "sk-crd.serviceName" -}}
{{- default (printf "%s" (include "sk-crd.fullname" .)) .Values.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "sk-crd.ingressName" -}}
{{- default (printf "%s" (include "sk-crd.fullname" .)) .Values.ingressName }}
{{- end }}

{{/*
Create the name of the cluster role for the server to access userdb namespaces when not same as deployment namespace
*/}}
{{- define "sk-crd.clusterRoleName" -}}
{{- default (printf "skas:%s-%s" .Release.Namespace (include "sk-crd.fullname" .)) .Values.clusterRoleName }}
{{- end }}

{{/*
Create the name of the Role for the server to access userdb when in same namespace
*/}}
{{- define "sk-crd.roleName" -}}
{{- default (printf "%s" (include "sk-crd.fullname" .)) .Values.roleName }}
{{- end }}

{{/*
Create the name of the Role for a manager to access userdb
*/}}
{{- define "sk-crd.editorRoleName" -}}
{{- default (printf "%s-editor" (include "sk-crd.fullname" .)) .Values.editorRoleName }}
{{- end }}
