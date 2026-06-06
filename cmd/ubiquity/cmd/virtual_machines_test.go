package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/virtualization"
)

func TestVirtualMachinesRenderProducesKubeVirtManifest(t *testing.T) {
	render := findCommand(virtualMachinesCmd, "render")
	if render == nil {
		t.Fatal("expected virtual-machines render subcommand")
	}
	oldOpts := vmOpts
	defer func() { vmOpts = oldOpts }()
	vmOpts = virtualization.VMRequest{
		Name: "trainer-ubuntu", Namespace: "tenant-a", OS: "ubuntu-24.04",
		Network: virtualization.NetworkRequest{Isolation: virtualization.NetworkIsolationMultus, Name: "tenant-a-rdma"},
		GPU:     virtualization.GPURequest{ResourceName: "nvidia.com/GA100_A100_PCIE_40GB", Count: 1},
	}
	out := captureStdout(t, func() {
		if err := render.RunE(render, []string{}); err != nil {
			t.Fatalf("virtual-machines render failed: %v", err)
		}
	})
	for _, required := range []string{"kind: VirtualMachine", "kind: DataVolume", "kind: NetworkAttachmentDefinition", "tenant-a-rdma", "nvidia.com/GA100_A100_PCIE_40GB", "ubuntu-24.04"} {
		if !strings.Contains(out, required) {
			t.Fatalf("render output missing %q:\n%s", required, out)
		}
	}
}

func TestVirtualMachinesApplyUsesServerDryRunByDefault(t *testing.T) {
	oldRunner := runVirtualMachinesKubectl
	oldOpts := vmOpts
	defer func() { runVirtualMachinesKubectl = oldRunner; vmOpts = oldOpts }()
	var gotArgs []string
	var gotStdin string
	runVirtualMachinesKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		gotArgs = append([]string{}, args...)
		gotStdin = string(stdin)
		return []byte("dry-run ok\n"), nil
	}
	vmOpts = virtualization.VMRequest{Name: "rocky-gpu", Namespace: "virtual-machines", OS: "rocky-9"}
	apply := findCommand(virtualMachinesCmd, "apply")
	if apply == nil {
		t.Fatal("expected virtual-machines apply subcommand")
	}
	out := captureStdout(t, func() {
		if err := apply.RunE(apply, []string{}); err != nil {
			t.Fatalf("virtual-machines apply failed: %v", err)
		}
	})
	if fmt.Sprint(gotArgs) != "[apply --dry-run=server -f -]" {
		t.Fatalf("apply should default to server dry-run, got args %v", gotArgs)
	}
	if !strings.Contains(gotStdin, "kind: VirtualMachine") || !strings.Contains(gotStdin, "rocky-9") {
		t.Fatalf("apply stdin missing VM manifest: %s", gotStdin)
	}
	if !strings.Contains(out, "dry-run ok") {
		t.Fatalf("expected kubectl output, got %q", out)
	}
}
