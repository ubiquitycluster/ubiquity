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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Tear down the Ubiquity cluster",
	Long:  `Destroys the cluster and associated resources. Use with caution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := provision.LoadState()
		if err != nil {
			return fmt.Errorf("reading provisioning state: %w", err)
		}
		if state == nil {
			fmt.Println("No provisioning state found.")
			return nil
		}

		env := state.Environment

		// Sandbox / k3d environments
		if env == "sandbox" || strings.Contains(env, "sandbox") {
			k3dCmd := exec.Command("k3d", "cluster", "delete", "ubiquity-dev")
			if err := k3dCmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 127 {
					fmt.Println("k3d not available, skipping cluster deletion.")
				} else {
					fmt.Printf("k3d cluster delete failed: %v\n", err)
				}
			} else {
				fmt.Println("k3d cluster deleted.")
			}
		}

		// Cloud providers
		cloudProviders := []string{"aws", "azure", "gcp", "openstack", "ovh"}
		for _, provider := range cloudProviders {
			if env == provider || strings.Contains(env, provider) {
				terraformCmd := exec.Command("terraform", "destroy", "-auto-approve")
				terraformCmd.Dir = fmt.Sprintf("cloud/%s", provider)
				if err := terraformCmd.Run(); err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 127 {
						fmt.Printf("terraform not available, skipping %s teardown.\n", provider)
					} else {
						fmt.Printf("terraform destroy for %s failed: %v\n", provider, err)
					}
				} else {
					fmt.Printf("Terraform destroyed %s resources.\n", provider)
				}
				break
			}
		}

		// Delete the state file
		if err := os.Remove(provision.StatePath()); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing state file: %w", err)
		}
		fmt.Println("Provisioning state cleared.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}