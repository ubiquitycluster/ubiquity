Implement the top-priority improvements identified in DEEP-ASSESSMENT.md at /home/ubuntu/ubiquity.

## Prerequisites
Go 1.22+, Helm, kubectl, gitleaks (install via `go install github.com/gitleaks/gitleaks/v8@latest`), kyverno CLI (install via `brew install kyverno` or download from GitHub releases).

Work in P0→P1→P2→P3 order. Commit after each item.

---

## P0 Items (must do)

### P0-1: Write ADR-001 through ADR-008

Create docs/reference/architecture/decision-records/ with a template and 8 ADRs.

First create the directory and template file:
```
mkdir -p docs/reference/architecture/decision-records
```

Create docs/reference/architecture/decision-records/template.md:
```markdown
# ADR-NNN: [Title]

**Status:** [Proposed | Accepted | Deprecated | Superseded]

**Date:** YYYY-MM-DD

## Context

What is the issue we're addressing? What options were considered?

## Decision

What was chosen and why?

## Consequences

- Positive: ...
- Negative: ...
- Neutral: ...
```

Create these 8 ADRs (each in its own file):

**ADR-001: Use Go CLI instead of Python**
- Context: Originally configured via Python scripts (scripts/configure, scripts/configure-sandbox). Needed a cross-platform, single-binary experience with no runtime dependencies.
- Decision: Rewrite as a Go CLI using cobra + viper. Go produces static binaries, has excellent CLI libraries, and the team knows Go.
- Consequences: Larger binary size (~8MB) but no Python/runtime deps. Reduced maintenance burden from 1,672 lines of Python to 834 lines of Go.

**ADR-002: Use k3s instead of full Kubernetes**
- Context: HPC clusters need lightweight K8s for resource-constrained nodes. Full K8s requires more RAM, more components.
- Decision: Use k3s (Rancher's lightweight distribution). Embedded etcd, single binary, supports standard K8s API.
- Consequences: Some edge features (certain CSI drivers, advanced networking) may need extra configuration. But standard K8s API compatibility means most tools work unmodified.

**ADR-003: Use Terraform for cloud provisioning**
- Context: Need to provision infrastructure across 5 cloud providers (AWS, Azure, GCP, OpenStack, OVH).
- Decision: Use Terraform. Mature multi-provider ecosystem, HCL is declarative, large community. Pulumi was considered but has fewer HPC-specific modules. CDKTF adds Python/TS complexity.
- Consequences: State management needed (backends). License change concerns (OpenTofu migration path noted).

**ADR-004: Use Kyverno over OPA/Gatekeeper**
- Context: Need Kubernetes admission control for security policies.
- Decision: Use Kyverno. Kubernetes-native (no custom DSL like OPA's Rego), supports policy mutation, simpler YAML-only policies. Gatekeeper was considered but requires Rego expertise.
- Consequences: Locked to K8s-specific policies (can't validate non-K8s artifacts). But for cluster-only policies, Kyverno is simpler.

**ADR-005: Use Helm chart per component**
- Context: Need idempotent, parameterized deployment of 20+ system components.
- Decision: Use Helm. Industry standard, integrates with ArgoCD, supports dependency management. Kustomize lacks template/parameter support.
- Consequences: Chart versioning overhead. Need to maintain Chart.yaml for each component. Use app-template pattern to reduce boilerplate (see ADR-NNN when created).

**ADR-006: Use Longhorn as primary storage**
- Context: Need distributed block storage for K8s workloads. Rook-Ceph was the alternative.
- Decision: Use Longhorn. Simpler operations (UI, no CRUSH map), built-in backup/DR, lighter resource requirements. Rook-Ceph considered too complex for typical HPC cluster sizes (<50 nodes).
- Consequences: Less mature than Ceph for very large clusters. But for typical ubiquity deployments (5-20 nodes), Longhorn is sufficient.

**ADR-007: Use ArgoCD over Flux**
- Context: Need GitOps operator for cluster management.
- Decision: Use ArgoCD. ApplicationSet for multi-cluster support, mature SSO integration, better UI. Flux v2 lacks ApplicationSet equivalent.
- Consequences: Heavier than Flux. But the ApplicationSet feature is critical for the multi-environment (sandbox/dev/prod) pattern.

**ADR-008: Use Bubbletea for TUI instead of plain text**
- Context: Need colored, interactive terminal output for status monitoring.
- Decision: Use charmbracelet/bubbletea + lipgloss. Rich terminal UX, cross-platform, composable components. Fallback to plain text when not in a TTY.
- Consequences: Adds 4 dependencies to go.mod (~500KB). --plain flag available for CI/pipe usage.

Also create docs/reference/architecture/decision-records/README.md linking all ADRs.

Commit: git add docs/reference/architecture/decision-records/ && git commit -m "docs: add ADR-001 through ADR-008 for key architectural decisions"

### P0-2: Create shared app-template Helm chart

Create charts/app-template/ that all apps/ individual charts can depend on. This eliminates massive boilerplate in every app chart.

Create charts/app-template/Chart.yaml:
```
apiVersion: v2
name: app-template
description: Shared application template for Ubiquity apps
type: application
version: 0.1.0
keywords:
  - ubiquity
  - template
home: https://github.com/ubiquitycluster/ubiquity
```

Create charts/app-template/values.yaml with all customizable fields:
```yaml
# Default values for app-template
# This is a YAML-formatted file.

nameOverride: ""
fullnameOverride: ""

image:
  repository: ""
  tag: latest
  pullPolicy: IfNotPresent

replicas: 1

service:
  enabled: true
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    hajimari.io/enable: "true"
  hosts:
    - host: ""
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  limits:
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

env: []
envFrom: []

persistence:
  enabled: false
  size: 10Gi
  storageClass: longhorn

probes:
  liveness:
    enabled: true
    path: /
  readiness:
    enabled: true
    path: /
```

Create charts/app-template/templates/ with standard K8s templates:
- deployment.yaml (Deployment with image, replicas, env, probes, resources, persistence)
- service.yaml (Service with port config)
- ingress.yaml (Ingress with hosts, TLS, annotations)
- pvc.yaml (PersistentVolumeClaim, conditional on persistence.enabled)
- _helpers.tpl (standard helm helper templates: name, fullname, labels)

All templates should use {{ .Values }} references so individual apps only need to override what differs.

Then convert ONE app (e.g., apps/hajimari) to use the template as a dependency:
- In apps/hajimari/Chart.yaml, add: dependencies: [{name: app-template, version: "0.1.0", repository: "file://../../charts/app-template"}]
- Replace apps/hajimari/values.yaml with minimal values that only set app-specific overrides
- Delete apps/hajimari/templates/* (no longer needed, templates come from the shared chart)

Commit: git add charts/ apps/hajimari/ && git commit -m "feat: create shared app-template Helm chart, convert hajimari to use it"

### P0-3: Add gitleaks to pre-commit

Add to .pre-commit-config.yaml:
```yaml
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks
```

Also add to .github/workflows/ci.yaml in the lint job a step:
```yaml
      - name: Run gitleaks
        uses: gitleaks/gitleaks-action@v2
```

Commit: git commit -m "sec: add gitleaks secrets scanning to pre-commit and CI" -a

---

## P1 Items (high priority)

### P1-1: Write Kyverno policy tests

For the chart at system/kyverno-policies/, create a test that validates policies against sample resources:

Create system/kyverno-policies/kyverno-test/ directory with:
- kyverno-test.yaml — Kyverno test manifest that defines test cases
- resources/ — sample K8s resources that should PASS the policies
- resources/privileged-pod.yaml — sample that SHOULD FAIL the disallow-privileged policy

Example kyverno-test.yaml structure:
```yaml
name: kyverno-policy-tests
policies:
  - ../templates/restricted-pod-security.yaml
  - ../templates/disallow-privileged-containers.yaml
  - ../templates/require-labels.yaml
resources:
  - resources/valid-pod.yaml
  - resources/privileged-pod.yaml
results:
  - policy: disallow-privileged-containers
    rule: privileged-containers
    resource: valid-pod
    result: pass
  - policy: disallow-privileged-containers
    rule: privileged-containers
    resource: privileged-pod
    result: fail
```

Run: kyverno test system/kyverno-policies/kyverno-test/
All tests must pass.

Commit: git add system/kyverno-policies/kyverno-test/ && git commit -m "test: add Kyverno policy tests against sample resources"

### P1-2: Add helm-diff CI pipeline

Create a script at scripts/helm-diff that compares Helm chart changes between branches:
```bash
#!/bin/bash
# Usage: ./scripts/helm-diff --repository URL --source BRANCH --target BRANCH --subpath PATH
# Shows what Helm values changed between feature and main

set -euo pipefail

REPO=""
SOURCE=""
TARGET="main"
SUBPATH="."

while [[ $# -gt 0 ]]; do
  case $1 in
    --repository) REPO="$2"; shift 2 ;;
    --source) SOURCE="$2"; shift 2 ;;
    --target) TARGET="$2"; shift 2 ;;
    --subpath) SUBPATH="$2"; shift 2 ;;
    *) echo "Unknown: $1"; exit 1 ;;
  esac
done

echo "Helm diff between $SOURCE and $TARGET in $SUBPATH"
cd "$(mktemp -d)"
git clone --depth=1 --branch "$SOURCE" "$REPO" source 2>/dev/null
git clone --depth=1 --branch "$TARGET" "$REPO" target 2>/dev/null

find "source/$SUBPATH" -name "values.yaml" | while read f; do
  rel="${f#source/}"
  if [ -f "target/$rel" ]; then
    echo "--- $rel ---"
    diff "target/$rel" "$f" || true
  fi
done
```

Make it executable: chmod +x scripts/helm-diff

Add a CI step in .github/workflows/ci.yaml for PRs:
```yaml
      - name: Helm diff
        if: github.event_name == 'pull_request'
        run: |
          ./scripts/helm-diff \
            --repository "${{ github.event.pull_request.head.repo.clone_url }}" \
            --source "${{ github.event.pull_request.head.ref }}" \
            --target "${{ github.event.pull_request.base.ref }}" \
            --subpath "system"
```

Commit: git add scripts/helm-diff .github/workflows/ci.yaml && git commit -m "ci: add helm-diff script to show PR value changes"

### P1-3: Chain ubiquity init → configure

In cmd/ubiquity/cmd/init.go, after creating the skeleton config, check if .env exists in the repo root. If not, offer to run `ubiquity configure -i`:

```go
// After creating skeleton, prompt to run config wizard
fmt.Print("Run configuration wizard now? [y/N]: ")
var response string
fmt.Scanln(&response)
if strings.ToLower(strings.TrimSpace(response)) == "y" {
    // Call configureCmd.RunE with --interactive flag
    configureCmd.Flags().Set("interactive", "true")
    return configureCmd.RunE(cmd, args)
}
```

Add "strings" to the imports.

Commit: git commit -m "feat: chain ubiquity init into configure wizard" -a

### P1-4: Add ubiquity info command

Create cmd/ubiquity/cmd/info.go with a cobra command that shows a summary of the cluster:

```go
var infoCmd = &cobra.Command{
    Use:   "info",
    Short: "Show cluster information summary",
    Long:  `Displays cluster version, K8s version, provisioning state, and installed components.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Show version info
        fmt.Printf("Ubiquity CLI: %s (commit: %s)\n", Version, Commit)
        
        // Try to load provisioning state
        state, _ := provision.LoadState()
        if state != nil {
            fmt.Printf("Environment: %s\n", state.Environment)
            fmt.Printf("Last updated: %s\n", state.UpdatedAt)
        }
        
        // Try kubectl version
        kubectlOut, err := exec.Command("kubectl", "version", "--short").Output()
        if err == nil {
            fmt.Printf("Kubernetes: %s", kubectlOut)
        }
        
        return nil
    },
}
```

Import "os/exec" and "github.com/ubiquitycluster/ubiquity/pkg/provision".

Commit: git add cmd/ubiquity/cmd/info.go && git commit -m "feat: add ubiquity info command for cluster summary"

### P1-5: Create CLI command reference doc

Create docs/reference/cli.md with auto-generated help text for all commands:

```markdown
# Ubiquity CLI Reference

## Overview

The `ubiquity` CLI is the primary entry point for managing Ubiquity clusters.
It replaces the legacy Makefile-based workflow.

## Installation

```bash
# From source
go install github.com/ubiquitycluster/ubiquity/cmd/ubiquity@latest

# With version info
make cli
sudo make install
```

## Commands
```

Then for each command (run `ubiquity COMMAND --help`), extract and include:
- Usage line
- Description
- Flags
- Example invocations

Include all 10+ commands: init, up, down, status, logs, retry, test, configure, version, info, integration.

Commit: git add docs/reference/cli.md && git commit -m "docs: add CLI command reference"

---

## P2 Items (medium priority)

### P2-1: Add mixed-line-ending and check-shebang pre-commit hooks

Add to .pre-commit-config.yaml:
```yaml
      - id: mixed-line-ending
        args: [--fix=lf]
      - id: check-shebang-scripts-are-executable
      - id: check-executables-have-shebangs
```

Commit: git commit -m "chore: add mixed-line-ending and shebang checks to pre-commit" -a

### P2-2: Add govulncheck to CI

Add to the test job in .github/workflows/ci.yaml:
```yaml
      - name: Govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
```

Commit: git commit -m "sec: add govulncheck vulnerability scanning to CI" -a

### P2-3: Create docs/FAQ.md

Create docs/FAQ.md with answers to common questions:

```markdown
# Frequently Asked Questions

## General

**Q: What is Ubiquity?**
A: An HPC cluster lifecycle platform using IaC and GitOps principles.

## Installation

**Q: What hardware do I need?**
A: See README.md — tested on 3× ThinkCentre M700 Tiny nodes.

**Q: Can I try it without hardware?**
A: Yes: `ubiquity up --sandbox` creates a local k3d cluster.

## Configuration

**Q: How do I change the domain?**
A: `ubiquity configure --domain mydomain.com`

**Q: How do I add a worker node?**
A: Update inventories and run `ubiquity retry metal`.

## Troubleshooting

**Q: A phase failed. How do I retry?**
A: `ubiquity status` shows failed phases, then `ubiquity retry <phase>`.

**Q: How do I tear down and start fresh?**
A: `ubiquity down` then `ubiquity up`.
```

Commit: git add docs/FAQ.md && git commit -m "docs: add FAQ page"

### P2-4: Add design tenets to architecture docs

Append a "Design Tenets" section to docs/architecture/overview.md:
- Idempotency everywhere
- Default secure
- Observable by default
- Tested in CI
- Single toolchain
- No generated artifacts in version control

Commit: git commit -m "docs: add design tenets to architecture overview" -a

---

## P3 Items (nice to have)

### P3-1: Add Ollama app

Create apps/ollama/ using the new app-template chart:
```yaml
# apps/ollama/Chart.yaml
apiVersion: v2
name: ollama
version: 0.1.0
dependencies:
  - name: app-template
    version: "0.1.0"
    repository: "file://../../charts/app-template"
```

```yaml
# apps/ollama/values.yaml
image:
  repository: ollama/ollama
  tag: latest

service:
  port: 11434
  targetPort: 11434

persistence:
  enabled: true
  size: 50Gi
  storageClass: longhorn

resources:
  limits:
    memory: 8Gi
    cpu: 4
  requests:
    memory: 4Gi
    cpu: 2
```

Commit: git add apps/ollama/ && git commit -m "feat: add Ollama app for local LLM inference"

### P3-2: Add ubiquity health command

Create cmd/ubiquity/cmd/health.go that checks:
- kubectl connectivity
- ArgoCD pods are running
- Core system components are healthy
- Storage is accessible

```go
var healthCmd = &cobra.Command{
    Use:   "health",
    Short: "Check cluster health",
    RunE: func(cmd *cobra.Command, args []string) error {
        checks := []struct{
            name string
            check func() error
        }{
            {"kubectl connectivity", func() error {
                return exec.Command("kubectl", "cluster-info").Run()
            }},
            {"ArgoCD server", func() error {
                return exec.Command("kubectl", "-n", "argocd", "get", "pod", "-l", "app.kubernetes.io/name=argocd-server").Run()
            }},
        }
        
        allPassed := true
        for _, c := range checks {
            fmt.Printf("  %s ... ", c.name)
            if err := c.check(); err != nil {
                fmt.Printf("FAIL (%v)\n", err)
                allPassed = false
            } else {
                fmt.Println("OK")
            }
        }
        
        if allPassed {
            fmt.Println("\nAll checks passed.")
        } else {
            fmt.Println("\nSome checks failed. Run 'ubiquity logs' for details.")
        }
        return nil
    },
}
```

Commit: git add cmd/ubiquity/cmd/health.go && git commit -m "feat: add ubiquity health command"

---

## Verification

After all items:
- go build ./... — PASS
- go test ./pkg/... ./cmd/... -count=1 — all green
- gitleaks detect --source . -v — no leaks found (or baseline created)
- kyverno test system/kyverno-policies/kyverno-test/ — all pass
- ubiquity version — prints version
- ubiquity info — shows cluster info
- ubiquity health — runs health checks
- ls docs/reference/architecture/decision-records/ADR-*.md — 8 files
- git log --oneline | wc -l — at least 15 new commits
