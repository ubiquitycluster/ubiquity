Take the Ubiquity project to the next level of community readiness at /home/ubuntu/ubiquity. Work in P0→P1→P2 order, commit after each item.

## Prerequisites
Go 1.22+, Docker, VS Code (for devcontainer testing), `pre-commit` installed. Install `asciinema` if doing the demo stretch goal.

---

## P0 Items (critical)

### P0-1: Create devcontainer for one-click development environment
Create .devcontainer/devcontainer.json that sets up everything a contributor needs:

```json
{
  "name": "Ubiquity Development",
  "image": "mcr.microsoft.com/devcontainers/go:1-1.24-bookworm",
  "features": {
    "ghcr.io/devcontainers/features/docker-in-docker:2": {},
    "ghcr.io/devcontainers/features/kubectl:1": {},
    "ghcr.io/devcontainers/features/helm:1": {}
  },
  "postCreateCommand": "sudo apt-get update && sudo apt-get install -y shellcheck yamllint age && pip install pre-commit ansible-lint && pre-commit install && go install github.com/goreleaser/goreleaser/v2@latest && go install github.com/getsops/sops/v3/cmd/sops@latest && go install golang.org/x/vuln/cmd/govulncheck@latest && make cli",
  "extensions": [
    "golang.go",
    "redhat.vscode-yaml",
    "ms-azuretools.vscode-docker",
    "timonwong.shellcheck"
  ],
  "settings": {
    "go.useLanguageServer": true,
    "yaml.schemas": {
      "https://raw.githubusercontent.com/khuedoan/homelab/main/.yamllint.yaml": ".yamllint.yaml"
    }
  }
}
```

Also create .devcontainer/README.md explaining how to use it (VS Code → Reopen in Container).

Commit: git add .devcontainer/ && git commit -m "dev: add devcontainer for one-click development environment"

### P0-2: Enable pre-commit.ci for automatic PR linting
Add `ci:` section to the top of .pre-commit-config.yaml:

```yaml
ci:
  autofix_commit_msg: "chore: [pre-commit.ci] auto fixes"
  autoupdate_commit_msg: "chore: [pre-commit.ci] autoupdate hooks"
  autoupdate_schedule: monthly
  submodules: false
```

Also add these missing Go-related hooks to the repos list:
```yaml
  - repo: https://github.com/tekwizely/pre-commit-golang
    rev: v1.0.0-rc.1
    hooks:
      - id: go-fmt
      - id: go-mod-tidy
```

Commit: git commit -m "ci: enable pre-commit.ci and add Go formatting hooks" -a

### P0-3: Create GitHub issue templates
Create .github/ISSUE_TEMPLATE/ directory with two templates:

.github/ISSUE_TEMPLATE/bug_report.md:
```markdown
---
name: Bug report
about: Create a report to help us improve
title: ''
labels: bug
assignees: ''
---

**Describe the bug**
A clear description of the issue.

**To reproduce**
Steps to reproduce the behavior:
1. Run '...'
2. See error

**Expected behavior**
What you expected to happen.

**Environment (please complete):**
- OS: [e.g. Ubuntu 24.04]
- Ubiquity version: [e.g. 1.0.0 or commit SHA]
- Deployment type: [sandbox / bare metal / cloud]
- Kubernetes: [e.g. k3s v1.30.4]

**Additional context**
Add any other context about the problem here.
```

.github/ISSUE_TEMPLATE/feature_request.md:
```markdown
---
name: Feature request
about: Suggest an idea for this project
title: ''
labels: enhancement
assignees: ''
---

**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
What you want to happen.

**Describe alternatives you've considered**
Other approaches you've thought about.

**Additional context**
Screenshots, links, or references.
```

Commit: git add .github/ISSUE_TEMPLATE/ && git commit -m "docs: add GitHub issue templates for bugs and feature requests"

### P0-4: Update CONTRIBUTING.md for Go CLI
Read CONTRIBUTING.md (256 lines). It still references the old Python scripts. Add a new "Development Setup" section at the top:

```markdown
## Development Setup

### Prerequisites
- Go 1.22+
- Docker (for sandbox testing)
- Helm
- kubectl
- pre-commit

### Quick start
```bash
# Clone and build
git clone https://github.com/ubiquitycluster/ubiquity.git
cd ubiquity
make cli          # Build the CLI binary
make installer    # Build the PXE installer (optional)

# Install pre-commit hooks
pre-commit install

# Test with sandbox
./ubiquity-cli up --sandbox --skip-security
```

### Project structure
- `cmd/ubiquity/` — Go CLI (main entry point)
- `pkg/` — Go packages (config, provision, network, cloud, tui)
- `tools/` — Go PXE installer
- `system/` — Helm charts for cluster components
- `platform/` — Helm charts for platform services
- `metal/` — Ansible roles for bare metal provisioning
- `cloud/` — Terraform modules for cloud providers
```

Keep the rest of the existing CONTRIBUTING.md content below this new section.

Also remove or update references to the Python scripts if any remain.

Commit: git commit -m "docs: update CONTRIBUTING.md for Go CLI development" -a

### P0-5: Create SECURITY.md
Create SECURITY.md at repo root:

```markdown
# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ |

## Reporting a Vulnerability

Please report security vulnerabilities by email to the security team
listed in SECURITY_CONTACTS. Do not open public issues for vulnerabilities.

You should receive a response within 48 hours. If not, please follow up.

## Disclosure Policy

- Vulnerability reported → acknowledgement within 48h
- Investigation → fix developed → release prepared
- Public disclosure after fix is released
```

Commit: git add SECURITY.md && git commit -m "docs: add SECURITY.md for vulnerability reporting"

---

## P1 Items (medium priority)

### P1-1: Fix `make dev` target
Read the Makefile. Find the `dev:` target. It currently points to the old pipeline:
```
dev: metal bootstrap wait post-install
```

Change it to:
```makefile
# Development workflow
dev: cli
	@echo "============================================"
	@echo "  Development environment ready"
	@echo "  Run: ./ubiquity-cli up --sandbox"
	@echo "============================================"
```

Also add `help:` target that lists available commands:
```makefile
# Show available targets
help:
	@echo "Available targets:"
	@echo "  cli         Build the ubiquity CLI binary"
	@echo "  installer   Build the PXE installer binary"
	@echo "  completions Generate shell completion files"
	@echo "  test        Run tests"
	@echo "  dev         Prepare development environment"
	@echo "  install     Install CLI to /usr/local/bin"
```

Commit: git commit -m "chore: fix make dev target and add help target" -a

### P1-2: Add Grafana dashboards as code
Create grafana/dashboards/ directory and add the Ubiquity cluster dashboard as a version-controlled JSON file.

First check if any dashboards are embedded in Helm values or monitoring configs.

Create grafana/dashboards/README.md:
```markdown
# Grafana Dashboards

This directory contains Grafana dashboards as code.
Dashboards are in JSON format and can be imported into Grafana.

## Usage

Import via Grafana UI or use grafanactl:
```
# Kubernetes
kubectl create configmap ubiquity-dashboard \
  --from-file=grafana/dashboards/cluster.json \
  -n monitoring
```
```

Create grafana/dashboards/cluster.json — a basic dashboard JSON that shows:
- Cluster status (nodes, pods, deployments)
- Resource usage (CPU, memory, disk)
- Phase provisioning status

Use a simple dashboard structure:
```json
{
  "__inputs": [],
  "__requires": [],
  "title": "Ubiquity Cluster",
  "uid": "ubiquity-cluster",
  "panels": [
    {
      "title": "Cluster Status",
      "type": "stat",
      "datasource": "Prometheus"
    }
  ],
  "schemaVersion": 38
}
```

Commit: git add grafana/ && git commit -m "feat: add Grafana dashboards as code"

### P1-3: Consolidate GitHub Actions workflows
The 4 workflow files may have been iteratively edited. Audit them for duplication.

Read .github/workflows/ci.yaml and check if:
- The lint job has duplicate steps that could be merged
- The validate job overlaps with test
- There are stale steps from removed features

Consolidate without removing functionality. The key principle: one workflow file per trigger type, not per step.

If ci.yaml has excessive trigger sections (16 was reported), simplify the on: section:
```yaml
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch: {}
```

Commit: git commit -m "ci: consolidate GitHub Actions workflow triggers" -a

### P1-4: Add backup/restore CLI commands
Create cmd/ubiquity/cmd/backup.go with a basic backup command:

```go
var backupCmd = &cobra.Command{
    Use:   "backup",
    Short: "Backup cluster state",
    Long: `Creates a backup of cluster state, including:
- All Kubernetes resources (via kubectl)
- Provisioning state
- Configuration files`,
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("Creating cluster backup...")
        // Create timestamped backup directory
        // Export kubectl resources
        // Copy provisioning state
        return nil
    },
}
```

And cmd/ubiquity/cmd/restore.go:
```go
var restoreCmd = &cobra.Command{
    Use:   "restore <backup-dir>",
    Short: "Restore cluster from backup",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Printf("Restoring from backup: %s\n", args[0])
        return nil
    },
}
```

Wire backup/restore as optional steps in the provisioning pipeline. Create docs/how-to-guides/backup-and-restore.md referencing these commands.

Commit: git add cmd/ubiquity/cmd/backup.go cmd/ubiquity/cmd/restore.go docs/how-to-guides/backup-and-restore.md && git commit -m "feat: add backup and restore CLI commands"

### P1-5: Wire Ollama into ubiquity ai command
Create cmd/ubiquity/cmd/ai.go that sends cluster state to a local Ollama instance:

```go
var aiCmd = &cobra.Command{
    Use:   "ai [prompt]",
    Short: "AI-assisted troubleshooting using Ollama",
    Long: `Sends cluster state and logs to a local Ollama LLM for diagnosis.
Requires Ollama running (apps/ollama/).`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Load provisioning state
        state, _ := provision.LoadState()
        // Build prompt from state + logs + user input
        prompt := "Analyze this cluster state:\n"
        if state != nil {
            prompt += state.Summary()
        }
        // Call Ollama API at localhost:11434
        resp, err := http.Post(
            "http://localhost:11434/api/generate",
            "application/json",
            strings.NewReader(fmt.Sprintf(`{"model":"llama3.2","prompt":"%s","stream":false}`, prompt)),
        )
        if err != nil {
            return fmt.Errorf("Ollama not available: %w", err)
        }
        // Parse and print response
        return nil
    },
}
```

Add net/http and strings to imports.

Commit: git add cmd/ubiquity/cmd/ai.go && git commit -m "feat: add ubiquity ai command for LLM-assisted troubleshooting"

---

## P2 Items (stretch)

### P2-1: Add Raspberry Pi / arm64 build target
Read .goreleaser.yaml. Add arm64 to the existing linux builds (already there), and add a Raspberry Pi-specific build note to README.

Update README.md "Hardware" section to mention Raspberry Pi:
```markdown
### Raspberry Pi (Experimental)
You can run Ubiquity on Raspberry Pi 4/5 clusters (64-bit OS required).
Build the CLI for arm64:
```
GOARCH=arm64 make cli
```
```

Commit: git commit -m "docs: document Raspberry Pi / arm64 support" -a

### P2-2: Add make demo target for asciicast recording
Install asciinema: sudo apt-get install -y asciinema

Add to Makefile:
```makefile
# Record a demo asciicast
demo:
	@echo "Recording demo... Press Ctrl+D when done."
	asciinema rec ubiquity-demo.cast -c "./ubiquity-cli up --sandbox --skip-security 2>&1 | head -20"
	@echo "Demo saved to ubiquity-demo.cast"
	@echo "Upload to https://asciinema.org or replay locally with 'asciinema play ubiquity-demo.cast'"
```

Commit: git commit -m "chore: add make demo target for asciicast recording" -a

### P2-3: Add ADR-012 for devcontainer, ADR-013 for pre-commit.ci
Create two new ADRs documenting the rationale:

ADR-012: Use devcontainer for reproducible development environments
- Context: New contributors spent significant time installing Go, Helm, kubectl, pre-commit, shellcheck, etc. before they could make their first contribution. Different team members had different tool versions.
- Decision: Use a devcontainer with pinned tool versions. VS Code's "Reopen in Container" automatically provisions the full environment.
- Consequences: Requires VS Code + Docker. But reduces onboarding from hours to minutes.

ADR-013: Use pre-commit.ci for automated linting
- Context: Pre-commit hooks were installed locally but not enforced in CI. PRs could bypass linting.
- Decision: Enable pre-commit.ci service to run pre-commit on every PR automatically.
- Consequences: Zero-config CI for linting. Frees up GitHub Actions compute for actual testing.

Update README.md to link both new ADRs.

Commit: git add docs/reference/architecture/decision-records/ADR-012-devvontainer.md docs/reference/architecture/decision-records/ADR-013-precommit-ci.md && git commit -m "docs: add ADR-012 for devcontainer and ADR-013 for pre-commit.ci"

---

## Verification

After all items:
- go build ./... — PASS
- go test ./pkg/... ./cmd/... -count=1 — all green
- ls .devcontainer/devcontainer.json — exists
- ls .github/ISSUE_TEMPLATE/bug_report.md — exists
- ls .github/ISSUE_TEMPLATE/feature_request.md — exists
- ls SECURITY.md — exists
- grep -q "pre-commit.ci" .pre-commit-config.yaml — configured
- ubiquity backup --help — works
- ubiquity restore --help — works
- ubiquity ai --help — works
- make help — lists targets
- make demo — records a demo or shows available
- ls grafana/dashboards/cluster.json — exists
- git log --oneline -15 — 15+ commits
