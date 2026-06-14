package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/cloud"
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

func TestCloudCollectReadinessOutputsEvidenceJSON(t *testing.T) {
	oldRunner := runCloudKubectl
	oldOpts := cloudOpts
	defer func() { runCloudKubectl = oldRunner; cloudOpts = oldOpts }()
	cloudOpts.ReadinessResources = []string{"datavolumes.cdi.kubevirt.io"}
	calls := []string{}
	runCloudKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		calls = append(calls, fmt.Sprint(args))
		switch fmt.Sprint(args) {
		case "[get crd -o json]":
			return []byte(`{"items":[{"metadata":{"name":"datavolumes.cdi.kubevirt.io"}}]}`), nil
		case "[get datavolumes.cdi.kubevirt.io -A -o json]":
			return []byte(`{"items":[{"kind":"DataVolume","metadata":{"namespace":"tenant-a","name":"ubuntu-root"},"status":{"conditions":[{"type":"Ready","status":"True","reason":"ImportSucceeded"}]}}]}`), nil
		case "[get configmaps -A -l ubiquity.ai/cloud-smoke=true -o json]":
			return []byte(`{"items":[{"metadata":{"name":"tenant-cluster-kubeconfig-present","labels":{"ubiquity.ai/cloud-smoke":"true"}},"data":{"passed":"true"}}]}`), nil
		default:
			return nil, fmt.Errorf("unexpected kubectl args %v", args)
		}
	}
	collect := findCommand(cloudCmd, "collect-readiness")
	if collect == nil {
		t.Fatal("expected cloud collect-readiness subcommand")
	}
	out := captureStdout(t, func() {
		if err := collect.RunE(collect, nil); err != nil {
			t.Fatalf("collect-readiness failed: %v", err)
		}
	})
	for _, required := range []string{"requiredCRDs", "requiredSmokeTests", "restore-drill-readable", "tenant-cluster-kubeconfig-present", "presentCRDs", "datavolumes.cdi.kubevirt.io", "DataVolume", "ubuntu-root", "ImportSucceeded"} {
		if !strings.Contains(out, required) {
			t.Fatalf("collector output missing %q:\n%s", required, out)
		}
	}
	if len(calls) != 3 {
		t.Fatalf("expected CRD, resource, and smoke-marker kubectl calls, got %v", calls)
	}
}

func TestParseCloudSmokeMarkersFromConfigMaps(t *testing.T) {
	markers, err := parseCloudSmokeMarkers([]byte(`{"items":[{"metadata":{"name":"tenant-cluster-kubeconfig-present","labels":{"ubiquity.ai/cloud-smoke":"true"}},"data":{"passed":"true"}},{"metadata":{"name":"tenant-cluster-api-reachable","labels":{"ubiquity.ai/cloud-smoke":"true"}},"data":{"passed":"true"}},{"metadata":{"name":"tenant-cluster-nodes-ready","labels":{"ubiquity.ai/cloud-smoke":"true"}},"data":{"passed":"false"}}]}`))
	if err != nil {
		t.Fatalf("parse smoke markers: %v", err)
	}
	if !markers["tenant-cluster-kubeconfig-present"] || !markers["tenant-cluster-api-reachable"] || markers["tenant-cluster-nodes-ready"] {
		t.Fatalf("unexpected marker map: %#v", markers)
	}
}

func TestParseKubectlResourceEvidenceConvertsPVCPhaseToBoundCondition(t *testing.T) {
	resources, err := parseKubectlResourceEvidence([]byte(`{"items":[{"kind":"PersistentVolumeClaim","metadata":{"namespace":"tenant-a","name":"ubuntu-dev-root"},"status":{"phase":"Bound"}}]}`))
	if err != nil {
		t.Fatalf("parse PVC evidence: %v", err)
	}
	if len(resources) != 1 || !cloudTestHasCondition(resources[0].Conditions, "Bound", "True") {
		t.Fatalf("expected PVC Bound=True condition, got %#v", resources)
	}
}

func cloudTestHasCondition(conditions []cloud.CloudCondition, typ, status string) bool {
	for _, condition := range conditions {
		if condition.Type == typ && condition.Status == status {
			return true
		}
	}
	return false
}

func TestCloudRenderBackupRestoreDrillProducesIsolatedRestore(t *testing.T) {
	manifest, err := renderCloudResource("restore-drill")
	if err != nil {
		t.Fatalf("render restore-drill returned error: %v", err)
	}
	for _, required := range []string{"kind: Restore", "restoreMethod: folder", "restore-drill", "readinessBoundary"} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestDefaultCloudReadinessResourcesIncludeExpandedManagedServices(t *testing.T) {
	resources := strings.Join(defaultCloudReadinessResources(), "\n")
	for _, required := range []string{"mariadbs.k8s.mariadb.com", "opensearchclusters.opensearch.opster.io", "tcproutes.gateway.networking.k8s.io"} {
		if !strings.Contains(resources, required) {
			t.Fatalf("default readiness resources missing %q in %s", required, resources)
		}
	}
}

func TestDefaultCloudReadinessResourcesIncludeRestoreDrillEvidence(t *testing.T) {
	resources := strings.Join(defaultCloudReadinessResources(), "\n")
	for _, required := range []string{"restores.k8up.io", "schedules.k8up.io"} {
		if !strings.Contains(resources, required) {
			t.Fatalf("default readiness resources missing %q in %s", required, resources)
		}
	}
}

func TestDefaultCloudReadinessResourcesIncludeKubeVirtBootEvidence(t *testing.T) {
	resources := strings.Join(defaultCloudReadinessResources(), "\n")
	for _, required := range []string{"datavolumes.cdi.kubevirt.io", "persistentvolumeclaims", "virtualmachines.kubevirt.io", "virtualmachineinstances.kubevirt.io"} {
		if !strings.Contains(resources, required) {
			t.Fatalf("default readiness resources missing %q in %s", required, resources)
		}
	}
}

func TestCloudProductionAuditChecklistCLI(t *testing.T) {
	audit := findCommand(cloudCmd, "audit-checklist")
	if audit == nil {
		t.Fatal("expected cloud audit-checklist subcommand")
	}
	out := captureStdout(t, func() {
		if err := audit.RunE(audit, nil); err != nil {
			t.Fatalf("audit-checklist failed: %v", err)
		}
	})
	for _, required := range []string{"production readiness audit checklist", "collect-readiness", "operator install-plan", "ready: true"} {
		if !strings.Contains(out, required) {
			t.Fatalf("audit checklist output missing %q:\n%s", required, out)
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
