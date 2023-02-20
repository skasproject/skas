


{{/*
Create the name of the configuration configmap
*/}}
{{- define "skAuth.configmapName" -}}
{{- default (printf "%s-auth-config" (include "skas.fullname" .)) .Values.skAuth.configmapName }}
{{- end }}


{{/*
Create the name of the role
*/}}
{{- define "skAuth.roleName" -}}
{{- default (printf "%s-auth" (include "skas.fullname" .)) .Values.skAuth.roleName }}
{{- end }}


{{/*
Compute the serverUrl for the kubeconfig.User.AuthServerUrl (Automatic client configuration)
*/}}
{{- define "skAuth.kubeconfig.authServerUrl" -}}
{{- if .Values.skAuth.exposure.ingress.enabled }}
{{- default (printf "https://%s" .Values.skAuth.exposure.ingress.host) .Values.skAuth.kubeconfig.user.authServerUrl }}
{{- else }}
{{- .Values.skAuth.kubeconfig.user.authServerUrl }}
{{- end }}
{{- end }}
