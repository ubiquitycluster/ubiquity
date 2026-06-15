{{- define "ai-platform-console.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ai-platform-console.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- include "ai-platform-console.name" . -}}
{{- end -}}
{{- end -}}

{{- define "ai-platform-console.labels" -}}
app.kubernetes.io/name: {{ include "ai-platform-console.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: ubiquity-ai-platform
app.kubernetes.io/component: frontend
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end -}}
