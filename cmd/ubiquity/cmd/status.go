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
	"net/http"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
	"github.com/ubiquitycluster/ubiquity/pkg/tui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cluster health summary",
	Long:  `Reads provisioning state and displays the current cluster status across all phases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := provision.LoadState()
		if err != nil {
			return fmt.Errorf("reading provisioning state: %w", err)
		}
		if state == nil {
			fmt.Println("No provisioning state found.")
			fmt.Println("Run 'ubiquity init' then 'ubiquity up' to start deployment.")
			return nil
		}

		plain, _ := cmd.Flags().GetBool("plain")
		if plain {
			fmt.Print(state.Summary())
		} else {
			tui.PrintStatus(state)
		}

		// Check PXE installer status
		checkInstallerStatus()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("plain", false, "plain text output (disable TUI)")
}

// checkInstallerStatus queries the PXE installer phone-home API if running.
func checkInstallerStatus() {
	// Check if ubiquity-installer binary exists
	if _, err := exec.LookPath("ubiquity-installer"); err != nil {
		return
	}

	// Query the phone-home API
	resp, err := http.Get("http://localhost:8080/status")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println()
	fmt.Println("PXE Installer:")
	fmt.Println("  API: http://localhost:8080/status")
	fmt.Println("  Status: running")
}