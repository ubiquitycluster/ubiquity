package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthNICOFlagRegisteredInHelpAndParses(t *testing.T) {
	cmd := findCommand(rootCmd, "health")
	if cmd == nil {
		t.Fatal("expected health command to be registered")
	}
	flag := cmd.Flags().Lookup("nico")
	if flag == nil {
		t.Fatal("expected --nico flag on health command")
	}
	if !strings.Contains(flag.Usage, "NVIDIA Infra Controller") {
		t.Fatalf("expected --nico usage to mention NVIDIA Infra Controller, got %q", flag.Usage)
	}

	help := cmd.UsageString()
	assertContains(t, help, "--nico")

	if err := cmd.Flags().Parse([]string{"--nico"}); err != nil {
		t.Fatalf("expected --nico to parse without unknown flag error: %v", err)
	}
	if err := cmd.Flags().Set("nico", "false"); err != nil {
		t.Fatalf("reset --nico: %v", err)
	}
}

func TestHealthNICOOnlyRendersNICOReadiness(t *testing.T) {
	kubectlLog := installFailingKubectl(t)

	cmd := findCommand(rootCmd, "health")
	if cmd == nil {
		t.Fatal("expected health command to be registered")
	}
	if err := cmd.Flags().Set("nico", "true"); err != nil {
		t.Fatalf("set --nico: %v", err)
	}
	defer cmd.Flags().Set("nico", "false")

	output := captureOutput(func() {
		if err := cmd.RunE(cmd, []string{}); err != nil {
			t.Fatalf("health --nico returned error: %v", err)
		}
	})

	assertContains(t, output, "NVIDIA Infra Controller bare-metal lifecycle readiness: NOT READY")
	assertContains(t, output, "policy: fail closed")
	for _, unwanted := range []string{
		"kubectl connectivity",
		"ArgoCD server",
		"All checks passed",
		"Some checks failed",
		"NVIDIA AI platform readiness",
		"NVIDIA AIStore data-plane readiness",
	} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("health --nico output unexpectedly contained %q:\n%s", unwanted, output)
		}
	}

	logBytes, err := os.ReadFile(kubectlLog)
	if err != nil {
		t.Fatalf("read fake kubectl log: %v", err)
	}
	log := string(logBytes)
	for _, unwanted := range []string{"cluster-info", "argocd", "gpu-operator", "aistore"} {
		if strings.Contains(log, unwanted) {
			t.Fatalf("health --nico unexpectedly ran generic check containing %q; kubectl log:\n%s", unwanted, log)
		}
	}
}

func installFailingKubectl(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "kubectl.log")
	scriptPath := filepath.Join(dir, "kubectl")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$KUBECTL_LOG\"\nexit 1\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake kubectl: %v", err)
	}
	t.Setenv("KUBECTL_LOG", logPath)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}
