package aiplatform

import (
	"strings"
	"testing"
)

func TestReleaseWorkflowGeneratesSBOMAndSignsArtifacts(t *testing.T) {
	workflow := mustRead(t, "../../.github/workflows/release.yaml")
	for _, required := range []string{
		"permissions:",
		"id-token: write",
		"anchore/sbom-action",
		"cosign-installer",
		"cosign sign-blob",
		"*.sbom.spdx.json",
		"checksums.txt",
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("release workflow must include %q", required)
		}
	}
}
