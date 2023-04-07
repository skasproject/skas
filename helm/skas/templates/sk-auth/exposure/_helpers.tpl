
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skAuth.certificateName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.exposure.external.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skAuth.certificateSecretName" -}}
{{- default (printf "%s-auth-cert" (include "skas.fullname" .)) .Values.skAuth.exposure.external.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skAuth.serviceName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.exposure.external.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "skAuth.ingressName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.exposure.external.ingressName }}
{{- end }}


