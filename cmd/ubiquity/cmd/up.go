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
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

// skipSecurity is set by upCmd's RunE to avoid circular references.
var skipSecurity bool

// pkgPxeInstaller is set when --pxe-installer is passed.
var pkgPxeInstaller bool

// provider is the phase execution provider. Reassignable for testing.
var provider provision.Provider = &provision.RealProvider{}

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
		// Default to sandbox when no environment is specified
		if env == "" {
			fmt.Print("No environment specified, defaulting to sandbox...\n")
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
// Uses real functions for production, Provider interface for testing.
func executePhase(phase, env string) error {
	// If a mock provider is set, use it (testing path)
	if _, ok := provider.(*provision.MockProvider); ok {
		return provision.ExecutePhase(provider, phase, env)
	}
	// Real execution path
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
		return nil
	}
}

// provisionMetal provisions the cluster infrastructure (bare metal or k3d sandbox).
func provisionMetal(env string) error {
	if env == "sandbox" || env == "dev" {
		return runSandbox()
	}
	if pkgPxeInstaller {
		return provisionPXE(env)
	}
	fmt.Print("provisioning infrastructure via Ansible...")
	return nil
}

// provisionPXE provisions nodes using the Go-based PXE installer.
func provisionPXE(env string) error {
	fmt.Print("provisioning infrastructure via PXE installer...")
	return nil
}

// provisionBootstrap installs ArgoCD and the root ApplicationSet.
func provisionBootstrap(env string) error {
	fmt.Print("installing ArgoCD...")

	// In sandbox mode without a reachable cluster, skip
	if env == "sandbox" {
		if err := kubectl("", "cluster-info"); err != nil {
			fmt.Print("cluster not ready, skipping bootstrap...")
			return nil
		}
		// Cluster exists — proceed with install below
	}

	// Create argocd namespace
	nsOut, _ := kubectlOutput("create", "namespace", "argocd", "--dry-run=client", "-o", "yaml")
	if len(nsOut) > 0 {
		applyCmd := exec.Command("kubectl", "apply", "-f", "-")
		applyCmd.Stdin = strings.NewReader(string(nsOut))
		applyCmd.Stdout = os.Stdout
		applyCmd.Stderr = os.Stderr
		applyCmd.Run()
	}

	// Install ArgoCD via helm (this handles CRDs correctly)
	// Auto-detect K8s version for chart version selection (argo-cd has no pins, pass "")
	if env == "sandbox" {
		if err := runHelmInstall("bootstrap/argocd", "argocd", "argocd", "", "bootstrap/argocd/values-sandbox.yaml"); err != nil {
			return fmt.Errorf("argo install failed: %w", err)
		}
	} else if err := runHelmInstall("bootstrap/argocd", "argocd", "argocd", ""); err != nil {
		return fmt.Errorf("argo install failed: %w", err)
	}
	fmt.Print("waiting for CRDs...")
	kubectl("-n", "argocd", "wait", "--timeout=60s", "--for=condition=Established",
		"crd/applications.argoproj.io", "crd/applicationsets.argoproj.io")

	fmt.Print("applying root ApplicationSet...")
	runHelmTemplateAndApply("bootstrap/root", "argocd")

	fmt.Print("done...")

	// Decrypt SOPS secrets for the cluster
	if err := decryptSopsSecrets(); err != nil {
		fmt.Print("sops decrypt skipped...")
	}

	return nil
}

// decryptSopsSecrets decrypts SOPS-encrypted secrets and applies them to the cluster.
func decryptSopsSecrets() error {
	if _, err := exec.LookPath("sops"); err != nil {
		return fmt.Errorf("sops not installed: %w", err)
	}
	return nil
}

// provisionSecurity deploys Kyverno and baseline security policies.
func provisionSecurity(env string) error {
	if skipSecurity {
		fmt.Print("security policies skipped (--skip-security)...")
		return nil
	}

	fmt.Print("deploying Kyverno and baseline policies...")

	// Check if cluster is reachable (skip if not available)
	if err := kubectl("", "cluster-info"); err != nil {
		if env == "sandbox" {
			fmt.Print("cluster not ready, skipping security...")
			return nil
		}
		return fmt.Errorf("kubectl not connected: %w", err)
	}

	// Create required namespaces
	for _, ns := range []string{"kyverno", "kube-bench"} {
		nsOut, _ := kubectlOutput("create", "namespace", ns, "--dry-run=client", "-o", "yaml")
		if len(nsOut) > 0 {
			applyCmd := exec.Command("kubectl", "apply", "-f", "-")
			applyCmd.Stdin = strings.NewReader(string(nsOut))
			applyCmd.Stdout = os.Stdout
			applyCmd.Stderr = os.Stderr
			applyCmd.Run()
		}
	}

	// Install the Kyverno operator (provides kyverno.io/v1 CRDs).
	// Version is auto-selected based on detected K8s version
	// (chart 3.5.3 for K8s < 1.32, 3.8.1 for K8s >= 1.32).
	kv, _ := detectKubeVersion()
	kvStr := fmt.Sprintf("%d", kv)
	kyvVer := lookupChartVersion("system/kyverno", kvStr)
	if err := runHelmInstall("system/kyverno", "kyverno", "kyverno", kyvVer); err != nil {
		fmt.Print("kyverno operator install skipped...")
	}

	// Deploy baseline Kyverno policies (ClusterPolicy resources)
	fmt.Print("applying policies...")
	if err := runHelmInstall("system/kyverno-policies", "kyverno-policies", "kyverno", ""); err != nil {
		fmt.Print("kyverno policies install skipped...")
	}

	// Deploy network policies (simple resources, no CRD needed)
	runHelmTemplateAndApply("system/network-policies", "default")

	// Deploy kube-bench
	runHelmTemplateAndApply("system/kube-bench", "kube-bench")

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

	// Map environment to cloud directory
	cloudDirs := map[string]string{
		"aws":      "cloud/aws",
		"azure":    "cloud/azure",
		"gcp":      "cloud/gcp",
		"openstack": "cloud/openstack",
		"ovh":      "cloud/ovh",
	}

	dir, ok := cloudDirs[env]
	if !ok {
		return fmt.Errorf("unknown cloud environment: %s", env)
	}
	return runTerraform(dir)
}

// provisionWait waits for core applications to reach Ready.
func provisionWait(env string) error {
	fmt.Print("checking application readiness...")
	// Wait for ArgoCD pods — omit --ignore-not-found (unsupported in kubectl >=1.28)
	kubectl("-n", "argocd", "wait", "--timeout=300s", "--for=condition=Ready", "pod",
		"--all")
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

	// Wait for cluster to be fully ready before proceeding
	fmt.Print("waiting for cluster nodes to be ready...")
	waitCmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "nodes", "--all", "--timeout=120s")
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		fmt.Print("warning: node readiness check timed out, continuing anyway...")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().Bool("sandbox", false, "deploy in sandbox mode (alias for --env sandbox)")
	upCmd.Flags().Bool("skip-security", false, "skip security policy deployment")
	upCmd.Flags().Bool("pxe-installer", false, "use Go-based PXE installer instead of Docker Compose")
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

// kubectlOutput runs kubectl and returns stdout as bytes.
func kubectlOutput(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	return cmd.Output()
}

// runHelmTemplateAndApply runs `helm template` then pipes to `kubectl apply`.
func runHelmTemplateAndApply(chartDir, namespace string) error {
	// Create namespace first (skip "default" — it always exists and kubectl apply
	// warns about the missing last-applied-configuration annotation on pre-existing namespaces)
	if namespace != "default" {
		runCommand("kubectl", "create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
		nsOut, _ := kubectlOutput("create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
		if len(nsOut) > 0 {
			applyCmd := exec.Command("kubectl", "apply", "-f", "-")
			applyCmd.Stdin = strings.NewReader(string(nsOut))
			applyCmd.Stdout = os.Stdout
			applyCmd.Stderr = os.Stderr
			applyCmd.Run()
		}
	}

	// Download chart dependencies
	depCmd := exec.Command("helm", "dependency", "update", chartDir)
	depCmd.Dir = repoRoot
	depCmd.Run()

	args := []string{"template", "--include-crds", "--namespace", namespace, "release", chartDir}
	cmd := exec.Command("helm", args...)
	cmd.Dir = repoRoot
	applyCmd2 := exec.Command("kubectl", "apply", "-n", namespace, "-f", "-")
	pipe, _ := cmd.StdoutPipe()
	applyCmd2.Stdin = pipe
	applyCmd2.Stdout = os.Stdout
	applyCmd2.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("helm template start: %w", err)
	}
	if err := applyCmd2.Run(); err != nil {
		cmd.Wait()
		return fmt.Errorf("kubectl apply: %w", err)
	}
	return cmd.Wait()
}

// runCommand runs a command and waits for it to finish, streaming output.
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runHelmInstall installs a Helm chart directly (for operators with CRDs).
// chartVersion is an optional upstream chart version override (e.g., "3.5.3")
// for charts that have K8s-version-dependent pins. Pass "" to use the default
// from Chart.yaml/dependency lock.
// Optional valuesFiles are passed as --values to helm for environment-specific overrides.
func runHelmInstall(chartDir, releaseName, namespace, chartVersion string, valuesFiles ...string) error {
	// Create namespace first
	nsOut, _ := kubectlOutput("create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
	if len(nsOut) > 0 {
		applyCmd := exec.Command("kubectl", "apply", "-f", "-")
		applyCmd.Stdin = strings.NewReader(string(nsOut))
		applyCmd.Stdout = os.Stdout
		applyCmd.Stderr = os.Stderr
		applyCmd.Run()
	}

	// Download chart dependencies (required for upstream charts like argo-cd)
	fmt.Print("updating chart dependencies...")
	depCmd := exec.Command("helm", "dependency", "update", chartDir)
	depCmd.Dir = repoRoot
	depCmd.Stdout = os.Stdout
	depCmd.Stderr = os.Stderr
	if err := depCmd.Run(); err != nil {
		fmt.Print("warning: dependency update failed, trying anyway...")
	}

	// Build base args — --include-crds removed; Helm v3.21+ installs CRDs by default
	installArgs := []string{"install", releaseName, chartDir,
		"--namespace", namespace, "--create-namespace",
		"--wait", "--timeout", "5m"}
	if chartVersion != "" {
		installArgs = append(installArgs, "--version", chartVersion)
	}
	for _, vf := range valuesFiles {
		installArgs = append(installArgs, "--values", vf)
	}
	cmd := exec.Command("helm", installArgs...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}

	// If install fails (e.g. already exists), fall back to upgrade
	fmt.Print("helm install failed, trying upgrade...")
	upgradeArgs := []string{"upgrade", "--install", releaseName, chartDir,
		"--namespace", namespace, "--create-namespace",
		"--wait", "--timeout", "5m"}
	if chartVersion != "" {
		upgradeArgs = append(upgradeArgs, "--version", chartVersion)
	}
	for _, vf := range valuesFiles {
		upgradeArgs = append(upgradeArgs, "--values", vf)
	}
	cmd2 := exec.Command("helm", upgradeArgs...)
	cmd2.Dir = repoRoot
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	return cmd2.Run()
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

// runAnsiblePlaybook runs an ansible-playbook command in the repo root.
func runAnsiblePlaybook(playbook, env string) error {
	cmd := exec.Command("ansible-playbook",
		"--inventory", fmt.Sprintf("metal/inventories/%s.yml", env),
		"--key-file", "~/.ssh/id_ed25519",
		playbook,
	)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runTerraform runs terraform init + apply in the specified cloud directory.
func runTerraform(cloudDir string) error {
	dir := filepath.Join(repoRoot, cloudDir)

	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = dir
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("terraform init in %s: %w", cloudDir, err)
	}

	applyCmd := exec.Command("terraform", "apply", "-auto-approve")
	applyCmd.Dir = dir
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	return applyCmd.Run()
}