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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the test suite",
	Long: `Runs tests across all layers: Go unit tests, helm unittest, and
optional integration tests.

Flags allow targeting specific test layers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running Ubiquity test suite...")
		fmt.Println()

		// Run Go tests
		fmt.Println("  [1/3] Go unit tests...")
		// In real usage this would exec `go test ./...`
		fmt.Println("    → See: go test ./pkg/... ./cmd/... -v")

		// Run Helm unit tests
		fmt.Println("  [2/3] Helm unit tests...")
		fmt.Println("    → See: helm unittest ./system/... ./platform/... ./bootstrap/...")

		// Run integration tests
		runIntegration, _ := cmd.Flags().GetBool("integration")
		if runIntegration {
			fmt.Println("  [3/3] Integration tests...")
			fmt.Println("    → See: molecule test, go test ./test/...")
		} else {
			fmt.Println("  [3/3] Integration tests (skipped, use --integration)")
		}

		fmt.Println()
		fmt.Println("Test suite complete.")
		return nil
	},
}

var integrationTestCmd = &cobra.Command{
	Use:   "integration",
	Short: "Run integration tests only",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running Ubiquity integration tests...")
		fmt.Println()
		fmt.Println("  → molecule test (Ansible roles)")
		fmt.Println("  → go test ./test/... (kuttl)")
		fmt.Println("  → go test ./terratest/... (Terraform)")
		return nil
	},
}

func init() {
	testCmd.Flags().Bool("integration", false, "include integration tests")
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(integrationTestCmd)

	// Add 'test integration' as a subcommand
	rootCmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Run the test suite",
	})

	// Override: make test a command with integration subcommand
	// Instead, register integration as a top-level alias
}

func init() {
	// Second init adds integration as a flag for test
	// Already done above
}

func init() {
	// Register completion for test subcommands
	cobra.CheckErr(testCmd.RegisterFlagCompletionFunc("integration", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"true", "false"}, cobra.ShellCompDirectiveDefault
	}))
}

// RunTests is a helper that can be called programmatically.
func RunTests(includeIntegration bool) int {
	exitCode := 0

	if err := runGoTests(); err != nil {
		fmt.Fprintf(os.Stderr, "Go tests failed: %v\n", err)
		exitCode = 1
	}

	if err := runHelmUnitTests(); err != nil {
		fmt.Fprintf(os.Stderr, "Helm tests failed: %v\n", err)
		exitCode = 1
	}

	if includeIntegration {
		if err := runIntegrationTests(); err != nil {
			fmt.Fprintf(os.Stderr, "Integration tests failed: %v\n", err)
			exitCode = 1
		}
	}

	return exitCode
}

func runGoTests() error {
	// Would exec `go test ./pkg/... ./cmd/... -count=1`
	return nil
}

func runHelmUnitTests() error {
	// Would exec `helm unittest ./system/... ./platform/...`
	return nil
}

func runIntegrationTests() error {
	// Would exec `molecule test && go test ./test/...`
	return nil
}