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

func TestCloudReadinessRequiresNamedSmokeTests(t *testing.T) {
	ev := readyCloudEvidence()
	ev.RequiredSmokeTests = []string{"postgres-connectivity", "restore-drill-readable"}
	ev.SmokeTests = map[string]bool{"postgres-connectivity": true}

	result := EvaluateCloudReadiness(ev)
	if result.Ready {
		t.Fatalf("expected missing required smoke test to fail closed")
	}
	if !containsReason(result.Reasons, "missing required smoke test restore-drill-readable") {
		t.Fatalf("expected missing smoke test reason, got %v", result.Reasons)
	}
}

func TestCloudReadinessRequiresRestoreDrillReadableSmoke(t *testing.T) {
	ev := readyCloudEvidence()
	ev.RequiredSmokeTests = RequiredCloudSmokeTests()
	ev.SmokeTests = map[string]bool{
		"postgres-connectivity":   true,
		"redis-connectivity":      true,
		"kafka-produce-consume":   true,
		"objectbucket-read-write": true,
	}

	result := EvaluateCloudReadiness(ev)
	if result.Ready {
		t.Fatalf("expected missing restore drill completion proof to fail closed")
	}
	if !containsReason(result.Reasons, "missing required smoke test restore-drill-readable") {
		t.Fatalf("expected restore drill smoke reason, got %v", result.Reasons)
	}
}

func TestCloudReadinessRequiresRestoreCompletionReadableDataAndMarker(t *testing.T) {
	ev := readyCloudEvidence()
	ev.RequiredSmokeTests = []string{"restore-drill-controller-succeeded", "restore-drill-readable", "cloud-restore-drill-smoke-passed"}
	ev.SmokeTests = map[string]bool{"restore-drill-controller-succeeded": true, "restore-drill-readable": true}

	result := EvaluateCloudReadiness(ev)
	if result.Ready {
		t.Fatalf("expected missing named restore drill marker to fail closed")
	}
	if !containsReason(result.Reasons, "missing required smoke test cloud-restore-drill-smoke-passed") {
		t.Fatalf("expected named restore marker reason, got %v", result.Reasons)
	}
	ev.SmokeTests["cloud-restore-drill-smoke-passed"] = true
	result = EvaluateCloudReadiness(ev)
	if !result.Ready {
		t.Fatalf("expected restore drill evidence to pass, got %v", result.Reasons)
	}
}

func TestCloudReadinessRequiresTenantClusterEvidence(t *testing.T) {
	ev := readyCloudEvidence()
	ev.RequiredCRDs = append(ev.RequiredCRDs, "clusters.cluster.x-k8s.io")
	ev.PresentCRDs = append(ev.PresentCRDs, "clusters.cluster.x-k8s.io")
	ev.Resources = append(ev.Resources, CloudResourceEvidence{Kind: "Cluster", Namespace: "tenant-a", Name: "tenant-a-dev", Conditions: []CloudCondition{{Type: "Ready", Status: "True"}}})
	ev.RequiredSmokeTests = []string{"tenant-cluster-kubeconfig-present", "tenant-cluster-api-reachable", "tenant-cluster-nodes-ready"}
	ev.SmokeTests = map[string]bool{"tenant-cluster-kubeconfig-present": true, "tenant-cluster-api-reachable": true}

	result := EvaluateCloudReadiness(ev)
	if result.Ready {
		t.Fatalf("expected missing tenant node readiness to fail closed")
	}
	if !containsReason(result.Reasons, "missing required smoke test tenant-cluster-nodes-ready") {
		t.Fatalf("expected tenant node readiness reason, got %v", result.Reasons)
	}
	ev.SmokeTests["tenant-cluster-nodes-ready"] = true
	result = EvaluateCloudReadiness(ev)
	if !result.Ready {
		t.Fatalf("expected tenant cluster evidence to pass, got %v", result.Reasons)
	}
}

func TestRequiredCloudSmokeTestsIncludeServiceAndRestoreMarkers(t *testing.T) {
	got := strings.Join(RequiredCloudSmokeTests(), "\n")
	for _, required := range append(AllManagedServiceSmokeTests(), "restore-drill-controller-succeeded", "restore-drill-readable", "cloud-restore-drill-smoke-passed", "tenant-cluster-kubeconfig-present", "tenant-cluster-api-reachable", "tenant-cluster-nodes-ready") {
		if !strings.Contains(got, required) {
			t.Fatalf("RequiredCloudSmokeTests missing %q in %s", required, got)
		}
	}
}

func TestCloudReadinessPassesWhenRequiredSmokeTestsPass(t *testing.T) {
	ev := readyCloudEvidence()
	ev.RequiredSmokeTests = RequiredCloudSmokeTests()
	ev.SmokeTests = map[string]bool{}
	for _, name := range ev.RequiredSmokeTests {
		ev.SmokeTests[name] = true
	}

	result := EvaluateCloudReadiness(ev)
	if !result.Ready {
		t.Fatalf("expected ready evidence with required smoke tests to pass, got %v", result.Reasons)
	}
}

func readyCloudEvidence() CloudReadinessEvidence {
	return CloudReadinessEvidence{
		RequiredCRDs: []string{"datavolumes.cdi.kubevirt.io"},
		PresentCRDs:  []string{"datavolumes.cdi.kubevirt.io"},
		Resources: []CloudResourceEvidence{
			{Kind: "DataVolume", Namespace: "tenant-a", Name: "ubuntu", Conditions: []CloudCondition{{Type: "Ready", Status: "True"}}},
		},
		SmokeTests: map[string]bool{},
	}
}

func containsReason(reasons []string, needle string) bool {
	return strings.Contains(strings.Join(reasons, "\n"), needle)
}
