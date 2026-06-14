package virtualization

import (
	"strings"
	"testing"
)

func TestRenderVMImageCatalogRecordsProfilesProvenanceAndReadinessBoundary(t *testing.T) {
	manifest, err := RenderVMImageCatalog(VMImageCatalogRequest{Name: "vm-image-catalog", Namespace: "ubiquity-system"})
	if err != nil {
		t.Fatalf("RenderVMImageCatalog returned error: %v", err)
	}
	for _, required := range []string{
		"kind: ConfigMap",
		"ubiquity.ai/vm-image-catalog: kubevirt",
		"ubuntu-24.04",
		"rocky-9",
		"windows-2022",
		"sourceURL:",
		"cloudInit:",
		"readinessBoundary: import-and-guest-boot-not-proven-by-catalog",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderVMImageCatalogRejectsBadNames(t *testing.T) {
	_, err := RenderVMImageCatalog(VMImageCatalogRequest{Name: "Bad_Name", Namespace: "ubiquity-system"})
	if err == nil || !strings.Contains(err.Error(), "DNS-compatible") {
		t.Fatalf("expected DNS validation error, got %v", err)
	}
}
