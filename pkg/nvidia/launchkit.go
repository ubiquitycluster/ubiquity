package nvidia

import (
	"encoding/json"
	"strings"
)

type RDMAProvenance string

const (
	RDMAProvenanceNICo         RDMAProvenance = "nico"
	RDMAProvenanceLaunchKit    RDMAProvenance = "k8s-launch-kit"
	RDMAProvenanceLocalKubectl RDMAProvenance = "local-kubectl"
)

type RDMADiscovery struct {
	NodeName   string         `json:"nodeName"`
	Resources  int            `json:"resources"`
	Interfaces []string       `json:"interfaces,omitempty"`
	Provenance RDMAProvenance `json:"provenance"`
}

type launchKitInventory struct {
	Nodes []struct {
		Name       string   `json:"name"`
		Interfaces []string `json:"interfaces"`
		RDMA       int      `json:"rdmaResources"`
	} `json:"nodes"`
}

func ParseLaunchKitRDMAInventory(data []byte) (map[string]RDMADiscovery, error) {
	parsed := launchKitInventory{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]RDMADiscovery{}, err
	}
	out := map[string]RDMADiscovery{}
	for _, node := range parsed.Nodes {
		if strings.TrimSpace(node.Name) == "" {
			continue
		}
		out[node.Name] = RDMADiscovery{NodeName: node.Name, Resources: node.RDMA, Interfaces: append([]string{}, node.Interfaces...), Provenance: RDMAProvenanceLaunchKit}
	}
	return out, nil
}

func SelectRDMAProvenance(node string, nicoResources int, launchKit map[string]RDMADiscovery, localKubectl map[string]int) RDMADiscovery {
	if nicoResources > 0 {
		return RDMADiscovery{NodeName: node, Resources: nicoResources, Provenance: RDMAProvenanceNICo}
	}
	if found, ok := launchKit[node]; ok && found.Resources > 0 {
		return found
	}
	if localKubectl[node] > 0 {
		return RDMADiscovery{NodeName: node, Resources: localKubectl[node], Provenance: RDMAProvenanceLocalKubectl}
	}
	return RDMADiscovery{NodeName: node}
}

func ValidateRDMAExpected(node string, expected bool, discovery RDMADiscovery) error {
	if expected && discovery.Resources <= 0 {
		return &MissingRDMAError{NodeName: node}
	}
	return nil
}

type MissingRDMAError struct{ NodeName string }

func (e *MissingRDMAError) Error() string {
	return "RDMA is expected but missing for node " + e.NodeName
}
