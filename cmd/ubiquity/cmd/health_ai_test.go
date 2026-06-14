package cmd

import (
	"strings"
	"testing"
)

func TestHealthAIFlagFailsClosedWhenNvidiaEvidenceMissing(t *testing.T) {
	cmd := findCommand(rootCmd, "health")
	if cmd == nil {
		t.Fatal("expected health command")
	}
	flag := cmd.Flags().Lookup("ai")
	if flag == nil {
		t.Fatal("expected --ai flag on health command")
	}
	oldArgs := cmd.Flags().Args()
	_ = oldArgs
	cmd.Flags().Set("ai", "true")
	defer cmd.Flags().Set("ai", "false")
	output := captureOutput(func() {
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Fatalf("expected health --ai to fail closed without live NVIDIA evidence")
		}
		if !strings.Contains(err.Error(), "NVIDIA AI platform is not ready") {
			t.Fatalf("expected AI readiness error, got %v", err)
		}
	})
	assertContains(t, output, "NVIDIA AI platform readiness: NOT READY")
	assertContains(t, output, "fail closed")
}
