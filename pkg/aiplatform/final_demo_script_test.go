package aiplatform

import (
	"strings"
	"testing"
)

func TestNvidiaAIPlatformFinalDemoScriptIsGatedAndCoversAcceptanceFlow(t *testing.T) {
	script := mustRead(t, "../../test/e2e/nvidia-ai-platform-final-demo.sh")
	for _, required := range []string{
		"UBIQUITY_RUN_NVIDIA_AI_FINAL_DEMO=true",
		"--dry-run",
		"provision",
		"reconcile",
		"schedule",
		"serve",
		"observe",
		"validate",
		"ubiquity ai-platform apply --profile ai-production --server-side",
		"ubiquity health --ai",
		"ubiquity info --ai",
		"nvidia-ai-final-demo-passed",
		"fail closed",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("final demo script missing %q", required)
		}
	}
}
