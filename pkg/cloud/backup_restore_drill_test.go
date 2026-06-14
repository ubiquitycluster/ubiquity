package cloud

import (
	"strings"
	"testing"
)

func TestRenderBackupRestoreDrillCreatesIsolatedK8upRestoreAndChecklist(t *testing.T) {
	manifest, err := RenderBackupRestoreDrill(BackupRestoreDrillRequest{
		Name:                 "tenant-a-daily",
		Namespace:            "tenant-a",
		DrillNamespace:       "tenant-a-restore-drill",
		RepositorySecretName: "tenant-a-backup-repo",
		Snapshot:             "latest",
	})
	if err != nil {
		t.Fatalf("RenderBackupRestoreDrill returned error: %v", err)
	}
	for _, required := range []string{
		"kind: Namespace",
		"name: tenant-a-restore-drill",
		"kind: Restore",
		"apiVersion: k8up.io/v1",
		"restoreMethod: folder",
		"snapshot: latest",
		"repositorySecretRef:",
		"name: tenant-a-backup-repo",
		"ubiquity.ai/restore-drill: tenant-a-daily",
		"kind: ConfigMap",
		"readinessBoundary: restore-object-rendered-not-restore-proof",
	} {
		if !strings.Contains(manifest, required) {
			t.Fatalf("manifest missing %q:\n%s", required, manifest)
		}
	}
}

func TestRenderBackupRestoreDrillRejectsProductionNamespace(t *testing.T) {
	_, err := RenderBackupRestoreDrill(BackupRestoreDrillRequest{Name: "tenant-a-daily", Namespace: "tenant-a", DrillNamespace: "tenant-a"})
	if err == nil || !strings.Contains(err.Error(), "isolated") {
		t.Fatalf("expected isolated namespace validation error, got %v", err)
	}
}
