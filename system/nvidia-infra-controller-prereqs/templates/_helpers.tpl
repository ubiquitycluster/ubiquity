{{- define "nicoPrereqs.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nicoPrereqs.fullname" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nicoPrereqs.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "nicoPrereqs.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: nvidia-infra-controller
{{- end -}}

{{- define "nicoPrereqs.provenanceAnnotations" -}}
ubiquity.dev/source-repo: NVIDIA/infra-controller
ubiquity.dev/source-component: nico-prereqs
ubiquity.dev/source-status: experimental-preview
{{- end -}}
