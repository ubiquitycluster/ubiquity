package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestCloudReadinessProofBundleScriptCoversEvidenceOutputs(t *testing.T) {
	content, err := os.ReadFile("../../test/e2e/cloud-readiness-proof-bundle.sh")
	if err != nil {
		t.Fatalf("read proof bundle script: %v", err)
	}
	script := string(content)
	for _, required := range []string{
		"--dry-run",
		"prerequisite contract",
		"operator provenance",
		"server-side dry-run output",
		"collected readiness JSON",
		"readiness report",
		"restore-drill evidence",
		"ubiquity cloud render prerequisites",
		"ubiquity cloud render operator-bundles",
		"ubiquity cloud apply service --dry-run",
		"ubiquity cloud collect-readiness",
		"ubiquity cloud readiness --readiness-file",
		"cloud-restore-drill-smoke-passed",
		"UBIQUITY_RUN_CLOUD_READINESS_PROOF",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("proof bundle script missing %q", required)
		}
	}
}
