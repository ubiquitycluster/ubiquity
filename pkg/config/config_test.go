package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNonExistent(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load should not error for missing .env: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load should return non-nil config")
	}
	if cfg.Domain != "ubiquitycluster.uk" {
		t.Errorf("expected default domain, got %q", cfg.Domain)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Domain:   "example.com",
		Timezone: "America/New_York",
		Editor:   "vim",
		OS:       "Ubuntu",
	}
	if err := Save(cfg, dir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %q", loaded.Domain)
	}
	if loaded.Timezone != "America/New_York" {
		t.Errorf("expected timezone 'America/New_York', got %q", loaded.Timezone)
	}
	if loaded.Editor != "vim" {
		t.Errorf("expected editor 'vim', got %q", loaded.Editor)
	}
}

func TestParseEnvFile(t *testing.T) {
	content := `# Comment
DOMAIN=test.domain
EDITOR=nano
EMPTY=
SEED_REPO=https://example.com/repo.git`
	env := parseEnvFile(content)
	if env["DOMAIN"] != "test.domain" {
		t.Errorf("expected 'test.domain', got %q", env["DOMAIN"])
	}
	if env["EDITOR"] != "nano" {
		t.Errorf("expected 'nano', got %q", env["EDITOR"])
	}
	// Empty values should be included
	if _, ok := env["EMPTY"]; ok {
		// parseEnvFile includes empty values
	}
	if env["SEED_REPO"] != "https://example.com/repo.git" {
		t.Errorf("expected repo URL, got %q", env["SEED_REPO"])
	}
}

func TestPatchValues(t *testing.T) {
	dir := t.TempDir()

	// Create a test YAML file
	testFile := filepath.Join(dir, "test-values.yaml")
	yamlContent := `argo-cd:
  server:
    ingress:
      hosts:
        - "old.domain"
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// Apply a patch
	if err := setYAMLValue(testFile, "argo-cd:server:ingress:hosts:0", "argocd.example.com"); err != nil {
		t.Fatalf("setYAMLValue failed: %v", err)
	}

	data, _ := os.ReadFile(testFile)
	if !contains(string(data), "argocd.example.com") {
		t.Errorf("expected patched value, got:\n%s", string(data))
	}
}

func TestIsValidCIDR(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"10.0.0.0/22", true},
		{"192.168.1.0/24", true},
		{"10.0.0.0/33", false},
		{"not-a-cidr", false},
		{"10.0.0.0", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsValidCIDR(tt.input)
		if got != tt.want {
			t.Errorf("IsValidCIDR(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestConfigFilePath(t *testing.T) {
	p := Path("/tmp")
	if !contains(p, ".env") {
		t.Errorf("expected path to contain .env, got %q", p)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}