package nodeinventory

import (
	"strings"
	"testing"
)

func TestRenderNICoOperatingSystemsForRockyRHELUbuntuAndCustom(t *testing.T) {
	inventory := NodeInventory{Images: []OSImage{
		{Name: "rocky-9.4-gpu", Family: OSFamilyRocky, Version: "9.4", Architecture: "x86_64", ImageURL: "https://images.example/rocky.iso", Checksum: "sha256:rocky", Provenance: "packer"},
		{Name: "rhel-9.4", Family: OSFamilyRHEL, Version: "9.4", Architecture: "x86_64", ImageURL: "https://images.example/rhel.iso", Checksum: "sha256:rhel", Provenance: "vendor"},
		{Name: "ubuntu-24.04-gpu", Family: OSFamilyUbuntu, Version: "24.04", Architecture: "x86_64", ImageURL: "https://images.example/ubuntu.iso", Checksum: "sha256:ubuntu", Provenance: "ubuntu-releases"},
		{Name: "custom-appliance", Family: OSFamilyCustom, Version: "1.0", Architecture: "x86_64", IPXEScript: "#!ipxe\necho custom\nboot http://custom/kernel", UserData: "#cloud-config\ncustom: true", Checksum: "sha256:custom", Provenance: "vendor"},
	}}

	objects, err := RenderNICoOperatingSystems(inventory)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if got, want := len(objects), 4; got != want {
		t.Fatalf("rendered %d objects, want %d", got, want)
	}

	byName := map[string]NICoOperatingSystem{}
	for _, object := range objects {
		byName[object.Metadata.Name] = object
		if object.Kind != "OperatingSystem" {
			t.Fatalf("%s kind = %q, want OperatingSystem", object.Metadata.Name, object.Kind)
		}
		if object.APIVersion == "" {
			t.Fatalf("%s missing API version", object.Metadata.Name)
		}
		if object.Spec.Checksum == "" || object.Spec.Provenance == "" {
			t.Fatalf("%s missing checksum/provenance in rendered spec", object.Metadata.Name)
		}
		for _, key := range []string{"ubiquity.ai/os-family", "ubiquity.ai/os-version", "ubiquity.ai/architecture", "ubiquity.ai/provenance", "ubiquity.ai/boot-mode", "ubiquity.ai/fallback-behavior"} {
			if object.Metadata.Labels[key] == "" || object.Spec.Labels[key] == "" {
				t.Fatalf("%s missing compatibility label %s: metadata=%v spec=%v", object.Metadata.Name, key, object.Metadata.Labels, object.Spec.Labels)
			}
		}
	}

	for _, name := range []string{"rocky-9.4-gpu", "rhel-9.4"} {
		obj := byName[name]
		assertContainsAll(t, name+" ipxe", obj.Spec.IPXEScript, "#!ipxe", "inst.ks=", obj.Spec.ImageURL)
		assertContainsAll(t, name+" userdata", obj.Spec.UserData, "# Kickstart", "text", "reboot")
	}

	ubuntu := byName["ubuntu-24.04-gpu"]
	assertContainsAll(t, "ubuntu ipxe", ubuntu.Spec.IPXEScript, "#!ipxe", "autoinstall", ubuntu.Spec.ImageURL)
	assertContainsAll(t, "ubuntu userdata", ubuntu.Spec.UserData, "#cloud-config", "autoinstall:", "version: 1")

	custom := byName["custom-appliance"]
	if custom.Spec.IPXEScript != "#!ipxe\necho custom\nboot http://custom/kernel" {
		t.Fatalf("custom ipxe was not passed through: %q", custom.Spec.IPXEScript)
	}
	if custom.Spec.UserData != "#cloud-config\ncustom: true" {
		t.Fatalf("custom user data was not passed through: %q", custom.Spec.UserData)
	}
}

func TestRenderNICoOperatingSystemsRejectsCustomWithoutUserBootData(t *testing.T) {
	inventory := NodeInventory{Images: []OSImage{{Name: "custom-broken", Family: OSFamilyCustom, Version: "1", Architecture: "x86_64", Checksum: "sha256:x", Provenance: "test"}}}

	errText := ""
	_, err := RenderNICoOperatingSystems(inventory)
	if err != nil {
		errText = err.Error()
	}
	if !strings.Contains(errText, "custom-broken") || !strings.Contains(errText, "IPXEScript") || !strings.Contains(errText, "UserData") {
		t.Fatalf("expected custom boot data error, got %q", errText)
	}
}

func assertContainsAll(t *testing.T, label, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("%s missing %q in:\n%s", label, want, got)
		}
	}
}
