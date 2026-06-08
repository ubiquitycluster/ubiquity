package cloud

import (
	"strings"
	"testing"
)

func TestCloudServiceSmokeScriptRecordsRequiredMarkers(t *testing.T) {
	script := mustRead(t, "../../test/e2e/cloud-service-smoke-tests.sh")
	for _, required := range RequiredCloudSmokeTests() {
		if !strings.Contains(script, required) {
			t.Fatalf("cloud service smoke script missing marker %q", required)
		}
	}
	for _, required := range []string{"psql", "redis-cli", "kafka-console-producer", "aws s3 cp", "kubectl get restore", "requiredSmokeTests", "smokeTests"} {
		if !strings.Contains(script, required) {
			t.Fatalf("cloud service smoke script missing command/evidence marker %q", required)
		}
	}
}
