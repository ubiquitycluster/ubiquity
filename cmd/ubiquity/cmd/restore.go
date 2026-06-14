package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <backup-dir>",
	Short: "Restore cluster from backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Restoring from backup: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}