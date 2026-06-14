package cloud

import (
	"strings"
	"testing"
)

func TestRenderBackupPolicyCreatesK8upScheduleSnapshotClassAndResourcePreset(t *testing.T) {
	manifest, err := RenderBackupPolicy(BackupPolicyRequest{
		Name:                 "tenant-a-daily",
		Namespace:            "tenant-a",
		Schedule:             "0 2 * * *",
		Retention:            "30d",
		RepositorySecretName: "tenant-a-backup-repo",
		SnapshotClass:        "longhorn-snapshots",
		PresetName:           "gpu-medium",
		PresetCPU:            "16",
		PresetMemory:         "128Gi",
		PresetGPU:            "1",
	})
	if err != nil {
		t.Fatalf("RenderBackupPolicy returned error: %v", err)
	}
	for _, required := range []string{
		"kind: Schedule", "apiVersion: k8up.io/v1", "backup:", "schedule: \"0 2 * * *\"", "keepDaily: 30",
		"repositorySecretRef:", "name: tenant-a-backup-repo",
		"kind: VolumeSnapshotClass", "driver: driver.longhorn.io", "deletionPolicy: Retain",
		"kind: ConfigMap", "name: gpu-medium-resource-preset", "cpu: \"16\"", "memory: 128Gi", "nvidia.com/gpu: \"1\"",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderBackupPolicyFailsClosedForUnsafeScheduleOrRetention(t *testing.T) {
	_, err := RenderBackupPolicy(BackupPolicyRequest{Name: "bad", Schedule: "@hourly", Retention: "30d", RepositorySecretName: "repo"})
	if err == nil || !strings.Contains(err.Error(), "five-field cron") {
		t.Fatalf("expected schedule validation error, got %v", err)
	}
	_, err = RenderBackupPolicy(BackupPolicyRequest{Name: "bad", Schedule: "0 2 * * *", Retention: "forever", RepositorySecretName: "repo"})
	if err == nil || !strings.Contains(err.Error(), "retention") {
		t.Fatalf("expected retention validation error, got %v", err)
	}
}
