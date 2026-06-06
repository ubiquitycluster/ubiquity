package cloud

import (
	"strings"
	"testing"
)

func TestEvaluateCloudReadinessFailsClosedWhenEvidenceIsMissing(t *testing.T) {
	result := EvaluateCloudReadiness(CloudReadinessEvidence{
		RequiredCRDs: []string{"datavolumes.cdi.kubevirt.io", "schedules.k8up.io"},
		PresentCRDs:  []string{"datavolumes.cdi.kubevirt.io"},
		Resources: []CloudResourceEvidence{
			{Kind: "DataVolume", Namespace: "tenant-a", Name: "ubuntu", Conditions: []CloudCondition{{Type: "Bound", Status: "True"}}},
			{Kind: "Schedule", Namespace: "tenant-a", Name: "daily"},
		},
		SmokeTests: map[string]bool{"restore-drill": false},
	})
	if result.Ready {
		t.Fatalf("missing CRDs, status, and smoke tests must fail closed: %+v", result)
	}
	for _, required := range []string{"missing CRD schedules.k8up.io", "Schedule tenant-a/daily has no conditions", "smoke test restore-drill did not pass"} {
		if !strings.Contains(strings.Join(result.Reasons, "\n"), required) {
			t.Fatalf("expected reason %q in %#v", required, result.Reasons)
		}
	}
}

func TestEvaluateCloudReadinessPassesOnlyWithCRDsReadyConditionsAndSmokeTests(t *testing.T) {
	result := EvaluateCloudReadiness(CloudReadinessEvidence{
		RequiredCRDs: []string{"datavolumes.cdi.kubevirt.io", "schedules.k8up.io"},
		PresentCRDs:  []string{"schedules.k8up.io", "datavolumes.cdi.kubevirt.io"},
		Resources: []CloudResourceEvidence{
			{Kind: "DataVolume", Namespace: "tenant-a", Name: "ubuntu", Conditions: []CloudCondition{{Type: "Ready", Status: "True"}}},
			{Kind: "Schedule", Namespace: "tenant-a", Name: "daily", Conditions: []CloudCondition{{Type: "Available", Status: "True"}}},
		},
		SmokeTests: map[string]bool{"restore-drill": true, "network-deny": true},
	})
	if !result.Ready {
		t.Fatalf("expected readiness with complete evidence, got %#v", result)
	}
	if len(result.Reasons) != 0 {
		t.Fatalf("ready result should not carry failure reasons: %#v", result.Reasons)
	}
}

func TestRenderCloudReadinessReportIsReviewerReadableAndMachineParseable(t *testing.T) {
	report := RenderCloudReadinessReport(EvaluateCloudReadiness(CloudReadinessEvidence{
		RequiredCRDs: []string{"datavolumes.cdi.kubevirt.io"},
		PresentCRDs:  []string{},
		SmokeTests:   map[string]bool{"vm-boot": false},
	}))
	for _, required := range []string{"ready: false", "missing CRD datavolumes.cdi.kubevirt.io", "smoke test vm-boot did not pass"} {
		if !strings.Contains(report, required) {
			t.Fatalf("report missing %q:\n%s", required, report)
		}
	}
}
