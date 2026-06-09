package cloud

import (
	"strings"
	"testing"
)

func TestProductionSecurityCIIncludesSBOMAndSigning(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/ci.yaml")
	script := mustRead(t, "../../test/e2e/sbom-and-signing-proof.sh")
	for path, content := range map[string]string{
		"ci":     workflow,
		"script": script,
	} {
		for _, required := range []string{"syft", "cyclonedx", "cosign sign", "cosign verify", "sbom.spdx.json", "sbom.cyclonedx.json", "--dry-run"} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s missing %q", path, required)
			}
		}
	}
}
