package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("findRepoRoot failed: %v", err)
	}
	if root == "" {
		t.Fatal("findRepoRoot returned empty path")
	}
	// Should find the Makefile
	makefile := filepath.Join(root, "Makefile")
	if _, err := os.Stat(makefile); os.IsNotExist(err) {
		t.Errorf("expected Makefile at %s", makefile)
	}
}

func TestHasExistingConfig(t *testing.T) {
	// In the repo dir, should find .env or return false
	root, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	// Should not panic
	_ = hasExistingConfig(root)
}

func TestConfigureCmdParseArgs(t *testing.T) {
	// Verify the configure command responds to --help
	configureCmd.SetArgs([]string{"--help"})
	err := configureCmd.Help()
	if err != nil {
		t.Logf("configure help: %v", err)
	}
}

func TestConfigFlagDefaultHelpMessage(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("config")
	if flag == nil {
		t.Fatal("expected --config flag")
	}
	if flag.Usage == "" {
		t.Error("expected usage message for --config flag")
	}
}

func TestRootCmdHasVersionFlag(t *testing.T) {
	flag := rootCmd.Flags().Lookup("version")
	if flag == nil {
		t.Fatal("expected --version flag on root command")
	}
	if flag.Shorthand != "v" {
		t.Errorf("expected shorthand 'v', got %q", flag.Shorthand)
	}
}

func TestAllCommandsHaveHelp(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden {
			continue
		}
		if cmd.Short == "" {
			t.Errorf("command %q has no Short description", cmd.Name())
		}
	}
}
