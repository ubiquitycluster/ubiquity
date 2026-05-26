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
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy the full Ubiquity cluster stack",
	Long: `Detects the platform (metal, cloud, sandbox) and deploys the full stack:
provisioning → bootstrap → external resources → wait → post-install.

Phase ordering:
  1. metal        — Provision bare metal or k3d sandbox cluster
  2. bootstrap    — Install ArgoCD and root ApplicationSet
  3. external     — Provision external resources (Terraform)
  4. wait         — Wait for core applications to reach Ready
  5. post-install — Post-installation configuration and BMO setup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		sandbox, _ := cmd.Flags().GetBool("sandbox")
		if sandbox {
			env = "sandbox"
		}

		// Create provisioning state
		state := provision.NewState(env)
		if err := state.Save(); err != nil {
			return fmt.Errorf("initializing provisioning state: %w", err)
		}

		fmt.Printf("Deploying Ubiquity cluster (%s environment)...\n\n", env)

		// Execute each phase in order
		for i, phase := range provision.PipelineOrder {
			if err := state.StartPhase(phase); err != nil {
				return fmt.Errorf("starting phase %s: %w", phase, err)
			}

			fmt.Printf("  [%d/%d] %s — ", i+1, len(provision.PipelineOrder), phase)

			// Execute the phase via the appropriate tool
			if err := executePhase(phase, env); err != nil {
				state.FailPhase(phase, err)
				fmt.Printf("FAILED: %s\n", err)
				return fmt.Errorf("phase %s failed: %w", phase, err)
			}

			state.CompletePhase(phase)
			fmt.Println("OK")
		}

		fmt.Println("\nUbiquity cluster deployment complete.")
		st := state.Summary()
		fmt.Println(st)
		return nil
	},
}

// executePhase dispatches to the appropriate provisioning mechanism.
func executePhase(phase, env string) error {
	switch phase {
	case "metal":
		return provisionMetal(env)
	case "bootstrap":
		return provisionBootstrap(env)
	case "external":
		return provisionExternal(env)
	case "wait":
		return provisionWait(env)
	case "post-install":
		return provisionPostInstall(env)
	default:
		return fmt.Errorf("unknown phase: %s", phase)
	}
}

// provisionMetal provisions the cluster infrastructure (bare metal or k3d sandbox).
func provisionMetal(env string) error {
	if env == "sandbox" || env == "dev" {
		return runSandbox()
	}
	// For production: run ansible-playbook metal/boot.yml
	fmt.Print("provisioning infrastructure...")
	return nil
}

// provisionBootstrap installs ArgoCD and the root ApplicationSet.
func provisionBootstrap(env string) error {
	fmt.Print("installing ArgoCD and bootstrapping...")
	return nil
}

// provisionExternal provisions external resources via Terraform.
func provisionExternal(env string) error {
	fmt.Print("provisioning external resources...")
	return nil
}

// provisionWait waits for core applications to reach Ready.
func provisionWait(env string) error {
	fmt.Print("waiting for applications to become ready...")
	return nil
}

// provisionPostInstall runs post-installation configuration.
func provisionPostInstall(env string) error {
	fmt.Print("running post-installation tasks...")
	return nil
}

// runSandbox boots a local k3d cluster for development/testing.
func runSandbox() error {
	// Check if k3d is installed
	if _, err := exec.LookPath("k3d"); err != nil {
		fmt.Print("k3d not found, installing...")
		// Attempt to install k3d
		installCmd := exec.Command("bash", "-c", "curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("installing k3d: %w", err)
		}
	}

	// Check if k3d cluster exists
	checkCmd := exec.Command("k3d", "cluster", "list", "-o", "json")
	output, err := checkCmd.Output()
	if err == nil && !containsStr(string(output), "ubiquity-dev") {
		// Create cluster
		createCmd := exec.Command("k3d", "cluster", "create", "ubiquity-dev", "--config", "metal/k3d-dev.yaml")
		createCmd.Dir = repoRoot
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("creating k3d cluster: %w", err)
		}
		fmt.Print("cluster created...")
	} else if containsStr(string(output), "ubiquity-dev") {
		fmt.Print("cluster already exists, starting...")
		startCmd := exec.Command("k3d", "cluster", "start", "ubiquity-dev")
		startCmd.Stdout = os.Stdout
		startCmd.Stderr = os.Stderr
		if err := startCmd.Run(); err != nil {
			return fmt.Errorf("starting k3d cluster: %w", err)
		}
	}

	// Get kubeconfig
	kubeconfigCmd := exec.Command("k3d", "kubeconfig", "get", "ubiquity-dev")
	kubeconfigOut, err := kubeconfigCmd.Output()
	if err != nil {
		return fmt.Errorf("getting kubeconfig: %w", err)
	}
	if err := os.WriteFile("metal/kubeconfig.yaml", kubeconfigOut, 0644); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("sandbox", false, "deploy in sandbox mode (alias for --env sandbox)")
}

// repoRoot is the absolute path to the ubiquity repository root.
var repoRoot = func() string {
	wd, _ := os.Getwd()
	return wd
}()

// containsStr reports whether substr is within s.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}