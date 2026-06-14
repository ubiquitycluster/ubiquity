package nodestatus

import "strings"

type CorrelationState string

const (
	CorrelationReady      CorrelationState = "ready"
	CorrelationIncomplete CorrelationState = "incomplete"
	CorrelationDegraded   CorrelationState = "degraded"
)

type GPUCorrelation struct {
	NodeName string           `json:"nodeName"`
	State    CorrelationState `json:"state"`
	Reason   string           `json:"reason"`
}

func CorrelateGPUEvidence(nodeName string, nicoGPUs, kubernetesGPUs int, migProfiles map[string]int, machineStatus string) GPUCorrelation {
	k8sHasAccelerator := kubernetesGPUs > 0 || len(migProfiles) > 0
	nicoHasGPU := nicoGPUs > 0
	machineHealthy := isHealthyMachineStatus(machineStatus)
	if nicoHasGPU && !k8sHasAccelerator {
		return GPUCorrelation{NodeName: nodeName, State: CorrelationIncomplete, Reason: "NICo reports GPU hardware but Kubernetes exposes no nvidia.com/gpu or nvidia.com/mig-* allocatable resources"}
	}
	if k8sHasAccelerator && !machineHealthy {
		return GPUCorrelation{NodeName: nodeName, State: CorrelationDegraded, Reason: "Kubernetes exposes NVIDIA accelerator resources but NICo machine status is unhealthy"}
	}
	return GPUCorrelation{NodeName: nodeName, State: CorrelationReady, Reason: "NICo and Kubernetes GPU evidence are consistent"}
}

func CorrelateGPUStatus(status NodeStatus) GPUCorrelation {
	return CorrelateGPUEvidence(status.Name, status.NICoGPUs, status.KubernetesGPUs, status.MIGProfiles, status.MachineStatus)
}

func isHealthyMachineStatus(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return true
	}
	for _, bad := range []string{"unhealthy", "failed", "error", "degraded"} {
		if strings.Contains(s, bad) {
			return false
		}
	}
	return true
}
