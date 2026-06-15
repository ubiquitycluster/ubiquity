package aiplatform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFrontendSnapshotCoversNCPReferenceLayers(t *testing.T) {
	profile, err := GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production): %v", err)
	}
	snapshot := BuildFrontendSnapshot(profile)
	if snapshot.Service != "ubiquity-ai-platform" {
		t.Fatalf("unexpected service %q", snapshot.Service)
	}
	for _, layer := range []string{"IaaS", "CaaS", "AI PaaS", "Workload Isolation", "Operations"} {
		if !snapshotHasLayer(snapshot, layer) {
			t.Fatalf("frontend snapshot missing NCP layer %q", layer)
		}
	}
	for _, required := range []string{"iaas-bare-metal-vm-lifecycle", "caas-gpu-kubernetes-substrate", "paas-serving-scheduling", "tenant-workload-isolation", "unified-frontend-service"} {
		if !snapshotHasCapability(snapshot, required) {
			t.Fatalf("frontend snapshot missing capability %q", required)
		}
	}
	if !strings.Contains(snapshot.ReadinessPolicy, "fail closed") {
		t.Fatalf("readiness policy must fail closed: %q", snapshot.ReadinessPolicy)
	}
}

func TestFrontendHandlerServesHTMLAndJSONAPIs(t *testing.T) {
	handler, err := NewFrontendHandler("ai-production")
	if err != nil {
		t.Fatalf("NewFrontendHandler: %v", err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d", resp.StatusCode)
	}
	buf := make([]byte, 8192)
	n, _ := resp.Body.Read(buf)
	html := string(buf[:n])
	for _, want := range []string{"Unified AI Platform Console", "IaaS", "CaaS", "AI PaaS", "Workload Isolation", "Platform API"} {
		if !strings.Contains(html, want) {
			t.Fatalf("HTML missing %q:\n%s", want, html)
		}
	}

	apiResp, err := http.Get(server.URL + "/api/platform")
	if err != nil {
		t.Fatalf("GET /api/platform: %v", err)
	}
	defer apiResp.Body.Close()
	var snapshot FrontendSnapshot
	if err := json.NewDecoder(apiResp.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode platform API: %v", err)
	}
	if snapshot.Profile.Name != "ai-production" || len(snapshot.Requirements) == 0 {
		t.Fatalf("unexpected API snapshot: %#v", snapshot)
	}
}

func TestFrontendHandlerFailsClosedForUnknownProfile(t *testing.T) {
	if _, err := NewFrontendHandler("unsupported"); err == nil {
		t.Fatal("unknown frontend profile should fail closed")
	}
}

func snapshotHasLayer(snapshot FrontendSnapshot, layer string) bool {
	for _, existing := range snapshot.Layers {
		if existing.Name == layer {
			return true
		}
	}
	return false
}

func snapshotHasCapability(snapshot FrontendSnapshot, id string) bool {
	for _, layer := range snapshot.Layers {
		for _, cap := range layer.Capabilities {
			if cap.ID == id {
				return true
			}
		}
	}
	return false
}
