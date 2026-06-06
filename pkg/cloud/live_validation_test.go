package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestCloudServerDryRunScriptFailsClosedAndCoversAllPrimitives(t *testing.T) {
	content, err := os.ReadFile("../../test/e2e/cloud-primitives-server-dry-run.sh")
	if err != nil {
		t.Fatalf("missing cloud e2e script: %v", err)
	}
	script := string(content)
	for _, required := range []string{
		"set -euo pipefail",
		"kubectl apply --dry-run=server -f -",
		"cloud render vm-disk", "cloud render vpc", "cloud render tenant-cluster", "cloud render service", "cloud render backup-policy", "cloud render prerequisites",
		"require_crd", "datavolumes.cdi.kubevirt.io", "schedules.k8up.io",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("script missing %q", required)
		}
	}
}

func TestCloudReadinessRunbookDocumentsReconciliationNotObjectExistence(t *testing.T) {
	content, err := os.ReadFile("../../docs/runbooks/cloud-readiness-validation.md")
	if err != nil {
		t.Fatalf("missing cloud readiness runbook: %v", err)
	}
	runbook := string(content)
	for _, required := range []string{"server-side dry-run", "CRD presence", "status condition", "restore drill", "fail closed", "not object existence"} {
		if !strings.Contains(runbook, required) {
			t.Fatalf("runbook missing %q", required)
		}
	}
}
