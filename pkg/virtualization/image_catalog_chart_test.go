package virtualization

import (
	"os"
	"strings"
	"testing"
)

func TestKubeVirtVMImageCatalogChartExistsAndDocumentsReadinessBoundary(t *testing.T) {
	for _, path := range []string{
		"../../platform/kubevirt-vm-image-catalog/Chart.yaml",
		"../../platform/kubevirt-vm-image-catalog/values.yaml",
		"../../platform/kubevirt-vm-image-catalog/templates/catalog.yaml",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected image catalog chart file %s: %v", path, err)
		}
		text := string(content)
		if strings.TrimSpace(text) == "" {
			t.Fatalf("%s must not be empty", path)
		}
	}
	values := mustReadFile(t, "../../platform/kubevirt-vm-image-catalog/values.yaml")
	for _, required := range []string{"ubuntu-24.04", "rocky-9", "windows-2022", "readinessBoundary:"} {
		if !strings.Contains(values, required) {
			t.Fatalf("image catalog values missing %q", required)
		}
	}
}
