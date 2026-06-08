package aiplatform

import (
	"strings"
	"testing"
)

func TestAIStoreDataPlaneSmokeScriptIsGatedAndRecordsMarkers(t *testing.T) {
	script := mustRead(t, "../../test/e2e/aistore-data-plane-smoke.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_AISTORE_SMOKE=true",
		"ais",
		"kubectl",
		"aistore-target-storage-proven",
		"aistore-bucket-smoke-test-passed",
		"aistore-gpu-artifact-read-passed",
		"aistore-metrics-proven",
		"not a generic PVC replacement",
		"exit 0",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("AIStore smoke script missing %q", required)
		}
	}
}
