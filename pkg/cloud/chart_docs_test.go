package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestVMDiskChartAndDocsExist(t *testing.T) {
	for _, path := range []string{
		"../../platform/kubevirt-vm-disks/Chart.yaml",
		"../../platform/kubevirt-vm-disks/values.yaml",
		"../../platform/kubevirt-vm-disks/templates/disk.yaml",
		"../../platform/tenant-vpc/Chart.yaml",
		"../../platform/tenant-vpc/values.yaml",
		"../../platform/tenant-vpc/templates/vpc.yaml",
		"../../docs/admin-guide/cloud-services.md",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
		if strings.TrimSpace(string(content)) == "" {
			t.Fatalf("%s must not be empty", path)
		}
	}
	values := mustRead(t, "../../platform/kubevirt-vm-disks/values.yaml")
	for _, required := range []string{"source:", "blank", "http", "pvc", "storageClassName"} {
		if !strings.Contains(values, required) {
			t.Fatalf("values missing %q", required)
		}
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(content)
}
