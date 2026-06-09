package aiplatform

import (
	"strings"
	"testing"
)

func TestNICoDay2LifecycleProofScriptIsGatedAndCoversRequiredOperations(t *testing.T) {
	script := mustRead(t, "../../test/e2e/nico-day2-lifecycle-proof.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_NICO_DAY2",
		"--dry-run",
		"ubiquity health --nico",
		"ubiquity nodes enroll",
		"ubiquity nodes inspect",
		"ubiquity nodes image",
		"ubiquity nodes reboot",
		"ubiquity nodes cordon",
		"ubiquity nodes drain",
		"ubiquity nodes maintenance",
		"ubiquity nodes status reconcile",
		"bmcStatus",
		"kubeletStatus",
		"gpuStatus",
		"rdmaStatus",
		"firmwareStatus",
		"imageStatus",
		"maintenanceState",
		"nico-day2-lifecycle-proof-passed",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("script missing %q", required)
		}
	}
}
