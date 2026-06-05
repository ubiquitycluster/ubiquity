package aiplatform

import "testing"

func TestParseAllocatableResourceByNodeReadsRDMAResources(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"name":"cpu-1"},"status":{"allocatable":{"cpu":"8"}}},
			{"metadata":{"name":"rdma-1"},"status":{"allocatable":{"nvidia.com/rdma":"4"}}},
			{"metadata":{"name":"rdma-zero"},"status":{"allocatable":{"nvidia.com/rdma":"0"}}}
		]
	}`)

	parsed, err := ParseAllocatableResourceByNode(input, "nvidia.com/rdma")
	if err != nil {
		t.Fatalf("ParseAllocatableResourceByNode returned unexpected error: %v", err)
	}
	if len(parsed) != 1 || parsed["rdma-1"] != 4 {
		t.Fatalf("expected only rdma-1 with 4 RDMA resources, got %#v", parsed)
	}
}

func TestParseNetworkAttachmentsReadsNamespacedNames(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"namespace":"default","name":"rdma-ipoib"}},
			{"metadata":{"namespace":"training","name":"gpu-rdma"}}
		]
	}`)

	parsed, err := ParseNetworkAttachments(input)
	if err != nil {
		t.Fatalf("ParseNetworkAttachments returned unexpected error: %v", err)
	}
	want := []string{"default/rdma-ipoib", "training/gpu-rdma"}
	if len(parsed) != len(want) {
		t.Fatalf("expected %d attachments, got %#v", len(want), parsed)
	}
	for i := range want {
		if parsed[i] != want[i] {
			t.Fatalf("expected attachment %d to be %q, got %#v", i, want[i], parsed)
		}
	}
}
