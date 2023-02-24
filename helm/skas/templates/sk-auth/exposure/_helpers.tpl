
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skAuth.certificateName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.exposure.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skAuth.certificateSecretName" -}}
{{- default (printf "%s-auth-cert" (include "skas.fullname" .)) .Values.skAuth.exposure.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skAuth.serviceName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.exposure.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "skAuth.ingressName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.exposure.ingressName }}
{{- end }}


