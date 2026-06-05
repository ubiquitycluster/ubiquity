{{- define "nicoRest.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nicoRest.fullname" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nicoRest.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "nicoRest.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: nvidia-infra-controller
{{- end -}}

{{- define "nicoRest.provenanceAnnotations" -}}
ubiquity.dev/source-repo: NVIDIA/infra-controller
ubiquity.dev/source-component: nico-rest
ubiquity.dev/source-status: experimental-preview
{{- end -}}

{{- define "nicoRest.image" -}}
{{- printf "%s/%s:%s" (.root.Values.image.registry | trimSuffix "/") .repository (.root.Values.image.tag | default .root.Chart.AppVersion) -}}
{{- end -}}
