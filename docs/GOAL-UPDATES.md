Audit and improve dependency freshness and test coverage across the Ubiquity project at /home/ubuntu/ubiquity. Work through all items, commit after each.

## Current state

- **20+ Go indirect dependencies** outdated (charmbracelet, viper, cobra sub-deps)
- **19 of 24 Helm charts** use unpinned upstream versions (version: 0.0.0)
- **No Dependabot** — only Renovate, which has a TODO to switch to YAML format
- **CLI test coverage: 26.1%** — well below 50% target
- **Packages with 0 tests**: pkg/cloud, pkg/tui
- **Functions with 0% coverage**: all provision* phase executors, all run* test functions, Execute(), Env()
- **Helm unittest coverage**: only 4 of ~24 charts have tests
- **Zero Go vulnerability scans** in CI (govulncheck was added to CI config but may never have run)

---

## P0 Items (must do)

### P0-1: Create Dependabot configuration
GitHub Dependabot is free and runs automatically. Create .github/dependabot.yml:

```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: "/"
    schedule:
      interval: weekly
      day: monday
      time: "09:00"
    labels:
      - dependencies
      - go
    open-pull-requests-limit: 10

  - package-ecosystem: gomod
    directory: "/tools"
    schedule:
      interval: weekly
      day: monday
      time: "09:00"
    labels:
      - dependencies
      - go
    open-pull-requests-limit: 5

  - package-ecosystem: docker
    directory: "/"
    schedule:
      interval: weekly
    labels:
      - dependencies
      - docker

  - package-ecosystem: github-actions
    directory: "/"
    schedule:
      interval: weekly
    labels:
      - dependencies
      - ci

  - package-ecosystem: "docker"
    directory: "/opus/Dockerfiles"
    schedule:
      interval: weekly
    labels:
      - dependencies
      - docker
```

Commit: git add .github/dependabot.yml && git commit -m "ci: add Dependabot for Go, Docker, and GitHub Actions updates"

### P0-2: Upgrade Helm chart dependencies to pinned versions

Many charts use version: 0.0.0 (unpinned). Audit and pin each:

For each chart with a `dependencies:` section in Chart.yaml, read the current chart and find the latest available version. This is safest to do by reading the dependency repository metadata. For each chart:

1. Read Chart.yaml
2. Note the dependency name and repository
3. Keep the existing version (don't change upstream versions, just ensure they're meaningfully pinned)

The critical charts to audit:
- system/descheduler/Chart.yaml — dependency version 0.27.0
- system/falco/Chart.yaml — dependency version 0.5.9
- system/vault/Chart.yaml — dependency version 1.15.6
- system/ingress-nginx/Chart.yaml
- system/kured/Chart.yaml
- system/longhorn-system/Chart.yaml
- system/metallb-system/Chart.yaml
- system/cert-manager/Chart.yaml
- system/monitoring-system/Chart.yaml
- platform/gitea/Chart.yaml
- platform/keycloak/Chart.yaml
- platform/onyxia/Chart.yaml
- platform/external-secrets/Chart.yaml
- bootstrap/argocd/Chart.yaml

For each one:
1. Check if version is set to "0.0.0" — if so, update it to match the dependency version
2. If version already pinned, leave it

Example fix (descheduler):
```yaml
# Before
version: 0.0.0
dependencies:
  - name: descheduler
    version: 0.27.0
# After
version: 0.27.0
```

Commit: git commit -m "chore: pin Helm chart versions to match dependency versions" -a

### P0-3: Fix Renovate YAML format
Read renovate.json5. The TODO says "switch to YAML". Move it to renovate.json (JSON, not JSON5) with proper Helm chart dependency detection:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "config:recommended"
  ],
  "labels": ["dependencies"],
  "rebaseWhen": "conflicted",
  "schedule": ["before 9am on Monday"],
  "helm-values": {
    "enabled": true,
    "fileMatch": ["values\\.yaml$"]
  },
  "packageRules": [
    {
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "all non-major dependencies",
      "groupSlug": "all-minor-patch",
      "automerge": true
    },
    {
      "matchManagers": ["helm-values"],
      "enabled": true
    },
    {
      "matchManagers": ["dockerfile"],
      "enabled": true
    },
    {
      "matchManagers": ["gomod"],
      "enabled": true,
      "groupName": "Go modules",
      "groupSlug": "go-deps"
    },
    {
      "matchManagers": ["github-actions"],
      "enabled": true,
      "groupName": "GitHub Actions",
      "groupSlug": "github-actions"
    },
    {
      "matchUpdateTypes": ["major"],
      "automerge": false,
      "labels": ["major-update"]
    }
  ],
  "regexManagers": [
    {
      "description": "Update Helm chart dependencies in Chart.yaml",
      "fileMatch": ["Chart\\.yaml$"],
      "matchStrings": ["version:\\s*(?<currentValue>\\d+\\.\\d+\\.\\d+)\n"],
      "datasourceTemplate": "helm",
      "depNameTemplate": "{{packageName}}"
    }
  ]
}
```

Remove the old renovate.json5.

Commit: git rm renovate.json5 && git add renovate.json && git commit -m "chore: migrate Renovate from JSON5 to JSON, add Helm chart dep matching"

---

## P1 Items (high priority)

### P1-1: Add Go vulnerability audit to CI
README notes govulncheck was added to CI but it may not be running correctly. Add a standalone vuln check workflow:

Create .github/workflows/vulncheck.yaml:
```yaml
name: Vulnerability audit
on:
  schedule:
    - cron: "0 6 * * 1"  # Every Monday 6am
  workflow_dispatch: {}

jobs:
  vulncheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - name: Run govulncheck
        run: govulncheck ./...
      - name: Run govulncheck on tools
        working-directory: tools
        run: govulncheck ./...
```

Commit: git add .github/workflows/vulncheck.yaml && git commit -m "sec: add weekly Go vulnerability audit workflow"

### P1-2: Add helm unittest to ALL system charts
Only 4 of ~24 charts have helm unittest tests. Add tests for every chart in system/:

For each chart in system/*/ that doesn't already have a tests/ directory:
- Create tests/basic_test.yaml with:
  - Renders without error
  - Creates expected number of resources
  - Key resources have correct metadata

Charts that need tests:
- system/descheduler/
- system/falco/
- system/vault/
- system/cert-manager/
- system/ingress-nginx/
- system/kured/
- system/longhorn-system/
- system/metallb-system/
- system/monitoring-system/
- system/cvmfs-csi/
- system/k8up-operator/
- system/nvidia-network-operator/
- system/baremetal-operator-system/

For each chart, create tests/basic_test.yaml:
```yaml
suite: test {{CHART_NAME}} basic rendering
templates:
  - "*.yaml"
tests:
  - it: should render at least one resource
    asserts:
      - hasDocuments:
          count: 1
```

Run: find system/ -name "tests" -type d | wc -l to confirm count increased.

Commit: git add system/*/tests/ && git commit -m "test: add basic helm unittest for all system charts"

### P1-3: Add Go unit tests for provision* phase executors
The up.go file has 8 provision* functions (provisionMetal, provisionPXE, provisionBootstrap, decryptSopsSecrets, provisionSecurity, provisionExternal, provisionWait, provisionPostInstall) all at 0% coverage.

Create cmd/ubiquity/cmd/up_test.go additions:
```go
func TestProvisionMetalCallsProvider(t *testing.T) {
    mock := &provision.MockProvider{}
    origProvider := provider
    provider = mock
    defer func() { provider = origProvider }()

    err := provisionMetal("sandbox")
    if err != nil {
        t.Fatalf("provisionMetal failed: %v", err)
    }
    if len(mock.Calls) == 0 {
        t.Error("expected at least one provider call")
    }
}
```

Add tests for each phase executor that just verify they don't panic and call the provider.

Target: Get up.go coverage from 0% to >20% for these functions.

Commit: git commit -m "test: add unit tests for provision phase executors" -a

### P1-4: Add tests for pkg/tui
Create pkg/tui/status_test.go that tests:
- RenderStatus works with a provision.State
- RenderStatus returns expected strings for different states
- PrintStatus doesn't panic with nil state

Target: pkg/tui coverage >50%.

Commit: git add pkg/tui/status_test.go && git commit -m "test: add tests for TUI status rendering"

---

## P2 Items (medium priority)

### P2-1: Add age-keygen install check to CI
Add a step to the lint job that verifies age-keygen and sops are installable:
```yaml
      - name: Check sops tools
        run: |
          which age-keygen || echo "age not installed"
          which sops || echo "sops not installed"
```

Commit: git commit -m "ci: add sops/age tool check to CI" -a

### P2-2: Add go mod verify to CI
Add a step to verify go.sum is not tampered with:
```yaml
      - name: Verify Go modules
        run: go mod verify
```

Commit: git commit -m "ci: add go mod verify step" -a

### P2-3: Add scheduled dependency freshness report
Create .github/workflows/deps-report.yaml that runs weekly and comments on issues:
```yaml
name: Dependency report
on:
  schedule:
    - cron: "0 8 * * 1"
  workflow_dispatch: {}
jobs:
  outdated:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - name: Check outdated Go deps
        run: |
          echo "## Outdated Go dependencies" >> $GITHUB_STEP_SUMMARY
          go list -u -m all 2>/dev/null | grep "\[" >> $GITHUB_STEP_SUMMARY || echo "None" >> $GITHUB_STEP_SUMMARY
```

Commit: git add .github/workflows/deps-report.yaml && git commit -m "ci: add weekly dependency freshness report"

### P2-4: Add Helm chart freshness check
Create scripts/helm-outdated that checks Chart.yaml dependency versions against latest:
```bash
#!/bin/bash
set -euo pipefail
# Check Helm chart upstream versions
# Usage: ./scripts/helm-outdated

find . -name "Chart.yaml" -not -path "./.git/*" | while read f; do
  dir=$(dirname "$f")
  deps=$(grep -A2 "dependencies:" "$f" 2>/dev/null | grep -E "name:|version:|repository:" | paste - - - 2>/dev/null)
  if [ -n "$deps" ]; then
    echo "=== $dir ==="
    echo "$deps"
  fi
done
```
Make executable: chmod +x scripts/helm-outdated

Commit: git add scripts/helm-outdated && git commit -m "chore: add helm-outdated script for dependency freshness" -a

---

## Verification

After all items:
- go build ./... — PASS
- go test ./pkg/... ./cmd/... -count=1 — all green
- go test ./cmd/ubiquity/cmd/... -cover | grep -o '[0-9.]*%'  — should be >30%
- find system/ -name tests -type d | wc -l — should be >8 (was 4)
- kubectl get --raw /apis/customresourcedefinitions/... || true (no-op, just verifying yaml parsable)
- ls .github/dependabot.yml — exists
- ls .github/workflows/vulncheck.yaml — exists
- renovate.json — exists (renovate.json5 removed)
- git log --oneline -10 — 10+ commits
