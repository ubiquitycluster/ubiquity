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

var logsCmd = &cobra.Command{
	Use:   "logs [phase]",
	Short: "Tail structured provisioning logs",
	Long: `Displays logs for a specific provisioning phase or all phases.
Logs are sourced from structured JSON files in .ubiquity/state.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		phase := "all"
		if len(args) > 0 {
			phase = args[0]
		}
		fmt.Printf("Showing logs for phase: %s\n", phase)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}