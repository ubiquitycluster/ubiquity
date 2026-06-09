package aiplatform

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestActivePlatformChartsIncludeValuesSchemas(t *testing.T) {
	root := filepath.Join("..", "..", "platform")
	var charts []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == "disabled" {
			return filepath.SkipDir
		}
		if entry.IsDir() || entry.Name() != "Chart.yaml" {
			return nil
		}
		charts = append(charts, filepath.Dir(path))
		return nil
	})
	if err != nil {
		t.Fatalf("walk platform charts: %v", err)
	}
	if len(charts) == 0 {
		t.Fatal("expected active platform charts")
	}
	for _, chart := range charts {
		t.Run(strings.TrimPrefix(chart, root+string(filepath.Separator)), func(t *testing.T) {
			schemaPath := filepath.Join(chart, "values.schema.json")
			content := mustRead(t, schemaPath)
			var schema map[string]any
			if err := json.Unmarshal([]byte(content), &schema); err != nil {
				t.Fatalf("values.schema.json must parse as JSON: %v", err)
			}
			if schema["$schema"] == "" {
				t.Fatal("values.schema.json must declare $schema")
			}
			if schema["type"] != "object" {
				t.Fatalf("values.schema.json must describe object values, got %v", schema["type"])
			}
		})
	}
}
