package nico

import (
	"strings"
	"testing"
)

func TestEvaluateReadinessAllReadyMockMode(t *testing.T) {
	s := ReadinessSnapshot{
		Workloads:      []WorkloadStatus{{Name: "nico-api", Ready: true}, {Name: "nico-rest-site-agent", Ready: true}},
		RESTAPIReady:   true,
		SiteAgentReady: true,
		Services:       allNICoServicesReady(),
	}
	res := EvaluateReadiness(s, ReadinessOptions{RealHardware: false})
	if !res.Ready {
		t.Fatalf("ready = false, failures: %#v", res.Failures)
	}
}

func TestEvaluateReadinessReportsMissingFoundations(t *testing.T) {
	s := ReadinessSnapshot{
		Workloads: []WorkloadStatus{{Name: "nico-api", Ready: false}, {Name: "nico-pxe", Ready: true}},
		Services:  map[string]bool{"nico-dhcp": true},
	}
	res := EvaluateReadiness(s, ReadinessOptions{RealHardware: true})
	if res.Ready {
		t.Fatalf("ready = true, want false")
	}
	joined := strings.Join(res.Failures, "\n")
	for _, want := range []string{"workload nico-api not ready", "REST API not ready", "site-agent not ready", "service nico-api not ready", "service nico-dns not ready", "service nico-ntp not ready", "service nico-pxe not ready", "service nico-bmc-proxy not ready", "service nico-hardware-health not ready", "service nico-rest-api not ready", "service nico-rest-site-agent not ready", "site not visible", "machine not visible"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("failures missing %q: %#v", want, res.Failures)
		}
	}
}

func TestEvaluateReadinessRequiresSiteAndMachineOnlyForRealHardware(t *testing.T) {
	s := ReadinessSnapshot{
		RESTAPIReady:   true,
		SiteAgentReady: true,
		Services:       allNICoServicesReady(),
	}
	if res := EvaluateReadiness(s, ReadinessOptions{RealHardware: false}); !res.Ready {
		t.Fatalf("mock readiness should not require site/machine: %#v", res.Failures)
	}
	if res := EvaluateReadiness(s, ReadinessOptions{RealHardware: true}); res.Ready {
		t.Fatalf("real hardware readiness should require site/machine")
	}
}

func TestChartComponentNamesReturnsNewWrapperComponentNames(t *testing.T) {
	joined := strings.Join(ChartComponentNames(), ",")
	for _, want := range []string{"nico-api", "nico-bmc-proxy", "nico-dhcp", "nico-dns", "nico-hardware-health", "nico-ntp", "nico-pxe", "nico-ssh-console-rs", "nico-rest-api", "nico-rest-site-agent"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("component list missing %q: %s", want, joined)
		}
	}
}

func allNICoServicesReady() map[string]bool {
	out := map[string]bool{}
	for _, name := range requiredFoundationServices {
		out[name] = true
	}
	return out
}
