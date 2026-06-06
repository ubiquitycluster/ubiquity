package virtualization

import (
	"strings"
	"testing"
)

func TestRenderVirtualMachineSupportsOSProfilesNetworkIsolationAndGPU(t *testing.T) {
	manifest, err := RenderVM(VMRequest{
		Name:         "trainer-ubuntu",
		Namespace:    "tenant-a",
		OS:           "ubuntu-24.04",
		CPUCores:     8,
		Memory:       "32Gi",
		DiskSize:     "120Gi",
		StorageClass: "fast-nvme",
		Network: NetworkRequest{
			Isolation: NetworkIsolationMultus,
			Name:      "tenant-a-rdma",
			CIDR:      "10.44.0.0/24",
			Gateway:   "10.44.0.1",
			Bridge:    "br-tenant-a",
		},
		GPU:              GPURequest{Enabled: true, ResourceName: "nvidia.com/GA100_A100_PCIE_40GB", Count: 1},
		SSHAuthorizedKey: "ssh-ed25519 AAAAexample reviewer@example",
	})
	if err != nil {
		t.Fatalf("RenderVM returned error: %v", err)
	}
	for _, required := range []string{
		"kind: Namespace",
		"name: tenant-a",
		"kind: NetworkAttachmentDefinition",
		"br-tenant-a",
		"kind: NetworkPolicy",
		"kind: DataVolume",
		"https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img",
		"kind: VirtualMachine",
		"runStrategy: Manual",
		"nvidia.com/GA100_A100_PCIE_40GB",
		"name: gpu0",
		"ssh-ed25519 AAAAexample reviewer@example",
		"ubiquity.ai/os-profile: ubuntu-24.04",
		"ubiquity.ai/network-isolation: multus",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("rendered manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderVirtualMachineSupportsDifferentOperatingSystems(t *testing.T) {
	cases := map[string]string{
		"ubuntu-24.04": "noble-server-cloudimg-amd64.img",
		"rocky-9":      "Rocky-9-GenericCloud-Base.latest.x86_64.qcow2",
		"windows-2022": "virtio-win",
	}
	for osName, expected := range cases {
		manifest, err := RenderVM(VMRequest{Name: "vm-" + strings.ReplaceAll(osName, ".", "-"), Namespace: "tenant-a", OS: osName})
		if err != nil {
			t.Fatalf("RenderVM(%s) returned error: %v", osName, err)
		}
		if !strings.Contains(manifest, expected) {
			t.Fatalf("%s manifest missing OS-specific source %q:\n%s", osName, expected, manifest)
		}
	}
}

func TestRenderVirtualMachineFailsClosedForUnknownOSOrUnsafeName(t *testing.T) {
	if _, err := RenderVM(VMRequest{Name: "bad", Namespace: "tenant-a", OS: "plan9"}); err == nil {
		t.Fatal("expected unknown OS to fail closed")
	}
	if _, err := RenderVM(VMRequest{Name: "Bad_Name", Namespace: "tenant-a", OS: "ubuntu-24.04"}); err == nil {
		t.Fatal("expected non-DNS VM name to fail closed")
	}
}
