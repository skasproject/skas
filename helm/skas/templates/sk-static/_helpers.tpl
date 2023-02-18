


{{/*
Create the name of the configuration configmap
*/}}
{{- define "skStatic.configmapName" -}}
{{- default (printf "%s-static-config" (include "skas.fullname" .)) .Values.skStatic.configmapName }}
{{- end }}


{{/*
Create the name of the usersDb configmap
*/}}
{{- define "skStatic.usersDbName" -}}
{{- default (printf "%s-users" (include "skas.fullname" .)) .Values.skStatic.usersDbName }}
{{- end }}

