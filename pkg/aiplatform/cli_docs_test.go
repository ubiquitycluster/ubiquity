package aiplatform

import (
	"strings"
	"testing"
)

func TestCLIReferenceDocumentsNvidiaAIUsage(t *testing.T) {
	doc := mustRead(t, "../../docs/reference/cli.md")
	for _, required := range []string{
		"### ai-platform",
		"ubiquity ai-platform --profile ai-production",
		"ubiquity test --sandbox-deploy",
		"NVIDIA AI sandbox deploy render validation",
		"does not require NVIDIA devices",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("CLI reference must document %q", required)
		}
	}
}
