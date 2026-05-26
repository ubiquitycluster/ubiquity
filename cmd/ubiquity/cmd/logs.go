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
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

var logsCmd = &cobra.Command{
	Use:   "logs [phase]",
	Short: "Tail structured provisioning logs",
	Long: `Displays logs for a specific provisioning phase or all phases.
Logs are sourced from structured JSON files in .ubiquity/state.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := provision.LoadState()
		if err != nil {
			return fmt.Errorf("reading provisioning state: %w", err)
		}
		if state == nil {
			fmt.Println("No provisioning state found. Run 'ubiquity up' first.")
			return nil
		}

		if len(args) == 0 {
			// Print all phases
			fmt.Printf("Provisioning Phases for environment: %s\n", state.Environment)
			fmt.Println(strings.Repeat("-", 60))
			fmt.Printf("%-18s %-12s %-12s %s\n", "Phase", "Status", "Duration", "Error")
			fmt.Println(strings.Repeat("-", 60))
			for _, p := range state.Phases {
				errMsg := p.Error
				if errMsg == "" {
					errMsg = "—"
				}
				dur := p.Duration
				if dur == "" {
					if p.Status == provision.PhasePending {
						dur = "—"
					} else if p.Status == provision.PhaseRunning {
						dur = "running…"
					} else {
						dur = "—"
					}
				}
				fmt.Printf("%-18s %-12s %-12s %s\n", p.Name, p.Status, dur, errMsg)
			}
		} else {
			// Print specific phase
			phaseName := args[0]
			found := false
			for _, p := range state.Phases {
				if p.Name == phaseName {
					found = true
					fmt.Printf("Phase: %s\n", p.Name)
					fmt.Printf("Status: %s\n", p.Status)
					if p.Duration != "" {
						fmt.Printf("Duration: %s\n", p.Duration)
					}
					if p.Error != "" {
						fmt.Printf("Error: %s\n", p.Error)
					}
					if p.StartedAt != nil {
						fmt.Printf("Started: %s\n", p.StartedAt.Format("2006-01-02 15:04:05 UTC"))
					}
					if p.EndedAt != nil {
						fmt.Printf("Ended: %s\n", p.EndedAt.Format("2006-01-02 15:04:05 UTC"))
					}
					if p.LogURL != "" {
						fmt.Printf("Log URL: %s\n", p.LogURL)
					}
					break
				}
			}
			if !found {
				fmt.Printf("Phase %q not found.\n", phaseName)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}