package cloud

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCommunityReadinessArtifactsCoverNextLevelPlan(t *testing.T) {
	devcontainer := mustRead(t, "../../.devcontainer/devcontainer.json")
	devcontainerReadme := mustRead(t, "../../.devcontainer/README.md")
	precommit := mustRead(t, "../../.pre-commit-config.yaml")
	bugReport := mustRead(t, "../../.github/ISSUE_TEMPLATE/bug_report.md")
	featureRequest := mustRead(t, "../../.github/ISSUE_TEMPLATE/feature_request.md")
	contributing := mustRead(t, "../../CONTRIBUTING.md")
	security := mustRead(t, "../../SECURITY.md")
	makefile := mustRead(t, "../../Makefile")
	readme := mustRead(t, "../../README.md")

	var devcontainerJSON map[string]any
	if err := json.Unmarshal([]byte(devcontainer), &devcontainerJSON); err != nil {
		t.Fatalf("devcontainer.json is not valid JSON: %v", err)
	}
	for _, required := range []string{"Ubiquity Development", "mcr.microsoft.com/devcontainers/go", "docker-in-docker", "kubectl", "helm", "pre-commit install", "govulncheck", "make cli"} {
		if !strings.Contains(devcontainer, required) {
			t.Fatalf("devcontainer missing %q", required)
		}
	}
	for _, required := range []string{"Reopen in Container", "Docker-in-Docker", "pre-commit", "govulncheck"} {
		if !strings.Contains(devcontainerReadme, required) {
			t.Fatalf("devcontainer README missing %q", required)
		}
	}
	for _, required := range []string{"autofix_commit_msg", "autoupdate_schedule: monthly", "tekwizely/pre-commit-golang", "go-fmt", "go-mod-tidy"} {
		if !strings.Contains(precommit, required) {
			t.Fatalf("pre-commit config missing %q", required)
		}
	}
	for path, content := range map[string]string{
		"bug_report":      bugReport,
		"feature_request": featureRequest,
	} {
		for _, required := range []string{"---", "labels:", "Additional context"} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s template missing %q", path, required)
			}
		}
	}
	for _, required := range []string{"Development Setup", "make cli", "./ubiquity-cli up --sandbox --skip-security", "cmd/ubiquity/", "cloud/"} {
		if !strings.Contains(contributing, required) {
			t.Fatalf("CONTRIBUTING.md missing %q", required)
		}
	}
	for _, required := range []string{"Security Policy", "Supported Versions", "Reporting a Vulnerability", "Disclosure Policy", "48 hours"} {
		if !strings.Contains(security, required) {
			t.Fatalf("SECURITY.md missing %q", required)
		}
	}
	for _, required := range []string{"dev: cli", "Available targets:", "completions", "asciinema rec ubiquity-demo.cast", "./ubiquity-cli up --sandbox --skip-security"} {
		if !strings.Contains(makefile, required) {
			t.Fatalf("Makefile missing %q", required)
		}
	}
	for _, required := range []string{"Raspberry Pi (Experimental)", "GOARCH=arm64 make cli"} {
		if !strings.Contains(readme, required) {
			t.Fatalf("README.md missing %q", required)
		}
	}
}

func TestContributorDocsDoNotContainHiddenFenceCharacters(t *testing.T) {
	for path, content := range map[string]string{
		"README.md":       mustRead(t, "../../README.md"),
		"CONTRIBUTING.md": mustRead(t, "../../CONTRIBUTING.md"),
	} {
		if strings.Contains(content, "\u200b") {
			t.Fatalf("%s contains zero-width spaces in rendered Markdown", path)
		}
	}
}
