package aiplatform

import (
	"strings"
	"testing"
)

func TestAIStoreEvaluationDocumentsLonghornReplacementBoundary(t *testing.T) {
	doc := mustRead(t, "../../docs/reference/nvidia-ai-platform/aistore-evaluation.md")
	for _, required := range []string{
		"NVIDIA/aistore",
		"NVIDIA/ais-k8s",
		"Longhorn",
		"replaces Longhorn for AI dataset/cache paths",
		"not a generic PVC replacement",
		"object storage",
		"S3-compatible",
		"PyTorch integration",
		"local PV",
		"ReadWriteOnce",
		"XFS recommended",
		"AIStore readiness evidence",
		"aistore-target-storage-proven",
		"aistore-bucket-smoke-test-passed",
		"aistore-gpu-artifact-read-passed",
		"aistore-metrics-proven",
		"reported separately from core GPU/NIM/KAI readiness",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("AIStore evaluation doc must include %q", required)
		}
	}
}
