package virtualization

import "testing"

func TestEvaluateVMReadinessFailsClosedUntilCDIPVCVMAndGuestEvidenceExist(t *testing.T) {
	status := EvaluateVMReadiness(VMReadinessEvidence{
		Namespace:                  "tenant-a",
		Name:                       "ubuntu-dev",
		DataVolumeImportReady:      true,
		PersistentVolumeClaimBound: true,
		VirtualMachineReady:        true,
		VirtualMachineRunning:      true,
		GuestAgentReady:            false,
	})
	if status.Ready {
		t.Fatalf("VM readiness must fail closed without guest health evidence: %#v", status)
	}
	if !containsReason(status.Reasons, "guest health evidence missing for VirtualMachine tenant-a/ubuntu-dev") {
		t.Fatalf("expected guest health failure reason, got %#v", status.Reasons)
	}
}

func TestEvaluateVMReadinessPassesWithCompleteEvidence(t *testing.T) {
	status := EvaluateVMReadiness(VMReadinessEvidence{
		Namespace:                  "tenant-a",
		Name:                       "ubuntu-dev",
		DataVolumeImportReady:      true,
		PersistentVolumeClaimBound: true,
		VirtualMachineReady:        true,
		VirtualMachineRunning:      true,
		GuestAgentReady:            true,
	})
	if !status.Ready {
		t.Fatalf("expected VM readiness with CDI/PVC/VM/guest evidence, got %#v", status)
	}
}

func TestEvaluateVMReadinessReportsMissingControllerEvidence(t *testing.T) {
	status := EvaluateVMReadiness(VMReadinessEvidence{Namespace: "tenant-a", Name: "ubuntu-dev"})
	for _, expected := range []string{
		"CDI import/clone has not succeeded for DataVolume tenant-a/ubuntu-dev-root",
		"PVC is not bound for VM tenant-a/ubuntu-dev",
		"VirtualMachine tenant-a/ubuntu-dev is not Ready",
		"VirtualMachine tenant-a/ubuntu-dev is not Running",
		"guest health evidence missing for VirtualMachine tenant-a/ubuntu-dev",
	} {
		if !containsReason(status.Reasons, expected) {
			t.Fatalf("expected reason %q in %#v", expected, status.Reasons)
		}
	}
}

func containsReason(reasons []string, expected string) bool {
	for _, reason := range reasons {
		if reason == expected {
			return true
		}
	}
	return false
}
