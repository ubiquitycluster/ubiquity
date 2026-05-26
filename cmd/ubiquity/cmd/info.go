package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show cluster information summary",
	Long:  `Displays CLI version, deployment environment, K8s version, and provisioning state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Ubiquity CLI: %s (commit: %s)\n", Version, Commit)
		fmt.Printf("Build date: %s\n", Date)

		state, err := provision.LoadState()
		if err == nil && state != nil {
			fmt.Printf("Environment: %s\n", state.Environment)
			fmt.Printf("Last updated: %s\n", state.UpdatedAt)
		}

		kubectlOut, err := exec.Command("kubectl", "version", "--short").Output()
		if err == nil {
			fmt.Printf("Kubernetes:\n%s", kubectlOut)
		} else {
			fmt.Println("Kubernetes: not connected")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
