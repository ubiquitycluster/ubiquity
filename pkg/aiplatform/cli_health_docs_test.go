package aiplatform

import (
	"strings"
	"testing"
)

func TestCLIReferenceDocumentsFocusedHealthReadinessCommands(t *testing.T) {
	doc := mustRead(t, "../../docs/reference/cli.md")
	for _, required := range []string{
		"ubiquity health --ai",
		"ubiquity health --aistore",
		"ubiquity health --nico",
		"fail closed",
		"AIStore data-plane",
		"not a generic PVC replacement",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("CLI reference missing %q", required)
		}
	}
}
