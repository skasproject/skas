
{{/*
Create the name of the configuration configmap
*/}}
{{- define "skHConf.configmapName" -}}
{{- default (printf "%s-hconf" (include "skas.fullname" .)) .Values.skHConf.configmapName }}
{{- end }}

{{/*
Create the name of the service account
*/}}
{{- define "skHConf.serviceAccountName" -}}
{{- default (printf "%s-hconf" (include "skas.fullname" .)) .Values.skHConf.serviceAccountName }}
{{- end }}

{{/*
Create the name of the clusterRole
*/}}
{{- define "skHConf.clusterRoleName" -}}
{{- default (printf "%s-hconf" (include "skas.fullname" .)) .Values.skHConf.clusterRoleName }}
{{- end }}

{{/*
Create the name of the monitor job
*/}}
{{- define "skHConf.monitorJobName" -}}
{{- default (printf "%s-job-hconf-monitor" (include "skas.fullname" .)) .Values.skHConf.monitorJobName }}
{{- end }}


