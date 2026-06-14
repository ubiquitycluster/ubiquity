package nodeinventory

import (
	"sort"

	"go.yaml.in/yaml/v3"
)

type ansibleInventory struct {
	All ansibleAll `yaml:"all"`
}

type ansibleAll struct {
	Vars  map[string]any            `yaml:"vars"`
	Hosts map[string]map[string]any `yaml:"hosts"`
}

func RenderAnsibleBootstrapInventory(inventory NodeInventory) (string, error) {
	if err := inventory.Validate(); err != nil {
		return "", err
	}

	hosts := make(map[string]map[string]any, len(inventory.Nodes))
	for _, node := range inventory.Nodes {
		host := map[string]any{
			"site":                   node.Site,
			"role":                   node.Role,
			"os_image_ref":           node.OSImageRef,
			"instance_type_ref":      node.InstanceTypeRef,
			"gpu_profile":            node.GPUProfile,
			"join_profile":           node.JoinProfile,
			"machine_selector":       sortedMap(node.MachineSelector),
			"labels":                 sortedMap(node.Labels),
			"bootstrap_only":         true,
			"node_lifecycle_backend": "nico",
		}
		hosts[node.Name] = host
	}

	payload := ansibleInventory{All: ansibleAll{
		Vars: map[string]any{
			"bootstrap_only":         true,
			"node_lifecycle_backend": "nico",
			"nico_day2_manager":      true,
		},
		Hosts: hosts,
	}}
	data, err := yaml.Marshal(payload)
	if err != nil {
		return "", err
	}
	return "# bootstrap-only\n# Day-2 node lifecycle is managed by NICo, not Ansible, unless fallback mode is explicit.\n" + string(data), nil
}

func sortedMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make(map[string]string, len(in))
	for _, key := range keys {
		out[key] = in[key]
	}
	return out
}
