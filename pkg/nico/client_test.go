package nico

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientBuildsOpenAPIResourcePaths(t *testing.T) {
	wantPaths := map[string]bool{
		"/v2/org/acme/nico/site":                 false,
		"/v2/org/acme/nico/machine":              false,
		"/v2/org/acme/nico/machine/node-1/power": false,
		"/v2/org/acme/nico/operating-system":     false,
		"/v2/org/acme/nico/instance":             false,
		"/v2/org/acme/nico/task/task-1":          false,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := wantPaths[r.URL.Path]; !ok {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		wantPaths[r.URL.Path] = true
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v2/org/acme/nico/task/task-1" || r.URL.Path == "/v2/org/acme/nico/machine/node-1/power" {
			_, _ = w.Write([]byte(`{"id":"task-1","status":"succeeded"}`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: "tok"}.WithDefaults())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := context.Background()
	_, _ = c.ListSites(ctx)
	_, _ = c.ListMachines(ctx)
	_, _ = c.ListOperatingSystems(ctx)
	_, _ = c.ListInstances(ctx)
	_, _ = c.GetTask(ctx, "task-1")
	_, _ = c.PowerMachine(ctx, "node-1", "off", "maintenance")

	for path, seen := range wantPaths {
		if !seen {
			t.Fatalf("path %s was not requested", path)
		}
	}
}

func TestClientEscapesOrgAPIAndResourceID(t *testing.T) {
	var got string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"id":"task/1","status":"succeeded"}`))
	}))
	defer ts.Close()

	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme/core", APIName: "nico api", Token: "tok"}.WithDefaults())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, _ = c.GetTask(context.Background(), "task/1")
	want := "/v2/org/acme%2Fcore/nico%20api/task/task%2F1"
	if got != want {
		t.Fatalf("escaped path = %q, want %q", got, want)
	}
}

func TestClientListMethodsDecodeBareArraysAndItemsEnvelope(t *testing.T) {
	responses := map[string]string{
		"/v2/org/acme/nico/site":              `{"items":[{"id":"site-1","name":"sf01"}]}`,
		"/v2/org/acme/nico/machine":           `[{"id":"machine-1","name":"cn01"}]`,
		"/v2/org/acme/nico/operating-system":  `{"items":[{"id":"os-1","name":"rocky"}]}`,
		"/v2/org/acme/nico/instance":          `{"items":[{"id":"inst-1","nodeName":"cn01"}]}`,
		"/v2/org/acme/nico/task":              `{"items":[{"id":"task-1","status":"running"}]}`,
		"/v2/org/acme/nico/machine-gpu-stats": `{"items":[{"machineId":"machine-1","count":8}]}`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := responses[r.URL.Path]
		if !ok {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()
	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: "bearer-secret"}.WithDefaults())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := context.Background()
	if sites, err := c.ListSites(ctx); err != nil || len(sites) != 1 || sites[0].ID != "site-1" {
		t.Fatalf("ListSites = %#v, %v", sites, err)
	}
	if machines, err := c.ListMachines(ctx); err != nil || len(machines) != 1 || machines[0].ID != "machine-1" {
		t.Fatalf("ListMachines = %#v, %v", machines, err)
	}
	if oses, err := c.ListOperatingSystems(ctx); err != nil || len(oses) != 1 || oses[0].ID != "os-1" {
		t.Fatalf("ListOperatingSystems = %#v, %v", oses, err)
	}
	if instances, err := c.ListInstances(ctx); err != nil || len(instances) != 1 || instances[0].ID != "inst-1" {
		t.Fatalf("ListInstances = %#v, %v", instances, err)
	}
	if tasks, err := c.ListTasks(ctx); err != nil || len(tasks) != 1 || tasks[0].ID != "task-1" {
		t.Fatalf("ListTasks = %#v, %v", tasks, err)
	}
	if stats, err := c.ListMachineGPUStats(ctx); err != nil || len(stats) != 1 || stats[0].Count != 8 {
		t.Fatalf("ListMachineGPUStats = %#v, %v", stats, err)
	}
}

func TestClientErrorsRedactResponseBodyAndTokenCommand(t *testing.T) {
	const token = "bearer-secret"
	const tokenCommand = "vault read nico/token"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `failed token bearer-secret command vault read nico/token`, http.StatusUnauthorized)
	}))
	defer ts.Close()
	c, err := NewClient(Config{BaseURL: ts.URL, Org: "acme", Token: token, TokenCommand: tokenCommand}.WithDefaults())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.ListMachines(context.Background())
	if err == nil {
		t.Fatal("expected request error")
	}
	msg := err.Error()
	if strings.Contains(msg, token) || strings.Contains(msg, tokenCommand) {
		t.Fatalf("error leaked secret material: %s", msg)
	}
	if !strings.Contains(msg, "<redacted>") {
		t.Fatalf("error did not show redaction marker: %s", msg)
	}
}
