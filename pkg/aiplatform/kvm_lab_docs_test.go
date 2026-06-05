package aiplatform

import (
	"os"
	"strings"
	"testing"
)

func TestNICoKVMPXELabArtifactsDefineRunnableVirtualBareMetalTier(t *testing.T) {
	doc := mustReadProjectFile(t, "docs/reference/nvidia-infra-controller/kvm-pxe-validation-lab.md")
	script := mustReadProjectFile(t, "test/e2e/nico-kvm-pxe-lab.sh")
	topology := mustReadProjectFile(t, "test/fixtures/nico-kvm-pxe/containerlab.yml")
	inventory := mustReadProjectFile(t, "test/fixtures/nico-kvm-pxe/inventory.yaml")
	makefile := mustReadProjectFile(t, "Makefile")

	for _, required := range []string{
		"qemu-bmc",
		"Redfish",
		"IPMI",
		"PXE",
		"multi-OS",
		"UBIQUITY_NICO_KVM_LAB=1",
		"physical hardware gate",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("KVM/PXE lab doc must explain %q", required)
		}
	}

	for _, required := range []string{
		"UBIQUITY_NICO_KVM_LAB",
		"qemu-bmc",
		"containerlab",
		"wait_for_redfish",
		"wait_for_ipmi",
		"nodes os apply",
		"nodes power",
		"--confirm",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("KVM/PXE lab script must include %q", required)
		}
	}

	for _, required := range []string{
		"ghcr.io/tjst-t/qemu-bmc",
		"IPMI_PASS",
		"VM_BOOT: \"n\"",
		"VM_NETWORKS",
		"8623:623/udp",
		"8443:443",
	} {
		if !strings.Contains(topology, required) {
			t.Fatalf("KVM/PXE topology fixture must include %q", required)
		}
	}

	for _, required := range []string{
		"rocky-9",
		"ubuntu-24.04",
		"custom-ipxe",
		"protocol: redfish",
		"passwordRef: env:QEMU_BMC_PASSWORD",
	} {
		if !strings.Contains(inventory, required) {
			t.Fatalf("KVM/PXE inventory fixture must include %q", required)
		}
	}

	if !strings.Contains(makefile, "nico-kvm-pxe-lab:") || !strings.Contains(makefile, "test/e2e/nico-kvm-pxe-lab.sh") {
		t.Fatal("Makefile must expose a nico-kvm-pxe-lab target for operators")
	}

	for _, forbidden := range []string{"client-key-data:", "BEGIN PRIVATE KEY", "UBIQUITY_NICO_TOKEN: "} {
		if strings.Contains(topology, forbidden) || strings.Contains(inventory, forbidden) || strings.Contains(script, forbidden) {
			t.Fatalf("KVM/PXE lab artifacts must not contain secret material matching %q", forbidden)
		}
	}
}

func mustReadProjectFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile("../../" + path)
	if err != nil {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
