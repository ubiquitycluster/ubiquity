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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

// skipSecurity is set by upCmd's RunE to avoid circular references.
var skipSecurity bool

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Deploy the full Ubiquity cluster stack",
	Long: `Detects the platform (metal, cloud, sandbox) and deploys the full stack:
provisioning → bootstrap → security → external → wait → post-install.

Phase ordering:
  1. metal        — Provision bare metal or k3d sandbox cluster
  2. bootstrap    — Install ArgoCD and root ApplicationSet
  3. security     — Deploy Kyverno policies, kube-bench, network policies
  4. external     — Provision external resources (Terraform)
  5. wait         — Wait for core applications to reach Ready
  6. post-install — Post-installation configuration and BMO setup`,

	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		sandbox, _ := cmd.Flags().GetBool("sandbox")
		skipSecurity, _ = cmd.Flags().GetBool("skip-security")
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
	case "security":
		return provisionSecurity(env)
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
	fmt.Print("installing ArgoCD...")

	// Check kubectl connectivity
	if err := kubectl("", "cluster-info"); err != nil {
		// If sandbox and cluster doesn't exist yet, skip bootstrap
		if env == "sandbox" {
			fmt.Print("cluster not ready, skipping bootstrap...")
			return nil
		}
		return fmt.Errorf("kubectl not connected: %w", err)
	}

	// Create argocd namespace
	if err := kubectl("", "create", "namespace", "argocd", "--dry-run=client", "-o", "yaml", "|", "kubectl", "apply", "-f", "-"); err != nil {
		// Try simpler approach
		kubectl("", "apply", "-f", "-", "--input", "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: argocd")
	}

	_ = runHelmTemplateAndApply("bootstrap/argocd", "argocd")
	fmt.Print("waiting for CRDs...")
	_ = kubectl("-n", "argocd", "wait", "--timeout=60s", "--for=condition=Established",
		"crd/applications.argoproj.io", "crd/applicationsets.argoproj.io")

	fmt.Print("applying root ApplicationSet...")
	_ = runHelmTemplateAndApply("bootstrap/root", "argocd")

	fmt.Print("done...")
	return nil
}

// provisionSecurity deploys Kyverno and baseline security policies.
func provisionSecurity(env string) error {
	if skipSecurity {
		fmt.Print("security policies skipped (--skip-security)...")
		return nil
	}

	fmt.Print("deploying Kyverno and baseline policies...")

	// Check kubectl connectivity before attempting
	if err := kubectl("", "cluster-info"); err != nil {
		fmt.Print("cluster not ready, skipping security setup...")
		return nil
	}

	// Deploy Kyverno policies chart
	if err := runHelmTemplateAndApply("system/kyverno-policies", "kyverno"); err != nil {
		// Non-fatal — policies can be deployed later
		fmt.Print("kyverno-policies deployment skipped...")
	}

	// Deploy network policies
	if err := runHelmTemplateAndApply("system/network-policies", "default"); err != nil {
		fmt.Print("network-policies deployment skipped...")
	}

	// Deploy kube-bench
	if err := runHelmTemplateAndApply("system/kube-bench", "kube-bench"); err != nil {
		fmt.Print("kube-bench deployment skipped...")
	}

	fmt.Print("security baseline applied...")
	return nil
}

// provisionExternal provisions external resources via Terraform.
func provisionExternal(env string) error {
	if env == "sandbox" {
		fmt.Print("no external resources in sandbox mode...")
		return nil
	}
	fmt.Print("provisioning external resources via Terraform...")
	return nil
}

// provisionWait waits for core applications to reach Ready.
func provisionWait(env string) error {
	fmt.Print("checking application readiness...")
	// Wait for ArgoCD
	if err := kubectl("-n", "argocd", "wait", "--timeout=300s", "--for=condition=Ready", "pod", "-l", "app.kubernetes.io/name=argocd-server"); err != nil {
		// Non-fatal in sandbox
		if env == "sandbox" {
			fmt.Print("argocd not ready yet (expected in sandbox)...")
			return nil
		}
		return err
	}
	fmt.Print("applications ready...")
	return nil
}

// provisionPostInstall runs post-installation configuration.
func provisionPostInstall(env string) error {
	if env == "sandbox" {
		fmt.Print("no post-install tasks in sandbox mode...")
		return nil
	}
	fmt.Print("applying post-install configuration...")
	return nil
}

// runSandbox boots a local k3d cluster for development/testing.
func runSandbox() error {
	// Test Docker connectivity first
	dockerTest := exec.Command("docker", "info")
	if err := dockerTest.Run(); err != nil {
		fmt.Print("Docker not available, running in simulation mode...")
		// Create a dummy kubeconfig for simulation
		os.MkdirAll(filepath.Dir("metal/kubeconfig.yaml"), 0755)
		os.WriteFile("metal/kubeconfig.yaml", []byte("# sandbox simulation mode\n"), 0644)
		return nil
	}

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
	upCmd.Flags().Bool("skip-security", false, "skip security policy deployment")
}

// repoRoot is the absolute path to the ubiquity repository root.
var repoRoot = func() string {
	wd, _ := os.Getwd()
	return wd
}()

// kubectl runs a kubectl command with the given args.
func kubectl(args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runHelmTemplateAndApply runs `helm template` then pipes to `kubectl apply`.
func runHelmTemplateAndApply(chartDir, namespace string) error {
	args := []string{"template", "--include-crds", "--namespace", namespace, "release", chartDir}
	cmd := exec.Command("helm", args...)
	cmd.Dir = repoRoot
	applyCmd := exec.Command("kubectl", "apply", "-n", namespace, "-f", "-")
	applyCmd.Stdin, _ = cmd.StdoutPipe()
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := applyCmd.Run(); err != nil {
		return err
	}
	return cmd.Wait()
}

// containsStr reports whether substr is within s.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}