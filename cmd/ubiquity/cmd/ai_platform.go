package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/aiplatform"
)

var aiPlatformProfile string

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
	aiPlatformCmd.Flags().StringVar(&aiPlatformProfile, "profile", "gpu-basic", fmt.Sprintf("AI platform profile (%s)", strings.Join(aiplatform.Names(), ", ")))
	rootCmd.AddCommand(aiPlatformCmd)
}
