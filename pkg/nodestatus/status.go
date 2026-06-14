package nodestatus

// NodeStatus is Ubiquity's live, joined view of a physical NICo machine,
// optional NICo instance, Kubernetes node, and NVIDIA AI resource readiness.
type NodeStatus struct {
	Name               string   `json:"name"`
	Site               string   `json:"site"`
	MachineID          string   `json:"machineID"`
	InstanceID         string   `json:"instanceID"`
	PowerState         string   `json:"powerState"`
	MachineStatus      string   `json:"machineStatus"`
	InstanceStatus     string   `json:"instanceStatus"`
	OSImage            string   `json:"osImage"`
	KubernetesNodeName string   `json:"kubernetesNodeName"`
	KubernetesReady    bool     `json:"kubernetesReady"`
	Cordoned           bool     `json:"cordoned"`
	Roles              []string `json:"roles"`
	// GPUs is kept for compatibility and reports the total visible GPU evidence.
	GPUs             int            `json:"gpus"`
	NICoGPUs         int            `json:"nicoGPUs"`
	KubernetesGPUs   int            `json:"kubernetesGPUs"`
	MIGProfiles      map[string]int `json:"migProfiles"`
	RDMAResources    int            `json:"rdmaResources"`
	NVIDIAReady      bool           `json:"nvidiaReady"`
	BMCStatus        string         `json:"bmcStatus"`
	KubeletStatus    string         `json:"kubeletStatus"`
	GPUStatus        string         `json:"gpuStatus"`
	RDMAStatus       string         `json:"rdmaStatus"`
	FirmwareStatus   string         `json:"firmwareStatus"`
	ImageStatus      string         `json:"imageStatus"`
	MaintenanceState string         `json:"maintenanceState"`
	ActiveTaskID     string         `json:"activeTaskID"`
	LastAction       string         `json:"lastAction"`
	Reason           string         `json:"reason"`
}

// KubernetesNodeEvidence is an intentionally small, fakeable Kubernetes model
// used by node-status aggregation. Production collectors can populate it from
// the Kubernetes Node API or the existing pkg/aiplatform kubectl JSON parsers.
type KubernetesNodeEvidence struct {
	Name          string         `json:"name"`
	Ready         bool           `json:"ready"`
	Cordoned      bool           `json:"cordoned"`
	Roles         []string       `json:"roles"`
	GPUs          int            `json:"gpus"`
	MIGProfiles   map[string]int `json:"migProfiles"`
	RDMAResources int            `json:"rdmaResources"`
	NVIDIAReady   bool           `json:"nvidiaReady"`
}

// CloneMIGProfiles returns a non-nil defensive copy for status output.
func CloneMIGProfiles(in map[string]int) map[string]int {
	out := map[string]int{}
	for k, v := range in {
		if v > 0 {
			out[k] = v
		}
	}
	return out
}
