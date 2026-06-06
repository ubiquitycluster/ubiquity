package virtualization

import (
	"os"
	"strings"
	"testing"
)

func TestKubeVirtVMChartRendersVMsOSProfilesNetworkAndGPU(t *testing.T) {
	for _, path := range []string{
		"../../platform/kubevirt-vms/Chart.yaml",
		"../../platform/kubevirt-vms/values.yaml",
		"../../platform/kubevirt-vms/templates/virtualmachine.yaml",
		"../../platform/kubevirt-vms/templates/datavolume.yaml",
		"../../platform/kubevirt-vms/templates/networkattachmentdefinition.yaml",
		"../../platform/kubevirt-vms/templates/networkpolicy.yaml",
		"../../platform/kubevirt-vms/templates/service.yaml",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected KubeVirt VM chart file %s: %v", path, err)
		}
		if strings.TrimSpace(string(content)) == "" {
			t.Fatalf("%s must not be empty", path)
		}
	}
	values := mustReadFile(t, "../../platform/kubevirt-vms/values.yaml")
	for _, required := range []string{"ubuntu-24.04", "rocky-9", "windows-2022", "networkIsolation:", "gpu:", "resourceName:", "instanceType:", "preference:", "attachmentMode:", "external:", "ports:"} {
		if !strings.Contains(values, required) {
			t.Fatalf("kubevirt-vms values missing %q", required)
		}
	}
	vmTemplate := mustReadFile(t, "../../platform/kubevirt-vms/templates/virtualmachine.yaml")
	if strings.Contains(vmTemplate, "dataVolumeTemplates:") {
		t.Fatal("VirtualMachine chart must reference the standalone DataVolume instead of rendering a duplicate dataVolumeTemplates entry with the same name")
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(content)
}
