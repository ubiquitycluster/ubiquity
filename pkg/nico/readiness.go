package nico

import "fmt"

type WorkloadStatus struct {
	Name  string
	Ready bool
}

type ReadinessSnapshot struct {
	Workloads      []WorkloadStatus
	RESTAPIReady   bool
	SiteAgentReady bool
	Services       map[string]bool
	SiteVisible    bool
	MachineVisible bool
}

type ReadinessOptions struct {
	RealHardware bool
}

type ReadinessResult struct {
	Ready    bool
	Failures []string
}

var requiredFoundationServices = []string{"nico-api", "nico-bmc-proxy", "nico-dhcp", "nico-dns", "nico-hardware-health", "nico-ntp", "nico-pxe", "nico-ssh-console-rs", "nico-rest-api", "nico-rest-site-agent"}

func ChartComponentNames() []string {
	return append([]string{}, requiredFoundationServices...)
}

func EvaluateReadiness(s ReadinessSnapshot, opts ReadinessOptions) ReadinessResult {
	var failures []string
	for _, w := range s.Workloads {
		if !w.Ready {
			failures = append(failures, fmt.Sprintf("workload %s not ready", w.Name))
		}
	}
	if !s.RESTAPIReady {
		failures = append(failures, "REST API not ready")
	}
	if !s.SiteAgentReady {
		failures = append(failures, "site-agent not ready")
	}
	for _, name := range requiredFoundationServices {
		if !s.Services[name] {
			failures = append(failures, fmt.Sprintf("service %s not ready", name))
		}
	}
	if opts.RealHardware {
		if !s.SiteVisible {
			failures = append(failures, "site not visible")
		}
		if !s.MachineVisible {
			failures = append(failures, "machine not visible")
		}
	}
	return ReadinessResult{Ready: len(failures) == 0, Failures: failures}
}
