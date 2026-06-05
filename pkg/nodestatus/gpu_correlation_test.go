package nodestatus

import "testing"

func TestCorrelateGPUEvidenceIncompleteWhenNICoGPUHasNoKubernetesGPUOrMIG(t *testing.T) {
	got := CorrelateGPUEvidence("cn01", 8, 0, nil, "provisioned")
	if got.State != CorrelationIncomplete {
		t.Fatalf("state = %s, want %s: %+v", got.State, CorrelationIncomplete, got)
	}
}

func TestCorrelateGPUEvidenceDegradedWhenKubernetesGPUHasUnhealthyNICoMachine(t *testing.T) {
	got := CorrelateGPUEvidence("cn02", 8, 8, nil, "unhealthy")
	if got.State != CorrelationDegraded {
		t.Fatalf("state = %s, want %s: %+v", got.State, CorrelationDegraded, got)
	}
}

func TestCorrelateGPUEvidenceAcceptsMIGAsKubernetesAcceleratorEvidence(t *testing.T) {
	got := CorrelateGPUEvidence("cn03", 8, 0, map[string]int{"nvidia.com/mig-1g.10gb": 7}, "provisioned")
	if got.State != CorrelationReady {
		t.Fatalf("state = %s, want %s: %+v", got.State, CorrelationReady, got)
	}
}

func TestCorrelateGPUStatusUsesSeparateNICoAndKubernetesCounts(t *testing.T) {
	got := CorrelateGPUStatus(NodeStatus{Name: "cn04", NICoGPUs: 8, KubernetesGPUs: 0, MachineStatus: "provisioned"})
	if got.State != CorrelationIncomplete {
		t.Fatalf("state = %s, want %s: %+v", got.State, CorrelationIncomplete, got)
	}
}
