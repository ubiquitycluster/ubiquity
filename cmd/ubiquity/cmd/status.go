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
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("plain", false, "plain text output (disable TUI)")
}