package cloud

import (
	"strings"
	"testing"
)

func TestRenderCloudProductionAuditChecklistCoversFailClosedEvidence(t *testing.T) {
	checklist := RenderCloudProductionAuditChecklist()
	for _, required := range []string{
		"# Ubiquity cloud production readiness audit checklist",
		"server-side dry-run",
		"collect-readiness",
		"required CRDs",
		"operator install-plan",
		"air-gap artifacts",
		"restore drill",
		"not object existence",
		"KubeVirt image catalog",
		"standalone disk attachments",
		"managed service readiness resources",
		"ready: true",
	} {
		if !strings.Contains(checklist, required) {
			t.Fatalf("audit checklist missing %q:\n%s", required, checklist)
		}
	}
}

func TestCloudProductionAuditChecklistIncludesEveryManagedServiceResource(t *testing.T) {
	checklist := RenderCloudProductionAuditChecklist()
	for _, resource := range AllManagedServiceReadinessResources() {
		if !strings.Contains(checklist, resource) {
			t.Fatalf("audit checklist missing readiness resource %q", resource)
		}
	}
}
