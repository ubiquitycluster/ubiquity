package aiplatform

import "encoding/json"

type deploymentList struct {
	Items []deploymentItem `json:"items"`
}

type deploymentItem struct {
	Metadata namedMetadata    `json:"metadata"`
	Status   deploymentStatus `json:"status"`
}

type namedMetadata struct {
	Name string `json:"name"`
}

type deploymentStatus struct {
	ReadyReplicas     int `json:"readyReplicas"`
	AvailableReplicas int `json:"availableReplicas"`
}

type daemonSetList struct {
	Items []daemonSetItem `json:"items"`
}

type daemonSetItem struct {
	Metadata namedMetadata   `json:"metadata"`
	Status   daemonSetStatus `json:"status"`
}

type daemonSetStatus struct {
	DesiredNumberScheduled int `json:"desiredNumberScheduled"`
	NumberReady            int `json:"numberReady"`
	NumberAvailable        int `json:"numberAvailable"`
}

func deploymentIsAvailable(item deploymentItem) bool {
	return item.Metadata.Name != "" && item.Status.ReadyReplicas > 0 && item.Status.AvailableReplicas > 0
}

func daemonSetIsReady(item daemonSetItem) bool {
	status := item.Status
	return item.Metadata.Name != "" && status.DesiredNumberScheduled > 0 && status.NumberReady >= status.DesiredNumberScheduled && status.NumberAvailable >= status.DesiredNumberScheduled
}

// ParseAvailableDeployments returns deployment names with at least one ready and
// available replica. Object existence alone is not readiness evidence.
func ParseAvailableDeployments(data []byte) (map[string]bool, error) {
	parsed := deploymentList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]bool{}, err
	}
	available := map[string]bool{}
	if len(parsed.Items) == 0 {
		item := deploymentItem{}
		if err := json.Unmarshal(data, &item); err != nil {
			return map[string]bool{}, err
		}
		if deploymentIsAvailable(item) {
			available[item.Metadata.Name] = true
		}
		return available, nil
	}
	for _, item := range parsed.Items {
		if deploymentIsAvailable(item) {
			available[item.Metadata.Name] = true
		}
	}
	return available, nil
}

// ParseReadyDaemonSets returns daemonset names only when the daemonset has a
// positive desired schedule count and every desired pod is both ready and
// available.
func ParseReadyDaemonSets(data []byte) (map[string]bool, error) {
	parsed := daemonSetList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return map[string]bool{}, err
	}
	ready := map[string]bool{}
	if len(parsed.Items) == 0 {
		item := daemonSetItem{}
		if err := json.Unmarshal(data, &item); err != nil {
			return map[string]bool{}, err
		}
		if daemonSetIsReady(item) {
			ready[item.Metadata.Name] = true
		}
		return ready, nil
	}
	for _, item := range parsed.Items {
		if daemonSetIsReady(item) {
			ready[item.Metadata.Name] = true
		}
	}
	return ready, nil
}
