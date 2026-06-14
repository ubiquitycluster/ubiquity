package aiplatform

// OrchestrationDecision records whether an upstream NVIDIA repository should be
// adopted, retained as a reference, or evaluated for a later profile slice.
type OrchestrationDecision string

const (
	OrchestrationDecisionAdopt     OrchestrationDecision = "adopt"
	OrchestrationDecisionReference OrchestrationDecision = "reference"
	OrchestrationDecisionEvaluate  OrchestrationDecision = "evaluate"
)

// OrchestrationAlternative captures the reviewer-visible decision for NVIDIA
// bare-metal orchestration repositories considered for this platform.
type OrchestrationAlternative struct {
	Name          string
	SourceRepo    string
	Decision      OrchestrationDecision
	ReplacesLocal bool
	Scope         string
	Evaluation    string
}

// BareMetalOrchestrationAlternatives returns the deterministic set of NVIDIA
// repository alternatives considered for bare-metal AI workload orchestration.
// Adopted entries should replace weaker local functionality; reference entries
// are source-backed guidance but are not treated as certification or a drop-in
// replacement for this project's sandbox/bootstrap path.
func BareMetalOrchestrationAlternatives() []OrchestrationAlternative {
	return []OrchestrationAlternative{
		{
			Name:          "gpu-operator",
			SourceRepo:    "https://github.com/NVIDIA/gpu-operator",
			Decision:      OrchestrationDecisionAdopt,
			ReplacesLocal: true,
			Scope:         "Kubernetes GPU node enablement",
			Evaluation:    "Adopt as the production default for drivers, container runtime, device plugin, GPU feature discovery, DCGM exporter, validator, and MIG Manager instead of bespoke node enablement.",
		},
		{
			Name:          "network-operator",
			SourceRepo:    "https://github.com/Mellanox/network-operator",
			Decision:      OrchestrationDecisionAdopt,
			ReplacesLocal: true,
			Scope:         "RDMA and secondary-network orchestration for GPU clusters",
			Evaluation:    "Adopt for RDMA profiles; sandbox mode must deploy only the control plane and CRDs so k3d networking is not broken by hardware-specific CNI operands.",
		},
		{
			Name:          "kai-scheduler",
			SourceRepo:    "https://github.com/NVIDIA/KAI-Scheduler",
			Decision:      OrchestrationDecisionEvaluate,
			ReplacesLocal: true,
			Scope:         "GPU-aware batch, training, inference, fairness, gang, and topology-aware scheduling",
			Evaluation:    "Evaluate as the preferred replacement for local priority/quota-only scheduling once wrapper chart, sandbox render/apply proof, and production fail-closed readiness checks are added.",
		},
		{
			Name:          "deepops",
			SourceRepo:    "https://github.com/NVIDIA/deepops",
			Decision:      OrchestrationDecisionReference,
			ReplacesLocal: false,
			Scope:         "Ansible/Kubespray/Slurm bare-metal cluster orchestration reference",
			Evaluation:    "Use as source-backed bare-metal reference for DGX/GPU cluster practices, Kubernetes via Kubespray, Slurm paths, OS support, and validation assumptions; do not claim it is a drop-in replacement for Ubiquity's k3d/Cluster API sandbox path.",
		},
		{
			Name:          "cloud-native-stack",
			SourceRepo:    "https://github.com/NVIDIA/cloud-native-stack",
			Decision:      OrchestrationDecisionReference,
			ReplacesLocal: false,
			Scope:         "NVIDIA Cloud Native Stack reference matrix and PoC installation playbooks",
			Evaluation:    "Use as source-backed component matrix and PoC reference, especially GPU Operator, Network Operator, NIM Operator, KAI Scheduler, Kubernetes, CNI, monitoring, and load-balancer versions; do not treat its basic non-HA Kubernetes install as production orchestration.",
		},
		{
			Name:          "k8s-dra-driver-gpu",
			SourceRepo:    "https://github.com/NVIDIA/k8s-dra-driver-gpu",
			Decision:      OrchestrationDecisionEvaluate,
			ReplacesLocal: false,
			Scope:         "Kubernetes Dynamic Resource Allocation for GPUs",
			Evaluation:    "Track as a future replacement for static GPU resource scheduling assumptions when Kubernetes DRA support is ready for the selected production profile.",
		},
	}
}
