
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skStatic.certificateName" -}}
{{- default (printf "%s-static" (include "skas.fullname" .)) .Values.skStatic.exposure.external.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skStatic.certificateSecretName" -}}
{{- default (printf "%s-static-cert" (include "skas.fullname" .)) .Values.skStatic.exposure.external.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skStatic.serviceName" -}}
{{- default (printf "%s-static" (include "skas.fullname" .)) .Values.skStatic.exposure.external.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "skStatic.ingressName" -}}
{{- default (printf "%s-static" (include "skas.fullname" .)) .Values.skStatic.exposure.external.ingressName }}
{{- end }}


