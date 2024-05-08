


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
Create the name of the token editor role
*/}}
{{- define "skAuth.editorRoleName" -}}
{{- default (printf "%s-auth-edit" (include "skas.fullname" .)) .Values.skAuth.editorRoleName }}
{{- end }}


{{/*
Compute the serverUrl for the kubeconfig.User.AuthServerUrl (Automatic client configuration)
*/}}
{{- define "skAuth.kubeconfig.authServerUrl" -}}
{{- if .Values.skAuth.exposure.external.ingress.enabled }}
{{- default (printf "https://%s" .Values.skAuth.exposure.external.ingress.host) .Values.skAuth.kubeconfig.user.authServerUrl }}
{{- else }}
{{- .Values.skAuth.kubeconfig.user.authServerUrl }}
{{- end }}
{{- end }}

{{/*
The namespace to store tokens
*/}}
{{- define "skAuth.tokenNamespace" -}}
{{- default .Release.Namespace .Values.skAuth.tokenNamespace }}
{{- end }}
