package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestPlatformOpsPolicyChartIncludesRestoreDrillContract(t *testing.T) {
	for _, path := range []string{
		"../../platform/platform-ops-policy/values.yaml",
		"../../platform/platform-ops-policy/templates/policy.yaml",
		"../../platform/platform-ops-policy/values.schema.json",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(content)
		for _, required := range []string{"restoreDrill", "restore-object-rendered-not-restore-proof"} {
			if !strings.Contains(text, required) {
				t.Fatalf("%s missing %q", path, required)
			}
		}
	}
}
