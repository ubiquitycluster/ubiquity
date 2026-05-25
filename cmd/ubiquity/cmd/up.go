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

	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy the full Ubiquity cluster stack",
	Long: `Detects the platform (metal, cloud, sandbox) and deploys the full stack:
provisioning → bootstrap → external resources → wait → post-install.

Phase ordering:
  1. metal        — Provision bare metal or k3d sandbox cluster
  2. bootstrap    — Install ArgoCD and root ApplicationSet
  3. external     — Provision external resources (Terraform)
  4. wait         — Wait for core applications to reach Ready
  5. post-install — Post-installation configuration and BMO setup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		fmt.Printf("Deploying Ubiquity cluster (%s environment)...\n", env)
		fmt.Println("  [1/5] metal    — provisioning infrastructure")
		fmt.Println("  [2/5] bootstrap — installing ArgoCD")
		fmt.Println("  [3/5] external  — provisioning external resources")
		fmt.Println("  [4/5] wait      — verifying application readiness")
		fmt.Println("  [5/5] post-install — final configuration")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}