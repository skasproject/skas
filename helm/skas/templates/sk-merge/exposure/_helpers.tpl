
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skMerge.certificateName" -}}
{{- default (printf "%s-merge" (include "skas.fullname" .)) .Values.skMerge.exposure.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skMerge.certificateSecretName" -}}
{{- default (printf "%s-merge-cert" (include "skas.fullname" .)) .Values.skMerge.exposure.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skMerge.serviceName" -}}
{{- default (printf "%s-merge" (include "skas.fullname" .)) .Values.skMerge.exposure.serviceName }}
{{- end }}

{{/*
Create the name of the ingress
*/}}
{{- define "skMerge.ingressName" -}}
{{- default (printf "%s-merge" (include "skas.fullname" .)) .Values.skMerge.exposure.ingressName }}
{{- end }}


