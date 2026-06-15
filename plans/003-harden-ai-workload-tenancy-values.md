# Plan 003: Harden AI workload tenancy values

> **Executor instructions**: Follow this plan step by step. Run every verification command and confirm the expected result before moving to the next step. If anything in the STOP conditions occurs, stop and report — do not improvise. When done, update the status row for this plan in `plans/README.md`.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- platform/ai-workload-tenancy/templates/gpu-quota.yaml platform/ai-workload-tenancy/values.schema.json platform/ai-workload-tenancy/tests/smoke_test.yaml platform/ai-workload-tenancy/values.yaml`
> If any listed files changed since this plan was written, compare the Current state excerpts against live code before proceeding; on mismatch, stop and report.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: MED
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

The AI workload tenancy chart renders tenant names directly into YAML scalar positions without quoting or validating the names as Kubernetes DNS labels. A malicious or malformed tenant value can inject extra YAML fields into rendered resources. This weakens a chart that is supposed to enforce multi-tenant isolation.

## Current state

- `platform/ai-workload-tenancy/templates/gpu-quota.yaml` — tenant names are unquoted in multiple metadata fields:

```yaml
4|apiVersion: v1
5|kind: Namespace
6|metadata:
7|  name: {{ $tenant.name }}
...
21|  name: ai-tenant-quota
22|  namespace: {{ $tenant.name }}
...
41|  name: ai-tenant-defaults
42|  namespace: {{ $tenant.name }}
```

The same pattern recurs for NetworkPolicy namespaces at lines 59, 71, 83, and 106.

- Reproduction captured during planning:

```text
helm template injected platform/ai-workload-tenancy -f malicious-values.yaml
kind: Namespace
metadata:
  name: evil
  annotations:
    injected: yes
```

The injected annotation came from a tenant name containing newline YAML.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Reproduce current injection | `helm template injected platform/ai-workload-tenancy -f /tmp/malicious-values.yaml` | initially renders injected YAML before fix |
| Schema validation | `helm lint platform/ai-workload-tenancy -f /tmp/malicious-values.yaml` | fails after schema hardening |
| Chart unit tests | `helm unittest platform/ai-workload-tenancy` | exits 0 |
| Chart lint | `helm lint platform/ai-workload-tenancy` | exits 0 |
| Go package checks | `go test ./pkg/... ./cmd/... -count=1` | exits 0 |

Run from the repo root.

## Scope

**In scope**:
- `platform/ai-workload-tenancy/templates/gpu-quota.yaml` for quoting tenant name scalar renderings.
- `platform/ai-workload-tenancy/values.schema.json` for Kubernetes DNS-label validation of tenant names and other string fields used in Kubernetes names if relevant.
- `platform/ai-workload-tenancy/tests/smoke_test.yaml` for a regression proving malicious names fail or safe names render correctly.
- `platform/ai-workload-tenancy/values.yaml` only if default values need schema-compatible adjustment.

**Out of scope**:
- Redesigning tenancy policy semantics.
- Changing tenant quota defaults.
- Adding admission controllers or runtime enforcement beyond Helm chart validation.

## Git workflow

- Branch: `improve/003-ai-workload-tenancy-value-hardening`
- Commit per logical step using the repo's commit style.
- Do not push or open a PR unless the operator instructed it.

## Steps

### Step 1: Add malicious value regression

Create a helm-unittest or lint/schema test fixture with a tenant name containing newline/colon content. The desired final behavior is schema validation failure before rendering.

Example malicious value:

```yaml
tenants:
  bad:
    name: "evil\n  annotations:\n    injected: yes"
    gpuQuota: 1
    rdmaQuota: 1
```

**Verify**: before the fix, `helm template` demonstrates injection or the test fails for the expected reason.

### Step 2: Validate tenant names as Kubernetes DNS labels

Update `values.schema.json` so tenant names must match a Kubernetes namespace-safe DNS label pattern:
- lowercase alphanumeric and `-`
- starts and ends with alphanumeric
- max length 63

Also inspect any other values interpolated into Kubernetes resource names or resource-key fragments and validate those if they are user-controlled.

**Verify**: `helm lint platform/ai-workload-tenancy -f /tmp/malicious-values.yaml` fails with schema validation.

### Step 3: Quote tenant name scalar renderings

Update `gpu-quota.yaml` so all tenant name scalar renderings use `| quote`, including Namespace metadata and every namespaced object. Quoting is defense-in-depth; schema validation is the primary prevention.

**Verify**: safe default render still contains the expected namespaces and quotas.

### Step 4: Run focused chart gates

Run chart lint and helm-unittest for `platform/ai-workload-tenancy`.

**Verify**: commands exit 0.

## Test plan

- Malicious tenant-name schema rejection fixture.
- Existing smoke test for default tenant resources.
- `helm lint platform/ai-workload-tenancy`
- `helm unittest platform/ai-workload-tenancy`
- `go test ./pkg/... ./cmd/... -count=1`

## Done criteria

- [ ] Malicious newline/annotation tenant name is rejected by schema/lint.
- [ ] Every `{{ $tenant.name }}` scalar in `gpu-quota.yaml` is quoted or otherwise safely rendered.
- [ ] Tenant name schema follows Kubernetes namespace naming constraints.
- [ ] Existing safe defaults render successfully.
- [ ] Focused Helm and Go gates exit 0.

## STOP conditions

Stop and report back if:

- Existing documented tenant names violate Kubernetes DNS-label rules and require a migration policy.
- Helm unittest cannot assert schema failure with the installed plugin version; use a shell-based `helm lint -f` regression instead and document why.
- Hardening reveals other user-controlled Kubernetes-name injection surfaces outside this chart.

## Maintenance notes

Keep schema validation and template quoting together. Quoting alone prevents YAML structure injection but can still render invalid Kubernetes object names; schema validation catches operator mistakes earlier.
