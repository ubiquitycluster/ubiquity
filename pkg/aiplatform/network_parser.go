package aiplatform

import "encoding/json"

type networkAttachmentList struct {
	Items []networkAttachmentItem `json:"items"`
}

type networkAttachmentItem struct {
	Metadata networkAttachmentMetadata `json:"metadata"`
}

type networkAttachmentMetadata struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// ParseNetworkAttachments extracts namespaced NetworkAttachmentDefinition names
// from `kubectl get network-attachment-definitions -A -o json` output. Empty
// lists produce no evidence so readiness fails closed.
func ParseNetworkAttachments(data []byte) ([]string, error) {
	parsed := networkAttachmentList{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return []string{}, err
	}

	attachments := []string{}
	for _, item := range parsed.Items {
		if item.Metadata.Name == "" {
			continue
		}
		namespace := item.Metadata.Namespace
		if namespace == "" {
			namespace = "default"
		}
		attachments = append(attachments, namespace+"/"+item.Metadata.Name)
	}
	return attachments, nil
}
