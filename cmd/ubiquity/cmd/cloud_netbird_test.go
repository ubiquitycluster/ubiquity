package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestCloudRenderNetBirdOverlayProducesFleetBundle(t *testing.T) {
	old := cloudOpts
	defer func() { cloudOpts = old }()
	cloudOpts.NetBird.ManagementCluster = "ubiquity-management"
	cloudOpts.NetBird.RegionalCluster = "spanish-fork-gpu-01"
	cloudOpts.NetBird.Region = "us-west"
	cloudOpts.NetBird.Site = "spanish-fork"
	cloudOpts.NetBird.StorageProvider = "vast"
	cloudOpts.NetBird.GPUClass = "h100"
	cmd := findCommand(cloudCmd, "render")
	if cmd == nil {
		t.Fatal("expected cloud render command")
	}
	out := captureStdout(t, func() {
		if err := cmd.RunE(cmd, []string{"netbird-overlay"}); err != nil {
			t.Fatalf("cloud render netbird-overlay failed: %v", err)
		}
	})
	for _, required := range []string{"kind: ApplicationSet", "ubiquity-regional-ai-platform", "spanish-fork-gpu-01", "ubiquity.io/region: us-west", "Geo DNS", "global load balancer"} {
		if !strings.Contains(out, required) {
			t.Fatalf("output missing %q:\n%s", required, out)
		}
	}
}

func TestCloudApplyNetBirdOverlayUsesServerSideDryRunByDefault(t *testing.T) {
	oldOpts := cloudOpts
	oldRunner := runCloudKubectl
	defer func() { cloudOpts = oldOpts; runCloudKubectl = oldRunner }()
	cloudOpts.NetBird.ManagementCluster = "ubiquity-management"
	cloudOpts.NetBird.RegionalCluster = "spanish-fork-gpu-01"
	cloudOpts.NetBird.Region = "us-west"
	cloudOpts.NetBird.Site = "spanish-fork"
	var gotArgs []string
	var gotStdin string
	runCloudKubectl = func(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
		gotArgs = append([]string{}, args...)
		gotStdin = string(stdin)
		return []byte("netbird overlay validated\n"), nil
	}
	cmd := findCommand(cloudCmd, "apply")
	if cmd == nil {
		t.Fatal("expected cloud apply command")
	}
	out := captureStdout(t, func() {
		if err := cmd.RunE(cmd, []string{"netbird-overlay"}); err != nil {
			t.Fatalf("cloud apply netbird-overlay failed: %v", err)
		}
	})
	if fmt.Sprint(gotArgs) != "[apply --dry-run=server -f -]" {
		t.Fatalf("unexpected kubectl args: %v", gotArgs)
	}
	for _, required := range []string{"netbird overlay validated", "kind: ApplicationSet", "PLACEHOLDER_BEARER_TOKEN_FROM_REMOTE_CLUSTER_SERVICE_ACCOUNT"} {
		content := out + gotStdin
		if !strings.Contains(content, required) {
			t.Fatalf("apply path missing %q; output=%q stdin=%q", required, out, gotStdin)
		}
	}
}
