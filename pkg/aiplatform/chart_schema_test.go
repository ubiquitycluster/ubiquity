package aiplatform

import (
	"encoding/json"
	"path/filepath"
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
