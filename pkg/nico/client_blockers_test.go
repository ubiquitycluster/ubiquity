package nico

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientDefaultHTTPTimeout(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://nico.example", Org: "acme", Token: "tok"}.WithDefaults())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.httpClient == nil || c.httpClient.Timeout <= 0 {
		t.Fatalf("default HTTP client timeout = %v, want positive", c.httpClient)
	}
}

func TestClientLiveMethodsUseExpectedHTTPVerbsAndPaths(t *testing.T) {
	want := map[string]string{
		"GET /v2/org/acme/nico/machine/node-a":     `{"id":"machine-1","name":"node-a"}`,
		"POST /v2/org/acme/nico/operating-system":  `{"id":"os-1","name":"ubuntu"}`,
		"POST /v2/org/acme/nico/instance":          `{"id":"inst-1","name":"node-a"}`,
		"DELETE /v2/org/acme/nico/instance/inst-1": `{}`,
	}
	seen := map[string]bool{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.EscapedPath()
		body, ok := want[key]
		if !ok {
			t.Fatalf("unexpected request %s", key)
		}
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("missing bearer token header")
		}
		seen[key] = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: "tok"}.WithDefaults(), WithHTTPClient(&http.Client{Timeout: time.Second}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := context.Background()
	if m, err := c.GetMachine(ctx, "node-a"); err != nil || m.Name != "node-a" {
		t.Fatalf("GetMachine = %#v, %v", m, err)
	}
	if os, err := c.CreateOperatingSystem(ctx, OperatingSystem{Name: "ubuntu"}); err != nil || os.ID != "os-1" {
		t.Fatalf("CreateOperatingSystem = %#v, %v", os, err)
	}
	if inst, err := c.CreateInstance(ctx, Instance{Name: "node-a"}); err != nil || inst.ID != "inst-1" {
		t.Fatalf("CreateInstance = %#v, %v", inst, err)
	}
	if err := c.DeleteInstance(ctx, "inst-1"); err != nil {
		t.Fatalf("DeleteInstance: %v", err)
	}
	for key := range want {
		if !seen[key] {
			t.Fatalf("request %s was not seen", key)
		}
	}
}
