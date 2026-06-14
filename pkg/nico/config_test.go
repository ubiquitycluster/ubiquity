package nico

import (
	"strings"
	"testing"
)

func TestConfigDefaultsAPIName(t *testing.T) {
	cfg := Config{BaseURL: "https://nico.example", Org: "acme", Token: "secret-token", Mode: ModeLive}.WithDefaults()
	if cfg.APIName != "nico" {
		t.Fatalf("APIName default = %q, want nico", cfg.APIName)
	}
}

func TestConfigValidateBaseURLByMode(t *testing.T) {
	if err := (Config{Mode: ModeLive, Org: "acme"}).WithDefaults().Validate(); err == nil {
		t.Fatalf("live mode without BaseURL should fail")
	}
	for _, mode := range []Mode{ModeMock, ModeOffline} {
		if err := (Config{Mode: mode, Org: "acme"}).WithDefaults().Validate(); err != nil {
			t.Fatalf("%s mode without BaseURL should validate: %v", mode, err)
		}
	}
}

func TestConfigStringAndDebugRedactSecrets(t *testing.T) {
	cfg := Config{BaseURL: "https://nico.example", Org: "acme", APIName: "custom", Token: "super-secret", TokenCommand: "pass show nico/token", Mode: ModeLive}
	for name, got := range map[string]string{"String": cfg.String(), "Debug": cfg.Debug()} {
		if strings.Contains(got, "super-secret") || strings.Contains(got, "pass show nico/token") {
			t.Fatalf("%s leaked secret material: %s", name, got)
		}
		if !strings.Contains(got, redactedValue) {
			t.Fatalf("%s should contain redaction marker, got %s", name, got)
		}
	}
}
