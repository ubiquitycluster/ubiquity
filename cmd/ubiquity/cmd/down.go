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

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Tear down the Ubiquity cluster",
	Long:  `Destroys the cluster and associated resources. Use with caution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Tearing down Ubiquity cluster...")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}