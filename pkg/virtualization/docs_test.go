package virtualization

import (
	"strings"
	"testing"
)

func TestKubeVirtAdminGuideDocumentsGPUAndNetworkIsolationBoundaries(t *testing.T) {
	doc := mustReadFile(t, "../../docs/admin-guide/kubevirt-virtual-machines.md")
	for _, required := range []string{
		"KubeVirt",
		"Containerized Data Importer",
		"ubuntu-24.04",
		"rocky-9",
		"windows-2022",
		"Multus",
		"NetworkAttachmentDefinition",
		"NetworkPolicy",
		"instance-type",
		"external-port",
		"boot-disk",
		"attach-disk",
		"standalone disk",
		"image-catalog",
		"import-and-guest-boot-not-proven-by-catalog",
		"virtual-machines readiness",
		"guest health evidence",
		"ubuntu-dev-guest-health-passed",
		"hostDevice",
		"NVIDIA GPU Operator",
		"permittedHostDevices",
		"not a GPU readiness or NVIDIA certification claim",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("KubeVirt VM guide missing %q", required)
		}
	}
}
