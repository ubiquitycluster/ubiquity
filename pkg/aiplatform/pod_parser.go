package aiplatform

import "encoding/json"

type podList struct {
	Items []podItem `json:"items"`
}

type podItem struct {
	Status podStatus `json:"status"`
}

type podStatus struct {
	Phase      string         `json:"phase"`
	Conditions []podCondition `json:"conditions"`
}

type podCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ParseReadyPodCount counts only pods that are Running and have Ready=True in
// `kubectl get pods -o json` output. Empty lists produce zero evidence because
// kubectl selector commands can succeed even when no pods match.
func ParseReadyPodCount(data []byte) (int, error) {
	parsed := podList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return 0, err
	}
	ready := 0
	for _, pod := range parsed.Items {
		if pod.Status.Phase != "Running" {
			continue
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				ready++
				break
			}
		}
	}
	return ready, nil
}
