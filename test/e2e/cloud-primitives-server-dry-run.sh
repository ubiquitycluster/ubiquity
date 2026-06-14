#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
UBIQUITY_BIN="${UBIQUITY_BIN:-$ROOT/bin/ubiquity}"

if [[ ! -x "$UBIQUITY_BIN" ]]; then
  (cd "$ROOT" && go build -o "$UBIQUITY_BIN" ./cmd/ubiquity)
fi

require_crd() {
  local crd="$1"
  kubectl get crd "$crd" >/dev/null 2>&1 || {
    echo "missing required CRD: $crd" >&2
    exit 20
  }
}

for crd in \
  datavolumes.cdi.kubevirt.io \
  virtualmachines.kubevirt.io \
  network-attachment-definitions.k8s.cni.cncf.io \
  objectbucketclaims.objectbucket.io \
  clusters.postgresql.cnpg.io \
  redisfailovers.databases.spotahome.com \
  kafkas.kafka.strimzi.io \
  projects.goharbor.io \
  clusters.cluster.x-k8s.io \
  schedules.k8up.io \
  volumesnapshotclasses.snapshot.storage.k8s.io; do
  require_crd "$crd"
done

server_dry_run() {
  local name="$1"
  shift
  echo "server-side dry-run: $name"
  "$UBIQUITY_BIN" "$@" | kubectl apply --dry-run=server -f -
}

server_dry_run prerequisites cloud render prerequisites
server_dry_run vm-disk cloud render vm-disk --name data-disk --namespace tenant-a --source blank
server_dry_run vpc cloud render vpc --name tenant-a --tenant tenant-a --cidr 10.60.0.0/24 --gateway 10.60.0.1 --bridge br-tenant-a
server_dry_run tenant-cluster cloud render tenant-cluster --name tenant-a-dev --namespace tenant-a --kubernetes-version v1.31.4
server_dry_run service-bucket cloud render service --name datasets --namespace tenant-a --service-type bucket --service-storage-class object-store
server_dry_run backup-policy cloud render backup-policy --name tenant-a-daily --namespace tenant-a --backup-repository-secret tenant-a-backup-repo

echo "cloud primitive server-side dry-run complete"
