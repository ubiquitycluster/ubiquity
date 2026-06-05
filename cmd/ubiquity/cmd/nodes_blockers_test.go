package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/nico"
)

type fakeNodesNICOClient struct {
	createdOS       []nico.OperatingSystem
	createdInstance []nico.Instance
	deleted         []string
	powerCalls      []nico.PowerRequest
	powerMachineIDs []string
	gotDeadline     bool
	sites           []nico.Site
	machines        []nico.Machine
	instances       []nico.Instance
	tasks           []nico.Task
	gpus            []nico.MachineGPUStats
}

func (f *fakeNodesNICOClient) GetMachine(ctx context.Context, name string) (nico.Machine, error) {
	if _, ok := ctx.Deadline(); ok {
		f.gotDeadline = true
	}
	return nico.Machine{ID: "machine-1", Name: name}, nil
}
func (f *fakeNodesNICOClient) CreateOperatingSystem(ctx context.Context, os nico.OperatingSystem) (nico.OperatingSystem, error) {
	f.createdOS = append(f.createdOS, os)
	os.ID = "os-1"
	return os, nil
}
func (f *fakeNodesNICOClient) CreateInstance(ctx context.Context, inst nico.Instance) (nico.Instance, error) {
	f.createdInstance = append(f.createdInstance, inst)
	inst.ID = "inst-1"
	return inst, nil
}
func (f *fakeNodesNICOClient) DeleteInstance(ctx context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}
func (f *fakeNodesNICOClient) PowerMachine(ctx context.Context, machineID, state, reason string) (nico.Task, error) {
	if _, ok := ctx.Deadline(); ok {
		f.gotDeadline = true
	}
	f.powerMachineIDs = append(f.powerMachineIDs, machineID)
	f.powerCalls = append(f.powerCalls, nico.PowerRequest{State: state, Reason: reason})
	return nico.Task{ID: "task-power-1", Status: nico.TaskRunning, MachineID: machineID, Action: "power " + state}, nil
}
func (f *fakeNodesNICOClient) ListSites(ctx context.Context) ([]nico.Site, error) {
	if _, ok := ctx.Deadline(); ok {
		f.gotDeadline = true
	}
	return f.sites, nil
}
func (f *fakeNodesNICOClient) ListMachines(ctx context.Context) ([]nico.Machine, error) {
	if _, ok := ctx.Deadline(); ok {
		f.gotDeadline = true
	}
	return f.machines, nil
}
func (f *fakeNodesNICOClient) ListInstances(ctx context.Context) ([]nico.Instance, error) {
	return f.instances, nil
}
func (f *fakeNodesNICOClient) ListTasks(ctx context.Context) ([]nico.Task, error) {
	return f.tasks, nil
}
func (f *fakeNodesNICOClient) ListMachineGPUStats(ctx context.Context) ([]nico.MachineGPUStats, error) {
	return f.gpus, nil
}
func (f *fakeNodesNICOClient) GetTask(ctx context.Context, id string) (nico.Task, error) {
	if _, ok := ctx.Deadline(); ok {
		f.gotDeadline = true
	}
	for _, task := range f.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return nico.Task{ID: id, Status: nico.TaskSucceeded}, nil
}

func captureNodesOutput(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), runErr
}

func TestNodesFlagsConfirmStringOSImageAliasAndSafetyFlags(t *testing.T) {
	cmd := findCommand(rootCmd, "nodes")
	for _, name := range []string{"confirm", "os-image", "force", "reason", "drain-confirmed", "storage-ack", "aistore-ack"} {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Fatalf("expected --%s persistent flag", name)
		}
	}
	if cmd.PersistentFlags().Lookup("confirm").Value.Type() != "string" {
		t.Fatalf("--confirm must be a string flag, got %s", cmd.PersistentFlags().Lookup("confirm").Value.Type())
	}
}

func TestNodesExactArgsAndPowerSyntax(t *testing.T) {
	if err := findCommand(nodesCmd, "remove").Args(nodesCmd, []string{}); err == nil {
		t.Fatalf("remove without node should fail exact args")
	}
	power := findCommand(nodesCmd, "power")
	if err := power.Args(power, []string{"node-a", "on"}); err != nil {
		t.Fatalf("power node-a on should validate: %v", err)
	}
	if err := power.Args(power, []string{"node-a", "invalid"}); err == nil || !strings.Contains(err.Error(), "on|off|reset") {
		t.Fatalf("invalid power state error = %v", err)
	}
}

func TestNodesSafetyRequiresExactConfirmation(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: true, Confirm: "wrong", DrainConfirmed: true}
	err := runNodesAction("remove", true)(nodesCmd, []string{"node-a"})
	if err == nil || !strings.Contains(err.Error(), "--confirm must exactly match node name") {
		t.Fatalf("expected exact confirmation safety error, got %v", err)
	}
}

func TestNodesLiveMutatingOperationCallsFakeableClient(t *testing.T) {
	oldOpts := nodeOpts
	oldFactory := newNodesNICOClient
	defer func() { nodeOpts = oldOpts; newNodesNICOClient = oldFactory }()
	fake := &fakeNodesNICOClient{
		machines:  []nico.Machine{{ID: "machine-1", Name: "node-a", Status: "provisioned"}},
		instances: []nico.Instance{{ID: "inst-1", NodeName: "node-a", MachineID: "machine-1", Status: "running"}},
	}
	newNodesNICOClient = func(cfg nico.Config) (nodesNICOClient, error) { return fake, nil }
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: false, Confirm: "node-a", DrainConfirmed: true, Site: "site-a"}
	t.Setenv("UBIQUITY_NICO_MODE", "live")
	t.Setenv("UBIQUITY_NICO_BASE_URL", "https://nico.example")
	t.Setenv("UBIQUITY_NICO_ORG", "acme")
	t.Setenv("UBIQUITY_NICO_TOKEN", "tok")
	if err := runNodesAction("remove", true)(nodesCmd, []string{"node-a"}); err != nil {
		t.Fatalf("remove live: %v", err)
	}
	if len(fake.deleted) != 1 || fake.deleted[0] != "inst-1" {
		t.Fatalf("DeleteInstance calls = %#v", fake.deleted)
	}
}

func TestNodesLivePowerCallsNICoMachinePowerTask(t *testing.T) {
	oldOpts := nodeOpts
	oldFactory := newNodesNICOClient
	defer func() { nodeOpts = oldOpts; newNodesNICOClient = oldFactory }()
	fake := &fakeNodesNICOClient{
		machines:  []nico.Machine{{ID: "machine-1", Name: "node-a", Status: "provisioned"}},
		instances: []nico.Instance{{ID: "inst-1", NodeName: "node-a", MachineID: "machine-1", Status: "running"}},
	}
	newNodesNICOClient = func(cfg nico.Config) (nodesNICOClient, error) { return fake, nil }
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: false, Confirm: "node-a", DrainConfirmed: true, Reason: "maintenance"}
	t.Setenv("UBIQUITY_NICO_MODE", "live")
	t.Setenv("UBIQUITY_NICO_BASE_URL", "https://nico.example")
	t.Setenv("UBIQUITY_NICO_ORG", "acme")
	t.Setenv("UBIQUITY_NICO_TOKEN", "tok")
	out, err := captureNodesOutput(t, func() error { return runNodesAction("power", true)(nodesCmd, []string{"node-a", "off"}) })
	if err != nil {
		t.Fatalf("power live: %v", err)
	}
	if len(fake.powerMachineIDs) != 1 || fake.powerMachineIDs[0] != "machine-1" {
		t.Fatalf("PowerMachine machine IDs = %#v", fake.powerMachineIDs)
	}
	if len(fake.powerCalls) != 1 || fake.powerCalls[0].State != "off" || fake.powerCalls[0].Reason != "maintenance" {
		t.Fatalf("PowerMachine calls = %#v", fake.powerCalls)
	}
	if !strings.Contains(out, "task-power-1") || !strings.Contains(out, "power off") {
		t.Fatalf("power output missing task/effect: %s", out)
	}
}

func TestNodesLiveTaskCallsFakeableClientAndRendersTask(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: false}
	fake := &fakeNodesNICOClient{tasks: []nico.Task{{ID: "task-7", Status: nico.TaskRunning, Action: "reinstall", MachineID: "machine-1"}}}
	out, err := captureNodesOutput(t, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), nodeLiveCommandTimeout)
		defer cancel()
		return runLiveNodesAction(ctx, fake, "task", "task-7", "", []string{"task-7"})
	})
	if err != nil {
		t.Fatalf("task live: %v", err)
	}
	for _, want := range []string{"task-7", "running", "reinstall", "machine-1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("task output %q missing %q", out, want)
		}
	}
	if !fake.gotDeadline {
		t.Fatalf("expected live task context to have deadline")
	}
}

func TestNodesLiveStatusCollectsNICoFiltersTargetAndUsesDeadline(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: false}
	fake := &fakeNodesNICOClient{
		sites:     []nico.Site{{ID: "site-1", Name: "sf01"}},
		machines:  []nico.Machine{{ID: "machine-1", Name: "cn01", SiteID: "site-1", Status: "ready"}, {ID: "machine-2", Name: "cn02", SiteID: "site-1", Status: "ready"}},
		instances: []nico.Instance{{ID: "inst-1", NodeName: "cn01", MachineID: "machine-1", Status: "running"}, {ID: "inst-2", NodeName: "cn02", MachineID: "machine-2", Status: "running"}},
	}
	out, err := captureNodesOutput(t, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), nodeLiveCommandTimeout)
		defer cancel()
		return runLiveNodesAction(ctx, fake, "status", "cn02", "", []string{"cn02"})
	})
	if err != nil {
		t.Fatalf("status live: %v", err)
	}
	if strings.Contains(out, "dry-run") {
		t.Fatalf("live status returned dry-run output: %s", out)
	}
	if !strings.Contains(out, "cn02") || strings.Contains(out, "cn01") {
		t.Fatalf("status output was not filtered to target: %s", out)
	}
	if !fake.gotDeadline {
		t.Fatalf("expected live status context to have deadline")
	}
}

func TestNodesLiveCreateActionsRenderCreatedPayloads(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: false, OS: "rocky"}
	for _, tc := range []struct {
		action, target string
		args           []string
		want           string
	}{
		{action: "add", target: "cn01", args: []string{"cn01"}, want: "inst-1"},
		{action: "reinstall", target: "cn01", args: []string{"cn01"}, want: "reinstall"},
		{action: "os apply", target: "rocky", args: []string{"rocky"}, want: "os-1"},
	} {
		fake := &fakeNodesNICOClient{}
		out, err := captureNodesOutput(t, func() error {
			ctx, cancel := context.WithTimeout(context.Background(), nodeLiveCommandTimeout)
			defer cancel()
			return runLiveNodesAction(ctx, fake, tc.action, tc.target, "", tc.args)
		})
		if err != nil {
			t.Fatalf("%s live: %v", tc.action, err)
		}
		if !strings.Contains(out, tc.want) {
			t.Fatalf("%s output %q missing created payload marker %q", tc.action, out, tc.want)
		}
	}
}
