package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestCoreOperationsRunbooksAreActionable(t *testing.T) {
	runbooks := map[string][]string{
		"../../docs/admin-guide/runbooks/cert-manager.md": {
			"## Scope",
			"## Health checks",
			"## Certificate renewal and failure response",
			"## Recovery procedure",
			"## Escalation and evidence",
			"kubectl get certificates",
			"kubectl describe certificaterequest",
		},
		"../../docs/admin-guide/runbooks/vault.md": {
			"## Scope",
			"## Health checks",
			"## Unseal and recovery",
			"## Credential rotation",
			"## Backup and restore",
			"## Escalation and evidence",
			"kubectl exec",
			"vault status",
		},
	}

	for path, required := range runbooks {
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(contentBytes)
		for _, forbidden := range []string{"TODO", "TBD", "FIXME"} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s still contains unresolved marker %q", path, forbidden)
			}
		}
		for _, needle := range required {
			if !strings.Contains(content, needle) {
				t.Fatalf("%s missing required runbook content %q", path, needle)
			}
		}
	}
}
