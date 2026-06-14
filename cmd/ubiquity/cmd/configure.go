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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/config"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Ubiquity cluster settings",
	Long: `Interactive configuration wizard for Ubiquity clusters.
Prompts for domain, networking, storage, and other settings,
then patches values across Helm charts and config files.

This replaces the Python scripts/configure and scripts/configure-sandbox.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive, _ := cmd.Flags().GetBool("interactive")
		domain, _ := cmd.Flags().GetString("domain")

		// Find repo root (current dir or parent with .git)
		repoRoot, err := findRepoRoot()
		if err != nil {
			return fmt.Errorf("finding repo root: %w", err)
		}

		// Load existing configuration
		cfg, err := config.Load(repoRoot)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if domain != "" {
			cfg.Domain = domain
		}

		if interactive || (domain == "" && !hasExistingConfig(repoRoot)) {
			scanner := config.NewScanner()
			config.RunInteractive(cfg, scanner)
		}

		// Save config
		if err := config.Save(cfg, repoRoot); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		// Patch values across the repo
		if err := config.PatchValues(cfg, repoRoot); err != nil {
			return fmt.Errorf("patching values: %w", err)
		}

		fmt.Printf("Ubiquity configuration complete for domain: %s\n", cfg.Domain)
		return nil
	},
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		gitPath := filepath.Join(dir, ".git")
		makePath := filepath.Join(dir, "Makefile")
		if _, err := os.Stat(gitPath); err == nil {
			if _, err := os.Stat(makePath); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding the repo
			return wd, nil
		}
		dir = parent
	}
}

func hasExistingConfig(repoRoot string) bool {
	_, err := os.Stat(repoRoot + "/.env")
	return err == nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.Flags().BoolP("interactive", "i", false, "run interactive configuration wizard")
	configureCmd.Flags().String("domain", "", "set cluster domain (non-interactive mode)")
}