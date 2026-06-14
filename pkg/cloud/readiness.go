package cloud

import (
	"fmt"
	"sort"
	"strings"
)

// CloudCondition is the minimal condition evidence required from controller status.
type CloudCondition struct {
	Type   string `json:"type" yaml:"type"`
	Status string `json:"status" yaml:"status"`
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// CloudResourceEvidence captures reviewer-visible reconciliation status for one resource.
type CloudResourceEvidence struct {
	Kind       string           `json:"kind" yaml:"kind"`
	Namespace  string           `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name       string           `json:"name" yaml:"name"`
	Conditions []CloudCondition `json:"conditions" yaml:"conditions"`
}

// CloudReadinessEvidence is the evidence bundle evaluated before claiming cloud readiness.
type CloudReadinessEvidence struct {
	RequiredCRDs       []string                `json:"requiredCRDs" yaml:"requiredCRDs"`
	PresentCRDs        []string                `json:"presentCRDs" yaml:"presentCRDs"`
	Resources          []CloudResourceEvidence `json:"resources" yaml:"resources"`
	RequiredSmokeTests []string                `json:"requiredSmokeTests,omitempty" yaml:"requiredSmokeTests,omitempty"`
	SmokeTests         map[string]bool         `json:"smokeTests" yaml:"smokeTests"`
	Metadata           map[string]string       `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// CloudReadinessResult is a fail-closed readiness decision.
type CloudReadinessResult struct {
	Ready   bool     `json:"ready" yaml:"ready"`
	Reasons []string `json:"reasons" yaml:"reasons"`
}

// EvaluateCloudReadiness fails closed unless all required CRDs, controller conditions, and smoke tests are present.
func EvaluateCloudReadiness(ev CloudReadinessEvidence) CloudReadinessResult {
	var reasons []string
	present := make(map[string]struct{}, len(ev.PresentCRDs))
	for _, crd := range ev.PresentCRDs {
		present[strings.TrimSpace(crd)] = struct{}{}
	}
	for _, crd := range ev.RequiredCRDs {
		crd = strings.TrimSpace(crd)
		if crd == "" {
			continue
		}
		if _, ok := present[crd]; !ok {
			reasons = append(reasons, "missing CRD "+crd)
		}
	}
	for _, res := range ev.Resources {
		id := resourceID(res)
		if strings.TrimSpace(res.Kind) == "" || strings.TrimSpace(res.Name) == "" {
			reasons = append(reasons, "resource evidence missing kind or name")
			continue
		}
		if len(res.Conditions) == 0 {
			reasons = append(reasons, fmt.Sprintf("%s has no conditions", id))
			continue
		}
		if !hasPositiveReadyCondition(res.Conditions) {
			reasons = append(reasons, fmt.Sprintf("%s lacks Ready/Available/Bound true condition", id))
		}
	}
	var smokeNames []string
	for name := range ev.SmokeTests {
		smokeNames = append(smokeNames, name)
	}
	sort.Strings(smokeNames)
	for _, name := range smokeNames {
		if !ev.SmokeTests[name] {
			reasons = append(reasons, "smoke test "+name+" did not pass")
		}
	}
	for _, name := range ev.RequiredSmokeTests {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		passed, ok := ev.SmokeTests[name]
		if !ok {
			reasons = append(reasons, "missing required smoke test "+name)
			continue
		}
		if !passed {
			reasons = append(reasons, "smoke test "+name+" did not pass")
		}
	}
	return CloudReadinessResult{Ready: len(reasons) == 0, Reasons: reasons}
}

// RenderCloudReadinessReport renders a stable, reviewer-readable readiness decision.
func RenderCloudReadinessReport(result CloudReadinessResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ready: %t\n", result.Ready)
	b.WriteString("reasons:\n")
	if len(result.Reasons) == 0 {
		b.WriteString("  []\n")
		return b.String()
	}
	for _, reason := range result.Reasons {
		fmt.Fprintf(&b, "  - %s\n", reason)
	}
	return b.String()
}

func hasPositiveReadyCondition(conditions []CloudCondition) bool {
	for _, cond := range conditions {
		if !strings.EqualFold(cond.Status, "true") {
			continue
		}
		switch strings.ToLower(cond.Type) {
		case "ready", "available", "bound", "reconciled", "completed", "succeeded":
			return true
		}
	}
	return false
}

func resourceID(res CloudResourceEvidence) string {
	if strings.TrimSpace(res.Namespace) == "" {
		return fmt.Sprintf("%s %s", res.Kind, res.Name)
	}
	return fmt.Sprintf("%s %s/%s", res.Kind, res.Namespace, res.Name)
}
