package virtualization

import (
	"strings"
	"testing"
)

func TestRenderVirtualMachineSupportsInstanceTypesHostDevicesAndExternalPorts(t *testing.T) {
	manifest, err := RenderVM(VMRequest{
		Name:         "vgpu-workstation",
		Namespace:    "tenant-a",
		OS:           "ubuntu-24.04",
		InstanceType: "gx-a100-medium",
		Preference:   "ubuntu-server",
		GPU: GPURequest{
			Enabled:      true,
			ResourceName: "nvidia.com/L40S-24Q",
			Count:        2,
			Mode:         GPUAttachmentHostDevice,
		},
		External: ExternalAccess{
			Enabled: true,
			Ports:   []int{22, 443},
		},
	})
	if err != nil {
		t.Fatalf("RenderVM returned error: %v", err)
	}
	for _, required := range []string{
		"instancetype:",
		"name: gx-a100-medium",
		"kind: VirtualMachineClusterInstancetype",
		"preference:",
		"name: ubuntu-server",
		"hostDevices:",
		"deviceName: nvidia.com/L40S-24Q",
		"kind: Service",
		"name: vgpu-workstation-external",
		"port: 22",
		"port: 443",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
	if strings.Contains(manifest, "gpus:") {
		t.Fatalf("hostDevice attachment mode must not render gpus block:\n%s", manifest)
	}
	if strings.Contains(manifest, "cpu:\n          cores") || strings.Contains(manifest, "resources:\n          requests:\n            memory") {
		t.Fatalf("instancetype-backed VM should not also render explicit CPU/memory sizing:\n%s", manifest)
	}
}

func TestRenderVirtualMachineRejectsUnsafeExternalPortsAndGPUAttachmentMode(t *testing.T) {
	_, err := RenderVM(VMRequest{Name: "bad", External: ExternalAccess{Enabled: true, Ports: []int{0}}})
	if err == nil || !strings.Contains(err.Error(), "external port") {
		t.Fatalf("expected invalid external port error, got %v", err)
	}
	_, err = RenderVM(VMRequest{Name: "bad", GPU: GPURequest{Enabled: true, ResourceName: "nvidia.com/gpu", Mode: GPUAttachmentMode("vfio")}})
	if err == nil || !strings.Contains(err.Error(), "GPU attachment mode") {
		t.Fatalf("expected invalid GPU attachment mode error, got %v", err)
	}
}
