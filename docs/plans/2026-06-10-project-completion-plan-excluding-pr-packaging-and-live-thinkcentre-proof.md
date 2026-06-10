# Project Completion Plan Excluding PR Packaging and Live ThinkCentre Proof

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Complete the remaining local, CI-safe project maturity work while explicitly excluding PR packaging/publishing and live ThinkCentre bare-metal proof.

**Architecture:** This plan converts the remaining roadmap into TDD slices: chart maturity, TODO debt reduction, CLI/root command coverage, NICo day-2 lifecycle proof, production operations documentation, Graphify maintenance, and generated/reference documentation freshness. Work should stay vendor-neutral, avoid proprietary branding/dependencies, and keep destructive or live-hardware behavior behind explicit safety gates.

**Tech Stack:** Go, Helm, helm-unittest, Bash, Markdown, Graphify, Kubernetes/k3d verification tooling.

**Explicitly Out of Scope:**
- Point 1: Packaging or splitting the 180 local commits into reviewable PRs.
- Point 6: Running live ThinkCentre/bare-metal deployment proof.

---

## Baseline State

The current audited baseline is:

- Branch: `main...origin/main [ahead 180]`
- Tracked graph: `graphify-out/graph.json`
- Graph report: `graphify-out/GRAPH_REPORT.md`
- Full Go suite passes: `go test ./pkg/... ./cmd/... -count=1`
- Generated Helm chart reference is current: `scripts/generate-helm-chart-reference.sh --check`
- Core services dry-run proof passes: `test/e2e/core-services-proof.sh --dry-run`
- Active charts: 60
- Active charts missing tests: 8
- Charts with placeholder version `0.0.0`: 2
- TODO/FIXME markers outside ignored/generated areas: 109 markers across 49 files

Before executing any task in this plan, run:

```bash
graphify query "current remaining implementation gaps chart maturity CLI coverage NICo production operations docs" --budget 2500
git status --short --branch
```

After any code or docs change, run:

```bash
graphify update .
```

and commit the resulting `graphify-out/` updates with the task commit when materially changed.

---

## Phase A: Helm Chart Maturity

### Task A1: Add a chart maturity contract test

**Objective:** Add a failing test that prevents active charts from missing tests or using placeholder versions.

**Files:**
- Create: `pkg/cloud/chart_maturity_contract_test.go`
- Read: chart directories under repository root

**Step 1: Write failing test**

Create `pkg/cloud/chart_maturity_contract_test.go` with a table-driven test that:

- walks all `Chart.yaml` files
- skips paths under `disabled/`, `.git/`, and `graphify-out/`
- requires each active chart directory to contain `tests/`
- requires `version` to exist and not equal `0.0.0`
- reports exact chart paths needing remediation

Suggested structure:

```go
package cloud_test

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "testing"
)

func TestActiveChartsHaveTestsAndNonPlaceholderVersions(t *testing.T) {
    root := repoRoot(t)
    versionRE := regexp.MustCompile(`(?m)^version:\\s*([^\\s]+)`)

    var missingTests []string
    var placeholderVersions []string

    err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        rel, err := filepath.Rel(root, path)
        if err != nil {
            return err
        }
        if d.IsDir() {
            switch {
            case rel == ".git", rel == "disabled", rel == "graphify-out":
                return filepath.SkipDir
            case strings.Contains(rel, string(filepath.Separator)+"disabled"+string(filepath.Separator)):
                return filepath.SkipDir
            }
            return nil
        }
        if filepath.Base(path) != "Chart.yaml" {
            return nil
        }

        dir := filepath.Dir(path)
        relDir, err := filepath.Rel(root, dir)
        if err != nil {
            return err
        }
        if _, err := os.Stat(filepath.Join(dir, "tests")); err != nil {
            missingTests = append(missingTests, relDir)
        }

        b, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        match := versionRE.FindSubmatch(b)
        if len(match) < 2 || string(match[1]) == "0.0.0" {
            placeholderVersions = append(placeholderVersions, relDir)
        }
        return nil
    })
    if err != nil {
        t.Fatalf("walk charts: %v", err)
    }

    if len(missingTests) > 0 {
        t.Fatalf("active charts missing tests directories: %v", missingTests)
    }
    if len(placeholderVersions) > 0 {
        t.Fatalf("active charts missing non-placeholder versions: %v", placeholderVersions)
    }
}

func repoRoot(t *testing.T) string {
    t.Helper()
    dir, err := os.Getwd()
    if err != nil {
        t.Fatal(err)
    }
    for {
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            return dir
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            t.Fatal("could not locate repo root")
        }
        dir = parent
    }
}
```

**Step 2: Verify failure**

Run:

```bash
go test ./pkg/cloud -run TestActiveChartsHaveTestsAndNonPlaceholderVersions -count=1
```

Expected: FAIL, listing these missing tests:

- `bootstrap/argocd`
- `charts/app-template`
- `storage/dragonfly-system`
- `monitoring/loki`
- `monitoring/monitoring-system`
- `apps/cloudcmd`
- `apps/hajimari`
- `apps/ollama`

and these placeholder versions:

- `bootstrap/root`
- `apps/hajimari`

**Step 3: Commit failing test**

```bash
git add pkg/cloud/chart_maturity_contract_test.go
git commit -m "test: add chart maturity contract"
```

---

### Task A2: Add minimal Helm unittest coverage for `bootstrap/argocd`

**Objective:** Give the ArgoCD bootstrap chart a smoke rendering test.

**Files:**
- Create: `bootstrap/argocd/tests/templates_test.yaml`
- Read: `bootstrap/argocd/Chart.yaml`
- Read: `bootstrap/argocd/templates/*`

**Step 1: Inspect rendered resources**

Run:

```bash
helm template bootstrap/argocd bootstrap/argocd | head -120
```

**Step 2: Create test file**

Create `bootstrap/argocd/tests/templates_test.yaml` asserting the chart renders at least its primary Application/ApplicationSet/namespace resources. Use exact `kind` and `metadata.name` values observed from the template command.

Skeleton:

```yaml
suite: bootstrap argocd rendering
chart:
  version: 0.1.0
templates:
  - templates/*.yaml
tests:
  - it: renders expected bootstrap resources
    asserts:
      - hasDocuments:
          count: 1
```

Adjust `hasDocuments` count and assertions to the actual chart output.

**Step 3: Verify**

Run:

```bash
helm unittest bootstrap/argocd
```

Expected: PASS.

**Step 4: Commit**

```bash
git add bootstrap/argocd/tests/templates_test.yaml
git commit -m "test: add bootstrap argocd chart unittest"
```

---

### Task A3: Add minimal Helm unittest coverage for `charts/app-template`

**Objective:** Ensure the reusable app template renders canonical Deployment/Service/Ingress/PVC behaviors.

**Files:**
- Create: `charts/app-template/tests/templates_test.yaml`
- Read: `charts/app-template/templates/*`
- Read: `charts/app-template/values.yaml`

**Step 1: Render with defaults**

```bash
helm template app-template charts/app-template
```

**Step 2: Write tests**

Assert the default chart renders expected base resources. Add a second test with inline values enabling ingress and persistence if supported.

**Step 3: Verify**

```bash
helm unittest charts/app-template
```

Expected: PASS.

**Step 4: Commit**

```bash
git add charts/app-template/tests/templates_test.yaml
git commit -m "test: add app-template chart unittest"
```

---

### Task A4: Add minimal Helm unittest coverage for application charts

**Objective:** Cover application charts that currently lack `tests/` directories.

**Files:**
- Create: `apps/cloudcmd/tests/templates_test.yaml`
- Create: `apps/hajimari/tests/templates_test.yaml`
- Create: `apps/ollama/tests/templates_test.yaml`

**Step 1: Render each chart**

```bash
helm template cloudcmd apps/cloudcmd
helm template hajimari apps/hajimari
helm template ollama apps/ollama
```

**Step 2: Write one smoke test per chart**

Each test should assert:

- expected number of rendered documents
- primary workload kind and name
- primary Service if present
- Ingress or PVC behavior if chart supports it

**Step 3: Verify each chart**

```bash
helm unittest apps/cloudcmd
helm unittest apps/hajimari
helm unittest apps/ollama
```

Expected: PASS.

**Step 4: Commit**

```bash
git add apps/cloudcmd/tests apps/hajimari/tests apps/ollama/tests
git commit -m "test: add application chart unittests"
```

---

### Task A5: Add minimal Helm unittest coverage for monitoring/storage charts

**Objective:** Cover the remaining monitoring and storage charts that lack tests.

**Files:**
- Create: `storage/dragonfly-system/tests/templates_test.yaml`
- Create: `monitoring/loki/tests/templates_test.yaml`
- Create: `monitoring/monitoring-system/tests/templates_test.yaml`

**Step 1: Render each chart**

```bash
helm template dragonfly-system storage/dragonfly-system
helm template loki monitoring/loki
helm template monitoring-system monitoring/monitoring-system
```

**Step 2: Write smoke tests**

Each test should assert the primary rendered resource kinds and names. If a chart is a wrapper or dependency-only chart, assert the wrapper-specific resources or expected dependency output.

**Step 3: Verify**

```bash
helm unittest storage/dragonfly-system
helm unittest monitoring/loki
helm unittest monitoring/monitoring-system
```

Expected: PASS.

**Step 4: Commit**

```bash
git add storage/dragonfly-system/tests monitoring/loki/tests monitoring/monitoring-system/tests
git commit -m "test: add monitoring and storage chart unittests"
```

---

### Task A6: Replace placeholder chart versions

**Objective:** Remove remaining `version: 0.0.0` placeholders.

**Files:**
- Modify: `bootstrap/root/Chart.yaml`
- Modify: `apps/hajimari/Chart.yaml`

**Step 1: Choose conservative chart versions**

Use `0.1.0` unless repo conventions indicate another version. Do not change `appVersion` unless it is also clearly placeholder and a source-backed upstream version is known.

**Step 2: Patch versions**

Replace:

```yaml
version: 0.0.0
```

with:

```yaml
version: 0.1.0
```

**Step 3: Verify contract now passes**

```bash
go test ./pkg/cloud -run TestActiveChartsHaveTestsAndNonPlaceholderVersions -count=1
```

Expected: PASS.

**Step 4: Refresh generated reference docs**

```bash
scripts/generate-helm-chart-reference.sh
scripts/generate-helm-chart-reference.sh --check
```

Expected: `docs/reference/helm-charts.md is current`.

**Step 5: Commit**

```bash
git add bootstrap/root/Chart.yaml apps/hajimari/Chart.yaml docs/reference/helm-charts.md
git commit -m "chore: replace placeholder chart versions"
```

---

### Task A7: Add CI guard for chart maturity

**Objective:** Ensure chart maturity contract runs in CI.

**Files:**
- Modify: `.github/workflows/ci.yaml`

**Step 1: Inspect current Go test job**

```bash
grep -n "go test" .github/workflows/ci.yaml
```

**Step 2: Ensure `go test ./pkg/... ./cmd/...` includes `pkg/cloud`**

If already present, no CI change is needed. If a narrower package list is used, add:

```yaml
- name: Chart maturity contract
  run: go test ./pkg/cloud -run TestActiveChartsHaveTestsAndNonPlaceholderVersions -count=1
```

**Step 3: Verify**

```bash
go test ./pkg/cloud -run TestActiveChartsHaveTestsAndNonPlaceholderVersions -count=1
git diff --check
```

Expected: PASS.

**Step 4: Commit if workflow changed**

```bash
git add .github/workflows/ci.yaml
git commit -m "ci: enforce chart maturity contract"
```

Skip commit if no workflow change was needed.

---

## Phase B: TODO/FIXME Debt Reduction

### Task B1: Add a TODO debt inventory document

**Objective:** Convert scattered TODOs into a reviewed, categorized backlog so markers can be removed or justified.

**Files:**
- Create: `docs/plans/2026-06-10-todo-debt-reduction.md`

**Step 1: Generate current inventory**

```bash
python3 - <<'PY'
from pathlib import Path
import re
root = Path('.')
skip = {'.git', 'graphify-out', 'disabled', 'node_modules', '.venv', 'vendor'}
for p in sorted(root.rglob('*')):
    if not p.is_file():
        continue
    if any(part in skip for part in p.parts):
        continue
    if p.suffix.lower() in {'.png','.jpg','.jpeg','.gif','.svg','.ico','.lock'}:
        continue
    text = p.read_text(errors='ignore')
    hits = [(i, line.strip()) for i, line in enumerate(text.splitlines(), 1) if re.search(r'\\b(TODO|FIXME)\\b', line, re.I)]
    if hits:
        print(p)
        for line_no, line in hits:
            print(f'  {line_no}: {line}')
PY
```

**Step 2: Write backlog categories**

Categorize markers into:

- production runbook gaps
- secret/credential handling gaps
- chart/config implementation gaps
- stale generated/update docs
- code TODOs requiring tests
- acceptable upstream TODO comments

**Step 3: Commit**

```bash
git add docs/plans/2026-06-10-todo-debt-reduction.md
git commit -m "docs: inventory todo debt reduction plan"
```

---

### Task B2: Resolve secret-handling TODOs in metal defaults

**Objective:** Remove ambiguity around `k3s_encryption_secret` and ensure production docs do not preserve inline secrets.

**Files:**
- Modify: `metal/group_vars/metal.yml`
- Modify: `metal/roles/k3s/defaults/main.yml`
- Modify or create tests under relevant Go/package or scripts if existing validation exists
- Modify docs if needed: `docs/admin-guide/deployment/production/configuration.md`

**Step 1: Inspect current values**

```bash
grep -RIn "k3s_encryption_secret\|TODO\|FIXME" metal/group_vars/metal.yml metal/roles/k3s/defaults/main.yml
```

**Step 2: Write a failing validation test or script check**

If a config validation test already exists, extend it. Otherwise add a small Go test in `pkg/cloud` that scans these files and rejects live-looking secret defaults outside documented fixture contexts.

Acceptance behavior:

- no raw production secret value committed
- placeholder fixture values are clearly labeled non-production
- production path points to vault/SOPS/external secret flow

**Step 3: Update files**

Replace TODO text with precise non-production fixture wording or external secret reference. Use `[REDACTED]` if removing any sensitive-looking value.

**Step 4: Verify**

```bash
go test ./pkg/... ./cmd/... -count=1
git diff --check
```

Expected: PASS.

**Step 5: Commit**

```bash
git add metal/group_vars/metal.yml metal/roles/k3s/defaults/main.yml docs/admin-guide/deployment/production/configuration.md pkg/cloud
git commit -m "chore: clarify metal secret handling defaults"
```

---

### Task B3: Resolve platform configuration TODOs with tests

**Objective:** Convert key platform TODOs into explicit behavior or documented non-goals.

**Files:**
- Modify: `platform/gitea/files/config/main.go`
- Modify: `platform/gitea/files/config/config.yaml`
- Modify: `platform/harbor/templates/harbor-config-overwrite-secret.yaml`
- Modify: `platform/external-secrets/templates/clustersecretstore.yaml`
- Modify: `platform/onyxia/values.yaml`
- Create or modify relevant tests under `pkg/cloud` or chart tests

**Step 1: Inspect TODOs**

```bash
grep -RIn "TODO\|FIXME" platform/gitea/files/config platform/harbor/templates platform/external-secrets/templates platform/onyxia/values.yaml
```

**Step 2: For each TODO, choose one resolution**

- implement the behavior with a test
- convert to a documented limitation with no TODO marker
- move to `docs/plans/2026-06-10-todo-debt-reduction.md` as tracked backlog

**Step 3: Add regression tests**

For charts, prefer Helm unittests. For Go config generation, add Go unit tests near the package or a contract test that validates expected output.

**Step 4: Verify**

```bash
helm unittest platform/gitea || true
helm unittest platform/harbor || true
helm unittest platform/external-secrets || true
go test ./pkg/... ./cmd/... -count=1
git diff --check
```

If a chart does not support `helm unittest` yet, document why in the commit message and cover it with a contract test.

**Step 5: Commit**

```bash
git add platform docs/plans/2026-06-10-todo-debt-reduction.md pkg/cloud
git commit -m "chore: resolve platform configuration todo debt"
```

---

## Phase C: CLI and Root Command Coverage

### Task C1: Add root command smoke tests

**Objective:** Remove the `cmd/ubiquity [no test files]` gap by adding root package tests.

**Files:**
- Create: `cmd/ubiquity/main_test.go`
- Read: `cmd/ubiquity/main.go`

**Step 1: Inspect root package**

```bash
sed -n '1,200p' cmd/ubiquity/main.go
```

**Step 2: Write smoke test**

Add tests that verify the package compiles and any exported root setup behavior is safe. If `main.go` only calls into `cmd.Execute()`, keep the test minimal and non-invasive.

Example:

```go
package main

import "testing"

func TestMainPackageCompiles(t *testing.T) {
    // This test intentionally documents that executable wiring lives in main,
    // while command behavior is covered in cmd/ubiquity/cmd tests.
}
```

Prefer a stronger test if root command construction is exported.

**Step 3: Verify**

```bash
go test ./cmd/ubiquity -count=1
```

Expected: PASS instead of `[no test files]`.

**Step 4: Commit**

```bash
git add cmd/ubiquity/main_test.go
git commit -m "test: add ubiquity root package smoke test"
```

---

### Task C2: Add CLI command registry coverage

**Objective:** Prove expected top-level commands are registered and destructive commands remain gated.

**Files:**
- Modify: `cmd/ubiquity/cmd/*_test.go`
- Read: `cmd/ubiquity/cmd/root.go`
- Read: command files under `cmd/ubiquity/cmd/`

**Step 1: Find root command constructor**

```bash
grep -RIn "func .*Command\|Use:" cmd/ubiquity/cmd | head -80
```

**Step 2: Write tests**

Add tests that assert expected command names exist, for example:

- `init`
- `up`
- `down`
- `test`
- `status`
- any NICo/node lifecycle commands present in the current code

Also assert destructive/live commands require explicit flags/env where applicable.

**Step 3: Verify**

```bash
go test ./cmd/ubiquity/cmd -run 'Test.*Command' -count=1
```

Expected: PASS.

**Step 4: Commit**

```bash
git add cmd/ubiquity/cmd
git commit -m "test: cover CLI command registry"
```

---

### Task C3: Add provisioning executor regression tests

**Objective:** Improve coverage around provisioning routing and sandbox/dev/cloud behavior.

**Files:**
- Modify: `cmd/ubiquity/cmd/up_test.go` or create focused test file
- Read: `cmd/ubiquity/cmd/up.go`

**Step 1: Inspect provisioning executor seams**

```bash
grep -n "executor\|sandbox\|cloud\|dev" cmd/ubiquity/cmd/up.go | head -120
```

**Step 2: Write failing tests**

Tests should cover:

- sandbox path does not require live cloud credentials
- cloud/live path fails closed without explicit live configuration
- local command execution uses injectable runners rather than shell interpolation
- destructive paths require explicit user intent

**Step 3: Implement minimal seams if needed**

If tests cannot inject behavior, introduce small interfaces/functions only where needed. Avoid large refactors.

**Step 4: Verify**

```bash
go test ./cmd/ubiquity/cmd -run 'Test.*Provision|Test.*Sandbox|Test.*Live' -count=1
go test ./pkg/... ./cmd/... -count=1
```

Expected: PASS.

**Step 5: Commit**

```bash
git add cmd/ubiquity/cmd
git commit -m "test: cover provisioning executor safety paths"
```

---

## Phase D: NICo Day-2 Lifecycle Proof Without Live ThinkCentre Execution

### Task D1: Add NICo lifecycle contract tests

**Objective:** Prove the repository models NICo as default day-2 node management and keeps BMO/Metal3 fallback-only.

**Files:**
- Create: `pkg/nico/lifecycle_contract_test.go` or extend existing `pkg/nico` tests
- Read: `pkg/nico/*`
- Read: `docs/plans/2026-06-09-nico-day2-lifecycle-proof.md`

**Step 1: Write failing tests**

Tests should assert docs/configs expose:

- NICo as default day-2 physical node lifecycle path
- live node/status reconciliation concepts
- multiple bootable OS images
- destructive-action safety boundaries
- BMO/Metal3 as fallback-only unless explicitly requested

**Step 2: Verify failure**

```bash
go test ./pkg/nico -run Test.*Lifecycle -count=1
```

Expected: FAIL until docs/config/code satisfy the contract.

**Step 3: Implement or document minimal behavior**

Update NICo package/docs to satisfy contracts without performing live hardware actions.

**Step 4: Verify**

```bash
go test ./pkg/nico -run Test.*Lifecycle -count=1
go test ./pkg/... ./cmd/... -count=1
```

Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/nico docs/plans/2026-06-09-nico-day2-lifecycle-proof.md
git commit -m "test: codify NICo lifecycle proof contracts"
```

---

### Task D2: Add NICo dry-run proof script

**Objective:** Provide a CI-safe proof path for NICo lifecycle workflows without live hardware.

**Files:**
- Create: `test/e2e/nico-lifecycle-proof.sh`
- Modify: `.github/workflows/ci.yaml` if appropriate
- Modify docs: `docs/admin-guide/runbooks/nvidia-infra-controller/nico-bmc-redfish.md`

**Step 1: Write dry-run script**

The script should:

- set `set -euo pipefail`
- refuse to run destructive/live actions unless explicit env variables are set
- validate expected docs/config files exist
- run relevant Go tests
- optionally render Helm/Kustomize resources if NICo charts/manifests exist

**Step 2: Verify locally**

```bash
bash test/e2e/nico-lifecycle-proof.sh --dry-run
```

Expected: PASS with clear proof output.

**Step 3: Add CI hook if runtime is low-risk**

Add a workflow step only if it is deterministic and does not require hardware credentials.

**Step 4: Commit**

```bash
git add test/e2e/nico-lifecycle-proof.sh .github/workflows/ci.yaml docs/admin-guide/runbooks/nvidia-infra-controller/nico-bmc-redfish.md
git commit -m "test: add NICo lifecycle dry-run proof"
```

---

## Phase E: Production Operations Documentation

### Task E1: Replace cert-manager runbook TODO with actionable runbook

**Objective:** Produce a usable cert-manager operations runbook.

**Files:**
- Modify: `docs/admin-guide/runbooks/cert-manager.md`

**Required sections:**

- scope and assumptions
- quick health checks
- certificate renewal checks
- issuer/clusterissuer debugging
- DNS-01/HTTP-01 troubleshooting if applicable
- rollback/safe remediation
- evidence to capture for support/review
- explicit secret handling guidance with no raw credentials

**Verification:**

```bash
grep -n "TODO\|FIXME" docs/admin-guide/runbooks/cert-manager.md && exit 1 || true
git diff --check
```

**Commit:**

```bash
git add docs/admin-guide/runbooks/cert-manager.md
git commit -m "docs: add cert-manager operations runbook"
```

---

### Task E2: Replace Vault runbook TODO with actionable runbook

**Objective:** Produce a usable Vault operations runbook.

**Files:**
- Modify: `docs/admin-guide/runbooks/vault.md`

**Required sections:**

- scope and supported deployment assumptions
- bootstrap and initialization boundaries
- unseal/recovery guidance
- Kubernetes auth checks
- policy/token lifecycle
- generated secret handling
- backup/restore considerations
- incident response and credential rotation
- no committed raw secrets

**Verification:**

```bash
grep -n "TODO\|FIXME" docs/admin-guide/runbooks/vault.md && exit 1 || true
git diff --check
```

**Commit:**

```bash
git add docs/admin-guide/runbooks/vault.md
git commit -m "docs: add Vault operations runbook"
```

---

### Task E3: Complete deployment production docs

**Objective:** Replace deployment TODOs with production-ready guidance that remains evidence-bounded.

**Files:**
- Modify: `docs/admin-guide/deployment/external-resources.md`
- Modify: `docs/admin-guide/deployment/post-installation.md`
- Modify: `docs/admin-guide/deployment/production/external-resources.md`
- Modify: `docs/admin-guide/deployment/production/post-installation.md`
- Modify: `docs/admin-guide/deployment/production/configuration.md`

**Required coverage:**

- external DNS/load balancer dependencies
- credential ownership and rotation
- post-install health checks
- backup/restore expectations
- upgrade/rollback expectations
- production-lite resource caveats where applicable
- clear separation between dry-run/local proof and live production proof

**Verification:**

```bash
grep -RIn "TODO\|FIXME" docs/admin-guide/deployment docs/admin-guide/deployment/production || true
git diff --check
```

Expected: no TODO/FIXME in touched production docs unless they point to a tracked plan item.

**Commit:**

```bash
git add docs/admin-guide/deployment docs/admin-guide/deployment/production
git commit -m "docs: complete production deployment operations guidance"
```

---

### Task E4: Complete PXE and administration tutorial TODOs

**Objective:** Replace remaining user-facing setup/tutorial TODOs.

**Files:**
- Modify: `docs/admin-guide/concepts/pxe-boot.md`
- Modify: `docs/admin-guide/administration/tutorials/install-pre-commit-hooks.md`
- Modify: `docs/admin-guide/administration/tutorials/updating-documentation.md`

**Verification:**

```bash
grep -RIn "TODO\|FIXME" docs/admin-guide/concepts/pxe-boot.md docs/admin-guide/administration/tutorials || true
git diff --check
```

Expected: no unresolved TODO/FIXME in touched files unless explicitly moved to the TODO debt plan.

**Commit:**

```bash
git add docs/admin-guide/concepts/pxe-boot.md docs/admin-guide/administration/tutorials
git commit -m "docs: complete PXE and administration tutorials"
```

---

## Phase F: Graphify Maintenance

### Task F1: Add Graphify workflow documentation to the repo

**Objective:** Document the project rule that codebase questions and code changes use the tracked graph first.

**Files:**
- Modify: `AGENTS.md`
- Create or modify: `docs/developers/graphify-workflow.md`
- Modify: `docs/index.md` if appropriate

**Required content:**

- check `graphify-out/graph.json` before answering project/codebase questions
- use `graphify query`, `graphify explain`, and `graphify path`
- run `graphify update .` after code changes
- commit material graph updates
- do not generate HTML visualization for graphs over 5,000 nodes without explicit approval
- show token cost in Graphify reports

**Verification:**

```bash
graphify query "Graphify workflow project questions update graph" --budget 1500
git diff --check
```

**Commit:**

```bash
git add AGENTS.md docs/developers/graphify-workflow.md docs/index.md
git commit -m "docs: document Graphify-first project workflow"
```

---

### Task F2: Add a Graphify freshness check

**Objective:** Make stale graph state visible in local/CI-safe checks.

**Files:**
- Create: `scripts/check-graphify-freshness.sh`
- Modify: `.github/workflows/ci.yaml` if appropriate
- Modify: `docs/developers/graphify-workflow.md`

**Step 1: Write script**

The script should:

- verify `graphify-out/graph.json` exists
- read the graph/report commit if available
- compare to `git rev-parse HEAD`
- warn or fail based on a flag, for example `--strict`
- avoid expensive extraction

**Step 2: Verify**

```bash
bash scripts/check-graphify-freshness.sh
bash scripts/check-graphify-freshness.sh --strict || true
```

**Step 3: Add CI warning-only check unless strict behavior is desired**

Prefer warning-only initially to avoid blocking unrelated changes.

**Step 4: Commit**

```bash
git add scripts/check-graphify-freshness.sh .github/workflows/ci.yaml docs/developers/graphify-workflow.md
git commit -m "chore: add Graphify freshness check"
```

---

## Phase G: Generated and Reference Documentation Freshness

### Task G1: Add generated-docs freshness contract

**Objective:** Ensure generated references remain current after chart/config changes.

**Files:**
- Create or modify: `pkg/cloud/generated_docs_contract_test.go`
- Read: `scripts/generate-helm-chart-reference.sh`
- Read: `docs/reference/helm-charts.md`

**Step 1: Write failing/stability test**

Add a Go test or script test that runs:

```bash
scripts/generate-helm-chart-reference.sh --check
```

If Go tests should not invoke scripts, document why and ensure CI invokes the script directly.

**Step 2: Verify**

```bash
scripts/generate-helm-chart-reference.sh --check
go test ./pkg/cloud -run Test.*Generated.*Docs -count=1
```

Expected: PASS.

**Step 3: Commit**

```bash
git add pkg/cloud/generated_docs_contract_test.go .github/workflows/ci.yaml
git commit -m "test: enforce generated docs freshness"
```

---

### Task G2: Refresh and validate all docs touched by this plan

**Objective:** Finish with a clean generated-doc/reference state.

**Files:**
- Modify as generated: `docs/reference/helm-charts.md`
- Modify as generated: `graphify-out/*`

**Step 1: Regenerate references**

```bash
scripts/generate-helm-chart-reference.sh
scripts/generate-helm-chart-reference.sh --check
```

**Step 2: Update graph after all changes**

```bash
graphify update .
graphify query "completed chart maturity CLI coverage NICo production operations docs" --budget 2000
```

**Step 3: Verify no unintended secrets**

```bash
grep -RInE "BEGIN (RSA|OPENSSH|EC|DSA) PRIVATE KEY|AKIA[0-9A-Z]{16}|aws_secret_access_key|api[_-]?key: .+|password: .+|token: .+" docs pkg cmd platform system metal graphify-out || true
```

Manually inspect any matches. Redact real secret values as `[REDACTED]`; do not redact harmless field names.

**Step 4: Commit**

```bash
git add docs/reference/helm-charts.md graphify-out
git commit -m "docs: refresh generated references and graph"
```

---

## Final Verification Gate

Run all of the following before considering the plan complete:

```bash
graphify query "project completion chart maturity CLI coverage NICo production operations docs Graphify maintenance" --budget 3000
go test ./pkg/... ./cmd/... -count=1
scripts/generate-helm-chart-reference.sh --check
test/e2e/core-services-proof.sh --dry-run
bash test/e2e/nico-lifecycle-proof.sh --dry-run
git diff --check
git status --short --branch
```

Expected outcomes:

- Graphify query returns updated completion/maturity nodes.
- Full Go suite passes.
- Helm chart reference is current.
- Core services proof passes.
- NICo lifecycle dry-run proof passes.
- No whitespace errors.
- Working tree is clean after final commit.

---

## Acceptance Criteria

This plan is complete when all in-scope remaining work is covered by tests/docs/proofs:

- All 60 active charts have `tests/` directories.
- No active chart uses `version: 0.0.0`.
- Chart maturity is enforced by tests/CI.
- TODO/FIXME debt is reduced or converted into a tracked backlog with no ambiguous production TODOs in critical docs/configs.
- `cmd/ubiquity` root package no longer reports `[no test files]`.
- CLI command registration and provisioning safety paths have regression tests.
- NICo day-2 lifecycle behavior has CI-safe contracts and a dry-run proof.
- Production cert-manager and Vault runbooks are actionable.
- Deployment production docs have no unresolved high-risk TODOs.
- Graphify-first workflow is documented in the repo.
- Graphify freshness checking exists.
- Generated/reference docs freshness is enforced.
- `graphify-out/` is updated after implementation and committed.

---

## Suggested Execution Order

1. Phase A: Helm chart maturity.
2. Phase C: CLI/root command coverage.
3. Phase D: NICo dry-run proof.
4. Phase E: production operations docs.
5. Phase B: TODO debt reduction across remaining config/docs.
6. Phase F: Graphify maintenance docs/checks.
7. Phase G: final generated docs and graph refresh.

This order front-loads CI-safe test coverage and leaves documentation cleanup and generated artifacts for the end.
