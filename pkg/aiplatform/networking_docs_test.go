package aiplatform

import (
	"strings"
	"testing"
)

func TestNetworkingArchitectureDocDocumentsReadinessBoundaries(t *testing.T) {
	doc := mustRead(t, "../../docs/architecture/networking.md")
	if strings.Contains(doc, "TODO") {
		t.Fatalf("networking architecture doc still contains TODO")
	}
	for _, required := range []string{
		"default-deny",
		"allow-dns",
		"NetworkAttachmentDefinition",
		"nvidia.com/rdma",
		"cloudflared",
		"readiness evidence",
		"not proof of service readiness",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("networking architecture doc missing %q", required)
		}
	}
}
