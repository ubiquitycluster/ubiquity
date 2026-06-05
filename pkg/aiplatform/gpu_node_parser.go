package aiplatform

import (
	"encoding/json"
	"strconv"
	"strings"
)

type nodeList struct {
	Items []nodeItem `json:"items"`
}

type nodeItem struct {
	Metadata nodeMetadata `json:"metadata"`
	Status   nodeStatus   `json:"status"`
}

type nodeMetadata struct {
	Name string `json:"name"`
}

type nodeStatus struct {
	Allocatable map[string]string `json:"allocatable"`
}

// ParseGPUAllocatableByNode extracts positive nvidia.com/gpu allocatable counts
// from `kubectl get nodes -o json` output. Missing, malformed, or zero values
// produce no readiness evidence so callers fail closed.
func ParseGPUAllocatableByNode(data []byte) (map[string]int, error) {
	return ParseAllocatableResourceByNode(data, "nvidia.com/gpu")
}

// ParseAllocatableResourceByNode extracts positive extended-resource allocatable
// counts from `kubectl get nodes -o json` output. It is used for NVIDIA GPU and
// RDMA resources so successful kubectl calls with empty/zero resources cannot
// produce false readiness evidence.
func ParseAllocatableResourceByNode(data []byte, resourceName string) (map[string]int, error) {
	parsed := nodeList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]int{}, err
	}

	resourcesByNode := map[string]int{}
	for _, node := range parsed.Items {
		if node.Metadata.Name == "" {
			continue
		}
		raw, ok := node.Status.Allocatable[resourceName]
		if !ok {
			continue
		}
		count, err := strconv.Atoi(raw)
		if err != nil || count <= 0 {
			continue
		}
		resourcesByNode[node.Metadata.Name] = count
	}
	return resourcesByNode, nil
}

// ParseMIGAllocatableByNode extracts positive nvidia.com/mig-* allocatable
// resources from node JSON. MIG-partitioned clusters often expose only MIG
// extended resources rather than nvidia.com/gpu, so these resources are valid
// accelerator capacity evidence for MIG profiles.
func ParseMIGAllocatableByNode(data []byte) (map[string]map[string]int, error) {
	parsed := nodeList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]map[string]int{}, err
	}
	migByNode := map[string]map[string]int{}
	for _, node := range parsed.Items {
		if node.Metadata.Name == "" {
			continue
		}
		for resourceName, raw := range node.Status.Allocatable {
			if !strings.HasPrefix(resourceName, "nvidia.com/mig-") {
				continue
			}
			count, err := strconv.Atoi(raw)
			if err != nil || count <= 0 {
				continue
			}
			if migByNode[node.Metadata.Name] == nil {
				migByNode[node.Metadata.Name] = map[string]int{}
			}
			migByNode[node.Metadata.Name][resourceName] = count
		}
	}
	return migByNode, nil
}
