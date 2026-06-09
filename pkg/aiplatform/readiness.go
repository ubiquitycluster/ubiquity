package aiplatform

import "fmt"

// ClusterSnapshot contains machine-checkable evidence collected from a cluster.
// Empty or missing evidence is treated as not-ready.
type ClusterSnapshot struct {
	GPUOperatorReady                 bool
	GPUDriverReady                   bool
	GPUContainerToolkitReady         bool
	GPUDevicePluginReady             bool
	GPUFeatureDiscoveryReady         bool
	GPUManagedDCGMExporterReady      bool
	GPUMIGManagerReady               bool
	GPUOperatorValidatorReady        bool
	DCGMMetricsScraped               bool
	GPUAllocatableByNode             map[string]int
	MIGProfilesByNode                map[string][]string
	MIGAllocatableByNode             map[string]map[string]int
	RDMAResourcesByNode              map[string]int
	NetworkAttachments               []string
	GPUCorrelationIssues             []string
	LastRDMASmokeTestPassed          bool
	NIMServicesReady                 int
	LastNIMSmokeTestPassed           bool
	KAISchedulerReady                bool
	KAIQueueReady                    bool
	LastKAISchedulingSmokeTestPassed bool
}

// CheckResult is one fail-closed readiness check.
type CheckResult struct {
	Name    string
	Ready   bool
	Message string
}

// ReadinessStatus is the aggregate NVIDIA AI platform readiness report.
type ReadinessStatus struct {
	Ready  bool
	Checks []CheckResult
}

// ChecksByName returns readiness checks keyed by name.
func (s ReadinessStatus) ChecksByName() map[string]CheckResult {
	checks := make(map[string]CheckResult, len(s.Checks))
	for _, check := range s.Checks {
		checks[check.Name] = check
	}
	return checks
}

// EvaluateReadiness evaluates whether the NVIDIA AI platform is proven ready.
func EvaluateReadiness(snapshot ClusterSnapshot) ReadinessStatus {
	checks := []CheckResult{
		boolCheck("gpu-operator", snapshot.GPUOperatorReady, "NVIDIA GPU Operator reports ready", "NVIDIA GPU Operator is missing or not ready"),
		boolCheck("gpu-driver", snapshot.GPUDriverReady, "NVIDIA GPU Operator driver operand reports ready", "NVIDIA GPU Operator driver operand is missing or not ready"),
		boolCheck("gpu-runtime-toolkit", snapshot.GPUContainerToolkitReady, "NVIDIA container runtime/toolkit operand reports ready", "NVIDIA container runtime/toolkit operand is missing or not ready"),
		boolCheck("device-plugin", snapshot.GPUDevicePluginReady, "NVIDIA device plugin reports ready", "NVIDIA device plugin is missing or not ready"),
		boolCheck("gpu-feature-discovery", snapshot.GPUFeatureDiscoveryReady, "GPU Feature Discovery reports ready", "GPU Feature Discovery is missing or not ready"),
		boolCheck("gpu-dcgm-exporter", snapshot.GPUManagedDCGMExporterReady, "NVIDIA GPU Operator managed DCGM exporter reports ready", "NVIDIA GPU Operator managed DCGM exporter is missing or not ready"),
		boolCheck("gpu-mig-manager", snapshot.GPUMIGManagerReady, "NVIDIA MIG Manager reports ready", "NVIDIA MIG Manager is missing or not ready"),
		boolCheck("gpu-validator", snapshot.GPUOperatorValidatorReady, "NVIDIA GPU Operator validator reports ready", "NVIDIA GPU Operator validator is missing or not ready"),
		gpuAllocatableCheck(snapshot.GPUAllocatableByNode, snapshot.MIGAllocatableByNode),
		gpuCorrelationCheck(snapshot.GPUCorrelationIssues),
		rdmaNetworkCheck(snapshot.RDMAResourcesByNode, snapshot.NetworkAttachments, snapshot.LastRDMASmokeTestPassed),
		boolCheck("dcgm-metrics", snapshot.DCGMMetricsScraped, "DCGM metrics are being scraped", "DCGM metrics are not proven scraped"),
		nimServingCheck(snapshot.NIMServicesReady, snapshot.LastNIMSmokeTestPassed),
		kaiSchedulerCheck(snapshot.KAISchedulerReady, snapshot.KAIQueueReady, snapshot.LastKAISchedulingSmokeTestPassed),
	}

	ready := true
	for _, check := range checks {
		if !check.Ready {
			ready = false
			break
		}
	}

	return ReadinessStatus{Ready: ready, Checks: checks}
}

func boolCheck(name string, ready bool, readyMessage, notReadyMessage string) CheckResult {
	if ready {
		return CheckResult{Name: name, Ready: true, Message: readyMessage}
	}
	return CheckResult{Name: name, Ready: false, Message: notReadyMessage}
}

func gpuAllocatableCheck(gpusByNode map[string]int, migByNode map[string]map[string]int) CheckResult {
	gpuTotal := 0
	for _, count := range gpusByNode {
		if count > 0 {
			gpuTotal += count
		}
	}
	migTotal := 0
	migProfiles := map[string]bool{}
	for _, resources := range migByNode {
		for profile, count := range resources {
			if count > 0 {
				migTotal += count
				migProfiles[profile] = true
			}
		}
	}
	if gpuTotal == 0 && migTotal == 0 {
		return CheckResult{Name: "gpu-allocatable", Ready: false, Message: "no allocatable NVIDIA GPUs or MIG resources found"}
	}
	if gpuTotal > 0 && migTotal > 0 {
		return CheckResult{Name: "gpu-allocatable", Ready: true, Message: fmt.Sprintf("%d allocatable NVIDIA GPUs and %d MIG instance(s) across %d MIG profile(s) found", gpuTotal, migTotal, len(migProfiles))}
	}
	if migTotal > 0 {
		return CheckResult{Name: "gpu-allocatable", Ready: true, Message: fmt.Sprintf("%d allocatable NVIDIA MIG instance(s) across %d MIG profile(s) found", migTotal, len(migProfiles))}
	}
	return CheckResult{Name: "gpu-allocatable", Ready: true, Message: fmt.Sprintf("%d allocatable NVIDIA GPUs found", gpuTotal)}
}

func gpuCorrelationCheck(issues []string) CheckResult {
	if len(issues) > 0 {
		return CheckResult{Name: "gpu-nico-kubernetes-correlation", Ready: false, Message: fmt.Sprintf("%d NICo/Kubernetes GPU correlation issue(s): %s", len(issues), issues[0])}
	}
	return CheckResult{Name: "gpu-nico-kubernetes-correlation", Ready: true, Message: "NICo GPU inventory and Kubernetes NVIDIA resources are consistent"}
}

func rdmaNetworkCheck(rdmaByNode map[string]int, attachments []string, smokePassed bool) CheckResult {
	if len(rdmaByNode) == 0 {
		return CheckResult{Name: "rdma-network", Ready: false, Message: "no allocatable NVIDIA RDMA resources found"}
	}
	total := 0
	for _, count := range rdmaByNode {
		if count > 0 {
			total += count
		}
	}
	if total == 0 {
		return CheckResult{Name: "rdma-network", Ready: false, Message: "RDMA nodes reported zero allocatable NVIDIA RDMA resources"}
	}
	if len(attachments) == 0 {
		return CheckResult{Name: "rdma-network", Ready: false, Message: "RDMA resources exist but no NetworkAttachmentDefinition evidence was found"}
	}
	if !smokePassed {
		return CheckResult{Name: "rdma-network", Ready: false, Message: "RDMA resources and NetworkAttachmentDefinitions exist but RDMA smoke test has not passed"}
	}
	return CheckResult{Name: "rdma-network", Ready: true, Message: fmt.Sprintf("%d allocatable NVIDIA RDMA resource(s), %d network attachment(s), and RDMA smoke test are proven", total, len(attachments))}
}

func nimServingCheck(readyServices int, smokePassed bool) CheckResult {
	if readyServices <= 0 {
		return CheckResult{Name: "nim-serving", Ready: false, Message: "no ready NVIDIA NIM service found"}
	}
	if !smokePassed {
		return CheckResult{Name: "nim-serving", Ready: false, Message: "NIM service exists but smoke test has not passed"}
	}
	return CheckResult{Name: "nim-serving", Ready: true, Message: fmt.Sprintf("%d NVIDIA NIM service(s) ready and smoke-tested", readyServices)}
}

func kaiSchedulerCheck(schedulerReady, queueReady, smokePassed bool) CheckResult {
	if !schedulerReady {
		return CheckResult{Name: "kai-scheduler", Ready: false, Message: "NVIDIA KAI Scheduler controllers are missing or not ready"}
	}
	if !queueReady {
		return CheckResult{Name: "kai-scheduler", Ready: false, Message: "KAI Scheduler queue evidence is missing"}
	}
	if !smokePassed {
		return CheckResult{Name: "kai-scheduler", Ready: false, Message: "KAI Scheduler exists but scheduling smoke test has not passed"}
	}
	return CheckResult{Name: "kai-scheduler", Ready: true, Message: "KAI Scheduler controllers, queues, and scheduling smoke test are proven"}
}
