


{{/*
Create the name of the configuration configmap
*/}}
{{- define "skMerge.configmapName" -}}
{{- default (printf "%s-merge-config" (include "skas.fullname" .)) .Values.skMerge.configmapName }}
{{- end }}


