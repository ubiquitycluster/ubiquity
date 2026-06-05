package nodeinventory

import (
	"strings"
	"testing"
)

func TestRenderAnsibleBootstrapInventoryFromNodeInventory(t *testing.T) {
	inventory := NodeInventory{
		Images: []OSImage{{Name: "rocky-9.4-gpu", Family: OSFamilyRocky, Version: "9.4", Architecture: "x86_64", ImageURL: "https://images.example/rocky.iso", Checksum: "sha256:rocky", Provenance: "packer"}},
		Nodes: []BareMetalNode{
			{Name: "cp01", Site: "sf01", Role: "control-plane", OSImageRef: "rocky-9.4-gpu", InstanceTypeRef: "mgmt-small", JoinProfile: "k3s-server", GPUProfile: "none", MachineSelector: map[string]string{"rack": "r1", "serial": "abc"}, Labels: map[string]string{"node-role.kubernetes.io/control-plane": "true"}},
			{Name: "cn01", Site: "sf01", Role: "worker", OSImageRef: "rocky-9.4-gpu", InstanceTypeRef: "gpu-h100", JoinProfile: "k3s-agent", GPUProfile: "h100-8g", MachineSelector: map[string]string{"rack": "r2"}},
		},
	}

	yaml, err := RenderAnsibleBootstrapInventory(inventory)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	assertContainsAll(t, "ansible inventory", yaml,
		"# bootstrap-only",
		"# Day-2 node lifecycle is managed by NICo, not Ansible, unless fallback mode is explicit.",
		"all:",
		"hosts:",
		"cp01:",
		"cn01:",
		"bootstrap_only: true",
		"node_lifecycle_backend: nico",
		"os_image_ref: rocky-9.4-gpu",
		"join_profile: k3s-server",
		"gpu_profile: h100-8g",
		"machine_selector:",
		"serial: abc",
	)
}

func TestRenderAnsibleBootstrapInventoryValidatesSource(t *testing.T) {
	_, err := RenderAnsibleBootstrapInventory(NodeInventory{Nodes: []BareMetalNode{{Name: "bad", OSImageRef: "missing"}}})
	if err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected validation error for missing image ref, got %v", err)
	}
}
