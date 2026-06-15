# Plan 001: Make Helm CI fail closed

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before moving to the next step. If anything in the STOP conditions occurs, stop and report — do not improvise. When done, update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- .github/workflows/ci.yaml system/monitoring-system/tests/basic_test.yaml system/monitoring-system/Chart.yaml`
> If any listed files changed since this plan was written, compare the Current state excerpts against live code before proceeding; on mismatch, stop and report.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: MED
- **Depends on**: none
- **Category**: tests/security
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

The CI pipeline currently downgrades Helm lint and Helm unittest failures into warnings. That can let broken or unsafe charts merge, including charts with security policy or monitoring coverage. A local `helm unittest system/monitoring-system` failure is currently masked by the workflow pattern.

## Current state

- `.github/workflows/ci.yaml` — Helm lint masks failures:

```yaml
126|      - name: Helm lint all charts
127|        run: |
128|          find . -name Chart.yaml -not -path "./.git/*" | while read chart; do
129|            dir=$(dirname "$chart")
130|            echo "Linting: $dir"
131|            helm lint "$dir" || echo "WARNING: $dir lint failed"
132|          done
```

- `.github/workflows/ci.yaml` — Helm unittest masks failures:

```yaml
197|      - name: Helm unittest
198|        run: |
199|          go install github.com/helm-unittest/helm-unittest@latest 2>/dev/null || true
200|          find . -name "Chart.yaml" -not -path "./.git/*" | while read chart; do
201|            dir=$(dirname "$chart")
202|            if [ -d "$dir/tests" ]; then
203|              echo "Testing: $dir"
204|              helm unittest "$dir" || echo "WARNING: $dir tests failed"
205|            fi
206|          done
```

- Reproduction captured during planning:

```text
helm unittest system/monitoring-system
FAIL test monitoring-system basic rendering ... Expected documents count 1, Actual 0
```

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Reproduce current failure | `helm unittest system/monitoring-system` | initially fails until chart test is fixed or narrowed |
| Focused CI shell check | `bash -n .github/workflows/ci.yaml` | may not apply because YAML is not shell; use workflow review instead |
| Helm tests | `helm unittest system/monitoring-system monitoring/monitoring-system platform/ai-platform-console system/nvidia-nic-configuration-operator platform/ai-workload-tenancy system/longhorn-system` | exits 0 |
| Helm lint | `helm lint system/monitoring-system monitoring/monitoring-system platform/ai-platform-console system/nvidia-nic-configuration-operator platform/ai-workload-tenancy system/longhorn-system` | exits 0 |
| Full gate | `go test ./pkg/... ./cmd/... -count=1 && make test` | exits 0 |

Run from the repo root.

## Scope

**In scope**:
- `.github/workflows/ci.yaml` to make Helm lint/unittest fail closed while still reporting all failures.
- `system/monitoring-system/tests/basic_test.yaml` if needed to stop asserting resources that dependency defaults intentionally do not render.
- `system/monitoring-system/Chart.yaml` only if required for metadata/test compatibility.

**Out of scope**:
- Rewriting every legacy chart.
- Changing release credentials or GitHub environment settings.
- Making hardware/live tests mandatory.

## Git workflow

- Branch: `improve/001-helm-ci-fail-closed`
- Commit per logical step using the repo's commit style.
- Do not push or open a PR unless the operator instructed it.

## Steps

### Step 1: Make CI collect Helm lint failures and exit non-zero

Replace the `helm lint "$dir" || echo ...` pattern with a loop that accumulates failures in a variable or temp file and exits `1` after the loop if any chart failed. Keep printing each chart name.

**Verify**: inspect `.github/workflows/ci.yaml` and confirm no `helm lint ... || echo "WARNING` bypass remains.

### Step 2: Make CI collect Helm unittest failures and exit non-zero

Replace the `helm unittest "$dir" || echo ...` pattern with the same fail-closed accumulator. Avoid failing immediately on the first chart if broad reporting is useful, but the step must exit non-zero when any chart test fails.

**Verify**: inspect `.github/workflows/ci.yaml` and confirm no `helm unittest ... || echo "WARNING` bypass remains.

### Step 3: Fix or narrow `system/monitoring-system` unit tests

Run `helm unittest system/monitoring-system`. If it fails because a dependency-only wrapper enumerates disabled subchart templates, narrow the test to the wrapper's rendered resources or values contract rather than requiring every disabled dependency template to emit one document.

**Verify**: `helm unittest system/monitoring-system` exits 0.

### Step 4: Run focused and normal gates

Run the Helm lint/unittest commands above, then Go and Makefile test gates.

**Verify**: all commands exit 0.

## Test plan

- `helm unittest system/monitoring-system`
- `helm unittest monitoring/monitoring-system platform/ai-platform-console system/nvidia-nic-configuration-operator platform/ai-workload-tenancy system/longhorn-system`
- `helm lint system/monitoring-system monitoring/monitoring-system platform/ai-platform-console system/nvidia-nic-configuration-operator platform/ai-workload-tenancy system/longhorn-system`
- `go test ./pkg/... ./cmd/... -count=1`
- `make test`

## Done criteria

- [ ] CI Helm lint step exits non-zero on any chart lint failure.
- [ ] CI Helm unittest step exits non-zero on any chart test failure.
- [ ] `helm unittest system/monitoring-system` exits 0.
- [ ] Focused Helm lint/unittest commands exit 0.
- [ ] Go and Makefile test gates exit 0.
- [ ] No files outside in-scope list are modified.

## STOP conditions

Stop and report back if:

- CI intentionally allows a documented chart failure and no allowlist policy exists.
- Fixing `system/monitoring-system` requires changing chart runtime behavior rather than test expectations.
- A required command is unavailable and no documented repo alternative exists.
- Verification fails for unrelated dirty worktree changes.

## Maintenance notes

If some legacy charts must remain non-blocking, encode that as an explicit allowlist with chart names and comments. Do not use blanket `|| echo WARNING` around all charts.
