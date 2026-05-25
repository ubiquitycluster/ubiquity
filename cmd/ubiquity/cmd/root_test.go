/*
Copyright © 2026 Ubiquity Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmdUse(t *testing.T) {
	if rootCmd.Use != "ubiquity" {
		t.Errorf("expected rootCmd.Use to be 'ubiquity', got %q", rootCmd.Use)
	}
}

func TestRootCmdSubcommands(t *testing.T) {
	expected := map[string]string{
		"init":   "init",
		"up":     "up",
		"down":   "down",
		"status": "status",
		"logs":   "logs [phase]",
	}

	registered := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		registered[cmd.Use] = true
	}

	for name, use := range expected {
		if !registered[use] {
			t.Errorf("expected subcommand %q with Use %q to be registered on rootCmd", name, use)
		}
	}
}

func TestRootCmdHasExpectedNumberOfSubcommands(t *testing.T) {
	// We expect exactly 5 subcommands: init, up, down, status, logs
	if got := len(rootCmd.Commands()); got != 5 {
		t.Errorf("expected rootCmd to have 5 subcommands, got %d", got)
	}
}

func TestEnvFlagDefault(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("env")
	if flag == nil {
		t.Fatal("expected --env persistent flag to exist on rootCmd")
	}
	if flag.DefValue != "sandbox" {
		t.Errorf("expected --env default to be 'sandbox', got %q", flag.DefValue)
	}
}

func TestEnvFlagParsing(t *testing.T) {
	cmd := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, args []string) error { return nil }}
	cmd.PersistentFlags().String("env", "sandbox", "deployment environment")

	// Override the default
	if err := cmd.PersistentFlags().Set("env", "prod"); err != nil {
		t.Fatalf("failed to set --env flag: %v", err)
	}

	val, err := cmd.PersistentFlags().GetString("env")
	if err != nil {
		t.Fatalf("failed to get --env flag: %v", err)
	}
	if val != "prod" {
		t.Errorf("expected --env to be 'prod', got %q", val)
	}
}

func TestConfigFlagExists(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("config")
	if flag == nil {
		t.Fatal("expected --config persistent flag to exist on rootCmd")
	}
}

func TestSubcommandRegistration(t *testing.T) {
	// Verify each subcommand has the correct Use field
	subcommands := rootCmd.Commands()
	uses := make(map[string]bool)
	for _, c := range subcommands {
		uses[c.Use] = true
	}

	tests := []struct {
		name  string
		use   string
		found bool
	}{
		{"init", "init", false},
		{"up", "up", false},
		{"down", "down", false},
		{"status", "status", false},
		{"logs", "logs [phase]", false},
	}

	for i := range tests {
		_, tests[i].found = uses[tests[i].use]
		if !tests[i].found {
			t.Errorf("subcommand %q not found with Use %q", tests[i].name, tests[i].use)
		}
	}
}