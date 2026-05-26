#!/usr/bin/env bash
#
# k3d-matrix.sh — Multi-version k3s sandbox test harness
#
# Creates a k3d cluster for each k3s version listed in k3s-versions.txt,
# runs `ubiquity up --sandbox` against it, checks for success markers,
# and produces a summary table.
#
# Usage:
#   ./test/k3d-matrix.sh            # run full matrix
#   K3S_IMAGE=rancher/k3s:v1.32.13-k3s1 ./test/k3d-matrix.sh  # run single version
#
# Results are written to test/k3d-matrix-results/<version-tag>/

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RESULTS_DIR="$ROOT_DIR/test/k3d-matrix-results"
mkdir -p "$RESULTS_DIR"

# Colours
GRN='\033[0;32m'
RED='\033[0;31m'
YLW='\033[1;33m'
RST='\033[0m'

PASS_MARK="${GRN}✓${RST}"
FAIL_MARK="${RED}✗${RST}"

UBIQUITY_CLI="${ROOT_DIR}/ubiquity-cli"
if [ ! -x "$UBIQUITY_CLI" ]; then
    echo "Building ubiquity-cli..."
    make -C "$ROOT_DIR" cli
fi

###############################################
# Determine which versions to test
###############################################
if [ -n "${K3S_IMAGE:-}" ]; then
    VERSIONS=("$K3S_IMAGE")
else
    mapfile -t VERSIONS < <(grep -v '^#' "$ROOT_DIR/test/k3s-versions.txt" | grep -v '^$')
fi

###############################################
# Helper: extract version tag from image name
###############################################
version_tag() {
    local img="$1"
    echo "$img" | sed 's|rancher/k3s:||' | tr '.' '-' | tr '/' '_'
}

###############################################
# Test a single version
###############################################
test_version() {
    local image="$1"
    local tag
    tag="$(version_tag "$image")"
    local log_dir="$RESULTS_DIR/$tag"
    mkdir -p "$log_dir"
    local log="$log_dir/ubiquity-up.log"
    local result=0

    echo -e "\n${YLW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RST}"
    echo -e "${YLW} Testing: $image${RST}"
    echo -e "${YLW} Log:    $log${RST}"
    echo -e "${YLW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RST}"

    # 1. Delete any previous cluster with this tag
    k3d cluster delete "ubiquity-test-$tag" 2>/dev/null || true

    # 2. Create cluster
    echo "  Creating cluster..."
    K3S_VERSION_TAG="$tag" \
    K3S_IMAGE="$image" \
        k3d cluster create \
        --config "$ROOT_DIR/test/k3d-matrix-config.yaml" \
        --wait --timeout 180s 2>&1 | tee "$log_dir/k3d-create.log" || {
        echo -e "  ${FAIL_MARK} Cluster creation failed"
        echo "$image: cluster_creation_failed" >> "$RESULTS_DIR/summary.tmp"
        k3d cluster delete "ubiquity-test-$tag" 2>/dev/null || true
        return 1
    }

    # 3. Wait for nodes
    if ! kubectl wait --for=condition=Ready nodes --all --timeout=60s 2>&1; then
        echo -e "  ${FAIL_MARK} Nodes never became ready"
        echo "$image: nodes_not_ready" >> "$RESULTS_DIR/summary.tmp"
        k3d cluster delete "ubiquity-test-$tag" 2>/dev/null || true
        return 1
    fi

    # 4. Run ubiquity up
    echo "  Running ubiquity up..."
    cd "$ROOT_DIR"
    set +e
    timeout 600 "$UBIQUITY_CLI" up --sandbox 2>&1 | tee "$log"
    local exit_code="${PIPESTATUS[0]}"
    set -e
    echo "  ubiquity up exit code: $exit_code"

    # 5. Check results
    local status="pass"
    local failures=""

    # 5a. Exit code check
    if [ "$exit_code" -ne 0 ]; then
        status="fail"
        failures="${failures}exit_code_${exit_code} "
    fi

    # 5b. Check for error markers in the log (excluding informational warnings)
    if grep -q "unknown flag" "$log" 2>/dev/null; then
        # Filter out known kubectl version --short warning (harmless)
        local uf
        uf=$(grep -c "unknown flag" "$log" || true)
        local kf
        kf=$(grep -c "kubectl version --short" "$log" || true)
        if [ "$((uf - kf))" -gt 0 ]; then
            status="fail"
            failures="${failures}unknown_flags "
        fi
    fi

    # 5c. Check for CRD failures
    if grep -q "no matches for kind" "$log" 2>/dev/null; then
        status="fail"
        failures="${failures}crd_mismatch "
    fi

    # 5d. Check ArgoCD pods came up (wait phase)
    if grep -q "pod/argocd-server.*condition met" "$log" 2>/dev/null; then
        :
    else
        status="fail"
        failures="${failures}argocd_not_ready "
    fi

    # 5e. Check Kyverno operator installed
    if grep -q "NAME: kyverno" "$log" 2>/dev/null; then
        :
    else
        # On K8s < 1.28, Kyverno won't install — that's expected
        local maj
        maj=$(echo "$tag" | cut -d- -f1 | cut -d. -f2)
        if [ "$maj" -ge 28 ] 2>/dev/null; then
            status="fail"
            failures="${failures}kyverno_not_installed "
        fi
    fi

    # 5f. Record result
    if [ "$status" = "pass" ]; then
        echo -e "  ${PASS_MARK} All checks passed"
        echo "$image: pass" >> "$RESULTS_DIR/summary.tmp"
    else
        echo -e "  ${FAIL_MARK} Failures: $failures"
        echo "$image: fail ${failures}" >> "$RESULTS_DIR/summary.tmp"
    fi

    # 6. Delete cluster
    echo "  Cleaning up..."
    k3d cluster delete "ubiquity-test-$tag" 2>/dev/null || true

    return 0
}

###############################################
# Main
###############################################
echo "k3d Multi-Version Test Matrix"
echo "=============================="
echo "Testing ${#VERSIONS[@]} version(s): ${VERSIONS[*]}"
echo ""

# Run tests sequentially (each needs full cluster)
for v in "${VERSIONS[@]}"; do
    test_version "$v"
done

# Summary
echo ""
echo "=============================="
echo "          SUMMARY"
echo "=============================="
if [ -f "$RESULTS_DIR/summary.tmp" ]; then
    while IFS= read -r line; do
        if echo "$line" | grep -q "pass$"; then
            echo -e "  ${PASS_MARK} $line"
        else
            echo -e "  ${FAIL_MARK} $line"
        fi
    done < "$RESULTS_DIR/summary.tmp"
    rm -f "$RESULTS_DIR/summary.tmp"
else
    echo "  No tests ran."
fi
echo ""
