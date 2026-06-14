package nodeinventory

import (
	"strings"
	"testing"
)

func TestValidateInventorySupportsMultiOSFamilies(t *testing.T) {
	inventory := NodeInventory{
		Images: []OSImage{
			{Name: "rocky-9.4-gpu", Family: OSFamilyRocky, Version: "9.4", Architecture: "x86_64", ImageURL: "https://images.example/rocky-9.4.iso", Checksum: "sha256:rocky", Provenance: "internal-packer-2026-06-05"},
			{Name: "rhel-9.4", Family: OSFamilyRHEL, Version: "9.4", Architecture: "x86_64", ImageURL: "https://images.example/rhel-9.4.iso", Checksum: "sha256:rhel", Provenance: "vendor-subscription-sync"},
			{Name: "ubuntu-24.04-gpu", Family: OSFamilyUbuntu, Version: "24.04", Architecture: "x86_64", ImageURL: "https://images.example/ubuntu-24.04.iso", Checksum: "sha256:ubuntu", Provenance: "ubuntu-releases"},
			{Name: "custom-appliance", Family: OSFamilyCustom, Version: "1.0", Architecture: "x86_64", IPXEScript: "#!ipxe\nboot", UserData: "custom", Checksum: "sha256:custom", Provenance: "signed-vendor-bundle"},
		},
		Nodes: []BareMetalNode{{Name: "cn01", Site: "sf01", OSImageRef: "rocky-9.4-gpu", Role: "worker", MachineSelector: map[string]string{"rack": "r1"}}},
	}

	if err := inventory.Validate(); err != nil {
		t.Fatalf("expected valid inventory, got %v", err)
	}
}

func TestValidateInventoryRequiresChecksumAndProvenanceForNonDevImages(t *testing.T) {
	inventory := NodeInventory{Images: []OSImage{{Name: "ubuntu-prod", Family: OSFamilyUbuntu, Version: "24.04", Architecture: "x86_64", ImageURL: "https://images.example/ubuntu.iso"}}}

	err := inventory.Validate()
	if err == nil {
		t.Fatal("expected missing checksum/provenance to fail")
	}
	message := err.Error()
	if !strings.Contains(message, "ubuntu-prod") || !strings.Contains(message, "checksum") || !strings.Contains(message, "provenance") {
		t.Fatalf("expected image name and missing fields in error, got %q", message)
	}
}

func TestValidateInventoryAllowsDevImagesWithoutChecksumOrProvenance(t *testing.T) {
	inventory := NodeInventory{Images: []OSImage{{Name: "dev-ubuntu", Family: OSFamilyUbuntu, Version: "24.04", Architecture: "x86_64", ImageURL: "http://lab/ubuntu.iso", Dev: true}}}

	if err := inventory.Validate(); err != nil {
		t.Fatalf("expected dev image to skip checksum/provenance requirement, got %v", err)
	}
}

func TestValidateInventoryRejectsUnknownFamilyAndMissingImageRefs(t *testing.T) {
	inventory := NodeInventory{
		Images: []OSImage{{Name: "weird", Family: "solaris", Version: "11", Architecture: "sparc", Dev: true}},
		Nodes:  []BareMetalNode{{Name: "cn01", Site: "sf01", OSImageRef: "missing"}},
	}

	err := inventory.Validate()
	if err == nil {
		t.Fatal("expected invalid inventory")
	}
	message := err.Error()
	for _, want := range []string{"unsupported OS family", "solaris", "cn01", "missing"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected error to contain %q, got %q", want, message)
		}
	}
}
