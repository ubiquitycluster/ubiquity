package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup cluster state",
	Long:  `Creates a backup of cluster state, including provisioning state and configuration files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Creating cluster backup...")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}