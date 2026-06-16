package aiplatform

import (
	"fmt"
	"sort"
)

// Capability describes a platform behavior that a profile provides.
type Capability string

const (
	CapabilityGPU                    Capability = "gpu"
	CapabilityRDMA                   Capability = "rdma"
	CapabilityServing                Capability = "serving"
	CapabilityTelemetry              Capability = "telemetry"
	CapabilityValidation             Capability = "validation"
	CapabilityStorage                Capability = "storage"
	CapabilityScheduler              Capability = "scheduler"
	CapabilityBareMetalOrchestration Capability = "bare-metal-orchestration"
	CapabilityVirtualization         Capability = "virtualization"
	CapabilityUnifiedFrontend        Capability = "unified-frontend"
)

// Component records the upstream source of a platform component and whether it
// replaces a weaker Ubiquity-local implementation.
type Component struct {
	Name                 string
	SourceRepo           string
	ChartName            string
	ChartRepository      string
	Version              string
	Namespace            string
	ReplacesLocal        bool
	ManagedByGPUOperator bool
	ProductionDefault    bool
	Optional             bool
	Notes                string
}

// Profile is a deterministic NVIDIA-backed AI platform target state.
type Profile struct {
	Name         string
	Description  string
	Capabilities []Capability
	Components   []Component
}

// ComponentsByName returns components keyed by component name.
func (p Profile) ComponentsByName() map[string]Component {
	components := make(map[string]Component, len(p.Components))
	for _, component := range p.Components {
		components[component.Name] = component
	}
	return components
}

// HasCapability reports whether a profile includes a capability.
func (p Profile) HasCapability(capability Capability) bool {
	for _, existing := range p.Capabilities {
		if existing == capability {
			return true
		}
	}
	return false
}

// Names returns the supported profile names in deterministic order.
func Names() []string {
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetProfile returns a supported NVIDIA AI platform profile, failing closed for
// unknown profile names so unsupported configurations never imply readiness.
func GetProfile(name string) (Profile, error) {
	profile, ok := profiles[name]
	if !ok {
		return Profile{}, fmt.Errorf("unknown AI platform profile %q; supported profiles: %v", name, Names())
	}
	return profile, nil
}

var gpuOperator = Component{
	Name:              "gpu-operator",
	SourceRepo:        "https://github.com/NVIDIA/gpu-operator",
	ChartName:         "gpu-operator",
	ChartRepository:   "https://helm.ngc.nvidia.com/nvidia",
	Version:           "v26.3.2",
	Namespace:         "gpu-operator",
	ReplacesLocal:     true,
	ProductionDefault: true,
	Notes:             "Primary source of truth for NVIDIA drivers, container runtime, device plugin, GPU feature discovery, DCGM exporter, validator, and MIG Manager.",
}

var dcgmExporter = Component{
	Name:                 "dcgm-exporter",
	SourceRepo:           "https://github.com/NVIDIA/dcgm-exporter",
	ChartName:            "dcgm-exporter",
	ChartRepository:      "https://helm.ngc.nvidia.com/nvidia",
	Version:              "managed-by-gpu-operator",
	Namespace:            "gpu-operator",
	ReplacesLocal:        true,
	ManagedByGPUOperator: true,
	ProductionDefault:    true,
	Notes:                "Replaces hand-authored Ubiquity DCGM DaemonSets; enabled through GPU Operator by default.",
}

var nimOperator = Component{
	Name:              "nim-operator",
	SourceRepo:        "https://github.com/NVIDIA/k8s-nim-operator",
	ChartName:         "k8s-nim-operator",
	ChartRepository:   "https://helm.ngc.nvidia.com/nvidia",
	Version:           "v2.0.0",
	Namespace:         "nim-operator",
	ReplacesLocal:     true,
	ProductionDefault: true,
	Notes:             "Production serving layer for NVIDIA NIM and NeMo microservices; replaces Ollama as production default.",
}

var nvidiaNetworkOperator = Component{
	Name:              "nvidia-network-operator",
	SourceRepo:        "https://github.com/Mellanox/network-operator",
	ChartName:         "network-operator",
	ChartRepository:   "https://helm.ngc.nvidia.com/nvidia",
	Version:           "v24.1.0",
	Namespace:         "nvidia-network-operator",
	ReplacesLocal:     false,
	ProductionDefault: true,
	Notes:             "Profile-driven RDMA, secondary-network, and NIC policy layer for GPU networking.",
}

var nvidiaNICConfigurationOperator = Component{
	Name:              "nvidia-nic-configuration-operator",
	SourceRepo:        "https://github.com/Mellanox/nic-configuration-operator",
	ChartName:         "nvidia-nic-configuration-operator",
	ChartRepository:   "file://system/nvidia-nic-configuration-operator",
	Version:           "9a42b0e",
	Namespace:         "network-operator",
	ReplacesLocal:     true,
	ProductionDefault: true,
	Notes:             "Production NIC configuration guardrail for NVIDIA ConnectX NIC templates, firmware storage boundaries, RoCE tuning, SR-IOV, and GPUDirect-safe NIC settings.",
}

var kaiScheduler = Component{
	Name:              "kai-scheduler",
	SourceRepo:        "https://github.com/NVIDIA/KAI-Scheduler",
	ChartName:         "kai-scheduler",
	ChartRepository:   "oci://ghcr.io/kai-scheduler/kai-scheduler",
	Version:           "v0.10.2",
	Namespace:         "kai-scheduler",
	ReplacesLocal:     true,
	ProductionDefault: true,
	Notes:             "Production AI workload scheduler for GPU-aware fairness, gang scheduling, topology-aware placement, DRA-aware scheduling, and batch/inference prioritization; replaces priority/quota-only local scheduling for production profiles.",
}

var aicr = Component{
	Name:              "aicr",
	SourceRepo:        "https://github.com/NVIDIA/aicr",
	Version:           "tracked",
	ReplacesLocal:     false,
	ProductionDefault: true,
	Notes:             "Recipe source for optimized, validated, reproducible GPU-accelerated AI runtime manifests.",
}

var aiStore = Component{
	Name:              "aistore",
	SourceRepo:        "https://github.com/NVIDIA/aistore",
	ChartName:         "ais-k8s",
	ChartRepository:   "https://github.com/NVIDIA/ais-k8s",
	Version:           "evaluated",
	Namespace:         "aistore",
	Optional:          true,
	ProductionDefault: false,
	Notes:             "Optional NVIDIA-maintained high-performance AI object store/cache via NVIDIA/ais-k8s; preferred over Longhorn for AI dataset/checkpoint/model artifact/cache paths when object/S3 semantics fit, but not a generic PVC/POSIX replacement.",
}

var stallscope = Component{
	Name:              "stallscope",
	SourceRepo:        "https://github.com/nshinde/stallscope",
	ChartName:         "stallscope",
	ChartRepository:   "file://platform/stallscope",
	Version:           "84b17513b9230ce66c2838bb4e6fe95f196a044c",
	Namespace:         "gpu-telemetry",
	ReplacesLocal:     false,
	ProductionDefault: true,
	Notes:             "GPU workload performance telemetry and stall classification from nvidia-smi, host /proc, RDMA counters, PFC pause counters, NCCL smoke tests, and Prometheus textfile metrics for slow/fail-risk diagnostics.",
}

var ollama = Component{
	Name:              "ollama",
	SourceRepo:        "https://github.com/ollama/ollama",
	ChartName:         "ollama",
	ChartRepository:   "file://apps/ollama",
	Version:           "optional",
	Namespace:         "ollama",
	Optional:          true,
	ProductionDefault: false,
	Notes:             "Retained as lightweight local diagnostics/lab app; not a production NVIDIA serving default.",
}

var kubevirt = Component{
	Name:              "kubevirt",
	SourceRepo:        "https://github.com/kubevirt/kubevirt",
	ChartName:         "kubevirt-vms",
	ChartRepository:   "file://platform/kubevirt-vms",
	Version:           "operator-managed",
	Namespace:         "kubevirt",
	ProductionDefault: true,
	Notes:             "VirtualMachine API for VM workloads on bootstrapped Kubernetes nodes. GPU-enabled VMs require KubeVirt permittedHostDevices plus NVIDIA GPU Operator/device-plugin resources.",
}

var cdi = Component{
	Name:              "containerized-data-importer",
	SourceRepo:        "https://github.com/kubevirt/containerized-data-importer",
	ChartName:         "kubevirt-vms",
	ChartRepository:   "file://platform/kubevirt-vms",
	Version:           "operator-managed",
	Namespace:         "cdi",
	ProductionDefault: true,
	Notes:             "CDI DataVolume imports OS images for Ubuntu, Rocky, Windows, and future operator-defined profiles.",
}

var multusCNI = Component{
	Name:              "multus-cni",
	SourceRepo:        "https://github.com/k8snetworkplumbingwg/multus-cni",
	ChartName:         "kubevirt-vms",
	ChartRepository:   "file://platform/kubevirt-vms",
	Version:           "operator-managed",
	Namespace:         "kube-system",
	ProductionDefault: true,
	Notes:             "Secondary network attachment provider for isolated VM networks and RDMA/SR-IOV capable data-plane networks.",
}

var profiles = map[string]Profile{
	"sandbox": {
		Name:         "sandbox",
		Description:  "CPU-only local or simulated path; no NVIDIA readiness claims.",
		Capabilities: []Capability{CapabilityValidation},
		Components:   []Component{ollama},
	},
	"gpu-basic": {
		Name:         "gpu-basic",
		Description:  "GPU Operator, DCGM telemetry, Stallscope workload stall telemetry, NIM sample serving, and fail-closed health checks.",
		Capabilities: []Capability{CapabilityGPU, CapabilityServing, CapabilityTelemetry, CapabilityValidation},
		Components:   []Component{gpuOperator, dcgmExporter, stallscope, nimOperator, aicr, ollama},
	},
	"gpu-rdma": {
		Name:         "gpu-rdma",
		Description:  "GPU Operator plus NVIDIA Network Operator/RDMA validation and Stallscope fabric stall telemetry.",
		Capabilities: []Capability{CapabilityGPU, CapabilityRDMA, CapabilityServing, CapabilityTelemetry, CapabilityValidation},
		Components:   []Component{gpuOperator, dcgmExporter, stallscope, nvidiaNetworkOperator, nvidiaNICConfigurationOperator, nimOperator, aicr, ollama},
	},
	"gpu-mig": {
		Name:         "gpu-mig",
		Description:  "GPU Operator with MIG profiles, scheduling validation, and Stallscope workload stall telemetry.",
		Capabilities: []Capability{CapabilityGPU, CapabilityServing, CapabilityTelemetry, CapabilityValidation, CapabilityScheduler},
		Components:   []Component{gpuOperator, dcgmExporter, stallscope, nimOperator, kaiScheduler, aicr, ollama},
	},
	"ai-production": {
		Name:         "ai-production",
		Description:  "Full NVIDIA-backed AI workload platform profile with GPU, RDMA, serving, observability, workload stall telemetry, storage/data-plane evaluation, KubeVirt virtualization, security, and E2E validation.",
		Capabilities: []Capability{CapabilityGPU, CapabilityRDMA, CapabilityServing, CapabilityTelemetry, CapabilityValidation, CapabilityStorage, CapabilityScheduler, CapabilityBareMetalOrchestration, CapabilityVirtualization, CapabilityUnifiedFrontend},
		Components:   []Component{gpuOperator, dcgmExporter, stallscope, nvidiaNetworkOperator, nvidiaNICConfigurationOperator, nimOperator, kaiScheduler, kubevirt, cdi, multusCNI, aicr, aiStore, ollama},
	},
}
