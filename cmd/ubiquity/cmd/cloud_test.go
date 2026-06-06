package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestCloudRenderVMDiskProducesStandaloneDisk(t *testing.T) {
	old := cloudOpts
	defer func() { cloudOpts = old }()
	cloudOpts.VMDisk.Name = "ubuntu-base"
	cloudOpts.VMDisk.Namespace = "tenant-a"
	cloudOpts.VMDisk.Source.Type = "http"
	cloudOpts.VMDisk.Source.URL = "https://images.example/ubuntu.qcow2"
	cmd := findCommand(cloudCmd, "render")
	if cmd == nil {
		t.Fatal("expected cloud render command")
	}
	out := captureStdout(t, func() {
		if err := cmd.RunE(cmd, []string{"vm-disk"}); err != nil {
			t.Fatalf("cloud render vm-disk failed: %v", err)
		}
	})
	for _, required := range []string{"kind: DataVolume", "ubuntu-base", "https://images.example/ubuntu.qcow2"} {
		if !strings.Contains(out, required) {
			t.Fatalf("output missing %q:\n%s", required, out)
		}
	}
}

func TestCloudApplyDefaultsToServerDryRun(t *testing.T) {
	oldRunner := runCloudKubectl
	old := cloudOpts
	defer func() { runCloudKubectl = oldRunner; cloudOpts = old }()
	cloudOpts.VMDisk.Name = "data-disk"
	var gotArgs []string
	var gotStdin string
	runCloudKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		gotArgs = append([]string{}, args...)
		gotStdin = string(stdin)
		return []byte("dry-run ok\n"), nil
	}
	cmd := findCommand(cloudCmd, "apply")
	if cmd == nil {
		t.Fatal("expected cloud apply command")
	}
	out := captureStdout(t, func() {
		if err := cmd.RunE(cmd, []string{"vm-disk"}); err != nil {
			t.Fatalf("cloud apply vm-disk failed: %v", err)
		}
	})
	if fmt.Sprint(gotArgs) != "[apply --dry-run=server -f -]" {
		t.Fatalf("apply should default to server dry-run, got %v", gotArgs)
	}
	if !strings.Contains(gotStdin, "kind: PersistentVolumeClaim") {
		t.Fatalf("expected disk manifest on stdin: %s", gotStdin)
	}
	if !strings.Contains(out, "dry-run ok") {
		t.Fatalf("expected kubectl output, got %q", out)
	}
}
