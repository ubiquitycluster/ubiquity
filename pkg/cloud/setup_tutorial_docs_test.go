package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestSetupTutorialDocsAreActionable(t *testing.T) {
	docs := map[string][]string{
		"../../docs/admin-guide/concepts/pxe-boot.md": {
			"DHCP",
			"TFTP",
			"HTTP",
			"BMC",
			"Troubleshooting",
			"not a live hardware proof",
		},
		"../../docs/admin-guide/administration/tutorials/install-pre-commit-hooks.md": {
			"pre-commit",
			"make git-hooks",
			"pre-commit run --all-files",
			"Troubleshooting",
		},
		"../../docs/admin-guide/administration/tutorials/updating-documentation.md": {
			"make docs",
			"Material for MkDocs",
			"Deployment",
			"DNS",
			"rollback",
			"dry-run/local proof",
		},
	}

	for path, required := range docs {
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
		lower := strings.ToLower(content)
		for _, needle := range required {
			if !strings.Contains(lower, strings.ToLower(needle)) {
				t.Fatalf("%s missing required tutorial content %q", path, needle)
			}
		}
	}
}
