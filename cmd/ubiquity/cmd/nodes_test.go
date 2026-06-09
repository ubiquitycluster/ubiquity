package cmd

import (
	"strings"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/nico"
)

func TestNodesCommandRegisteredWithSubcommands(t *testing.T) {
	cmd := findCommand(rootCmd, "nodes")
	if cmd == nil {
		t.Fatal("expected nodes command")
	}
	if cmd.Flags().Lookup("backend") == nil && cmd.PersistentFlags().Lookup("backend") == nil {
		t.Fatal("expected --backend flag")
	}
	for _, name := range []string{"list", "status", "os", "add", "enroll", "inspect", "image", "cordon", "drain", "maintenance", "remove", "reinstall", "reboot", "power", "task"} {
		if findCommand(cmd, name) == nil {
			t.Fatalf("expected nodes %s command", name)
		}
	}
	osCmd := findCommand(cmd, "os")
	for _, name := range []string{"list", "apply"} {
		if findCommand(osCmd, name) == nil {
			t.Fatalf("expected nodes os %s command", name)
		}
	}
	statusCmd := findCommand(cmd, "status")
	if findCommand(statusCmd, "reconcile") == nil {
		t.Fatalf("expected nodes status reconcile command")
	}
}

func TestNodesDefaultBackendNICOAndAbsentConfigFailsClosed(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "table", DryRun: false}
	t.Setenv("UBIQUITY_NICO_BASE_URL", "")
	t.Setenv("UBIQUITY_NICO_MODE", "")
	err := requireNodeBackend(nico.Config{})
	if err == nil || !strings.Contains(err.Error(), "NICo config absent") {
		t.Fatalf("expected fail-closed NICo config error, got %v", err)
	}
}

func TestNodesDestructiveRequiresConfirm(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: true, Confirm: ""}
	err := runNodesAction("remove", true)(nodesCmd, []string{"node-a"})
	if err == nil || !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("expected --confirm error, got %v", err)
	}
}

func TestNodesJSONRedactsSecrets(t *testing.T) {
	t.Setenv("UBIQUITY_NICO_TOKEN", "super-secret-token")
	out := redactNodeOutput(`{"token":"super-secret-token"}`)
	if strings.Contains(out, "super-secret-token") || !strings.Contains(out, "<redacted>") {
		t.Fatalf("secret was not redacted: %s", out)
	}
}

func TestNICOHealthRendererFailsClosed(t *testing.T) {
	out := renderNICOReadinessStatus(nico.EvaluateReadiness(nico.ReadinessSnapshot{Services: map[string]bool{}}, nico.ReadinessOptions{}))
	if !strings.Contains(out, "NOT READY") || !strings.Contains(out, "fail closed") {
		t.Fatalf("expected fail-closed NOT READY text, got %q", out)
	}
	if strings.Contains(out, "bare-metal lifecycle readiness: READY") {
		t.Fatalf("renderer claimed lifecycle ready without NICo readiness: %q", out)
	}
}
