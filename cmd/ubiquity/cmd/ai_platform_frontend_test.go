package cmd

import (
	"net/http"
	"testing"

	"github.com/ubiquitycluster/ubiquity/pkg/aiplatform"
)

func TestAIPlatformRenderIncludesUnifiedFrontendConsole(t *testing.T) {
	profile, err := aiplatform.GetProfile("ai-production")
	if err != nil {
		t.Fatalf("GetProfile(ai-production): %v", err)
	}
	manifest := renderAIPlatformManifest(profile)
	assertContains(t, manifest, "name: ai-platform-ai-platform-console")
	assertContains(t, manifest, "path: platform/ai-platform-console")
	assertContains(t, manifest, "namespace: ai-platform")
}

func TestAIPlatformServeSubcommandUsesUnifiedFrontend(t *testing.T) {
	serve := findCommand(aiPlatformCmd, "serve")
	if serve == nil {
		t.Fatal("expected ai-platform serve subcommand")
	}
	aiPlatformProfile = "ai-production"
	aiPlatformServeListen = "127.0.0.1:0"
	defer func() {
		aiPlatformProfile = "gpu-basic"
		aiPlatformServeListen = ":8080"
		serve.Flags().Set("listen", ":8080")
	}()
	serve.Flags().Set("listen", "127.0.0.1:0")

	oldRunner := runAIPlatformHTTPServer
	defer func() { runAIPlatformHTTPServer = oldRunner }()
	var gotAddr string
	var gotHandler http.Handler
	runAIPlatformHTTPServer = func(addr string, handler http.Handler) error {
		gotAddr = addr
		gotHandler = handler
		return nil
	}
	output := captureOutput(func() {
		if err := serve.RunE(serve, []string{}); err != nil {
			t.Fatalf("serve failed: %v", err)
		}
	})
	if gotAddr != "127.0.0.1:0" {
		t.Fatalf("listen address = %q", gotAddr)
	}
	if gotHandler == nil {
		t.Fatal("serve did not build the frontend handler")
	}
	assertContains(t, output, "serving Ubiquity AI Platform frontend")
	assertContains(t, output, "profile ai-production")
}
