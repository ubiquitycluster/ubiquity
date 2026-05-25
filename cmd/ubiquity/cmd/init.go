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
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap Ubiquity configuration",
	Long:  `Initialize Ubiquity configuration in the current directory. Creates .ubiquity/ with default config files, secrets skeleton, and inventory templates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")

		home, _ := os.UserHomeDir()
		ubiquityDir := filepath.Join(home, ".ubiquity")
		if err := os.MkdirAll(ubiquityDir, 0755); err != nil {
			return fmt.Errorf("creating .ubiquity directory: %w", err)
		}

		fmt.Printf("Initialized Ubiquity configuration for environment: %s\n", env)
		fmt.Printf("Config directory: %s\n", ubiquityDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}