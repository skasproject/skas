
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skCrd.certificateName" -}}
{{- default (printf "%s-crd" (include "skas.fullname" .)) .Values.skCrd.exposure.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skCrd.certificateSecretName" -}}
{{- default (printf "%s-crd-cert" (include "skas.fullname" .)) .Values.skCrd.exposure.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skCrd.serviceName" -}}
{{- default (printf "%s-crd" (include "skas.fullname" .)) .Values.skCrd.exposure.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "skCrd.ingressName" -}}
{{- default (printf "%s-crd" (include "skas.fullname" .)) .Values.skCrd.exposure.ingressName }}
{{- end }}


