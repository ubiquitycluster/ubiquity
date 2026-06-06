package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

func TestCloudRenderPrerequisitesProducesCRDContract(t *testing.T) {
	old := cloudOpts
	defer func() { cloudOpts = old }()
	cloudOpts.VMDisk.Name = "cloud-prereqs"
	manifest, err := renderCloudResource("prerequisites")
	if err != nil {
		t.Fatalf("render prerequisites returned error: %v", err)
	}
	for _, required := range []string{"kind: ConfigMap", "datavolumes.cdi.kubevirt.io", "serverSideDryRunRequired"} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestCloudRenderOperatorBundlesProducesInstallPlan(t *testing.T) {
	manifest, err := renderCloudResource("operator-bundles")
	if err != nil {
		t.Fatalf("render operator-bundles returned error: %v", err)
	}
	for _, required := range []string{"kind: ConfigMap", "kubevirt-cdi", "airgapArtifact"} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestCloudReadinessEvaluatesEvidenceFile(t *testing.T) {
	old := cloudOpts
	defer func() { cloudOpts = old }()
	path := filepath.Join(t.TempDir(), "readiness.json")
	if err := os.WriteFile(path, []byte(`{
  "requiredCRDs": ["datavolumes.cdi.kubevirt.io"],
  "presentCRDs": [],
  "smokeTests": {"restore-drill": false}
}`), 0o600); err != nil {
		t.Fatalf("write evidence: %v", err)
	}
	cloudOpts.ReadinessFile = path
	cmd := findCommand(cloudCmd, "readiness")
	if cmd == nil {
		t.Fatal("expected cloud readiness command")
	}
	out := captureStdout(t, func() {
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("cloud readiness failed: %v", err)
		}
	})
	for _, required := range []string{"ready: false", "missing CRD datavolumes.cdi.kubevirt.io", "smoke test restore-drill did not pass"} {
		if !strings.Contains(out, required) {
			t.Fatalf("readiness output missing %q:\n%s", required, out)
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
