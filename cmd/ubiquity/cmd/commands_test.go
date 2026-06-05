package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "version")
	if cmd == nil {
		t.Fatal("expected version command to be registered")
	}
	if cmd.Use != "version" {
		t.Errorf("expected Use 'version', got %q", cmd.Use)
	}
}

func TestVersionJSONFlag(t *testing.T) {
	cmd := findCommand(rootCmd, "version")
	if cmd == nil {
		t.Fatal("expected version command")
	}
	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("expected --json flag on version command")
	}
}

func TestVersionFlagOnRoot(t *testing.T) {
	// The --version flag is set up in the root command's persistent flags
	// via cobra.OnInitialize. Check it exists among all flags.
	flag := rootCmd.Flags().Lookup("version")
	if flag == nil {
		// Also check persistent flags
		flag = rootCmd.PersistentFlags().Lookup("version")
	}
	if flag == nil {
		t.Fatal("expected --version flag on root command")
	}
}

func TestConfigureCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "configure")
	if cmd == nil {
		t.Fatal("expected configure command to be registered")
	}
	flag := cmd.Flags().Lookup("domain")
	if flag == nil {
		t.Fatal("expected --domain flag on configure command")
	}
	flag = cmd.Flags().Lookup("interactive")
	if flag == nil {
		t.Fatal("expected --interactive flag on configure command")
	}
}

func TestRetryCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "retry")
	if cmd == nil {
		t.Fatal("expected retry command to be registered")
	}
	if len(cmd.ValidArgs) == 0 {
		t.Error("expected ValidArgs to be set on retry command")
	}
}

func TestStatusCmdPlainFlag(t *testing.T) {
	cmd := findCommand(rootCmd, "status")
	if cmd == nil {
		t.Fatal("expected status command")
	}
	flag := cmd.Flags().Lookup("plain")
	if flag == nil {
		t.Fatal("expected --plain flag on status command")
	}
}

func TestUpSkipSecurityFlag(t *testing.T) {
	cmd := findCommand(rootCmd, "up")
	if cmd == nil {
		t.Fatal("expected up command")
	}
	for _, name := range []string{"skip-security", "sandbox", "metal-bootstrap-backend", "node-lifecycle-backend", "nico-values", "nico-rest-values", "nico-site"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("expected --%s flag on up command", name)
		}
	}
}

func TestIntegrationCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "integration")
	if cmd == nil {
		t.Fatal("expected integration command to be registered")
	}
}

func TestInitCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "init")
	if cmd == nil {
		t.Fatal("expected init command")
	}
}

func TestDownCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "down")
	if cmd == nil {
		t.Fatal("expected down command")
	}
}

func TestLogsCmdRegistered(t *testing.T) {
	cmd := findCommand(rootCmd, "logs")
	if cmd == nil {
		t.Fatal("expected logs command")
	}
}

func TestVersionCmdExecute(t *testing.T) {
	buf := captureOutput(func() {
		versionCmd.RunE(versionCmd, []string{})
	})
	if len(buf) == 0 {
		t.Error("expected version output")
	}
}

func TestStatusCmdExecuteWithNoState(t *testing.T) {
	// Without a state file, status should print a help message
	buf := captureOutput(func() {
		statusCmd.RunE(statusCmd, []string{})
	})
	if len(buf) == 0 {
		t.Error("expected status output")
	}
}

func TestLogsCmdExecuteWithNoState(t *testing.T) {
	buf := captureOutput(func() {
		logsCmd.RunE(logsCmd, []string{})
	})
	if len(buf) == 0 {
		t.Error("expected logs output")
	}
}

func TestDownCmdExecuteWithNoState(t *testing.T) {
	buf := captureOutput(func() {
		downCmd.RunE(downCmd, []string{})
	})
	if len(buf) == 0 {
		t.Error("expected down output")
	}
}

func TestRetryCmdExecute(t *testing.T) {
	// Retry with a phase arg — should not panic
	err := retryCmd.RunE(retryCmd, []string{"metal"})
	if err != nil {
		// Expected to error if no state file exists
		t.Logf("retry expected error: %v", err)
	}
}

func TestInitCmdExecute(t *testing.T) {
	initCmd.RunE(initCmd, []string{})
}

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

func TestUpHelpExecute(t *testing.T) {
	cmd := findCommand(rootCmd, "up")
	if cmd == nil {
		t.Fatal("expected up command")
	}
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("up --help failed: %v", err)
	}
	out := cmd.Long + "\n" + cmd.Flags().Lookup("node-lifecycle-backend").Usage
	for _, want := range []string{"node lifecycle integration", "NICo", "fallback/migration-only"} {
		if !strings.Contains(out, want) {
			t.Fatalf("up help missing %q: %s", want, out)
		}
	}
}
