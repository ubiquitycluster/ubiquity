package nodestatus

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/ubiquitycluster/ubiquity/pkg/aiplatform"
)

type kubernetesNodeList struct {
	Items []kubernetesNodeItem `json:"items"`
}

type kubernetesNodeItem struct {
	Metadata struct {
		Name   string            `json:"name"`
		Labels map[string]string `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		Unschedulable bool `json:"unschedulable"`
	} `json:"spec"`
	Status struct {
		Allocatable map[string]string         `json:"allocatable"`
		Conditions  []kubernetesNodeCondition `json:"conditions"`
	} `json:"status"`
}

type kubernetesNodeCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ParseKubernetesNodeEvidence extracts the small status/safety model used by
// NICo node aggregation from `kubectl get nodes -o json`. It reuses the
// aiplatform parsers for NVIDIA GPU, MIG, and RDMA allocatable resources, then
// adds Kubernetes Ready, roles, and cordon state. Missing resources produce no
// positive evidence so callers do not claim readiness without live proof.
func ParseKubernetesNodeEvidence(data []byte) (map[string]KubernetesNodeEvidence, error) {
	parsed := kubernetesNodeList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]KubernetesNodeEvidence{}, err
	}
	gpusByNode, err := aiplatform.ParseGPUAllocatableByNode(data)
	if err != nil {
		return map[string]KubernetesNodeEvidence{}, err
	}
	migByNode, err := aiplatform.ParseMIGAllocatableByNode(data)
	if err != nil {
		return map[string]KubernetesNodeEvidence{}, err
	}
	rdmaByNode, err := aiplatform.ParseAllocatableResourceByNode(data, "nvidia.com/rdma")
	if err != nil {
		return map[string]KubernetesNodeEvidence{}, err
	}

	out := map[string]KubernetesNodeEvidence{}
	for _, node := range parsed.Items {
		name := strings.TrimSpace(node.Metadata.Name)
		if name == "" {
			continue
		}
		mig := CloneMIGProfiles(migByNode[name])
		rdma := rdmaByNode[name]
		if rdma == 0 {
			rdma = sumAllocatableContaining(node.Status.Allocatable, "rdma")
		}
		gpu := gpusByNode[name]
		out[name] = KubernetesNodeEvidence{
			Name:          name,
			Ready:         nodeReady(node.Status.Conditions),
			Cordoned:      node.Spec.Unschedulable,
			Roles:         nodeRoles(node.Metadata.Labels),
			GPUs:          gpu,
			MIGProfiles:   mig,
			RDMAResources: rdma,
			NVIDIAReady:   gpu > 0 || len(mig) > 0,
		}
	}
	return out, nil
}

func nodeReady(conditions []kubernetesNodeCondition) bool {
	for _, condition := range conditions {
		if condition.Type == "Ready" {
			return strings.EqualFold(condition.Status, "True")
		}
	}
	return false
}

func nodeRoles(labels map[string]string) []string {
	roles := []string{}
	seen := map[string]bool{}
	for key, value := range labels {
		if strings.HasPrefix(key, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(key, "node-role.kubernetes.io/")
			if role == "" && value != "" {
				role = value
			}
			if role != "" && !seen[role] {
				seen[role] = true
				roles = append(roles, role)
			}
		}
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	sort.Strings(roles)
	return roles
}

func sumAllocatableContaining(resources map[string]string, needle string) int {
	total := 0
	for name, raw := range resources {
		if !strings.Contains(strings.ToLower(name), needle) {
			continue
		}
		count, err := strconv.Atoi(raw)
		if err == nil && count > 0 {
			total += count
		}
	}
	return total
}
