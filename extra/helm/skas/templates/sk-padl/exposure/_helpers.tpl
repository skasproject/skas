
{{/*
Create the name of the cert-manager certificate
*/}}
{{- define "skPadl.certificateName" -}}
{{- default (printf "%s-padl" (include "skas.fullname" .)) .Values.skPadl.exposure.certificateName }}
{{- end }}

{{/*
Create the name of the secret hosting the server certificate
*/}}
{{- define "skPadl.certificateSecretName" -}}
{{- default (printf "%s-padl-cert" (include "skas.fullname" .)) .Values.skPadl.exposure.certificateSecretName }}
{{- end }}

{{/*
Create the name of the service
*/}}
{{- define "skPadl.serviceName" -}}
{{- default (printf "%s-padl" (include "skas.fullname" .)) .Values.skPadl.exposure.serviceName }}
{{- end }}


{{/*
Create the name of the loadBalancer
*/}}
{{- define "skPadl.loadBalancerName" -}}
{{- default (printf "%s-padl-lb" (include "skas.fullname" .)) .Values.skPadl.exposure.loadBalancerName }}
{{- end }}


{{/*
Compute the binding ldap port for pod (Must be a non system port)
*/}}
{{- define "skPadl.pod.port" -}}
{{- "6363" }}
{{- end }}
