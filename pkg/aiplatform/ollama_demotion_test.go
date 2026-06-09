package aiplatform

import (
	"regexp"
	"strings"
	"testing"
)

func TestOllamaChartIsDiagnosticsOnlyAndDisabledByDefault(t *testing.T) {
	values := mustRead(t, "../../apps/ollama/values.yaml")
	for _, required := range []string{
		"enabled: false",
		"mode: diagnostics-only",
		"productionDefault: false",
		"notProductionServing: true",
	} {
		if !strings.Contains(values, required) {
			t.Fatalf("apps/ollama/values.yaml must include %q to keep Ollama diagnostics-only", required)
		}
	}
	if regexp.MustCompile(`(?m)^enabled:\s+true\s*$`).MatchString(values) {
		t.Fatal("apps/ollama top-level enabled must not be true by default")
	}
}

func TestAICommandDescribesOllamaAsLocalDiagnosticsOnly(t *testing.T) {
	cmdSource := mustRead(t, "../../cmd/ubiquity/cmd/ai.go")
	for _, required := range []string{
		"local diagnostics",
		"not the production AI serving layer",
		"NIM Operator",
	} {
		if !strings.Contains(cmdSource, required) {
			t.Fatalf("ubiquity ai command source must include %q", required)
		}
	}
}
