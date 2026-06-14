package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/aiplatform"
)

var aiPlatformProfile string
var aiPlatformApplyDryRun bool
var aiPlatformApplyServerSide bool

var runAIPlatformKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdin = bytes.NewReader(stdin)
	return cmd.CombinedOutput()
}

var aiPlatformCmd = &cobra.Command{
	Use:   "ai-platform",
	Short: "Show NVIDIA-backed AI platform profile plan",
	Long: `Shows the declarative NVIDIA-backed AI platform profile, including
source-backed NVIDIA components, replacement decisions, and readiness claims.
This command does not claim NVIDIA approval or certification.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, err := aiplatform.GetProfile(aiPlatformProfile)
		if err != nil {
			return err
		}

		fmt.Printf("Profile: %s\n", profile.Name)
		fmt.Printf("Description: %s\n", profile.Description)
		fmt.Printf("Capabilities: %s\n", formatCapabilities(profile.Capabilities))
		fmt.Println("Components:")
		for _, component := range profile.Components {
			fmt.Printf("  - %s\n", component.Name)
			if component.SourceRepo != "" {
				fmt.Printf("    source: %s\n", component.SourceRepo)
			}
			if component.ChartRepository != "" {
				fmt.Printf("    chart: %s from %s\n", component.ChartName, component.ChartRepository)
			}
			if component.ReplacesLocal {
				fmt.Println("    replacement: replaces weaker Ubiquity-local functionality")
			}
			if component.ManagedByGPUOperator {
				fmt.Println("    management: managed by NVIDIA GPU Operator")
			}
			if component.Name == "ollama" && !component.ProductionDefault {
				fmt.Println("    Ollama: optional diagnostics only")
			}
		}
		fmt.Println("Bare-metal orchestration alternatives:")
		for _, alternative := range aiplatform.BareMetalOrchestrationAlternatives() {
			fmt.Printf("  - %s\n", alternative.Name)
			fmt.Printf("    source: %s\n", alternative.SourceRepo)
			fmt.Printf("    decision: %s\n", alternative.Decision)
			fmt.Printf("    scope: %s\n", alternative.Scope)
			if alternative.ReplacesLocal {
				fmt.Println("    replacement: replaces weaker Ubiquity-local functionality when enabled")
			}
			fmt.Printf("    evaluation: %s\n", alternative.Evaluation)
		}
		fmt.Println("Storage alternatives:")
		for _, alternative := range aiplatform.StorageAlternatives() {
			fmt.Printf("  - %s\n", alternative.Name)
			fmt.Printf("    source: %s\n", alternative.Source)
			fmt.Printf("    decision: %s\n", alternative.Decision)
			fmt.Printf("    scope: %s\n", alternative.Scope)
			if alternative.ReplacesLonghornForAIData {
				fmt.Println("    replacement: replaces Longhorn for AI dataset/cache paths")
			}
			if !alternative.ReplacesGenericPVCs {
				fmt.Println("    boundary: not a generic PVC replacement")
			}
			fmt.Printf("    rationale: %s\n", alternative.Rationale)
		}
		fmt.Println("Readiness policy: fail closed until GPU, runtime, telemetry, and serving evidence is proven.")
		fmt.Println("Approval policy: no NVIDIA approval/certification claim without explicit evidence.")
		return nil
	},
}

func formatCapabilities(capabilities []aiplatform.Capability) string {
	parts := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		parts = append(parts, string(capability))
	}
	return strings.Join(parts, ", ")
}

func init() {
	aiPlatformCmd.PersistentFlags().StringVar(&aiPlatformProfile, "profile", "gpu-basic", fmt.Sprintf("AI platform profile (%s)", strings.Join(aiplatform.Names(), ", ")))
	renderCmd := &cobra.Command{Use: "render", Short: "Render profile manifests", Args: cobra.NoArgs, RunE: runAIPlatformRender}
	applyCmd := &cobra.Command{Use: "apply", Short: "Apply rendered profile manifests", Args: cobra.NoArgs, RunE: runAIPlatformApply}
	applyCmd.Flags().BoolVar(&aiPlatformApplyDryRun, "dry-run", true, "use kubectl server-side dry-run instead of mutating the cluster")
	applyCmd.Flags().BoolVar(&aiPlatformApplyServerSide, "server-side", true, "use kubectl server-side apply for GitOps resources")
	aiPlatformCmd.AddCommand(renderCmd, applyCmd)
	rootCmd.AddCommand(aiPlatformCmd)
}

func runAIPlatformRender(cmd *cobra.Command, args []string) error {
	profile, err := aiplatform.GetProfile(aiPlatformProfile)
	if err != nil {
		return err
	}
	fmt.Print(renderAIPlatformManifest(profile))
	return nil
}

func runAIPlatformApply(cmd *cobra.Command, args []string) error {
	profile, err := aiplatform.GetProfile(aiPlatformProfile)
	if err != nil {
		return err
	}
	kubectlArgs := []string{"apply"}
	if aiPlatformApplyDryRun {
		kubectlArgs = append(kubectlArgs, "--dry-run=server")
	} else if aiPlatformApplyServerSide {
		kubectlArgs = append(kubectlArgs, "--server-side")
	}
	kubectlArgs = append(kubectlArgs, "-f", "-")
	out, err := runAIPlatformKubectl(cmd.Context(), kubectlArgs, []byte(renderAIPlatformManifest(profile)))
	if len(out) > 0 {
		fmt.Print(string(out))
	}
	return err
}

func renderAIPlatformManifest(profile aiplatform.Profile) string {
	var b strings.Builder
	b.WriteString("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: ubiquity-ai-platform-profile\n  namespace: ubiquity-system\n  labels:\n    app.kubernetes.io/name: ubiquity-ai-platform\n    ubiquity.ai/profile: \"")
	b.WriteString(profile.Name)
	b.WriteString("\"\n    not-nvidia-certified: \"true\"\ndata:\n  profile: ")
	b.WriteString(profile.Name)
	b.WriteString("\n  description: |\n    ")
	b.WriteString(profile.Description)
	b.WriteString("\n  readiness-policy: \"fail closed until GPU, runtime, telemetry, and serving evidence is proven\"\n  approval-policy: \"no NVIDIA approval/certification claim without explicit evidence\"\n  components: |\n")
	for _, component := range profile.Components {
		b.WriteString("    - name: ")
		b.WriteString(component.Name)
		b.WriteString("\n")
		if component.SourceRepo != "" {
			b.WriteString("      source: ")
			b.WriteString(component.SourceRepo)
			b.WriteString("\n")
		}
		if component.ChartName != "" {
			b.WriteString("      chart: ")
			b.WriteString(component.ChartName)
			b.WriteString("\n")
		}
		if component.ChartRepository != "" {
			b.WriteString("      chartRepository: ")
			b.WriteString(component.ChartRepository)
			b.WriteString("\n")
		}
		if component.Namespace != "" {
			b.WriteString("      namespace: ")
			b.WriteString(component.Namespace)
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("      replacesLocal: %t\n      productionDefault: %t\n", component.ReplacesLocal, component.ProductionDefault))
	}
	b.WriteString(renderAIPlatformGitOpsApplications(profile))
	return b.String()
}

type aiPlatformGitOpsTarget struct {
	Name      string
	Path      string
	Namespace string
}

func renderAIPlatformGitOpsApplications(profile aiplatform.Profile) string {
	targets := aiPlatformGitOpsTargets(profile)
	if len(targets) == 0 {
		return ""
	}
	var b strings.Builder
	for _, target := range targets {
		b.WriteString("---\napiVersion: argoproj.io/v1alpha1\nkind: Application\nmetadata:\n  name: ai-platform-")
		b.WriteString(target.Name)
		b.WriteString("\n  namespace: argocd\n  labels:\n    app.kubernetes.io/part-of: ubiquity-ai-platform\n    ubiquity.ai/profile: \"")
		b.WriteString(profile.Name)
		b.WriteString("\"\nspec:\n  project: default\n  source:\n    repoURL: https://github.com/ubiquitycluster/ubiquity\n    targetRevision: HEAD\n    path: ")
		b.WriteString(target.Path)
		b.WriteString("\n  destination:\n    server: https://kubernetes.default.svc\n    namespace: ")
		b.WriteString(target.Namespace)
		b.WriteString("\n  syncPolicy:\n    automated:\n      prune: true\n      selfHeal: true\n    syncOptions:\n      - CreateNamespace=true\n")
	}
	return b.String()
}

func aiPlatformGitOpsTargets(profile aiplatform.Profile) []aiPlatformGitOpsTarget {
	components := profile.ComponentsByName()
	targets := []aiPlatformGitOpsTarget{}
	add := func(componentName, appName, path, namespace string) {
		if _, ok := components[componentName]; ok {
			targets = append(targets, aiPlatformGitOpsTarget{Name: appName, Path: path, Namespace: namespace})
		}
	}
	add("gpu-operator", "nvidia-gpu-operator", "system/nvidia-gpu-operator", "gpu-operator")
	add("nvidia-network-operator", "nvidia-network-operator", "system/nvidia-network-operator", "nvidia-network-operator")
	add("nim-operator", "nim-operator", "platform/nim-operator", "nim-operator")
	add("kai-scheduler", "kai-scheduler", "platform/kai-scheduler", "kai-scheduler")
	add("stallscope", "stallscope", "platform/stallscope", "gpu-telemetry")
	if profile.HasCapability(aiplatform.CapabilityServing) || profile.HasCapability(aiplatform.CapabilityScheduler) {
		targets = append(targets, aiPlatformGitOpsTarget{Name: "ai-workload-tenancy", Path: "platform/ai-workload-tenancy", Namespace: "ai-workloads"})
	}
	return targets
}
