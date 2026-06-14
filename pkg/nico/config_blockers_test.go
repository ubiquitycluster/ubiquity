package nico

import (
	"context"
	"strings"
	"testing"
)

func TestConfigIncludesDeploymentFieldsAndValidatesLiveStrictly(t *testing.T) {
	cfg := Config{
		BaseURL:    "https://nico.example",
		Org:        "acme",
		SiteName:   "site-a",
		Token:      "tok",
		TokenEnv:   "NICO_TOKEN",
		ConfigPath: "/tmp/nico.yaml",
		Mode:       ModeLive,
	}.WithDefaults()
	if cfg.SiteName != "site-a" || cfg.TokenEnv != "NICO_TOKEN" || cfg.ConfigPath != "/tmp/nico.yaml" {
		t.Fatalf("deployment fields not preserved: %#v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid live config rejected: %v", err)
	}

	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{"invalid-mode", Config{Mode: Mode("bogus"), BaseURL: "https://nico.example", Org: "acme", Token: "tok"}, "mode"},
		{"bad-url", Config{Mode: ModeLive, BaseURL: "://bad", Org: "acme", Token: "tok"}, "base URL"},
		{"missing-org", Config{Mode: ModeLive, BaseURL: "https://nico.example", Token: "tok"}, "org"},
		{"missing-token", Config{Mode: ModeLive, BaseURL: "https://nico.example", Org: "acme"}, "token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.WithDefaults().Validate()
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) {
				t.Fatalf("Validate() = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestConfigAllowsUnauthenticatedOnlyOutsideLiveUnlessExplicit(t *testing.T) {
	for _, mode := range []Mode{ModeMock, ModeOffline} {
		if err := (Config{Mode: mode}).WithDefaults().Validate(); err != nil {
			t.Fatalf("%s without token should validate: %v", mode, err)
		}
	}
	if err := (Config{Mode: ModeLive, BaseURL: "https://nico.example", Org: "acme", AllowUnauthenticated: true}).WithDefaults().Validate(); err != nil {
		t.Fatalf("explicit unauthenticated live config should validate: %v", err)
	}
}

func TestTokenCommandResolutionUsesInjectableRunnerWithoutShell(t *testing.T) {
	var got []string
	cfg := Config{Mode: ModeLive, BaseURL: "https://nico.example", Org: "acme", TokenCommand: "pass show nico/token"}.WithDefaults()
	token, err := cfg.ResolveToken(context.Background(), TokenResolverFunc(func(ctx context.Context, argv []string) (string, error) {
		got = append([]string(nil), argv...)
		return "resolved-token\n", nil
	}))
	if err != nil {
		t.Fatalf("ResolveToken: %v", err)
	}
	if token != "resolved-token" {
		t.Fatalf("token = %q", token)
	}
	want := []string{"pass", "show", "nico/token"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("argv = %#v, want %#v", got, want)
	}
}

func TestTokenCommandRejectsShellMetacharacters(t *testing.T) {
	cfg := Config{Mode: ModeLive, BaseURL: "https://nico.example", Org: "acme", TokenCommand: "echo token; rm -rf /"}.WithDefaults()
	_, err := cfg.ResolveToken(context.Background(), TokenResolverFunc(func(context.Context, []string) (string, error) {
		t.Fatal("runner should not be called for unsafe command")
		return "", nil
	}))
	if err == nil || !strings.Contains(err.Error(), "unsafe") {
		t.Fatalf("expected unsafe token command error, got %v", err)
	}
}
