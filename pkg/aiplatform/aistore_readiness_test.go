package aiplatform

import "testing"

func TestEvaluateAIStoreReadinessFailsClosedUntilAllEvidenceExists(t *testing.T) {
	status := EvaluateAIStoreReadiness(AIStoreSnapshot{
		OperatorReady:         true,
		CRDsEstablished:       true,
		ClusterReady:          true,
		ProxyPodsReady:        true,
		TargetPodsReady:       true,
		TargetPVCsBound:       true,
		BucketSmokeTestPassed: true,
	})

	if status.Ready {
		t.Fatal("AIStore readiness should fail closed without GPU workload read evidence")
	}
	check := status.ChecksByName()["aistore-gpu-artifact-read"]
	if check.Ready || check.Message == "" {
		t.Fatalf("expected missing GPU artifact read evidence to be explicit, got %#v", check)
	}
}

func TestEvaluateAIStoreReadinessPassesWithCompleteEvidence(t *testing.T) {
	status := EvaluateAIStoreReadiness(AIStoreSnapshot{
		OperatorReady:         true,
		CRDsEstablished:       true,
		ClusterReady:          true,
		ProxyPodsReady:        true,
		TargetPodsReady:       true,
		TargetPVCsBound:       true,
		BucketSmokeTestPassed: true,
		GPUArtifactReadPassed: true,
		MetricsAvailable:      true,
	})

	if !status.Ready {
		t.Fatalf("expected AIStore readiness with complete evidence, got %#v", status.Checks)
	}
}
