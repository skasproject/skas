
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skLdap.certificateName" -}}
{{- default (printf "%s-ldap" (include "skas.fullname" .)) .Values.skLdap.exposure.external.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skLdap.certificateSecretName" -}}
{{- default (printf "%s-ldap-cert" (include "skas.fullname" .)) .Values.skLdap.exposure.external.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skLdap.serviceName" -}}
{{- default (printf "%s-ldap" (include "skas.fullname" .)) .Values.skLdap.exposure.external.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "skLdap.ingressName" -}}
{{- default (printf "%s-ldap" (include "skas.fullname" .)) .Values.skLdap.exposure.external.ingressName }}
{{- end }}


