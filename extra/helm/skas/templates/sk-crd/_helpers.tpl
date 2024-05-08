


{{/*
Create the name of the configuration configmap
*/}}
{{- define "skCrd.configmapName" -}}
{{- default (printf "%s-crd-config" (include "skas.fullname" .)) .Values.skCrd.configmapName }}
{{- end }}


{{/*
Create the name of the role
*/}}
{{- define "skCrd.roleName" -}}
{{- default (printf "%s-crd" (include "skas.fullname" .)) .Values.skCrd.roleName }}
{{- end }}


{{/*
Create the name of the user editor role
*/}}
{{- define "skCrd.editorRoleName" -}}
{{- default (printf "%s-crd-edit" (include "skas.fullname" .)) .Values.skCrd.editorRoleName }}
{{- end }}


{{/*
The namespace to store users and groupBinding
*/}}
{{- define "skCrd.userDbNamespace" -}}
{{- default .Release.Namespace .Values.skCrd.userDbNamespace }}
{{- end }}
