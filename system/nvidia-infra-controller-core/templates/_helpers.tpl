{{- define "nicoCore.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nicoCore.fullname" -}}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "nicoCore.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "nicoCore.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: nvidia-infra-controller
{{- end -}}

{{- define "nicoCore.provenanceAnnotations" -}}
ubiquity.dev/source-repo: NVIDIA/infra-controller
ubiquity.dev/source-component: nico-core
ubiquity.dev/source-status: experimental-preview
{{- end -}}

{{- define "nicoCore.componentImage" -}}
{{- $root := .root -}}
{{- $component := .component -}}
{{- if $component.image.registry -}}
{{- printf "%s/%s:%s" ($component.image.registry | trimSuffix "/") $component.image.repository ($component.image.tag | default $root.Chart.AppVersion) -}}
{{- else if $root.Values.global.imageRegistry -}}
{{- printf "%s/%s:%s" ($root.Values.global.imageRegistry | trimSuffix "/") $component.image.repository ($component.image.tag | default $root.Chart.AppVersion) -}}
{{- else -}}
{{- printf "%s:%s" $component.image.repository ($component.image.tag | default $root.Chart.AppVersion) -}}
{{- end -}}
{{- end -}}
