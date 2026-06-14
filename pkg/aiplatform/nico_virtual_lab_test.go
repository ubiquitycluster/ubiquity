package aiplatform

import (
	"strings"
	"testing"
)

func TestVirtualBareMetalNICoValidationScriptExistsAndIsGated(t *testing.T) {
	script := mustRead(t, "../../test/e2e/nico-virtual-bare-metal-lab.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_NICO_VIRTUAL_BARE_METAL_E2E=true",
		"qemu-system-x86_64",
		"redfish",
		"sushy-tools",
		"PXE",
		"kubectl apply --server-side",
		"nodes status",
		"nodes power",
		"nodes apply-os",
		"destructive actions remain dry-run",
		"exit 0",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("virtual NICo lab script missing %q", required)
		}
	}
}
