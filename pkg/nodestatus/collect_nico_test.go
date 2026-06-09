package nodestatus

import (
	"context"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/nico"
)

type fakeNICo struct {
	sites    []nico.Site
	machines []nico.Machine
	inst     []nico.Instance
	tasks    []nico.Task
	gpus     []nico.MachineGPUStats
}

func (f fakeNICo) ListSites(context.Context) ([]nico.Site, error)         { return f.sites, nil }
func (f fakeNICo) ListMachines(context.Context) ([]nico.Machine, error)   { return f.machines, nil }
func (f fakeNICo) ListInstances(context.Context) ([]nico.Instance, error) { return f.inst, nil }
func (f fakeNICo) ListTasks(context.Context) ([]nico.Task, error)         { return f.tasks, nil }
func (f fakeNICo) ListMachineGPUStats(context.Context) ([]nico.MachineGPUStats, error) {
	return f.gpus, nil
}

func TestCollectNICoAggregatesMachineInstanceKubernetesAndNVIDIAEvidence(t *testing.T) {
	src := fakeNICo{
		sites:    []nico.Site{{ID: "site-1", Name: "sf01"}},
		machines: []nico.Machine{{ID: "mach-1", Name: "cn01-bmc", SiteID: "site-1", PowerState: "on", Status: "provisioned"}},
		inst:     []nico.Instance{{ID: "inst-1", Name: "cn01", NodeName: "cn01", MachineID: "mach-1", Status: "running", OSImage: "rocky-9.4-gpu"}},
		tasks:    []nico.Task{{ID: "task-1", MachineID: "mach-1", Status: nico.TaskRunning, Action: "reinstall"}},
		gpus:     []nico.MachineGPUStats{{MachineID: "mach-1", Count: 8}},
	}
	got, err := CollectNICo(context.Background(), src, Evidence{KubernetesNodes: map[string]KubernetesNodeEvidence{
		"cn01": {Name: "cn01", Ready: true, Roles: []string{"worker"}, GPUs: 8, RDMAResources: 4, NVIDIAReady: true},
	}})
	if err != nil {
		t.Fatalf("CollectNICo returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d statuses, want 1", len(got))
	}
	n := got[0]
	if n.Name != "cn01" || n.Site != "sf01" || n.MachineID != "mach-1" || n.InstanceID != "inst-1" {
		t.Fatalf("unexpected identity fields: %+v", n)
	}
	if !n.KubernetesReady || !n.NVIDIAReady || n.GPUs != 8 || n.NICoGPUs != 8 || n.KubernetesGPUs != 8 || n.RDMAResources != 4 || n.ActiveTaskID != "task-1" || n.LastAction != "reinstall" {
		t.Fatalf("aggregation missed readiness/resource/task evidence: %+v", n)
	}
	if n.BMCStatus != "power:on" || n.KubeletStatus != "Ready" || n.GPUStatus != "ready" || n.RDMAStatus != "ready" || n.ImageStatus != "image:rocky-9.4-gpu" || n.MaintenanceState != "active-task:reinstall" || n.FirmwareStatus != "unknown" {
		t.Fatalf("long-term status aggregation fields missing or wrong: %+v", n)
	}
}

func TestCollectNICoDoesNotOverwriteNICoGPUsWithKubernetesEvidence(t *testing.T) {
	src := fakeNICo{
		machines: []nico.Machine{{ID: "mach-1", Name: "cn01", Status: "provisioned"}},
		inst:     []nico.Instance{{ID: "inst-1", NodeName: "cn01", MachineID: "mach-1"}},
		gpus:     []nico.MachineGPUStats{{MachineID: "mach-1", Count: 8}},
	}
	got, err := CollectNICo(context.Background(), src, Evidence{KubernetesNodes: map[string]KubernetesNodeEvidence{
		"cn01": {Name: "cn01", Ready: true, GPUs: 0, MIGProfiles: map[string]int{"nvidia.com/mig-1g.10gb": 7}},
	}})
	if err != nil {
		t.Fatalf("CollectNICo returned error: %v", err)
	}
	n := got[0]
	if n.NICoGPUs != 8 || n.KubernetesGPUs != 0 || n.GPUs != 8 {
		t.Fatalf("GPU evidence was not kept separate: %+v", n)
	}
}

func TestCollectNICoIncludesSpareMachineWithoutKubernetesEvidence(t *testing.T) {
	got, err := CollectNICo(context.Background(), fakeNICo{machines: []nico.Machine{{ID: "mach-spare", Name: "cn03", PowerState: "off", Status: "available"}}}, Evidence{})
	if err != nil {
		t.Fatalf("CollectNICo returned error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "cn03" || got[0].KubernetesReady || got[0].MIGProfiles == nil {
		t.Fatalf("unexpected spare status: %+v", got)
	}
}
