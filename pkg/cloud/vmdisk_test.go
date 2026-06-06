package cloud

import (
	"strings"
	"testing"
)

func TestRenderVMDiskSupportsBlankImportAndCloneSources(t *testing.T) {
	cases := []struct {
		name     string
		req      VMDiskRequest
		required []string
	}{
		{
			name:     "blank",
			req:      VMDiskRequest{Name: "data-disk", Namespace: "tenant-a", Size: "100Gi", StorageClass: "fast-nvme", Source: VMDiskSource{Type: VMDiskSourceBlank}},
			required: []string{"kind: PersistentVolumeClaim", "name: data-disk", "storage: 100Gi", "storageClassName: fast-nvme", "ubiquity.ai/disk-source: blank"},
		},
		{
			name:     "import",
			req:      VMDiskRequest{Name: "ubuntu-base", Namespace: "tenant-a", Size: "40Gi", Source: VMDiskSource{Type: VMDiskSourceHTTP, URL: "https://images.example/ubuntu.qcow2"}},
			required: []string{"kind: DataVolume", "http:", "url: https://images.example/ubuntu.qcow2", "ubiquity.ai/disk-source: http"},
		},
		{
			name:     "clone",
			req:      VMDiskRequest{Name: "clone-disk", Namespace: "tenant-a", Size: "40Gi", Source: VMDiskSource{Type: VMDiskSourcePVC, PVCName: "ubuntu-base", PVCNamespace: "golden-images"}},
			required: []string{"kind: DataVolume", "pvc:", "name: ubuntu-base", "namespace: golden-images", "ubiquity.ai/disk-source: pvc"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manifest, err := RenderVMDisk(tc.req)
			if err != nil {
				t.Fatalf("RenderVMDisk returned error: %v", err)
			}
			for _, required := range tc.required {
				if !strings.Contains(manifest, required) {
					t.Fatalf("manifest missing %q:\n%s", required, manifest)
				}
			}
		})
	}
}

func TestRenderVMDiskFailsClosedForMissingSourceData(t *testing.T) {
	_, err := RenderVMDisk(VMDiskRequest{Name: "bad", Source: VMDiskSource{Type: VMDiskSourceHTTP}})
	if err == nil || !strings.Contains(err.Error(), "http source URL") {
		t.Fatalf("expected missing http URL error, got %v", err)
	}
	_, err = RenderVMDisk(VMDiskRequest{Name: "bad", Source: VMDiskSource{Type: VMDiskSourcePVC}})
	if err == nil || !strings.Contains(err.Error(), "source PVC name") {
		t.Fatalf("expected missing PVC error, got %v", err)
	}
}
