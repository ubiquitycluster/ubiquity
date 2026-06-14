package aiplatform

import (
	"strings"
	"testing"
)

func TestKAISchedulerSmokeScriptIsGatedAndRecordsEvidence(t *testing.T) {
	script := mustRead(t, "../../test/e2e/kai-scheduler-smoke.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_KAI_SMOKE=true",
		"kai-scheduling-smoke-test-passed",
		"queues.scheduling.run.ai",
		"kubectl apply --server-side",
		"kubectl wait",
		"--dry-run",
		"fail closed",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("KAI smoke script missing %q", required)
		}
	}
}
