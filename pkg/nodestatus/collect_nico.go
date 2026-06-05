package nodestatus

import (
	"context"
	"sort"
	"strings"

	"github.com/ubiquitycluster/ubiquity/pkg/nico"
)

type NICoSource interface {
	ListSites(context.Context) ([]nico.Site, error)
	ListMachines(context.Context) ([]nico.Machine, error)
	ListInstances(context.Context) ([]nico.Instance, error)
	ListTasks(context.Context) ([]nico.Task, error)
	ListMachineGPUStats(context.Context) ([]nico.MachineGPUStats, error)
}

type Evidence struct {
	KubernetesNodes map[string]KubernetesNodeEvidence
}

func CollectNICo(ctx context.Context, source NICoSource, evidence Evidence) ([]NodeStatus, error) {
	if evidence.KubernetesNodes == nil {
		evidence.KubernetesNodes = map[string]KubernetesNodeEvidence{}
	}

	sites, err := source.ListSites(ctx)
	if err != nil {
		return nil, err
	}
	machines, err := source.ListMachines(ctx)
	if err != nil {
		return nil, err
	}
	instances, err := source.ListInstances(ctx)
	if err != nil {
		return nil, err
	}
	tasks, err := source.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	gpuStats, err := source.ListMachineGPUStats(ctx)
	if err != nil {
		return nil, err
	}

	siteNameByID := map[string]string{}
	for _, site := range sites {
		if site.ID != "" {
			siteNameByID[site.ID] = firstNonEmpty(site.Name, site.ID)
		}
	}

	instByMachine := map[string]nico.Instance{}
	for _, inst := range instances {
		if inst.MachineID != "" {
			instByMachine[inst.MachineID] = inst
		}
	}

	activeTaskByMachine := map[string]nico.Task{}
	for _, task := range tasks {
		if !isActiveTask(string(task.Status)) || task.MachineID == "" {
			continue
		}
		activeTaskByMachine[task.MachineID] = task
	}

	gpuByMachine := map[string]nico.MachineGPUStats{}
	for _, stat := range gpuStats {
		if stat.MachineID != "" {
			gpuByMachine[stat.MachineID] = stat
		}
	}

	out := make([]NodeStatus, 0, len(machines))
	for _, machine := range machines {
		inst := instByMachine[machine.ID]
		name := firstNonEmpty(inst.NodeName, inst.Name, machine.Name, machine.ID)
		k8s := lookupK8sEvidence(evidence.KubernetesNodes, inst.NodeName, inst.Name, machine.Name, machine.ID)
		nicoGPU := gpuByMachine[machine.ID]
		status := NodeStatus{
			Name:               name,
			Site:               firstNonEmpty(machine.SiteName, siteNameByID[machine.SiteID], machine.SiteID),
			MachineID:          machine.ID,
			InstanceID:         inst.ID,
			PowerState:         machine.PowerState,
			MachineStatus:      machine.Status,
			InstanceStatus:     inst.Status,
			OSImage:            firstNonEmpty(inst.OSImage, inst.OSID),
			KubernetesNodeName: k8s.Name,
			KubernetesReady:    k8s.Ready,
			Cordoned:           k8s.Cordoned,
			Roles:              append([]string{}, k8s.Roles...),
			GPUs:               maxInt(nicoGPU.Count, k8s.GPUs),
			NICoGPUs:           nicoGPU.Count,
			KubernetesGPUs:     k8s.GPUs,
			MIGProfiles:        CloneMIGProfiles(k8s.MIGProfiles),
			RDMAResources:      k8s.RDMAResources,
			NVIDIAReady:        k8s.NVIDIAReady,
			LastAction:         firstNonEmpty(machine.LastAction, inst.LastAction),
			Reason:             firstNonEmpty(machine.Reason, inst.Reason),
		}
		if task, ok := activeTaskByMachine[machine.ID]; ok {
			status.ActiveTaskID = task.ID
			status.LastAction = firstNonEmpty(task.Action, status.LastAction)
			status.Reason = firstNonEmpty(task.Error, status.Reason)
		}
		out = append(out, status)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Site == out[j].Site {
			return out[i].Name < out[j].Name
		}
		return out[i].Site < out[j].Site
	})
	return out, nil
}

func lookupK8sEvidence(nodes map[string]KubernetesNodeEvidence, keys ...string) KubernetesNodeEvidence {
	for _, key := range keys {
		if key == "" {
			continue
		}
		if node, ok := nodes[key]; ok {
			return node
		}
	}
	return KubernetesNodeEvidence{}
}

func isActiveTask(status string) bool {
	s := strings.ToLower(status)
	return s == "pending" || s == "running" || s == "in_progress" || s == "in-progress"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
