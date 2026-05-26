package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

type healthCheck struct {
	name  string
	check func() error
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check cluster health",
	Long:  `Runs health checks against the cluster: kubectl connectivity, core components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checks := []healthCheck{
			{"kubectl connectivity", func() error {
				return exec.Command("kubectl", "cluster-info").Run()
			}},
			{"ArgoCD server", func() error {
				return exec.Command("kubectl", "-n", "argocd", "get", "pod", "-l", "app.kubernetes.io/name=argocd-server").Run()
			}},
		}

		allPassed := true
		for _, c := range checks {
			fmt.Printf("  %s ... ", c.name)
			if err := c.check(); err != nil {
				fmt.Printf("FAIL (%v)\n", err)
				allPassed = false
			} else {
				fmt.Println("OK")
			}
		}

		if allPassed {
			fmt.Println("\nAll checks passed.")
		} else {
			fmt.Println("\nSome checks failed. Run 'ubiquity logs' for details.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
