package virtualization

import "fmt"

// VMReadinessEvidence captures live KubeVirt/CDI/PVC/guest signals for one VM.
// Rendering a VM or image catalog does not populate this evidence.
type VMReadinessEvidence struct {
	Namespace                  string
	Name                       string
	BootDiskName               string
	DataVolumeImportReady      bool
	PersistentVolumeClaimBound bool
	VirtualMachineReady        bool
	VirtualMachineRunning      bool
	GuestAgentReady            bool
}

// VMReadinessStatus is a fail-closed VM readiness decision.
type VMReadinessStatus struct {
	Ready   bool
	Reasons []string
}

// EvaluateVMReadiness fails closed until CDI import/clone, PVC binding, VM phase,
// VM Ready condition, and guest health evidence are all present.
func EvaluateVMReadiness(ev VMReadinessEvidence) VMReadinessStatus {
	var reasons []string
	id := fmt.Sprintf("%s/%s", defaultString(ev.Namespace, "default"), defaultString(ev.Name, "unknown"))
	bootDisk := ev.BootDiskName
	if bootDisk == "" {
		bootDisk = defaultString(ev.Name, "unknown") + "-root"
	}
	if !ev.DataVolumeImportReady {
		reasons = append(reasons, fmt.Sprintf("CDI import/clone has not succeeded for DataVolume %s/%s", defaultString(ev.Namespace, "default"), bootDisk))
	}
	if !ev.PersistentVolumeClaimBound {
		reasons = append(reasons, fmt.Sprintf("PVC is not bound for VM %s", id))
	}
	if !ev.VirtualMachineReady {
		reasons = append(reasons, fmt.Sprintf("VirtualMachine %s is not Ready", id))
	}
	if !ev.VirtualMachineRunning {
		reasons = append(reasons, fmt.Sprintf("VirtualMachine %s is not Running", id))
	}
	if !ev.GuestAgentReady {
		reasons = append(reasons, fmt.Sprintf("guest health evidence missing for VirtualMachine %s", id))
	}
	return VMReadinessStatus{Ready: len(reasons) == 0, Reasons: reasons}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
