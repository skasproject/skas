


{{/*
Create the name of the configuration configmap
*/}}
{{- define "skLdap.configmapName" -}}
{{- default (printf "%s-ldap-config" (include "skas.fullname" .)) .Values.skLdap.configmapName }}
{{- end }}
