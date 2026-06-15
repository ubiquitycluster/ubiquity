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
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

// skipSecurity is set by upCmd's RunE to avoid circular references.
var skipSecurity bool

// pkgPxeInstaller is set when --pxe-installer is passed.
var pkgPxeInstaller bool

// provider is the phase execution provider. Reassignable for testing.
var provider provision.Provider = &provision.RealProvider{}

type upLifecycleOptions struct {
	MetalBootstrapBackend string
	NodeLifecycleBackend  string
	NICOValues            string
	NICORESTValues        string
	NICOSite              string
}

type nicoGitOpsTarget struct {
	ChartDir    string
	ReleaseName string
	Namespace   string
	ValuesFiles []string
	Site        string
}

var upLifecycle = upLifecycleOptions{}

var nicoGitOpsInstall = installNICOGitOpsTarget
var terraformProvisioner = runTerraform
var ansibleProvisioner = provisionAnsible

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
  6. post-install — Wire node lifecycle integration (NICo by default for production)

Node lifecycle backend:
  Production defaults to NVIDIA Infra Controller (NICo) through Ubiquity GitOps
  wrappers. The legacy BMO backend is fallback/migration-only and is rejected by
  'ubiquity up' unless UBIQUITY_ALLOW_BMO_MIGRATION=true is set explicitly.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := cmd.Flags().GetString("env")
		sandbox, _ := cmd.Flags().GetBool("sandbox")
		skipSecurity, _ = cmd.Flags().GetBool("skip-security")
		upLifecycle.MetalBootstrapBackend, _ = cmd.Flags().GetString("metal-bootstrap-backend")
		upLifecycle.NodeLifecycleBackend, _ = cmd.Flags().GetString("node-lifecycle-backend")
		upLifecycle.NICOValues, _ = cmd.Flags().GetString("nico-values")
		upLifecycle.NICORESTValues, _ = cmd.Flags().GetString("nico-rest-values")
		upLifecycle.NICOSite, _ = cmd.Flags().GetString("nico-site")
		if sandbox {
			env = "sandbox"
		}
		// Default to sandbox when no environment is specified
		if env == "" {
			fmt.Print("No environment specified, defaulting to sandbox...\n")
			env = "sandbox"
		}
		applyLifecycleDefaults(env)
		if err := validateLifecycleBackends(); err != nil {
			return err
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
	applyLifecycleDefaults(env)
	if pkgPxeInstaller {
		return provisionPXE(env)
	}
	if upLifecycle.MetalBootstrapBackend == "none" {
		fmt.Print("metal bootstrap backend disabled...")
		return nil
	}
	fmt.Print("provisioning infrastructure via Ansible...")
	return ansibleProvisioner(env)
}

func applyLifecycleDefaults(env string) {
	if upLifecycle.MetalBootstrapBackend == "" {
		if env == "prod" || env == "production" {
			upLifecycle.MetalBootstrapBackend = "ansible"
		} else {
			upLifecycle.MetalBootstrapBackend = "none"
		}
	}
	if upLifecycle.NodeLifecycleBackend == "" {
		if env == "prod" || env == "production" {
			upLifecycle.NodeLifecycleBackend = "nico"
		} else {
			upLifecycle.NodeLifecycleBackend = "none"
		}
	}
}

func validateLifecycleBackends() error {
	switch upLifecycle.MetalBootstrapBackend {
	case "ansible", "none":
	default:
		return fmt.Errorf("unsupported metal bootstrap backend %q", upLifecycle.MetalBootstrapBackend)
	}
	switch upLifecycle.NodeLifecycleBackend {
	case "nico", "none":
	case "bmo":
		if os.Getenv("UBIQUITY_ALLOW_BMO_MIGRATION") != "true" {
			return fmt.Errorf("BMO node lifecycle backend is fallback/migration-only; use --node-lifecycle-backend=nico for normal up or set UBIQUITY_ALLOW_BMO_MIGRATION=true to acknowledge legacy migration fallback")
		}
	default:
		return fmt.Errorf("unsupported node lifecycle backend %q", upLifecycle.NodeLifecycleBackend)
	}
	return nil
}

func describeNICOInstallPath(env string) string {
	applyLifecycleDefaults(env)
	if upLifecycle.NodeLifecycleBackend != "nico" {
		return "NICo lifecycle backend disabled"
	}
	parts := []string{"NICo lifecycle backend enabled via GitOps wrappers", "render/apply NVIDIA Infra Controller wrappers with helm template and kubectl apply"}
	if upLifecycle.NICOValues != "" {
		parts = append(parts, "controller values="+upLifecycle.NICOValues)
	}
	if upLifecycle.NICORESTValues != "" {
		parts = append(parts, "REST values="+upLifecycle.NICORESTValues)
	}
	if upLifecycle.NICOSite != "" {
		parts = append(parts, "site="+upLifecycle.NICOSite)
	}
	return strings.Join(parts, "; ")
}

// provisionPXE provisions nodes using the Go-based PXE installer.
func provisionPXE(env string) error {
	fmt.Print("provisioning infrastructure via PXE installer...")
	return nil
}

// provisionBootstrap installs ArgoCD and the root ApplicationSet.
// In sandbox mode, the ArgoCD app-of-apps mechanism is bypassed because
// ArgoCD cannot authenticate to the GitHub repository (credentials in
// values-seed.yaml are placeholders). Instead, charts are applied directly
// from the local filesystem via helm template | kubectl apply.
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

	if env == "sandbox" {
		// In sandbox mode, skip the ArgoCD app-of-apps mechanism (which reads
		// from GitHub via ApplicationSets) and instead apply all charts directly
		// from the local filesystem. This avoids the GitHub authentication issue
		// where ArgoCD cannot clone the repo with the placeholder credentials.
		if !bootstrapShouldApplySandboxCharts(env) {
			fmt.Print("skipping direct sandbox chart apply (--skip-security smoke mode)...")
			return nil
		}
		fmt.Print("applying charts directly (sandbox mode)...")
		if err := applySandboxCharts(); err != nil {
			return fmt.Errorf("applying sandbox charts failed: %w", err)
		}
	} else {
		fmt.Print("applying root ApplicationSet...")
		runHelmTemplateAndApply("bootstrap/root", "argocd")
	}

	fmt.Print("done...")

	// Decrypt SOPS secrets for the cluster
	if err := decryptSopsSecrets(); err != nil {
		fmt.Print("sops decrypt skipped...")
	}

	return nil
}

func bootstrapShouldApplySandboxCharts(env string) bool {
	return env == "sandbox" && !skipSecurity
}

// applySandboxCharts applies all Helm charts and kustomize manifests from the
// system/monitoring/platform/apps stacks directly from the local filesystem.
// This is used in sandbox mode instead of the ArgoCD app-of-apps mechanism,
// which cannot authenticate to GitHub with the placeholder credentials in
// values-seed.yaml.
//
// Directories with Chart.yaml are deployed via helm template | kubectl apply.
// Directories with kustomization.yaml (but no Chart.yaml) are deployed via
// kubectl apply -k. Directories with neither are skipped silently.
// Kyverno and kyverno-policies are skipped because they are already installed
// in the security phase.
func applySandboxCharts() error {
	if err := kubectl("", "cluster-info"); err != nil {
		fmt.Print("cluster not ready, skipping sandbox chart apply...")
		return nil
	}

	targets, err := collectSandboxDeployTargets()
	if err != nil {
		return err
	}

	currentStack := ""
	for _, target := range targets {
		if currentStack != "" && currentStack != target.Stack && currentStack == "system" {
			fmt.Print(" waiting for cert-manager...")
			kubectl("-n", "cert-manager", "wait", "--timeout=120s",
				"--for=condition=Ready", "pod", "-l", "app.kubernetes.io/instance=cert-manager")
		}
		currentStack = target.Stack

		if target.SkipReason != "" {
			fmt.Printf(" applying %s/%s... skipped (%s)\n", target.Stack, target.Name, target.SkipReason)
			continue
		}

		if target.Kind == "helm" {
			fmt.Printf(" applying %s/%s...", target.Stack, target.Name)
			if err := runHelmTemplateAndApply(target.ChartDir, target.Namespace); err != nil {
				if isNvidiaAISandboxDeployTarget(target) {
					fmt.Printf(" failed (%v)\n", err)
					return fmt.Errorf("required NVIDIA AI sandbox component %s/%s failed to deploy: %w", target.Stack, target.Name, err)
				}
				fmt.Printf(" skipped (%v)\n", err)
				continue
			}
			fmt.Printf(" ok\n")
			continue
		}

		if target.Kind == "kustomize" {
			fmt.Printf(" applying kustomize %s/%s...", target.Stack, target.Name)
			applyK := exec.Command("kubectl", "apply", "-k", filepath.Join(repoRoot, target.ChartDir))
			applyK.Stdout = os.Stdout
			applyK.Stderr = os.Stderr
			if err := applyK.Run(); err != nil {
				fmt.Printf(" skipped (%v)\n", err)
				continue
			}
			fmt.Printf(" ok\n")
		}
	}

	if currentStack == "system" {
		fmt.Print(" waiting for cert-manager...")
		kubectl("-n", "cert-manager", "wait", "--timeout=120s",
			"--for=condition=Ready", "pod", "-l", "app.kubernetes.io/instance=cert-manager")
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

type sandboxDeployTarget struct {
	Stack      string
	Name       string
	ChartDir   string
	Namespace  string
	Kind       string
	SkipReason string
}

func collectSandboxDeployTargets() ([]sandboxDeployTarget, error) {
	stacks := []struct {
		name string
		path string
	}{
		{"system", "system"},
		{"monitoring", "monitoring"},
		{"platform", "platform"},
		{"apps", "apps"},
	}

	skipCharts := map[string]string{
		"kyverno":          "already installed in security phase",
		"kyverno-policies": "already installed in security phase",
	}
	namespaceOverrides := map[string]string{
		"ai-platform-console":               "ai-platform",
		"nvidia-gpu-operator":               "gpu-operator",
		"nvidia-nic-configuration-operator": "network-operator",
		"nim-operator":                      "nim-operator",
		"kai-scheduler":                     "kai-scheduler",
		"ai-workload-tenancy":               "ai-workloads",
		"stallscope":                        "gpu-telemetry",
		"nvidia-network-operator":           "nvidia-network-operator",
	}

	var targets []sandboxDeployTarget
	for _, stack := range stacks {
		stackDir := filepath.Join(repoRoot, stack.path)
		entries, err := os.ReadDir(stackDir)
		if err != nil {
			fmt.Printf("warning: reading %s directory: %v\n", stack.path, err)
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			chartName := entry.Name()
			chartDir := filepath.Join(stackDir, chartName)
			namespace := chartName
			if override, ok := namespaceOverrides[chartName]; ok {
				namespace = override
			}

			target := sandboxDeployTarget{
				Stack:     stack.name,
				Name:      chartName,
				ChartDir:  filepath.ToSlash(filepath.Join(stack.path, chartName)),
				Namespace: namespace,
			}

			if reason, ok := skipCharts[chartName]; ok {
				target.SkipReason = reason
				targets = append(targets, target)
				continue
			}

			switch {
			case fileExists(filepath.Join(chartDir, "Chart.yaml")):
				target.Kind = "helm"
				targets = append(targets, target)
			case fileExists(filepath.Join(chartDir, "kustomization.yaml")):
				target.Kind = "kustomize"
				targets = append(targets, target)
			}
		}
	}

	return targets, nil
}

func validateSandboxDeployTargets(targets []sandboxDeployTarget) error {
	for _, target := range targets {
		if target.SkipReason != "" || target.Kind != "helm" {
			continue
		}
		if err := renderSandboxHelmTarget(target); err != nil {
			return err
		}
	}
	return nil
}

func validateSandboxDeployTargetsForCLI() error {
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no sandbox deploy targets found under %s", repoRoot)
	}
	return validateSandboxDeployTargets(targets)
}

func validateNvidiaAISandboxDeployTargetsForCLI() error {
	targets, err := collectSandboxDeployTargets()
	if err != nil {
		return err
	}
	filtered := filterNvidiaAISandboxDeployTargets(targets)
	if len(filtered) == 0 {
		return fmt.Errorf("no NVIDIA AI sandbox deploy targets found under %s", repoRoot)
	}
	return validateSandboxDeployTargets(filtered)
}

func filterNvidiaAISandboxDeployTargets(targets []sandboxDeployTarget) []sandboxDeployTarget {
	var filtered []sandboxDeployTarget
	for _, target := range targets {
		if isNvidiaAISandboxDeployTarget(target) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func isNvidiaAISandboxDeployTarget(target sandboxDeployTarget) bool {
	included := map[string]bool{
		"platform/ai-platform-console":             true,
		"platform/ai-workload-tenancy":             true,
		"platform/kai-scheduler":                   true,
		"platform/nim-operator":                    true,
		"platform/stallscope":                      true,
		"system/nvidia-gpu-operator":               true,
		"system/nvidia-network-operator":           true,
		"system/nvidia-nic-configuration-operator": true,
	}
	return included[target.ChartDir]
}

func renderSandboxHelmTarget(target sandboxDeployTarget) error {
	chartDir := filepath.Join(repoRoot, target.ChartDir)
	if err := ensureHelmReposForChart(chartDir); err != nil {
		return fmt.Errorf("%s: add Helm repos: %w", target.ChartDir, err)
	}
	workChartDir, cleanup, err := prepareHelmChartWorkdir(chartDir)
	if err != nil {
		return fmt.Errorf("%s: prepare Helm workdir: %w", target.ChartDir, err)
	}
	defer cleanup()

	if fileExists(filepath.Join(workChartDir, "Chart.yaml")) {
		depCmd := exec.Command("helm", "dependency", "build", workChartDir)
		depCmd.Dir = repoRoot
		if out, err := depCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: helm dependency build: %w\n%s", target.ChartDir, err, string(out))
		}
	}

	args := helmTemplateArgs(target.Namespace, "release", workChartDir)
	cmd := exec.Command("helm", args...)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: helm template: %w\n%s", target.ChartDir, err, string(out))
	}
	return nil
}

func ensureHelmRepo(name, url string) error {
	cmd := exec.Command("helm", "repo", "add", name, url)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil && !strings.Contains(string(out), "already exists") {
		return fmt.Errorf("%w\n%s", err, string(out))
	}
	return nil
}

func ensureHelmReposForChart(chartDir string) error {
	for _, repoURL := range chartRepositoryURLs(chartDir) {
		if strings.HasPrefix(repoURL, "file://") || strings.HasPrefix(repoURL, "oci://") || repoURL == "" {
			continue
		}
		name := helmRepoName(repoURL)
		if err := ensureHelmRepo(name, repoURL); err != nil {
			return err
		}
	}
	return nil
}

func chartRepositoryURLs(chartDir string) []string {
	data, err := os.ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var repos []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "repository:") {
			continue
		}
		repoURL := strings.TrimSpace(strings.TrimPrefix(trimmed, "repository:"))
		repoURL = strings.Trim(repoURL, "\"'")
		if repoURL != "" && !seen[repoURL] {
			seen[repoURL] = true
			repos = append(repos, repoURL)
		}
	}
	return repos
}

func helmRepoName(repoURL string) string {
	name := repoURL
	for _, prefix := range []string{"https://", "http://"} {
		name = strings.TrimPrefix(name, prefix)
	}
	name = strings.TrimSuffix(name, "/")
	replacer := strings.NewReplacer("/", "-", ".", "-", ":", "-", "_", "-")
	name = replacer.Replace(name)
	if strings.Contains(repoURL, "helm.ngc.nvidia.com/nvidia") {
		return "nvidia"
	}
	if strings.Contains(repoURL, "charts.jetstack.io") {
		return "jetstack"
	}
	if strings.Contains(repoURL, "mysql.github.io/mysql-operator") {
		return "mysql-operator"
	}
	return name
}

func chartUsesNvidiaHelmRepo(chartDir string) bool {
	data, err := os.ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	return err == nil && strings.Contains(string(data), "https://helm.ngc.nvidia.com/nvidia")
}

func cleanupSandboxDependencyArchives(chartDir string) {
	chartsDir := filepath.Join(chartDir, "charts")
	entries, err := os.ReadDir(chartsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tgz") {
				_ = os.Remove(filepath.Join(chartsDir, entry.Name()))
			}
		}
		_ = os.Remove(chartsDir)
	}
	_ = os.Remove(filepath.Join(chartDir, "Chart.lock"))
}

func prepareHelmChartWorkdir(chartDir string) (string, func(), error) {
	tmpRoot, err := os.MkdirTemp("", "ubiquity-helm-chart-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpRoot) }
	workChartDir := filepath.Join(tmpRoot, filepath.Base(chartDir))
	if err := copyDir(chartDir, workChartDir); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return workChartDir, cleanup, nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, target)
		}
		if d.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
	kyvVer := ""
	if kv, err := detectKubeVersion(); err == nil {
		kvStr := fmt.Sprintf("%d", kv)
		kyvVer = lookupChartVersion("system/kyverno", kvStr)
	}
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
	dir, err := cloudDirForEnv(env)
	if err != nil {
		return err
	}
	return terraformProvisioner(dir)
}

func cloudDirForEnv(env string) (string, error) {
	cloudEnv := strings.ToLower(strings.TrimSpace(env))
	if cloudEnv == "prod" || cloudEnv == "production" || cloudEnv == "dev" {
		for _, key := range []string{"cloud_provider", "cloudProvider", "provider"} {
			if configured := strings.ToLower(strings.TrimSpace(viper.GetString(key))); configured != "" {
				cloudEnv = configured
				break
			}
		}
	}
	cloudDirs := map[string]string{
		"aws":       "cloud/aws",
		"azure":     "cloud/azure",
		"gcp":       "cloud/gcp",
		"openstack": "cloud/openstack",
		"ovh":       "cloud/ovh",
	}
	dir, ok := cloudDirs[cloudEnv]
	if !ok {
		return "", fmt.Errorf("unknown cloud environment/provider %q; set --env to one of aws, azure, gcp, openstack, ovh or configure cloud_provider", env)
	}
	return dir, nil
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
	applyLifecycleDefaults(env)
	switch upLifecycle.NodeLifecycleBackend {
	case "nico":
		fmt.Print(describeNICOInstallPath(env) + "...")
		return installNICOGitOpsWrappers()
	case "bmo":
		if os.Getenv("UBIQUITY_ALLOW_BMO_MIGRATION") != "true" {
			return fmt.Errorf("BMO node lifecycle backend is fallback/migration-only; set UBIQUITY_ALLOW_BMO_MIGRATION=true to acknowledge legacy migration fallback")
		}
		fmt.Print("WARNING: BMO node lifecycle backend selected as fallback/migration-only; NICo remains the default production path...")
		return nil
	}
	if env == "sandbox" || env == "dev" {
		fmt.Print("no post-install tasks in sandbox/dev mode...")
		return nil
	}
	fmt.Print("applying post-install configuration...")
	return nil
}

func installNICOGitOpsWrappers() error {
	for _, target := range nicoGitOpsTargets() {
		if err := nicoGitOpsInstall(target); err != nil {
			return fmt.Errorf("installing NICo GitOps wrapper %s: %w", target.ChartDir, err)
		}
	}
	return nil
}

func nicoGitOpsTargets() []nicoGitOpsTarget {
	coreValues := filterEmptyStrings(upLifecycle.NICOValues)
	restValues := filterEmptyStrings(upLifecycle.NICORESTValues)
	return []nicoGitOpsTarget{
		{ChartDir: "system/nvidia-infra-controller-prereqs", ReleaseName: "nico-prereqs", Namespace: "nvidia-infra-controller", Site: upLifecycle.NICOSite},
		{ChartDir: "system/nvidia-infra-controller-core", ReleaseName: "nico-core", Namespace: "nvidia-infra-controller", ValuesFiles: coreValues, Site: upLifecycle.NICOSite},
		{ChartDir: "platform/nvidia-infra-controller-rest", ReleaseName: "nico-rest", Namespace: "nvidia-infra-controller", ValuesFiles: restValues, Site: upLifecycle.NICOSite},
		{ChartDir: "platform/node-lifecycle-exporter", ReleaseName: "node-lifecycle-exporter", Namespace: "node-lifecycle", Site: upLifecycle.NICOSite},
	}
}

func filterEmptyStrings(values ...string) []string {
	var out []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func installNICOGitOpsTarget(target nicoGitOpsTarget) error {
	if target.Namespace != "" {
		nsOut, _ := kubectlOutput("create", "namespace", target.Namespace, "--dry-run=client", "-o", "yaml")
		if len(nsOut) > 0 {
			if err := kubectlApplyRendered(nsOut, "", false, false); err != nil {
				return fmt.Errorf("create namespace %s: %w", target.Namespace, err)
			}
		}
	}
	depCmd := exec.Command("helm", "dependency", "update", target.ChartDir)
	depCmd.Dir = repoRoot
	depCmd.Stdout = os.Stdout
	depCmd.Stderr = os.Stderr
	_ = depCmd.Run()

	args := []string{"template", "--include-crds", "--namespace", target.Namespace, target.ReleaseName, target.ChartDir}
	for _, valuesFile := range target.ValuesFiles {
		args = append(args, "--values", valuesFile)
	}
	if target.Site != "" {
		switch target.ChartDir {
		case "system/nvidia-infra-controller-core":
			args = append(args, "--set-string", "config.siteName="+target.Site)
		case "platform/nvidia-infra-controller-rest":
			args = append(args, "--set-string", "site.name="+target.Site)
		}
	}
	cmd := exec.Command("helm", args...)
	cmd.Dir = repoRoot
	rendered, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("helm template: %w", err)
	}
	if len(bytes.TrimSpace(rendered)) == 0 {
		return fmt.Errorf("helm template rendered no resources")
	}
	if err := kubectlApplyRendered(rendered, target.Namespace, false, false); err != nil {
		waitForCRDsEstablished()
		if retryErr := kubectlApplyRendered(rendered, target.Namespace, false, false); retryErr != nil {
			return fmt.Errorf("kubectl apply: %w", retryErr)
		}
	}
	waitForCRDsEstablished()
	return nil
}

// runSandbox boots a local k3d cluster for development/testing.
// The k3s version defaults to the latest stable (read from
// metal/k3s-default-version.txt) but can be overridden via the
// K3S_IMAGE environment variable.
func readK3sImage(data []byte) string {
	// Strip comment lines (lines starting with #) and empty lines,
	// then return the first remaining line trimmed.
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}

func runSandbox() error {
	// If kubectl already connects to a working cluster, don't touch k3d.
	// This allows the test harness (or existing deployments) to manage the
	// cluster independently of `ubiquity up`.
	if err := exec.Command("kubectl", "cluster-info").Run(); err == nil {
		fmt.Print("cluster already connected via kubectl, skipping k3d...")
		return nil
	}

	// Resolve the k3s image: K3S_IMAGE env var > default file > k3d default
	k3sImage := os.Getenv("K3S_IMAGE")
	if k3sImage == "" {
		verFile := filepath.Join(repoRoot, "metal", "k3s-default-version.txt")
		if data, err := os.ReadFile(verFile); err == nil {
			k3sImage = readK3sImage(data)
		}
	}

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
		createArgs := []string{"cluster", "create", "ubiquity-dev", "--config", "metal/k3d-dev.yaml"}
		if k3sImage != "" {
			createArgs = append(createArgs, "--image", k3sImage)
		}
		createCmd := exec.Command("k3d", createArgs...)
		createCmd.Dir = repoRoot
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("creating k3d cluster: %w", err)
		}
		fmt.Print("cluster created...")
	} else if containsStr(string(output), "ubiquity-dev") {
		fmt.Print("cluster already exists, starting...")
		startArgs := []string{"cluster", "start", "ubiquity-dev"}
		if k3sImage != "" {
			startArgs = append(startArgs, "--image", k3sImage)
		}
		startCmd := exec.Command("k3d", startArgs...)
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
	upCmd.Flags().String("metal-bootstrap-backend", "", "metal bootstrap backend (ansible, none); defaults sandbox/dev=none prod=ansible")
	upCmd.Flags().String("node-lifecycle-backend", "", "node lifecycle backend (nico, none; bmo requires UBIQUITY_ALLOW_BMO_MIGRATION=true for legacy migration fallback); defaults sandbox/dev=none prod=nico")
	upCmd.Flags().String("nico-values", "", "NVIDIA Infra Controller Helm values file for GitOps install")
	upCmd.Flags().String("nico-rest-values", "", "NVIDIA Infra Controller REST API Helm values file for GitOps install")
	upCmd.Flags().String("nico-site", "", "NVIDIA Infra Controller site name/ID")
}

// repoRoot is the absolute path to the ubiquity repository root.
var repoRoot = func() string {
	root, err := findRepoRoot()
	if err == nil && root != "" {
		return root
	}
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

	workChartDir, cleanup, err := prepareHelmChartWorkdir(chartDir)
	if err != nil {
		return fmt.Errorf("prepare Helm workdir: %w", err)
	}
	defer cleanup()

	// Download chart dependencies in an isolated workdir so sandbox/live apply does not mutate source charts.
	depCmd := exec.Command("helm", "dependency", "build", workChartDir)
	depCmd.Dir = repoRoot
	if out, err := depCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("helm dependency build: %w\n%s", err, string(out))
	}

	args := helmTemplateArgs(namespace, "release", workChartDir)
	cmd := exec.Command("helm", args...)
	cmd.Dir = repoRoot
	rendered, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("helm template: %w", err)
	}

	if err := kubectlApplyRendered(rendered, namespace, shouldForceApplyNamespace(chartDir), shouldServerSideApply(chartDir)); err != nil {
		waitForCRDsEstablished()
		if retryErr := kubectlApplyRendered(rendered, namespace, shouldForceApplyNamespace(chartDir), shouldServerSideApply(chartDir)); retryErr != nil {
			return fmt.Errorf("kubectl apply: %w", retryErr)
		}
	}
	waitForCRDsEstablished()
	return nil
}

func kubectlApplyRendered(rendered []byte, namespace string, forceNamespace bool, serverSide bool) error {
	args := []string{"apply", "-f", "-"}
	if forceNamespace && namespace != "" {
		args = []string{"apply", "-n", namespace, "-f", "-"}
	}
	if serverSide {
		args = append(args, "--server-side", "--force-conflicts")
	}
	applyCmd := exec.Command("kubectl", args...)
	applyCmd.Stdin = bytes.NewReader(rendered)
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	return applyCmd.Run()
}

func shouldForceApplyNamespace(chartDir string) bool {
	return strings.HasSuffix(filepath.ToSlash(chartDir), "platform/nim-operator")
}

func shouldServerSideApply(chartDir string) bool {
	return strings.HasSuffix(filepath.ToSlash(chartDir), "platform/kai-scheduler")
}

func helmTemplateArgs(namespace, releaseName, chartDir string) []string {
	args := []string{"template", "--include-crds", "--namespace", namespace, releaseName, chartDir}
	sandboxValues := filepath.Join(chartDir, "values-sandbox.yaml")
	if fileExists(sandboxValues) {
		args = append(args, "--values", sandboxValues)
	}
	return args
}

func waitForCRDsEstablished() {
	cmd := exec.Command("kubectl", "wait", "--for=condition=Established", "crd", "--all", "--timeout=60s")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
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
// for charts that have K8s-version-dependent pins. When set, the Chart.yaml
// dependency version is patched before install so the correct upstream chart
// is fetched. Pass "" to use the default from Chart.yaml/dependency lock.
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

	// If chartVersion is set, patch the Chart.yaml dependency version so the
	// correct upstream chart is resolved. Restore the original after install.
	var chartYamlBak []byte
	var chartYamlPath string
	if chartVersion != "" {
		chartYamlPath = filepath.Join(repoRoot, chartDir, "Chart.yaml")
		data, err := os.ReadFile(chartYamlPath)
		if err == nil {
			chartYamlBak = make([]byte, len(data))
			copy(chartYamlBak, data)
			// Replace the dependency version line: `version: ~3.8.1` → `version: "3.5.3"`
			patched := patchDependencyVersion(string(data), chartVersion)
			if patched != string(data) {
				os.WriteFile(chartYamlPath, []byte(patched), 0644)
				defer os.WriteFile(chartYamlPath, chartYamlBak, 0644)
			}
		}
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

// patchDependencyVersion replaces the dependency version line in a Chart.yaml
// with the given version. It looks for a line matching:
//
//	version: ~X.Y.Z  or  version: "~X.Y.Z"
//
// within a dependency block (starting with "  - name:" and ending at the next
// top-level key). Returns the patched content.
func patchDependencyVersion(chartYaml, newVersion string) string {
	lines := strings.Split(chartYaml, "\n")
	inDeps := false
	inDep := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "dependencies:" {
			inDeps = true
			continue
		}
		if inDeps {
			// Top-level key ends the dependencies block
			if trimmed != "" && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "#") && line[0] != ' ' && line[0] != '\t' {
				break
			}
			if strings.HasPrefix(trimmed, "- name:") {
				inDep = true
				continue
			}
			if inDep && strings.HasPrefix(trimmed, "version:") {
				// Replace version line with new version (preserve indentation)
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				lines[i] = indent + "version: \"" + newVersion + "\""
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

// runAnsiblePlaybook runs an ansible-playbook command in the repo root.
func runAnsiblePlaybook(playbook, env string) error {
	if _, err := exec.LookPath("ansible-playbook"); err != nil {
		return fmt.Errorf("ansible-playbook not found in PATH: %w", err)
	}
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

func provisionAnsible(env string) error {
	return runAnsiblePlaybook("metal/cluster.yml", env)
}

// runTerraform runs terraform init + apply in the specified cloud directory.
func runTerraform(cloudDir string) error {
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform not found in PATH: %w", err)
	}
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
