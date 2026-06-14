package aiplatform

import (
	"os"
	"strings"
	"testing"
)

func TestLegacyDCGMExporterTemplatesAreDisabledByDefault(t *testing.T) {
	for _, path := range []string{
		"../../system/monitoring-system/templates/dgcm-exporter.yaml",
		"../../monitoring/monitoring-system/templates/dgcm-exporter.yaml",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		text := string(content)
		if !strings.Contains(text, "{{- if .Values.legacyDcgmExporter.enabled }}") {
			t.Fatalf("%s must gate the legacy hand-authored DCGM exporter behind legacyDcgmExporter.enabled", path)
		}
		if !strings.Contains(text, "{{- end }}") {
			t.Fatalf("%s must close the legacy DCGM exporter gate", path)
		}
	}
}

func TestGPUOperatorValuesEnableManagedDCGMExporter(t *testing.T) {
	content, err := os.ReadFile("../../system/nvidia-gpu-operator/values.yaml")
	if err != nil {
		t.Fatalf("failed to read GPU Operator values: %v", err)
	}
	text := string(content)
	for _, required := range []string{
		"dcgmExporter:",
		"enabled: true",
		"serviceMonitor:",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("GPU Operator values must include %q to make NVIDIA-managed DCGM exporter the default", required)
		}
	}
}
