package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/nico"
	"github.com/ubiquitycluster/ubiquity/pkg/nodeinventory"
	"github.com/ubiquitycluster/ubiquity/pkg/nodestatus"
	"go.yaml.in/yaml/v3"
)

const (
	nodeBackendNICO        = "nico"
	nodeBackendBMO         = "bmo"
	nodeBackendNone        = "none"
	nodeLiveCommandTimeout = 30 * time.Second
)

type nodeCommandOptions struct {
	Backend             string
	Output              string
	DryRun              bool
	Confirm             string
	Site                string
	OS                  string
	Power               string
	Inventory           string
	Force               bool
	Reason              string
	DrainConfirmed      bool
	StorageAcknowledged bool
	AIStoreAcknowledged bool
}

type nodesNICOClient interface {
	GetMachine(context.Context, string) (nico.Machine, error)
	CreateOperatingSystem(context.Context, nico.OperatingSystem) (nico.OperatingSystem, error)
	CreateInstance(context.Context, nico.Instance) (nico.Instance, error)
	DeleteInstance(context.Context, string) error
	PowerMachine(context.Context, string, string, string) (nico.Task, error)
	ListSites(context.Context) ([]nico.Site, error)
	ListMachines(context.Context) ([]nico.Machine, error)
	ListInstances(context.Context) ([]nico.Instance, error)
	ListTasks(context.Context) ([]nico.Task, error)
	ListMachineGPUStats(context.Context) ([]nico.MachineGPUStats, error)
	GetTask(context.Context, string) (nico.Task, error)
}

var newNodesNICOClient = func(cfg nico.Config) (nodesNICOClient, error) {
	return nico.NewClient(cfg)
}

var runNodesKubectl = func(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "kubectl", args...).Output()
}

var nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "table", DryRun: true}

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Operate bare-metal node lifecycle through NVIDIA Infra Controller",
	Long:  "Safe wrappers for NVIDIA Infra Controller-backed node inventory, status, OS and lifecycle tasks. Defaults to NICo, not BMO.",
}

func init() {
	rootCmd.AddCommand(nodesCmd)
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.Backend, "backend", nodeBackendNICO, "node lifecycle backend (nico, bmo, none)")
	nodesCmd.PersistentFlags().StringVarP(&nodeOpts.Output, "output", "o", "table", "output format (table, json)")
	nodesCmd.PersistentFlags().BoolVar(&nodeOpts.DryRun, "dry-run", true, "describe the NICo operation without mutating hardware")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.Confirm, "confirm", "", "confirm destructive node lifecycle action by exact node name")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.Site, "site", "", "NICo site name or ID")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.OS, "os", "", "operating system image/name for apply/reinstall")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.OS, "os-image", "", "alias for --os")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.Power, "state", "", "power state for power command")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.Inventory, "inventory", "", "Ubiquity NodeInventory YAML for NICo OS apply/add/reinstall operations")
	nodesCmd.PersistentFlags().BoolVar(&nodeOpts.Force, "force", false, "override selected safety gates")
	nodesCmd.PersistentFlags().StringVar(&nodeOpts.Reason, "reason", "", "required reason when --force is used")
	nodesCmd.PersistentFlags().BoolVar(&nodeOpts.DrainConfirmed, "drain-confirmed", false, "acknowledge node has been cordoned/drained")
	nodesCmd.PersistentFlags().BoolVar(&nodeOpts.StorageAcknowledged, "storage-ack", false, "acknowledge local persistent volume storage risk")
	nodesCmd.PersistentFlags().BoolVar(&nodeOpts.AIStoreAcknowledged, "aistore-ack", false, "acknowledge AIStore target data risk")

	for _, spec := range []struct {
		use  string
		args cobra.PositionalArgs
		run  func(*cobra.Command, []string) error
	}{
		{"list", cobra.NoArgs, runNodesList},
		{"status [node]", cobra.MaximumNArgs(1), runNodesStatus},
		{"add <node>", cobra.ExactArgs(1), runNodesAction("add", false)},
		{"drain <node>", cobra.ExactArgs(1), runNodesAction("drain", true)},
		{"remove <node>", cobra.ExactArgs(1), runNodesAction("remove", true)},
		{"reinstall <node>", cobra.ExactArgs(1), runNodesAction("reinstall", true)},
		{"power <name> on|off|reset", validatePowerArgs, runNodesAction("power", true)},
		{"task <task-id>", cobra.ExactArgs(1), runNodesTask},
	} {
		nodesCmd.AddCommand(&cobra.Command{Use: spec.use, Args: spec.args, RunE: spec.run})
	}

	osCmd := &cobra.Command{Use: "os", Short: "Manage NICo Operating System objects"}
	osCmd.AddCommand(&cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: runNodesOSList})
	osCmd.AddCommand(&cobra.Command{Use: "apply [image]", Args: cobra.MaximumNArgs(1), RunE: runNodesAction("os apply", false)})
	nodesCmd.AddCommand(osCmd)
}

func validatePowerArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(2)(cmd, args); err != nil {
		return err
	}
	switch args[1] {
	case "on", "off", "reset":
		return nil
	default:
		return fmt.Errorf("power syntax: power <name> on|off|reset")
	}
}

func runNodesList(cmd *cobra.Command, args []string) error {
	cfg := nicoConfigFromEnv(nodeOpts.Site)
	if err := requireNodeBackend(cfg); err != nil {
		return err
	}
	if cfg.Mode == nico.ModeMock || nodeOpts.DryRun {
		return renderNodeRows([]map[string]string{{"name": "mock-node-1", "backend": "nico", "status": "dry-run", "site": nodeOpts.Site}}, nodeOpts.Output)
	}
	client, err := nico.NewClient(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), nodeLiveCommandTimeout)
	defer cancel()
	statuses, err := collectLiveNodeStatuses(ctx, client)
	if err != nil {
		return err
	}
	rows := make([]map[string]string, 0, len(statuses))
	for _, s := range statuses {
		rows = append(rows, map[string]string{"name": s.Name, "id": s.MachineID, "backend": "nico", "status": s.MachineStatus, "site": s.Site})
	}
	return renderNodeRows(rows, nodeOpts.Output)
}

func runNodesStatus(cmd *cobra.Command, args []string) error {
	return runNodesAction("status", false)(cmd, args)
}
func runNodesTask(cmd *cobra.Command, args []string) error {
	return runNodesAction("task", false)(cmd, args)
}

func runNodesOSList(cmd *cobra.Command, args []string) error {
	cfg := nicoConfigFromEnv(nodeOpts.Site)
	if err := requireNodeBackend(cfg); err != nil {
		return err
	}
	if cfg.Mode == nico.ModeMock || nodeOpts.DryRun {
		return renderNodeRows([]map[string]string{{"name": "mock-os", "backend": "nico", "status": "dry-run"}}, nodeOpts.Output)
	}
	client, err := nico.NewClient(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), nodeLiveCommandTimeout)
	defer cancel()
	oses, err := client.ListOperatingSystems(ctx)
	if err != nil {
		return err
	}
	rows := make([]map[string]string, 0, len(oses))
	for _, os := range oses {
		rows = append(rows, map[string]string{"name": os.Name, "id": os.ID, "backend": "nico"})
	}
	return renderNodeRows(rows, nodeOpts.Output)
}

func runNodesAction(action string, destructive bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg := nicoConfigFromEnv(nodeOpts.Site)
		if err := requireNodeBackend(cfg); err != nil {
			return err
		}
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		powerState := nodeOpts.Power
		if action == "power" && len(args) > 1 {
			powerState = args[1]
		}
		if destructive && (cfg.Mode == nico.ModeMock || nodeOpts.DryRun) {
			if err := evaluateNodeActionSafety(action, target, powerState, nodestatus.NodeStatus{Name: target}, nil); err != nil {
				return err
			}
		}
		if cfg.Mode == nico.ModeMock || nodeOpts.DryRun {
			payload := map[string]string{"action": action, "backend": "nico", "mode": string(cfg.WithDefaults().Mode), "target": target, "site": nodeOpts.Site, "os": nodeOpts.OS, "powerState": powerState, "effect": "dry-run/mock safe wrapper; no hardware mutation performed"}
			return renderNodeRows([]map[string]string{payload}, nodeOpts.Output)
		}
		client, err := newNodesNICOClient(cfg)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), nodeLiveCommandTimeout)
		defer cancel()
		if destructive {
			statuses, err := collectLiveNodeStatuses(ctx, client)
			if err != nil {
				return err
			}
			resolved, err := resolveNodeTargetStatus(statuses, target)
			if err != nil {
				return err
			}
			if err := evaluateNodeActionSafety(action, resolved.Name, powerState, resolved, statuses); err != nil {
				return err
			}
		}
		return runLiveNodesAction(ctx, client, action, target, powerState, args)
	}
}

func evaluateNodeActionSafety(action, target, powerState string, node nodestatus.NodeStatus, statuses []nodestatus.NodeStatus) error {
	op := nodestatus.Operation(action)
	if action == "power" {
		switch powerState {
		case "off":
			op = nodestatus.OperationPowerOff
		case "reset":
			op = nodestatus.OperationReset
		default:
			op = nodestatus.Operation("power-" + powerState)
		}
	}
	if node.Name == "" {
		node.Name = target
	}
	cpReady, cpTotal := controlPlaneCounts(statuses)
	if len(statuses) == 0 {
		cpReady, cpTotal = 1, 1
	}
	storage := nodeStorageSafety{}
	if len(statuses) > 0 {
		storage = collectNodeStorageSafety(context.Background(), node.Name)
	}
	_, err := nodestatus.EvaluateSafety(nodestatus.SafetyRequest{Operation: op, NodeName: node.Name, Confirm: nodeOpts.Confirm, Force: nodeOpts.Force, ForceReason: nodeOpts.Reason, DrainConfirmed: nodeOpts.DrainConfirmed, StorageAcknowledged: nodeOpts.StorageAcknowledged, AIStoreAcknowledged: nodeOpts.AIStoreAcknowledged}, nodestatus.SafetyClusterState{Node: node, ControlPlaneReady: cpReady, ControlPlaneTotal: cpTotal, HasLocalPV: storage.HasLocalPV, HasAIStoreTargetData: storage.HasAIStoreTargetData})
	return err
}

func runLiveNodesAction(ctx context.Context, client nodesNICOClient, action, target, powerState string, args []string) error {
	switch action {
	case "add":
		if strings.TrimSpace(nodeOpts.Inventory) != "" {
			inst, err := createInstanceFromInventory(ctx, client, nodeOpts.Inventory, target)
			if err != nil {
				return err
			}
			return renderNodeRows([]map[string]string{instanceNodeRow(inst, action)}, nodeOpts.Output)
		}
		inst, err := client.CreateInstance(ctx, nico.Instance{Name: target, NodeName: target, OSImage: nodeOpts.OS})
		if err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{instanceNodeRow(inst, action)}, nodeOpts.Output)
	case "remove":
		statuses, err := collectLiveNodeStatuses(ctx, client)
		if err != nil {
			return err
		}
		resolved, err := resolveNodeTargetStatus(statuses, target)
		if err != nil {
			return err
		}
		if resolved.InstanceID == "" {
			return fmt.Errorf("node target %q resolved to machine %q but has no NICo instance ID; refusing DeleteInstance", target, resolved.MachineID)
		}
		if err := client.DeleteInstance(ctx, resolved.InstanceID); err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{{"name": resolved.Name, "id": resolved.InstanceID, "backend": "nico", "status": "deleted", "effect": "delete instance requested"}}, nodeOpts.Output)
	case "reinstall":
		inst, err := client.CreateInstance(ctx, nico.Instance{Name: target, NodeName: target, OSImage: nodeOpts.OS, LastAction: "reinstall"})
		if err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{instanceNodeRow(inst, action)}, nodeOpts.Output)
	case "os apply":
		if strings.TrimSpace(nodeOpts.Inventory) != "" {
			oses, err := loadNICoOperatingSystemsFromInventory(nodeOpts.Inventory)
			if err != nil {
				return err
			}
			rows := make([]map[string]string, 0, len(oses))
			for _, osSpec := range oses {
				created, err := client.CreateOperatingSystem(ctx, osSpec)
				if err != nil {
					return err
				}
				rows = append(rows, map[string]string{"name": firstNonEmptyString(created.Name, created.Metadata.Name, osSpec.Name), "id": created.ID, "backend": "nico", "status": "created", "effect": "operating-system applied from inventory"})
			}
			return renderNodeRows(rows, nodeOpts.Output)
		}
		image := nodeOpts.OS
		if image == "" && len(args) > 0 {
			image = args[0]
		}
		if image == "" {
			return fmt.Errorf("nodes os apply requires an image argument or --inventory")
		}
		os, err := client.CreateOperatingSystem(ctx, nico.OperatingSystem{Name: image})
		if err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{{"name": os.Name, "id": os.ID, "backend": "nico", "status": "created", "effect": "operating-system applied"}}, nodeOpts.Output)
	case "power":
		statuses, err := collectLiveNodeStatuses(ctx, client)
		if err != nil {
			return err
		}
		resolved, err := resolveNodeTargetStatus(statuses, target)
		if err != nil {
			return err
		}
		if resolved.MachineID == "" {
			return fmt.Errorf("node target %q has no NICo machine ID; refusing power operation", target)
		}
		task, err := client.PowerMachine(ctx, resolved.MachineID, powerState, nodeOpts.Reason)
		if err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{{"name": resolved.Name, "id": task.ID, "backend": "nico", "status": string(task.Status), "effect": "power " + powerState, "machineId": resolved.MachineID}}, nodeOpts.Output)
	case "drain":
		statuses, err := collectLiveNodeStatuses(ctx, client)
		if err != nil {
			return err
		}
		resolved, err := resolveNodeTargetStatus(statuses, target)
		if err != nil {
			return err
		}
		k8sNode := firstNonEmptyString(resolved.KubernetesNodeName, resolved.Name, target)
		if err := drainKubernetesNode(ctx, k8sNode); err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{{"name": resolved.Name, "id": resolved.InstanceID, "backend": "nico", "status": "drained", "effect": "cordoned/drained", "machineId": resolved.MachineID}}, nodeOpts.Output)
	case "status":
		statuses, err := collectLiveNodeStatuses(ctx, client)
		if err != nil {
			return err
		}
		rows := make([]map[string]string, 0, len(statuses))
		for _, s := range statuses {
			if target != "" && target != s.Name && target != s.MachineID && target != s.InstanceID {
				continue
			}
			rows = append(rows, map[string]string{"name": s.Name, "id": s.MachineID, "instance": s.InstanceID, "backend": "nico", "status": firstNonEmptyString(s.MachineStatus, s.InstanceStatus), "site": s.Site, "effect": s.LastAction})
		}
		if target != "" && len(rows) == 0 {
			return fmt.Errorf("node %q not found in NICo status", target)
		}
		return renderNodeRows(rows, nodeOpts.Output)
	case "task":
		task, err := client.GetTask(ctx, target)
		if err != nil {
			return err
		}
		return renderNodeRows([]map[string]string{{"name": task.ID, "id": task.ID, "backend": "nico", "status": string(task.Status), "effect": task.Action, "machineId": task.MachineID, "error": task.Error}}, nodeOpts.Output)
	default:
		return fmt.Errorf("nodes action %q is not implemented", action)
	}
}

func collectLiveNodeStatuses(ctx context.Context, client nodesNICOClient) ([]nodestatus.NodeStatus, error) {
	return nodestatus.CollectNICo(ctx, client, collectNodeKubernetesEvidence(ctx))
}

func loadNICoOperatingSystemsFromInventory(path string) ([]nico.OperatingSystem, error) {
	inventory, err := loadNodeInventory(path)
	if err != nil {
		return nil, err
	}
	rendered, err := nodeinventory.RenderNICoOperatingSystems(inventory)
	if err != nil {
		return nil, err
	}
	out := make([]nico.OperatingSystem, 0, len(rendered))
	for _, osSpec := range rendered {
		out = append(out, nicoOperatingSystemFromRendered(osSpec))
	}
	return out, nil
}

func loadNodeInventory(path string) (nodeinventory.NodeInventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nodeinventory.NodeInventory{}, fmt.Errorf("read node inventory %q: %w", path, err)
	}
	var inventory nodeinventory.NodeInventory
	if err := yaml.Unmarshal(data, &inventory); err != nil {
		return nodeinventory.NodeInventory{}, fmt.Errorf("parse node inventory %q: %w", path, err)
	}
	if err := inventory.Validate(); err != nil {
		return nodeinventory.NodeInventory{}, err
	}
	return inventory, nil
}

func nicoOperatingSystemFromRendered(osSpec nodeinventory.NICoOperatingSystem) nico.OperatingSystem {
	return nico.OperatingSystem{
		Name:       osSpec.Metadata.Name,
		APIVersion: osSpec.APIVersion,
		Kind:       osSpec.Kind,
		Metadata: nico.ObjectMetadata{
			Name:   osSpec.Metadata.Name,
			Labels: osSpec.Metadata.Labels,
		},
		Spec: nico.OperatingSystemSpec{
			Family:       osSpec.Spec.Family,
			Version:      osSpec.Spec.Version,
			Architecture: osSpec.Spec.Architecture,
			ImageURL:     osSpec.Spec.ImageURL,
			Checksum:     osSpec.Spec.Checksum,
			Provenance:   osSpec.Spec.Provenance,
			IPXEScript:   osSpec.Spec.IPXEScript,
			UserData:     osSpec.Spec.UserData,
			Labels:       osSpec.Spec.Labels,
		},
	}
}

func createInstanceFromInventory(ctx context.Context, client nodesNICOClient, path, target string) (nico.Instance, error) {
	inventory, err := loadNodeInventory(path)
	if err != nil {
		return nico.Instance{}, err
	}
	rendered, err := nodeinventory.RenderNICoOperatingSystems(inventory)
	if err != nil {
		return nico.Instance{}, err
	}
	osByName := map[string]nico.OperatingSystem{}
	for _, osSpec := range rendered {
		osByName[osSpec.Metadata.Name] = nicoOperatingSystemFromRendered(osSpec)
	}
	for _, node := range inventory.Nodes {
		if node.Name != target {
			continue
		}
		osSpec, ok := osByName[node.OSImageRef]
		if !ok {
			return nico.Instance{}, fmt.Errorf("node %q references unknown OS image %q", target, node.OSImageRef)
		}
		if _, err := client.CreateOperatingSystem(ctx, osSpec); err != nil {
			return nico.Instance{}, err
		}
		return client.CreateInstance(ctx, nico.Instance{Name: node.Name, NodeName: node.Name, OSImage: node.OSImageRef, InstanceTypeRef: node.InstanceTypeRef, GPUProfile: node.GPUProfile, JoinProfile: node.JoinProfile, MachineSelector: node.MachineSelector, Labels: node.Labels, LastAction: "add"})
	}
	return nico.Instance{}, fmt.Errorf("node %q not found in inventory %q", target, path)
}

func collectNodeKubernetesEvidence(ctx context.Context) nodestatus.Evidence {
	output, err := runNodesKubectl(ctx, "get", "nodes", "-o", "json")
	if err != nil {
		return nodestatus.Evidence{KubernetesNodes: map[string]nodestatus.KubernetesNodeEvidence{}}
	}
	nodes, err := nodestatus.ParseKubernetesNodeEvidence(output)
	if err != nil {
		return nodestatus.Evidence{KubernetesNodes: map[string]nodestatus.KubernetesNodeEvidence{}}
	}
	return nodestatus.Evidence{KubernetesNodes: nodes}
}

func drainKubernetesNode(ctx context.Context, nodeName string) error {
	if strings.TrimSpace(nodeName) == "" {
		return fmt.Errorf("cannot drain empty Kubernetes node name")
	}
	if _, err := runNodesKubectl(ctx, "cordon", nodeName); err != nil {
		return fmt.Errorf("cordon Kubernetes node %q: %w", nodeName, err)
	}
	if _, err := runNodesKubectl(ctx, "drain", nodeName, "--ignore-daemonsets", "--delete-emptydir-data", "--timeout=10m"); err != nil {
		return fmt.Errorf("drain Kubernetes node %q: %w", nodeName, err)
	}
	return nil
}

type nodeStorageSafety struct {
	HasLocalPV           bool
	HasAIStoreTargetData bool
}

func collectNodeStorageSafety(ctx context.Context, nodeName string) nodeStorageSafety {
	if strings.TrimSpace(nodeName) == "" {
		return nodeStorageSafety{}
	}
	out := nodeStorageSafety{}
	if data, err := runNodesKubectl(ctx, "get", "pv", "-o", "json"); err == nil {
		lower := strings.ToLower(string(data))
		out.HasLocalPV = strings.Contains(string(data), nodeName) && (strings.Contains(lower, "local") || strings.Contains(lower, "nodeaffinity"))
	}
	if data, err := runNodesKubectl(ctx, "get", "pods", "-A", "-l", "app.kubernetes.io/component=target", "-o", "json"); err == nil {
		lower := strings.ToLower(string(data))
		out.HasAIStoreTargetData = strings.Contains(string(data), nodeName) && (strings.Contains(lower, "aistore") || strings.Contains(lower, "ais"))
	}
	return out
}

func resolveNodeTargetStatus(statuses []nodestatus.NodeStatus, target string) (nodestatus.NodeStatus, error) {
	matches := []nodestatus.NodeStatus{}
	for _, status := range statuses {
		if target == status.Name || target == status.MachineID || target == status.InstanceID || target == status.KubernetesNodeName {
			matches = append(matches, status)
		}
	}
	if len(matches) == 0 {
		return nodestatus.NodeStatus{}, fmt.Errorf("node target %q not found in live NICo/Kubernetes status; refusing lifecycle action", target)
	}
	if len(matches) > 1 {
		return nodestatus.NodeStatus{}, fmt.Errorf("node target %q is ambiguous across %d live NICo/Kubernetes statuses; use exact instance ID or machine ID", target, len(matches))
	}
	return matches[0], nil
}

func controlPlaneCounts(statuses []nodestatus.NodeStatus) (ready, total int) {
	for _, status := range statuses {
		if !nodeHasControlPlaneRole(status) {
			continue
		}
		total++
		if status.KubernetesReady {
			ready++
		}
	}
	return ready, total
}

func nodeHasControlPlaneRole(status nodestatus.NodeStatus) bool {
	for _, role := range status.Roles {
		r := strings.ToLower(role)
		if r == "control-plane" || r == "master" || strings.Contains(r, "control-plane") {
			return true
		}
	}
	return false
}

func instanceNodeRow(inst nico.Instance, action string) map[string]string {
	return map[string]string{"name": firstNonEmptyString(inst.NodeName, inst.Name), "id": inst.ID, "backend": "nico", "status": firstNonEmptyString(inst.Status, "created"), "effect": action, "os": firstNonEmptyString(inst.OSImage, inst.OSID)}
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func requireNodeBackend(cfg nico.Config) error {
	switch nodeOpts.Backend {
	case nodeBackendNICO:
		cfg = cfg.WithDefaults()
		if cfg.Mode == nico.ModeLive && strings.TrimSpace(cfg.BaseURL) == "" {
			return fmt.Errorf("NICo config absent: set UBIQUITY_NICO_BASE_URL or use UBIQUITY_NICO_MODE=mock/--dry-run; refusing to claim lifecycle readiness")
		}
		return nil
	case nodeBackendBMO:
		return fmt.Errorf("BMO backend is not the default and is not wired for this NICo lifecycle path")
	case nodeBackendNone:
		return fmt.Errorf("node lifecycle backend disabled")
	default:
		return fmt.Errorf("unsupported node lifecycle backend %q", nodeOpts.Backend)
	}
}

func nicoConfigFromEnv(site string) nico.Config {
	mode := nico.Mode(os.Getenv("UBIQUITY_NICO_MODE"))
	if mode == "" && nodeOpts.DryRun {
		mode = nico.ModeMock
	}
	return nico.Config{BaseURL: os.Getenv("UBIQUITY_NICO_BASE_URL"), Org: os.Getenv("UBIQUITY_NICO_ORG"), SiteName: site, APIName: os.Getenv("UBIQUITY_NICO_API"), Token: os.Getenv("UBIQUITY_NICO_TOKEN"), TokenEnv: "UBIQUITY_NICO_TOKEN", TokenCommand: os.Getenv("UBIQUITY_NICO_TOKEN_COMMAND"), ConfigPath: os.Getenv("UBIQUITY_NICO_CONFIG"), Mode: mode}
}

func renderNodeRows(rows []map[string]string, output string) error {
	if output == "json" {
		b, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(redactNodeOutput(string(b)))
		return nil
	}
	if output != "table" {
		return fmt.Errorf("unsupported output %q", output)
	}
	for _, row := range rows {
		fmt.Printf("%s\t%s\t%s\t%s\n", row["name"], row["backend"], row["status"], row["effect"])
	}
	return nil
}

func redactNodeOutput(s string) string {
	for _, secret := range []string{os.Getenv("UBIQUITY_NICO_TOKEN"), os.Getenv("NICO_TOKEN")} {
		if secret != "" {
			s = strings.ReplaceAll(s, secret, "<redacted>")
		}
	}
	return s
}
