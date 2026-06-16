package aiplatform

// NCPRequirement records a source-backed reference-platform capability and the
// Ubiquity-owned evidence that must exist before the platform claims parity for
// that capability. These are intentionally capability statements, not copied
// implementation logic from any external project.
type NCPRequirement struct {
	ID               string
	Layer            string
	Capability       string
	UbiquityEvidence []string
	ReadinessSignal  string
}

// NCPRequirements returns the deterministic acceptance map for the
// ai-production profile. The map follows the NVIDIA Cloud Accelerator layered
// model: IaaS, CaaS, AI PaaS, workload isolation, and operations.
func NCPRequirements() []NCPRequirement {
	return []NCPRequirement{
		{
			ID:         "iaas-bare-metal-vm-lifecycle",
			Layer:      "IaaS",
			Capability: "Bare metal and VM lifecycle with GPU-aware placement, OS image provenance, and fail-closed node readiness.",
			UbiquityEvidence: []string{
				"pkg/nodeinventory",
				"pkg/nodestatus",
				"pkg/virtualization",
				"system/nvidia-nic-configuration-operator",
			},
			ReadinessSignal: "node status joins NICo inventory, Kubernetes resource evidence, firmware/image state, and VM readiness without claiming readiness on missing evidence",
		},
		{
			ID:         "caas-gpu-kubernetes-substrate",
			Layer:      "CaaS",
			Capability: "Kubernetes substrate with NVIDIA GPU Operator managed driver, runtime toolkit, device plugin, GPU feature discovery, MIG Manager, validator, and DCGM exporter.",
			UbiquityEvidence: []string{
				"system/nvidia-gpu-operator",
				"pkg/aiplatform/readiness.go",
				"ubiquity health --ai",
			},
			ReadinessSignal: "gpu-operator, gpu-driver, gpu-runtime-toolkit, device-plugin, gpu-feature-discovery, gpu-mig-manager, gpu-validator, gpu-dcgm-exporter, and gpu-allocatable checks are ready",
		},
		{
			ID:         "caas-rdma-networking",
			Layer:      "CaaS",
			Capability: "High-performance NVIDIA networking with RDMA resources, secondary NetworkAttachmentDefinitions, and NIC configuration guardrails.",
			UbiquityEvidence: []string{
				"system/nvidia-network-operator",
				"system/nvidia-nic-configuration-operator",
				"test/e2e/nvidia-rdma-smoke.sh",
			},
			ReadinessSignal: "rdma-network check requires nvidia.com/rdma allocatable resources, NetworkAttachmentDefinitions, and rdma-network-smoke-test-passed evidence",
		},
		{
			ID:         "paas-serving-scheduling",
			Layer:      "AI PaaS",
			Capability: "Production inference and batch scheduling with NIM Operator backed serving and KAI Scheduler backed GPU workload placement.",
			UbiquityEvidence: []string{
				"platform/nim-operator",
				"platform/kai-scheduler",
				"test/e2e/nim-gpu-serving-smoke.sh",
				"test/e2e/kai-scheduler-smoke.sh",
			},
			ReadinessSignal: "nim-serving and kai-scheduler checks require live service, queue, and smoke-test ConfigMap evidence",
		},
		{
			ID:         "tenant-workload-isolation",
			Layer:      "Workload Isolation",
			Capability: "Tenant namespaces use restricted pod security labels, ResourceQuota, LimitRange defaults, and default-deny NetworkPolicies with only explicit tenant-local and DNS egress allowances.",
			UbiquityEvidence: []string{
				"platform/ai-workload-tenancy",
				"platform/tenant-kubernetes-cluster",
				"platform/tenant-vpc",
			},
			ReadinessSignal: "tenant isolation manifests render per-tenant namespaces, quotas, limits, and network policies before workloads are scheduled",
		},
		{
			ID:         "observability-validation",
			Layer:      "Operations",
			Capability: "GPU telemetry, fail-closed health reporting, gated E2E validation, and final demonstration of provision, reconcile, schedule, serve, observe, and validate phases.",
			UbiquityEvidence: []string{
				"ubiquity info --ai",
				"ubiquity health --ai",
				"test/e2e/nvidia-ai-platform-final-demo.sh",
			},
			ReadinessSignal: "final demo evidence marker is written only after all gated substrate, scheduling, serving, networking, observability, and health checks pass",
		},
		{
			ID:         "unified-frontend-service",
			Layer:      "Operations",
			Capability: "Single project-native frontend and JSON API for tenant compute, operator capability, component, evidence, and readiness views across the reference platform.",
			UbiquityEvidence: []string{
				"pkg/aiplatform/frontend.go",
				"cmd/ubiquity/cmd/ai_platform.go",
				"platform/ai-platform-console",
			},
			ReadinessSignal: "frontend serves /, /api/platform, /api/requirements, /api/profiles, and /healthz from Ubiquity-owned profile and requirement data without weakening fail-closed runtime readiness boundaries",
		},
	}
}
