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

var rootCmd = &cobra.Command{
	Use:   "ubiquity",
	Short: "HPC Cluster Lifecycle Platform",
	Long: `Ubiquity provisions, operates, and updates self-hosted HPC clusters
using Infrastructure as Code and GitOps principles.

Complete documentation: https://ubiquitycluster.github.io/ubiquity`,
}

// Execute adds all child commands to the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file (default is .ubiquity.yaml in project root)")
	rootCmd.PersistentFlags().String("env", "sandbox", "deployment environment (sandbox, dev, prod)")
}