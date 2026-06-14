package cloud

import (
	"os"
	"strings"
	"testing"
)

func TestGraphifyWorkflowDocumentationAndFreshnessCheck(t *testing.T) {
	docBytes, err := os.ReadFile("../../docs/developers/graphify-workflow.md")
	if err != nil {
		t.Fatalf("read Graphify workflow doc: %v", err)
	}
	doc := string(docBytes)
	for _, required := range []string{
		"graphify-out/graph.json",
		"graphify query",
		"graphify explain",
		"graphify path",
		"graphify update .",
		"commit material graph updates",
		"5,000 nodes",
		"token cost",
		"scripts/check-graphify-freshness.sh",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("Graphify workflow doc missing %q", required)
		}
	}

	agentsBytes, err := os.ReadFile("../../AGENTS.md")
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	agents := string(agentsBytes)
	for _, required := range []string{
		"docs/developers/graphify-workflow.md",
		"show token cost",
		"HTML visualization",
	} {
		if !strings.Contains(agents, required) {
			t.Fatalf("AGENTS.md missing Graphify rule %q", required)
		}
	}

	scriptBytes, err := os.ReadFile("../../scripts/check-graphify-freshness.sh")
	if err != nil {
		t.Fatalf("read Graphify freshness script: %v", err)
	}
	script := string(scriptBytes)
	for _, required := range []string{
		"set -euo pipefail",
		"graphify-out/graph.json",
		"graphify-out/GRAPH_REPORT.md",
		"git rev-parse HEAD",
		"--strict",
		"GRAPHIFY_FRESHNESS_STRICT",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("Graphify freshness script missing %q", required)
		}
	}
}
