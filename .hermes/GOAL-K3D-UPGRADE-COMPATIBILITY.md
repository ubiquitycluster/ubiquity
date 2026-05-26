# K3d Sandbox Upgrade & Chart Compatibility Goal

> **For Hermes:** Execute phases in order, commit after each phase. Use `subagent-driven-development` for parallel workstreams where possible.

**Goal:** Upgrade the k3d sandbox from k3s v1.23.4 to the latest stable (v1.32.13-k3s1), add a multi-version test harness covering the last 3 k3s minor releases, fault-fix incompatible Helm charts, build a compatibility matrix, and template chart selection to auto-adapt to the Kubernetes version.

**Prerequisites:** kubectl, helm, k3d v5.8.3, Docker, Go toolchain, git

**Repo:** `/home/ubuntu/ubiquity`

---

## Execution Approach

Work in phases, making a git commit after each phase completes.

Phase 1: Upgrade k3s version + verify ubiquity up works
Phase 2: Multi-version test harness (last 3 k3s releases)
Phase 3: Chart compatibility matrix + automated version selection
Phase 4: Test all 3 k3s releases end-to-end

---

## What To Build

### 1. K3s Version Upgrade & Smoke Test

**Upgrade `metal/k3d-dev.yaml`:**
- Change `image: docker.io/rancher/k3s:v1.23.4-k3s1` to `docker.io/rancher/k3s:v1.32.13-k3s1`
- Delete the existing k3d cluster (`k3d cluster delete ubiquity-dev`)
- Run `ubiquity up --sandbox` and verify it completes without errors

**Expected behavior after upgrade:**
- All redis-ha pods schedule (single replicas from values-sandbox.yaml) — already working
- Kyverno operator installs successfully via `system/kyverno` chart wrapper (K8s 1.32 supports VAP)
- `kyverno-policies` ClusterPolicy resources are created successfully
- `kubectl wait --for=condition=Ready pod --all` in provisionWait works without `--ignore-not-found`
- ArgoCD CRDs (`applications.argoproj.io`, `applicationsets.argoproj.io`) get established

### 2. Multi-Version Test Harness

**Create `Makefile` targets** in `test/`:

```
.PHONY: test-k3d-matrix
test-k3d-matrix: test-k3d-v1.30 test-k3d-v1.31 test-k3d-v1.32

.PHONY: test-k3d-v1.30
test-k3d-v1.30:
    K3S_IMAGE=rancher/k3s:v1.30.14-k3s2 k3d cluster create ubiquity-dev --config metal/k3d-dev.yaml
    # run ubiquity up --sandbox --skip-security
    # verify exit code 0
    k3d cluster delete ubiquity-dev

.PHONY: test-k3d-v1.31
test-k3d-v1.31: (same pattern, image: rancher/k3s:v1.31.14-k3s1)

.PHONY: test-k3d-v1.32
test-k3d-v1.32: (same pattern, image: rancher/k3s:v1.32.13-k3s1)
```

**Create `test/k3d-matrix.sh`** — a shell script that:
1. Reads a list of k3s versions from `test/k3s-versions.txt`
2. Creates a k3d cluster for each version using a template config
3. Runs `ubiquity up --sandbox` against it
4. Captures stdout/stderr to `test/k3d-matrix-results/`
5. Verifies exit code and checks for specific success markers
6. Destroys the cluster
7. Produces a summary table at the end

**Create `test/k3s-versions.txt`**:
```
rancher/k3s:v1.30.14-k3s2
rancher/k3s:v1.31.14-k3s1
rancher/k3s:v1.32.13-k3s1
```

**Create `test/k3d-matrix-config.yaml`** — templatized k3d config where `K3S_IMAGE` env var sets the k3s image:

```yaml
apiVersion: k3d.io/v1alpha4
kind: Simple
metadata:
  name: ubiquity-dev-${K3S_VERSION_TAG}
image: ${K3S_IMAGE}
servers: 1
agents: 1
options:
  k3s:
    extraArgs:
      - arg: --disable=traefik
        nodeFilters:
          - server:*
      - arg: --disable-cloud-controller
        nodeFilters:
          - server:*
ports:
  - port: 80:80
    nodeFilters:
      - loadbalancer
  - port: 443:443
    nodeFilters:
      - loadbalancer
```

**Test harness success criteria per version:**
- `ubiquity up --sandbox` exits with code 0
- ArgoCD pods all become Ready within 5 minutes
- Kyverno pods all become Ready (when applicable for K8s 1.28+)
- Kyverno ClusterPolicy resources get created
- Redis-ha server statefulset has 1/1 ready
- No "unknown flag" errors in output
- No "resource mapping not found" errors (CRD validation)

### 3. Compatibility Matrix

**Create `docs/compatibility/k3s-helm-matrix.yaml`** with this format:

```yaml
# Ubiquity Helm Chart Compatibility Matrix
# Each entry records which component versions work with which K8s versions.
k3s_versions:
  v1.30.14:
    components:
      argo-cd:
        chart_version: "6.7.17"
        app_version: "v2.10.8"
        notes: ""
      kyverno:
        chart_version: "3.5.3"
        app_version: "v1.15.3"
        notes: "3.8.x requires selectableFields (K8s 1.29+); pinned to 3.5.x"
      kyverno-policies:
        status: "works"
        notes: "Requires Kyverno CRDs installed first"
  v1.31.14:
    ...
  v1.32.13:
    ...
```

**Also create `docs/compatibility/COMPATIBILITY.md`** with human-readable documentation of:
- Which component charts are pinned and why
- The K8s feature gates each pin relies on
- How to add a new K8s version to the matrix
- How to verify compatibility after a chart update
- ADR-format entries for each pinning decision

### 4. Templated Helm Chart Version Selection

**Problem:** Currently each Helm chart in `bootstrap/` and `system/` has its dependency version hardcoded in `Chart.yaml`. When the k3s version changes, some charts need different version pins (e.g., kyverno 3.8.x requires K8s 1.29+).

**Solution:** Create a version selection mechanism that reads the k3s version and picks the right chart version.

**Approach A (preferred — template the Chart.yaml):**
- Create `test/k3d-matrix.sh` which sets `K3S_IMAGE` and runs the full harness
- The harness reads `k3s-versions.txt`, extracts the minor version, and passes it to Helm via `--set global.kubeVersionOverride=<detected>`
- Each chart that needs version selection uses `{{ .Capabilities.KubeVersion }}` in its templates
- OR: subcharts declare version ranges in their `condition` field

**Approach B (simpler — go code):**
- Modify `provisionBootstrap()` and `provisionSecurity()` in `cmd/ubiquity/cmd/up.go`
- Before installing each chart, detect the K8s version via `kubectl version -o json`
- Look up the compatible chart version in the compatibility matrix
- Pass `--version <compatible-version>` to `helm install`/`helm upgrade`

**Implement Approach B** since it doesn't require restructuring chart dependencies:

```go
// detectKubeVersion returns the K8s minor version (e.g., 1.31, 1.32)
func detectKubeVersion() (float64, error) {
    out, err := kubectlOutput("version", "-o", "json")
    if err != nil {
        return 0, fmt.Errorf("detecting k8s version: %w", err)
    }
    var v struct {
        ServerVersion struct {
            Major string `json:"major"`
            Minor string `json:"minor"`
        } `json:"serverVersion"`
    }
    if err := json.Unmarshal(out, &v); err != nil {
        return 0, fmt.Errorf("parsing k8s version: %w", err)
    }
    minor, _ := strconv.Atoi(v.ServerVersion.Minor)
    return float64(minor), nil
}
```

Then in `provisionBootstrap` and `provisionSecurity`, call this before `runHelmInstall` and pass `--version` for charts with multiple compatibility paths.

**Create `cmd/ubiquity/cmd/compat.go`** — holds the compatibility matrix in Go code:

```go
package cmd

type ChartCompat struct {
    Chart     string   // chart directory
    Name      string   // helm release name
    Namespace string   // target namespace
    Pins      map[string]string  // k8s minor version -> chart version
}

var compatMatrix = []ChartCompat{
    {
        Chart: "system/kyverno",
        Name: "kyverno",
        Namespace: "kyverno",
        Pins: map[string]string{
            "23": "3.3.0",   // k3s v1.23 — oldest compat
            "30": "3.5.3",   // k3s v1.30
            "31": "3.5.3",   // k3s v1.31
            "32": "3.8.1",   // k3s v1.32 — latest, full compat
        },
    },
}

func lookupChartVersion(chartDir, kubeMinor string) string {
    for _, c := range compatMatrix {
        if strings.Contains(c.Chart, chartDir) {
            if v, ok := c.Pins[kubeMinor]; ok {
                return v
            }
            // Fallback: find the highest pin <= current version
            ...
        }
    }
    return "" // use chart default
}
```

---

## Success Criteria (all must be true)

1. `k3d cluster create ubiquity-dev --config metal/k3d-dev.yaml` successfully creates a cluster with k3s v1.32.13
2. `ubiquity up --sandbox` completes with exit 0 and no warnings/errors on k3s v1.32.13
3. `make test-k3d-matrix` or `bash test/k3d-matrix.sh` passes for all 3 k3s versions
4. `docs/compatibility/k3s-helm-matrix.yaml` exists with entries for all 3 K8s versions
5. `docs/compatibility/COMPATIBILITY.md` exists with ADRs for each chart pin
6. `cmd/ubiquity/cmd/compat.go` exists with auto-detection logic and compat matrix
7. Kyverno operator installs successfully on k3s v1.32.13 (VAP available)
8. `go test ./cmd/ubiquity/cmd/...` passes (unit tests)
9. `go build ./...` passes
10. No charts use `--include-crds` flag anywhere
11. No `--ignore-not-found` flag anywhere
12. All redis-ha pods are ready on all 3 k3s versions (pinned to 1 replica)
