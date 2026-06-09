package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchitectureADRAuditCoversRequiredDecisions(t *testing.T) {
	adrDir := "../../docs/reference/architecture/decision-records"
	index := mustRead(t, filepath.Join(adrDir, "README.md"))
	required := map[string][]string{
		"ADR-001-go-cli.md":       {"Go CLI", "Python"},
		"ADR-002-k3s.md":          {"k3s", "full Kubernetes"},
		"ADR-003-terraform.md":    {"Terraform", "cloud provisioning"},
		"ADR-004-kyverno.md":      {"Kyverno", "OPA", "Gatekeeper"},
		"ADR-005-helm.md":         {"Helm chart per component", "Kustomize"},
		"ADR-006-longhorn.md":     {"Longhorn", "primary storage"},
		"ADR-007-argocd.md":       {"ArgoCD", "Flux"},
		"ADR-008-bubbletea.md":    {"Bubbletea", "TUI"},
		"ADR-010-installer.md":    {"PXE", "Go"},
		"ADR-011-sops.md":         {"SOPS", "secrets"},
		"ADR-012-devcontainer.md": {"devcontainer", "reproducible"},
		"ADR-013-precommit-ci.md": {"pre-commit.ci", "lint"},
		"ADR-014-nico.md":         {"NVIDIA Infra Controller", "NICo", "Metal3"},
	}
	for file, tokens := range required {
		content := mustRead(t, filepath.Join(adrDir, file))
		if !strings.Contains(index, file) {
			t.Fatalf("ADR index does not link %s", file)
		}
		for _, token := range tokens {
			if !strings.Contains(content, token) {
				t.Fatalf("%s missing %q", file, token)
			}
		}
	}
}

func TestArchitectureNetworkingDocHasNoTODOs(t *testing.T) {
	content := mustRead(t, "../../docs/architecture/networking.md")
	for _, forbidden := range []string{"TODO", "TBD", "FIXME"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("networking architecture doc still contains unresolved marker %q", forbidden)
		}
	}
	for _, required := range []string{"control plane", "tenant", "storage", "ingress", "NetworkPolicy", "Cilium"} {
		if !strings.Contains(content, required) {
			t.Fatalf("networking architecture doc missing %q", required)
		}
	}
}

func TestHelmChartReferenceIsGeneratedAndCurrent(t *testing.T) {
	script := mustRead(t, "../../scripts/generate-helm-chart-reference.sh")
	reference := mustRead(t, "../../docs/reference/helm-charts.md")
	for _, required := range []string{"AUTO-GENERATED", "find system platform", "Chart.yaml", "values.yaml", "--check"} {
		if !strings.Contains(script+"\n"+reference, required) {
			t.Fatalf("helm chart reference contract missing %q", required)
		}
	}
	for _, chart := range []string{"platform/cloud-prerequisites", "platform/hpc-ubiq", "system/kyverno-policies", "system/falco-rules"} {
		if !strings.Contains(reference, chart) {
			t.Fatalf("helm chart reference missing %s", chart)
		}
	}
}

func TestCloudNVIDIAReadinessDocsUseEvidenceBoundaries(t *testing.T) {
	paths := []string{
		"../../docs/admin-guide/nvidia-ai-platform.md",
		"../../docs/runbooks/cloud-readiness-validation.md",
		"../../docs/runbooks/cloud-production-readiness-audit.md",
		"../../docs/admin-guide/kubevirt-virtual-machines.md",
		"../../docs/reference/nvidia-infra-controller/os-images.md",
	}
	for _, path := range paths {
		content := mustRead(t, path)
		for _, required := range []string{"live proof", "not NVIDIA approved", "not NVIDIA certified", "approval evidence"} {
			if !strings.Contains(strings.ToLower(content), strings.ToLower(required)) {
				t.Fatalf("%s missing boundary language %q", path, required)
			}
		}
	}
	badClaims := []string{"NVIDIA approved", "NVIDIA-approved", "NVIDIA certified", "NVIDIA-certified"}
	err := filepath.WalkDir("../../docs", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		content := mustRead(t, path)
		for _, claim := range badClaims {
			if strings.Contains(content, claim) && !strings.Contains(strings.ToLower(content), "not "+strings.ToLower(claim)) && !strings.Contains(strings.ToLower(content), "without") {
				t.Fatalf("%s contains unsupported claim %q", path, claim)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestKustomizeHelmRelationshipIsDocumented(t *testing.T) {
	doc := mustRead(t, "../../docs/architecture/kustomize-helm.md")
	overview := mustRead(t, "../../docs/architecture/overview.md")
	combined := doc + "\n" + overview
	for _, required := range []string{"platform/hpc-ubiq", "Kustomize", "Helm", "boundary", "environment-specific patches", "component packaging"} {
		if !strings.Contains(combined, required) {
			t.Fatalf("Kustomize/Helm relationship docs missing %q", required)
		}
	}
}
