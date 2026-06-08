package cmd

import (
	"strings"
	"testing"
)

func TestHealthAIStoreFlagFailsClosedWhenEvidenceMissing(t *testing.T) {
	cmd := findCommand(rootCmd, "health")
	if cmd == nil {
		t.Fatal("expected health command")
	}
	flag := cmd.Flags().Lookup("aistore")
	if flag == nil {
		t.Fatal("expected --aistore flag on health command")
	}
	if err := cmd.Flags().Set("aistore", "true"); err != nil {
		t.Fatalf("set --aistore: %v", err)
	}
	defer cmd.Flags().Set("aistore", "false")
	output := captureOutput(func() {
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Fatalf("expected health --aistore to fail closed without live AIStore evidence")
		}
		if !strings.Contains(err.Error(), "NVIDIA AIStore data-plane is not ready") {
			t.Fatalf("expected AIStore readiness error, got %v", err)
		}
	})
	assertContains(t, output, "NVIDIA AIStore data-plane readiness: NOT READY")
	assertContains(t, output, "not a generic PVC replacement")
	assertContains(t, output, "fail closed for AIStore claims")
}
