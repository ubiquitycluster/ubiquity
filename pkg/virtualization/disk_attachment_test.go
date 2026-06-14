package virtualization

import (
	"strings"
	"testing"
)

func TestRenderVirtualMachineAttachesStandaloneDisksAndBootDisk(t *testing.T) {
	manifest, err := RenderVM(VMRequest{
		Name:      "ai-notebook",
		Namespace: "tenant-a",
		OS:        "ubuntu-24.04",
		BootDisk:  "golden-ubuntu",
		DataDisks: []DiskAttachment{
			{Name: "datasets", PVCName: "datasets-pvc"},
			{Name: "checkpoints", PVCName: "checkpoints-pvc"},
		},
	})
	if err != nil {
		t.Fatalf("RenderVM returned error: %v", err)
	}
	for _, required := range []string{
		"name: bootdisk",
		"claimName: golden-ubuntu",
		"name: datasets",
		"claimName: datasets-pvc",
		"name: checkpoints",
		"claimName: checkpoints-pvc",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
	if strings.Contains(manifest, "name: rootdisk\n          dataVolume:") || strings.Contains(manifest, "kind: DataVolume") {
		t.Fatalf("booting from an existing standalone disk must not render a new root DataVolume:\n%s", manifest)
	}
}

func TestRenderVirtualMachineRejectsUnsafeDiskAttachmentNames(t *testing.T) {
	_, err := RenderVM(VMRequest{Name: "bad", DataDisks: []DiskAttachment{{Name: "Bad_Name", PVCName: "pvc"}}})
	if err == nil || !strings.Contains(err.Error(), "disk attachment") {
		t.Fatalf("expected disk attachment validation error, got %v", err)
	}
}
