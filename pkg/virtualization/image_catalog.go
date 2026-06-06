package virtualization

import (
	"fmt"
	"strings"
)

// VMImageCatalogRequest renders a provenance catalog for importable VM images.
type VMImageCatalogRequest struct {
	Name      string
	Namespace string
}

// RenderVMImageCatalog renders a reviewer-visible catalog of OS image profiles.
func RenderVMImageCatalog(req VMImageCatalogRequest) (string, error) {
	if req.Name == "" {
		req.Name = "vm-image-catalog"
	}
	if req.Namespace == "" {
		req.Namespace = "ubiquity-system"
	}
	if !dnsName.MatchString(req.Name) || !dnsName.MatchString(req.Namespace) {
		return "", fmt.Errorf("VM image catalog name and namespace must be DNS-compatible")
	}
	var b strings.Builder
	fmt.Fprintf(&b, `apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
  labels:
    ubiquity.ai/vm-image-catalog: kubevirt
data:
  readinessBoundary: import-and-guest-boot-not-proven-by-catalog
  profiles: |
`, req.Name, req.Namespace)
	for _, profile := range OSProfiles() {
		fmt.Fprintf(&b, "    - name: %s\n", profile.Name)
		fmt.Fprintf(&b, "      displayName: %s\n", profile.DisplayName)
		fmt.Fprintf(&b, "      sourceURL: %s\n", profile.SourceURL)
		fmt.Fprintf(&b, "      cloudInit: %t\n", profile.CloudInit)
		fmt.Fprintf(&b, "      notes: %s\n", profile.Notes)
	}
	return b.String(), nil
}
