package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestDeploymentProductionDocsAreActionable(t *testing.T) {
	docs := map[string][]string{
		"../../docs/admin-guide/deployment/external-resources.md": {
			"API token",
			"external DNS",
			"load balancer",
			"credential ownership",
			"rotation",
			"backup",
		},
		"../../docs/admin-guide/deployment/post-installation.md": {
			"credential ownership",
			"rotation",
			"post-install health checks",
			"backup and restore",
			"upgrade and rollback",
			"dry-run/local proof",
		},
		"../../docs/admin-guide/deployment/production/external-resources.md": {
			"API token",
			"external DNS",
			"load balancer",
			"credential ownership",
			"rotation",
			"backup",
		},
		"../../docs/admin-guide/deployment/production/post-installation.md": {
			"credential ownership",
			"rotation",
			"post-install health checks",
			"backup and restore",
			"upgrade and rollback",
			"dry-run/local proof",
		},
		"../../docs/admin-guide/deployment/production/configuration.md": {
			"example input",
			"production-lite",
			"dry-run/local proof",
			"live production proof",
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
				t.Fatalf("%s missing required production guidance %q", path, needle)
			}
		}
	}
}
