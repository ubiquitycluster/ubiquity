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
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ChartCompat records the compatible upstream chart version for a given
// Kubernetes minor version. Charts that have no K8s version constraints
// (e.g., argo-cd) are not listed here — they use the default version
// from Chart.yaml.
type ChartCompat struct {
	// Chart is the chart directory path (e.g., "system/kyverno").
	Chart string
	// Name is the helm release name (e.g., "kyverno").
	Name string
	// Namespace is the target namespace.
	Namespace string
	// Pins maps a K8s minor version (e.g., "30", "31", "32") to the
	// upstream chart version (e.g., "3.5.3", "3.8.1").
	Pins map[string]string
}

// compatMatrix defines the version pinning for all managed Helm charts.
// Update this when adding new K8s versions or updating chart pins.
//
// See docs/compatibility/ for the full compatibility documentation and ADRs.
var compatMatrix = []ChartCompat{
	{
		Chart:     "system/kyverno",
		Name:      "kyverno",
		Namespace: "kyverno",
		Pins: map[string]string{
			"30": "3.5.3", // K8s 1.30 — lacks selectableFields in CRDs
			"31": "3.5.3", // K8s 1.31 — lacks selectableFields in CRDs
			"32": "3.8.1", // K8s 1.32 — full feature support
		},
	},
}

// detectKubeVersion returns the Kubernetes minor version as an integer
// (e.g., 32 for v1.32.x) by querying the cluster via kubectl.
func detectKubeVersion() (int, error) {
	cmd := exec.Command("kubectl", "version", "-o", "json")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("kubectl version failed: %w", err)
	}

	var v struct {
		ServerVersion struct {
			Major string `json:"major"`
			Minor string `json:"minor"`
		} `json:"serverVersion"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		return 0, fmt.Errorf("parsing kubectl version output: %w", err)
	}

	// Strip "+" suffix that k3s sometimes appends (e.g., "32+" → 32)
	minor := strings.TrimRight(v.ServerVersion.Minor, "+")
	return strconv.Atoi(minor)
}

// lookupChartVersion finds the best chart version for the given chart
// directory and K8s minor version. It first looks for an exact match.
// If none exists, it falls back to the highest pin that is <= the
// running version. Returns empty string if no pin is found (caller
// should use the chart's default version).
func lookupChartVersion(chartDir, kubeMinor string) string {
	for _, c := range compatMatrix {
		if !strings.Contains(c.Chart, chartDir) {
			continue
		}
		// Exact match
		if v, ok := c.Pins[kubeMinor]; ok {
			return v
		}
		// Fallback: find highest pin <= running version
		runningVer, err := strconv.Atoi(kubeMinor)
		if err != nil {
			return ""
		}
		var bestVersion int
		var bestPin string
		for kv, cv := range c.Pins {
			pinVer, err := strconv.Atoi(kv)
			if err != nil {
				continue
			}
			if pinVer <= runningVer && pinVer > bestVersion {
				bestVersion = pinVer
				bestPin = cv
			}
		}
		return bestPin
	}
	return ""
}

// kubeMinorFromVersion parses a major.minor.patch string and returns
// the minor version as a string (e.g., "1.32.13" → "32").
func kubeMinorFromVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// sortedChartPins returns the K8s minor versions for a chart sorted
// ascending. Used for display and testing.
func sortedChartPins(chart string) []string {
	for _, c := range compatMatrix {
		if c.Chart == chart {
			keys := make([]string, 0, len(c.Pins))
			for k := range c.Pins {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		}
	}
	return nil
}
