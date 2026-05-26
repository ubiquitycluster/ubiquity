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
)

var retryCmd = &cobra.Command{
	Use:   "retry [phase]",
	Short: "Retry a failed provisioning phase",
	Long: `Re-runs a specific provisioning phase that previously failed.
Use 'ubiquity status' to see which phases need attention.

Valid phases: metal, bootstrap, external, wait, post-install`,
	Args: cobra.ExactArgs(1),
	ValidArgs: []string{"metal", "bootstrap", "external", "wait", "post-install"},
	RunE: func(cmd *cobra.Command, args []string) error {
		phase := args[0]

		state, err := provision.LoadState()
		if err != nil {
			return fmt.Errorf("reading provisioning state: %w", err)
		}
		if state == nil {
			fmt.Println("No provisioning state found. Run 'ubiquity up' first.")
			return nil
		}

		fmt.Printf("Retrying phase: %s\n", phase)
		if err := state.StartPhase(phase); err != nil {
			return fmt.Errorf("starting phase %s: %w", phase, err)
		}

		if err := executePhase(phase, state.Environment); err != nil {
			state.FailPhase(phase, err)
			return fmt.Errorf("phase %s failed: %w", phase, err)
		}

		state.CompletePhase(phase)
		fmt.Printf("Phase %s completed successfully.\n", phase)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(retryCmd)
}