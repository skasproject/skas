


{{/*
Create the name of the configuration configmap
*/}}
{{- define "skPadl.configmapName" -}}
{{- default (printf "%s-padl-config" (include "skas.fullname" .)) .Values.skPadl.configmapName }}
{{- end }}
