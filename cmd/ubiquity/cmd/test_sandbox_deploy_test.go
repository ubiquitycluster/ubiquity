package cmd

import "testing"

func TestTestCmdSupportsSandboxDeployValidation(t *testing.T) {
	flag := testCmd.Flags().Lookup("sandbox-deploy")
	if flag == nil {
		t.Fatal("expected ubiquity test --sandbox-deploy flag")
	}
	if flag.DefValue != "false" {
		t.Fatalf("expected --sandbox-deploy default false, got %q", flag.DefValue)
	}
}
