package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/nico"
	"github.com/ubiquitycluster/ubiquity/pkg/nodestatus"
)

func TestCollectNodeKubernetesEvidenceUsesKubectlNodesJSON(t *testing.T) {
	old := runNodesKubectl
	defer func() { runNodesKubectl = old }()
	runNodesKubectl = func(ctx context.Context, args ...string) ([]byte, error) {
		if strings.Join(args, " ") != "get nodes -o json" {
			t.Fatalf("unexpected kubectl args: %#v", args)
		}
		return []byte(`{"items":[{"metadata":{"name":"cn01","labels":{"node-role.kubernetes.io/worker":""}},"status":{"allocatable":{"nvidia.com/gpu":"8","nvidia.com/rdma":"4"},"conditions":[{"type":"Ready","status":"True"}]}}]}`), nil
	}
	got := collectNodeKubernetesEvidence(context.Background()).KubernetesNodes["cn01"]
	if !got.Ready || got.GPUs != 8 || got.RDMAResources != 4 || !got.NVIDIAReady || len(got.Roles) != 1 || got.Roles[0] != "worker" {
		t.Fatalf("live Kubernetes evidence not wired: %+v", got)
	}
}

func TestCollectNodeKubernetesEvidenceFailsClosedWhenKubectlUnavailable(t *testing.T) {
	old := runNodesKubectl
	defer func() { runNodesKubectl = old }()
	runNodesKubectl = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, errors.New("kubectl: not found")
	}
	got := collectNodeKubernetesEvidence(context.Background())
	if len(got.KubernetesNodes) != 0 {
		t.Fatalf("kubectl failure must not create positive Kubernetes claims: %#v", got.KubernetesNodes)
	}
}

func TestResolveNodeTargetStatusFailsClosedForAmbiguousOrMissingTargets(t *testing.T) {
	statuses := []nodestatus.NodeStatus{{Name: "dup", MachineID: "m1", InstanceID: "i1"}, {Name: "dup", MachineID: "m2", InstanceID: "i2"}}
	if _, err := resolveNodeTargetStatus(statuses, "dup"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous failure, got %v", err)
	}
	if _, err := resolveNodeTargetStatus(statuses, "missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found failure, got %v", err)
	}
}

func TestRunLiveRemoveResolvesNodeNameToInstanceID(t *testing.T) {
	old := nodeOpts
	defer func() { nodeOpts = old }()
	nodeOpts = nodeCommandOptions{Backend: nodeBackendNICO, Output: "json", DryRun: false}
	fake := &fakeNodesNICOClient{
		machines:  []nico.Machine{{ID: "machine-1", Name: "cn01", Status: "provisioned"}},
		instances: []nico.Instance{{ID: "inst-1", NodeName: "cn01", MachineID: "machine-1", Status: "running"}},
	}
	_, err := captureNodesOutput(t, func() error {
		return runLiveNodesAction(context.Background(), fake, "remove", "cn01", "", []string{"cn01"})
	})
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(fake.deleted) != 1 || fake.deleted[0] != "inst-1" {
		t.Fatalf("DeleteInstance should receive resolved instance ID, got %#v", fake.deleted)
	}
}

func TestControlPlaneCountsUseLiveRolesAndReadyEvidence(t *testing.T) {
	ready, total := controlPlaneCounts([]nodestatus.NodeStatus{
		{Name: "cp01", KubernetesReady: true, Roles: []string{"control-plane"}},
		{Name: "cp02", KubernetesReady: false, Roles: []string{"master"}},
		{Name: "cn01", KubernetesReady: true, Roles: []string{"worker"}},
	})
	if ready != 1 || total != 2 {
		t.Fatalf("control-plane counts = %d/%d, want 1/2", ready, total)
	}
}
