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
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time injected via ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the Ubiquity CLI version, git commit, build date, and Go runtime info.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			info := map[string]string{
				"version":    Version,
				"commit":     Commit,
				"date":       Date,
				"go_version": runtime.Version(),
				"os":         runtime.GOOS,
				"arch":       runtime.GOARCH,
			}
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling version: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Ubiquity CLI %s\n", Version)
		fmt.Printf("  Commit: %s\n", Commit)
		fmt.Printf("  Built:  %s\n", Date)
		fmt.Printf("  Go:     %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().Bool("json", false, "output version as JSON")
}