{{- define "kubevirt-vms.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubevirt-vms.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "kubevirt-vms.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubevirt-vms.labels" -}}
app.kubernetes.io/name: {{ include "kubevirt-vms.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: ubiquity-virtual-machines
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
ubiquity.ai/os-profile: {{ .Values.vm.os | quote }}
ubiquity.ai/network-isolation: {{ .Values.vm.networkIsolation | quote }}
{{- end -}}

{{- define "kubevirt-vms.osProfile" -}}
{{- $profile := index .Values.vm.osProfiles .Values.vm.os -}}
{{- if not $profile -}}
{{- fail (printf "unknown vm.os %q; configure vm.osProfiles for this operating system" .Values.vm.os) -}}
{{- end -}}
{{- end -}}
