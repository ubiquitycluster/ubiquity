package aiplatform

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestNvidiaAIWrapperChartsIncludeValuesSchemas(t *testing.T) {
	charts := []string{
		"platform/kai-scheduler",
		"platform/nim-operator",
		"platform/stallscope",
		"system/nvidia-gpu-operator",
		"system/nvidia-network-operator",
	}

	for _, chart := range charts {
		t.Run(chart, func(t *testing.T) {
			path := filepath.Join("../..", chart, "values.schema.json")
			content := mustRead(t, path)
			var schema map[string]any
			if err := json.Unmarshal([]byte(content), &schema); err != nil {
				t.Fatalf("schema must be valid JSON: %v", err)
			}
			if schema["$schema"] == "" {
				t.Fatalf("schema %s missing $schema", path)
			}
			if schema["type"] != "object" {
				t.Fatalf("schema %s must describe an object, got %v", path, schema["type"])
			}
		})
	}
}

func TestAIWorkloadTenancyRejectsUnsafeTenantNames(t *testing.T) {
	schemaPath := filepath.Join("../..", "platform/ai-workload-tenancy/values.schema.json")
	content := mustRead(t, schemaPath)
	for _, required := range []string{
		`"maxLength": 63`,
		`"pattern": "^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$"`,
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("tenant schema missing %s", required)
		}
	}

	templatePath := filepath.Join("../..", "platform/ai-workload-tenancy/templates/gpu-quota.yaml")
	template := mustRead(t, templatePath)
	if strings.Contains(template, "namespace: {{ $tenant.name }}") || strings.Contains(template, "name: {{ $tenant.name }}") {
		t.Fatalf("tenant name is rendered without quote/fail guard in %s", templatePath)
	}
	for _, required := range []string{
		`regexMatch "^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$"`,
		`{{ $tenant.name | quote }}`,
	} {
		if !strings.Contains(template, required) {
			t.Fatalf("tenant template missing hardening snippet %s", required)
		}
	}
}
