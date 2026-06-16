package aiplatform

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"
)

// FrontendSnapshot is the project-native API contract for the unified AI
// platform frontend. It intentionally exposes Ubiquity's capability model and
// readiness boundaries rather than proxying another project's data model.
type FrontendSnapshot struct {
	Service           string           `json:"service"`
	Profile           Profile          `json:"profile"`
	SupportedProfiles []string         `json:"supportedProfiles"`
	Layers            []FrontendLayer  `json:"layers"`
	Requirements      []NCPRequirement `json:"requirements"`
	ReadinessPolicy   string           `json:"readinessPolicy"`
	ApprovalPolicy    string           `json:"approvalPolicy"`
}

type FrontendLayer struct {
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Capabilities []FrontendCapability `json:"capabilities"`
}

type FrontendCapability struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Evidence    []string `json:"evidence"`
	Readiness   string   `json:"readiness"`
}

const (
	frontendReadinessPolicy = "fail closed until live GPU, runtime, network, serving, isolation, telemetry, and validation evidence is present"
	frontendApprovalPolicy  = "reference-platform evidence only; no vendor approval or certification claim without explicit external approval evidence"
)

// NewFrontendHandler returns an HTTP handler for the unified AI platform
// frontend. The handler is side-effect free and safe for local previews and
// Kubernetes deployment behind an ingress/controller.
func NewFrontendHandler(profileName string) (http.Handler, error) {
	profile, err := GetProfile(profileName)
	if err != nil {
		return nil, err
	}
	snapshot := BuildFrontendSnapshot(profile)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(RenderFrontendHTML(snapshot)))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"ubiquity-ai-platform"}`))
	})
	mux.HandleFunc("/api/platform", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, snapshot)
	})
	mux.HandleFunc("/api/requirements", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, snapshot.Requirements)
	})
	mux.HandleFunc("/api/profiles", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, snapshot.SupportedProfiles)
	})
	return mux, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func BuildFrontendSnapshot(profile Profile) FrontendSnapshot {
	requirements := NCPRequirements()
	layersByName := map[string][]NCPRequirement{}
	for _, requirement := range requirements {
		layersByName[requirement.Layer] = append(layersByName[requirement.Layer], requirement)
	}
	order := []string{"IaaS", "CaaS", "AI PaaS", "Workload Isolation", "Operations"}
	layers := make([]FrontendLayer, 0, len(order))
	for _, layerName := range order {
		reqs := layersByName[layerName]
		if len(reqs) == 0 {
			continue
		}
		capabilities := make([]FrontendCapability, 0, len(reqs))
		for _, requirement := range reqs {
			capabilities = append(capabilities, FrontendCapability{
				ID:          requirement.ID,
				Title:       titleFromID(requirement.ID),
				Description: requirement.Capability,
				Evidence:    append([]string(nil), requirement.UbiquityEvidence...),
				Readiness:   requirement.ReadinessSignal,
			})
		}
		layers = append(layers, FrontendLayer{Name: layerName, Description: layerDescription(layerName), Capabilities: capabilities})
	}
	return FrontendSnapshot{
		Service:           "ubiquity-ai-platform",
		Profile:           profile,
		SupportedProfiles: Names(),
		Layers:            layers,
		Requirements:      requirements,
		ReadinessPolicy:   frontendReadinessPolicy,
		ApprovalPolicy:    frontendApprovalPolicy,
	}
}

func layerDescription(layer string) string {
	switch layer {
	case "IaaS":
		return "Tenant-consumable bare-metal and VM lifecycle with sanitized, evidence-backed readiness."
	case "CaaS":
		return "Managed Kubernetes substrate for GPU, RDMA, runtime, telemetry, and network services."
	case "AI PaaS":
		return "Cloud-native AI serving and scheduling control plane for inference, batch, and training workloads."
	case "Workload Isolation":
		return "Tenant boundary controls: namespaces, quotas, limits, and default-deny traffic policy."
	case "Operations":
		return "Operator evidence for health, observability, validation, backup boundaries, and final demos."
	default:
		return "Reference-platform capability layer."
	}
}

func titleFromID(id string) string {
	parts := strings.Split(id, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

// RenderFrontendHTML renders a deterministic single-page console. The page uses
// the JSON API for machine consumers and server-rendered content for static
// GitOps previews, so it works without a JavaScript build pipeline.
func RenderFrontendHTML(snapshot FrontendSnapshot) string {
	profileNames := append([]string(nil), snapshot.SupportedProfiles...)
	sort.Strings(profileNames)
	var b strings.Builder
	b.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<title>Ubiquity AI Platform</title><style>")
	b.WriteString("body{margin:0;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,Segoe UI,sans-serif;background:#0b1020;color:#e5eefc}a{color:#93c5fd}.shell{max-width:1180px;margin:0 auto;padding:32px}.hero{border:1px solid #27364f;background:linear-gradient(135deg,#111827,#172554);border-radius:24px;padding:28px;box-shadow:0 20px 80px #0008}.badge{display:inline-block;padding:6px 10px;border-radius:999px;background:#1e3a8a;color:#bfdbfe;font-size:12px;text-transform:uppercase;letter-spacing:.08em}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:18px;margin-top:24px}.card{background:#111827;border:1px solid #27364f;border-radius:18px;padding:18px}.card h3{margin:.2rem 0 .5rem}.muted{color:#9ca3af}.evidence{font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-size:12px;background:#020617;border-radius:10px;padding:10px;overflow:auto}.policy{border-left:4px solid #f59e0b;padding-left:14px}.api{display:flex;gap:12px;flex-wrap:wrap;margin-top:16px}.api a{background:#172554;border:1px solid #3b82f6;border-radius:999px;padding:8px 12px;text-decoration:none}</style></head><body><main class=\"shell\">")
	b.WriteString("<section class=\"hero\"><span class=\"badge\">Unified AI Platform Console</span>")
	b.WriteString("<h1>Ubiquity AI Platform</h1><p class=\"muted\">Project-native frontend for NCP-style infrastructure, Kubernetes, AI platform, tenant isolation, and operations evidence.</p>")
	b.WriteString("<p><strong>Profile:</strong> ")
	b.WriteString(html.EscapeString(snapshot.Profile.Name))
	b.WriteString(" — ")
	b.WriteString(html.EscapeString(snapshot.Profile.Description))
	b.WriteString("</p><p><strong>Supported profiles:</strong> ")
	b.WriteString(html.EscapeString(strings.Join(profileNames, ", ")))
	b.WriteString("</p><div class=\"api\"><a href=\"/api/platform\">Platform API</a><a href=\"/api/requirements\">Requirement API</a><a href=\"/healthz\">Health</a></div></section>")
	b.WriteString("<section class=\"grid\">")
	for _, layer := range snapshot.Layers {
		b.WriteString("<article class=\"card\"><h2>")
		b.WriteString(html.EscapeString(layer.Name))
		b.WriteString("</h2><p class=\"muted\">")
		b.WriteString(html.EscapeString(layer.Description))
		b.WriteString("</p>")
		for _, cap := range layer.Capabilities {
			b.WriteString("<div><h3>")
			b.WriteString(html.EscapeString(cap.Title))
			b.WriteString("</h3><p>")
			b.WriteString(html.EscapeString(cap.Description))
			b.WriteString("</p><p class=\"muted\"><strong>Readiness:</strong> ")
			b.WriteString(html.EscapeString(cap.Readiness))
			b.WriteString("</p><div class=\"evidence\">")
			b.WriteString(html.EscapeString(strings.Join(cap.Evidence, "\n")))
			b.WriteString("</div></div>")
		}
		b.WriteString("</article>")
	}
	b.WriteString("</section><section class=\"card policy\"><h2>Evidence boundary</h2><p>")
	b.WriteString(html.EscapeString(snapshot.ReadinessPolicy))
	b.WriteString("</p><p class=\"muted\">")
	b.WriteString(html.EscapeString(snapshot.ApprovalPolicy))
	b.WriteString("</p></section></main></body></html>")
	return b.String()
}

func FrontendListenDescription(addr, profile string) string {
	return fmt.Sprintf("serving Ubiquity AI Platform frontend on %s with profile %s", addr, profile)
}
