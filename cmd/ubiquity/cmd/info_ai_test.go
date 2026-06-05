package cmd

import "testing"

func TestInfoIncludesNvidiaAIReadinessPolicy(t *testing.T) {
	output := captureOutput(func() {
		if err := infoCmd.RunE(infoCmd, []string{}); err != nil {
			t.Fatalf("info command failed: %v", err)
		}
	})
	assertContains(t, output, "NVIDIA AI platform readiness:")
	assertContains(t, output, "fail closed")
}
