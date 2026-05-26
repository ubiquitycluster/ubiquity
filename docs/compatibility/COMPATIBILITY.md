# Ubiquity Kubernetes Compatibility Guide

## Overview

This document tracks which component versions are compatible with which
Kubernetes versions in the Ubiquity stack. The sandbox test matrix
(`test/k3d-matrix.sh`) validates these combinations automatically.

## Compatibility Matrix

| K8s Version | k3s Tag                 | ArgoCD    | Kyverno        | kyverno-policies |
|-------------|-------------------------|-----------|----------------|------------------|
| 1.30        | rancher/k3s:v1.30.14    | 6.7.17    | 3.5.3 (1.15)   | 0.1.0            |
| 1.31        | rancher/k3s:v1.31.14    | 6.7.17    | 3.5.3 (1.15)   | 0.1.0            |
| 1.32        | rancher/k3s:v1.32.13    | 6.7.17    | 3.8.1 (1.18)   | 0.1.0            |

See `k3s-helm-matrix.yaml` for the machine-readable version.

## Version Selection Logic

The `compat` package in `cmd/ubiquity/cmd/compat.go` detects the running
Kubernetes version and selects the correct chart version automatically.

### How it works

1. `detectKubeVersion()` calls `kubectl version -o json` and parses the
   server version's minor version number.
2. `lookupChartVersion(chartDir, kubeMinor)` looks up the compatibility
   matrix for the given chart and K8s minor version.
3. If an exact match exists, that chart version is used. If not, the
   highest pin that is <= the running version is selected as a fallback.
4. The chart version is passed to `helm install` / `helm upgrade` via
   `--version <chart-version>`.

### When version selection triggers

- **Kyverno operator**: K8s 1.30 and 1.31 need chart 3.5.3 (app v1.15.3)
  because chart 3.8.x uses `selectableFields` in CRDs, which requires
  K8s 1.29+. However, K8s 1.30 and 1.31 *do* have ValidatingAdmissionPolicy
  (beta in 1.28, stable in 1.30). K8s 1.32 gets the full 3.8.1 chart.
- **ArgoCD**: No version selection needed — upstream chart is compatible
  with all K8s 1.20+ versions.

## ADRs

### ADR-001: Kyverno chart pinning

**Status:** Active

**Context:** The upstream Kyverno chart v3.8.x includes `selectableFields`
in CRD definitions, a CRD feature stable since K8s 1.29. When running on
K8s 1.30 or 1.31, these CRDs are rejected because the K8s API server doesn't
recognize the field. Kyverno v3.5.x uses an older CRD format without
`selectableFields` and is compatible with K8s 1.28+.

**Decision:** Pin Kyverno chart version based on detected K8s version.
K8s < 1.32 → chart 3.5.3 (app 1.15.3). K8s >= 1.32 → chart 3.8.1 (app 1.18.1).

**Consequences:**
- K8s 1.30/1.31 users get an older Kyverno but still have full policy
  enforcement via ValidatingAdmissionPolicy.
- K8s 1.32 users get the latest Kyverno with all features.
- When K8s 1.30 and 1.31 reach EOL, the pin can be removed.

### ADR-002: validate.forbid → validate.deny migration

**Status:** Active

**Context:** Kyverno v1.12+ removed the `validate.forbid` syntax in favor
of `validate.deny` with JMESPath conditions. The `restricted-pod-security`
policy used `forbid` for host namespace rules.

**Decision:** Migrated from:
```yaml
validate:
  forbid:
    value:
      spec:
        hostNetwork: true
```
to:
```yaml
validate:
  deny:
    conditions:
      any:
      - key: "{{ request.object.spec.hostNetwork || 'false' }}"
        operator: Equals
        value: "true"
```

The `{{ "{{" }}` and `{{ "}}" }}` escaping ensures Helm's template engine
passes the Kyverno JMESPath expression through unmodified.

**Consequences:**
- Compatible with Kyverno v1.12+ (all versions we deploy).
- Slightly more verbose but more expressive — allows complex conditions.

### ADR-003: nodeSelector master → control-plane

**Status:** Active

**Context:** K8s 1.25 removed the `node-role.kubernetes.io/master` label
from control-plane nodes. Starting with K8s 1.32, using the label generates
deprecation warnings during `kubectl apply`.

**Decision:** Changed all `nodeSelector` entries from
`node-role.kubernetes.io/master: 'true'` to
`node-role.kubernetes.io/control-plane: 'true'`. This applies to all
charts in bootstrap/, system/, platform/, monitoring/, and apps/.

**Consequences:**
- All components now target the correct label on K8s 1.25+.
- K8s 1.24 and earlier clusters won't match the label (but we no longer
  support those for sandbox deployments).

## How to Add a New K8s Version

1. Add the image tag to `test/k3s-versions.txt`
2. Run `make test-k3d-matrix` to check compatibility
3. If a chart fails, identify the feature gap (e.g., missing CRD field)
4. Find a compatible chart version and add a pin to `compat.go`
5. Add entries to `k3s-helm-matrix.yaml` and this table
6. Run the matrix again to confirm all versions pass

## How to Verify After a Chart Update

1. Run `make test-k3d-matrix` — this tests the current code against all
   supported K8s versions
2. Check test/k3d-matrix-results/ for individual logs
3. Verify no new CRD validation errors, unknown flags, or RBAC denials
4. If a chart update breaks a previously working version, add a version pin
