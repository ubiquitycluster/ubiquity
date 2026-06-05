package aiplatform

import "testing"

func TestParseGPUAllocatableByNodeFailsClosedForEmptyNodeList(t *testing.T) {
	parsed, err := ParseGPUAllocatableByNode([]byte(`{"items":[]}`))
	if err != nil {
		t.Fatalf("ParseGPUAllocatableByNode returned unexpected error: %v", err)
	}
	if len(parsed) != 0 {
		t.Fatalf("expected no GPU allocatable evidence for empty node list, got %#v", parsed)
	}
}

func TestParseGPUAllocatableByNodeReadsOnlyPositiveNvidiaGPUCapacity(t *testing.T) {
	input := []byte(`{
		"items": [
			{"metadata":{"name":"cpu-1"},"status":{"allocatable":{"cpu":"8"}}},
			{"metadata":{"name":"gpu-1"},"status":{"allocatable":{"nvidia.com/gpu":"8"}}},
			{"metadata":{"name":"gpu-zero"},"status":{"allocatable":{"nvidia.com/gpu":"0"}}}
		]
	}`)

	parsed, err := ParseGPUAllocatableByNode(input)
	if err != nil {
		t.Fatalf("ParseGPUAllocatableByNode returned unexpected error: %v", err)
	}
	if len(parsed) != 1 || parsed["gpu-1"] != 8 {
		t.Fatalf("expected only gpu-1 with 8 GPUs, got %#v", parsed)
	}
}

func TestParseGPUAllocatableByNodeMalformedJSONFailsClosed(t *testing.T) {
	parsed, err := ParseGPUAllocatableByNode([]byte(`not-json`))
	if err == nil {
		t.Fatal("expected malformed JSON to return an error")
	}
	if len(parsed) != 0 {
		t.Fatalf("expected malformed JSON to produce no GPU evidence, got %#v", parsed)
	}
}
