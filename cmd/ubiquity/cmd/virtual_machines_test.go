package cmd

import (
	"context"
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

func TestVirtualMachinesRenderImageCatalogProducesProfiles(t *testing.T) {
	catalog := findCommand(virtualMachinesCmd, "image-catalog")
	if catalog == nil {
		t.Fatal("expected virtual-machines image-catalog subcommand")
	}
	out := captureStdout(t, func() {
		if err := catalog.RunE(catalog, []string{}); err != nil {
			t.Fatalf("virtual-machines image-catalog failed: %v", err)
		}
	})
	for _, required := range []string{"kind: ConfigMap", "ubuntu-24.04", "readinessBoundary"} {
		if !strings.Contains(out, required) {
			t.Fatalf("image catalog output missing %q:\n%s", required, out)
		}
	}
}

func TestVirtualMachinesApplyUsesServerSideDryRun(t *testing.T) {
	oldRunner := runVirtualMachinesKubectl
	oldOpts := vmOpts
	defer func() { runVirtualMachinesKubectl = oldRunner; vmOpts = oldOpts }()
	var capturedArgs []string
	var capturedStdin string
	runVirtualMachinesKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		capturedArgs = append([]string{}, args...)
		capturedStdin = string(stdin)
		return []byte("server dry-run ok\n"), nil
	}
	vmOpts = virtualization.VMRequest{Name: "rocky-gpu", Namespace: "virtual-machines", OS: "rocky-9"}
	apply := findCommand(virtualMachinesCmd, "apply")
	if apply == nil {
		t.Fatal("expected virtual-machines apply subcommand")
	}
	output := captureOutput(func() {
		if err := apply.RunE(apply, []string{}); err != nil {
			t.Fatalf("virtual-machines apply failed: %v", err)
		}
	})
	assertContains(t, output, "server dry-run ok")
	if strings.Join(capturedArgs, " ") != "apply --dry-run=server -f -" {
		t.Fatalf("expected server-side dry-run apply, got %v", capturedArgs)
	}
	assertContains(t, capturedStdin, "kind: VirtualMachine")
}

func TestVirtualMachinesReadinessCollectsCDIPVCVMAndGuestEvidence(t *testing.T) {
	oldRunner := runVirtualMachinesKubectl
	oldOpts := vmOpts
	defer func() { runVirtualMachinesKubectl = oldRunner; vmOpts = oldOpts }()
	vmOpts = virtualization.VMRequest{Name: "ubuntu-dev", Namespace: "tenant-a", OS: "ubuntu-24.04"}
	runVirtualMachinesKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "-n tenant-a get datavolume ubuntu-dev-root -o jsonpath={.status.conditions[?(@.type==\"Ready\")].status}":
			return []byte("True"), nil
		case "-n tenant-a get pvc ubuntu-dev-root -o jsonpath={.status.phase}":
			return []byte("Bound"), nil
		case "-n tenant-a get virtualmachine ubuntu-dev -o jsonpath={.status.conditions[?(@.type==\"Ready\")].status}":
			return []byte("True"), nil
		case "-n tenant-a get virtualmachineinstance ubuntu-dev -o jsonpath={.status.phase}":
			return []byte("Running"), nil
		case "-n tenant-a get configmap ubuntu-dev-guest-health-passed":
			return []byte("ok"), nil
		default:
			t.Fatalf("unexpected kubectl args: %v", args)
		}
		return nil, nil
	}
	readiness := findCommand(virtualMachinesCmd, "readiness")
	if readiness == nil {
		t.Fatal("expected virtual-machines readiness subcommand")
	}
	output := captureOutput(func() {
		if err := readiness.RunE(readiness, []string{}); err != nil {
			t.Fatalf("virtual-machines readiness failed: %v", err)
		}
	})
	assertContains(t, output, "ready: true")
	assertContains(t, output, "guest health evidence present")
}
