# Deploying the core-services bundle: sandbox and ThinkCentre production profile

This guide shows how to deploy Ubiquity with the `system/core-services` bundle in two modes:

1. sandbox mode on a laptop/workstation with Docker and k3d
2. production profile on the original small-node hardware class: 3 x Lenovo ThinkCentre M700 Tiny nodes

The ThinkCentre profile is a real bare-metal production-style deployment path for a small cluster, but it is capacity-constrained. Treat it as suitable for home lab, edge, development, and low-risk internal services unless you have separately validated workload capacity, backups, restore drills, monitoring, physical redundancy, and operational support.

## What gets deployed

The core-services bundle is a thin ArgoCD Application orchestration chart. It does not replace Ubiquity's per-component charts. It wires together the base services normally expected in a usable cluster:

- cert-manager
- cilium
- external-secrets
- longhorn
- network-policies
- kyverno and kyverno-policies
- falco
- monitoring-system
- ingress-nginx
- metrics-server
- node-feature-discovery
- node-problem-detector
- snapshot-controller
- vertical-pod-autoscaler
- kubescape
- optional local-path-provisioner
- optional Velero, gated by explicit backup bucket configuration

The bundle intentionally uses ArgoCD Applications and Helm charts only. It does not use the ignored GitOps controller path.

## Hardware assumptions

### Sandbox host

Recommended minimum:

- 4+ CPU cores
- 16 GB RAM minimum; 32 GB recommended
- 60+ GB free disk
- Docker or a Docker-compatible runtime
- k3d, kubectl, Helm, Go, and Git available locally

The original documentation used a Lenovo ThinkPad P16s G1 as the sandbox/bootstrap host, but any comparable Linux/macOS workstation is fine.

### ThinkCentre production-profile cluster

Original small-node hardware class:

- 3 x Lenovo ThinkCentre M700 Tiny
- Intel Core i5-6600T
- 16 GB RAM each
- 500 GB SSD each
- 1 GbE switch, such as Netgear GS305E

Recommended layout:

| Node | Role | Example IP | Notes |
| --- | --- | --- | --- |
| `cp1` | k3s server/control-plane | `10.1.0.11` | also runs workloads on small clusters |
| `cp2` | k3s server/control-plane | `10.1.0.12` | also runs workloads on small clusters |
| `cp3` | k3s server/control-plane | `10.1.0.13` | also runs workloads on small clusters |
| bootstrap host | operator workstation/PXE host | `10.1.0.10` | runs CLI, Ansible, PXE, kubectl |

Important ThinkCentre caveat: most M700 Tiny units do not have server-class IPMI/BMC. Use Wake-on-LAN plus manual BIOS/PXE configuration, or boot them from USB once and then manage them over SSH. If your units do have out-of-band management, populate the IPMI fields in the inventory; otherwise prefer `wol: true` and prepare BIOS boot order manually.

## Common prerequisites

Run these on the bootstrap workstation.

```bash
git clone https://github.com/ubiquitycluster/ubiquity.git
cd ubiquity

git submodule update --init --recursive
make cli
```

Install local tools if they are not already present:

```bash
# macOS example
brew install go helm kubectl k3d ansible sops age

# Linux examples vary by distro; at minimum install:
# go, helm, kubectl, docker, ansible, sops, age, git
```

Verify the implementation before deploying:

```bash
go test ./pkg/... ./cmd/... -count=1
helm lint system/core-services
helm template core-services system/core-services --namespace argocd >/tmp/core-services.yaml
helm unittest system/core-services
test/e2e/core-services-proof.sh --dry-run
```

Expected final proof line:

```text
core-services-proof-passed
```

## Sandbox deployment

Sandbox mode creates a local k3d cluster and deploys the stack from the local repository. This is the fastest way to validate the bundle and UI flows without touching the ThinkCentre nodes.

### 1. Start from a clean local state

```bash
# optional cleanup if you already have a sandbox cluster
k3d cluster delete ubiquity-dev || true
rm -f metal/kubeconfig.yaml
```

### 2. Bring up sandbox

Preferred CLI path:

```bash
./ubiquity-cli up --sandbox
```

Equivalent legacy Make target:

```bash
make sandbox
```

The sandbox k3d config exposes HTTP/HTTPS through the k3d load balancer and writes kubeconfig to:

```text
metal/kubeconfig.yaml
```

Set your shell to use it:

```bash
export KUBECONFIG="$PWD/metal/kubeconfig.yaml"
kubectl cluster-info
kubectl get nodes -o wide
```

### 3. Validate the core-services chart locally

```bash
test/e2e/core-services-proof.sh --dry-run
helm template core-services system/core-services --namespace argocd >/tmp/core-services.yaml
kubectl apply --dry-run=server -f /tmp/core-services.yaml
```

If the ArgoCD CRDs are installed in the sandbox, the server-side dry-run should validate the Application objects. If the cluster does not have the CRDs yet, finish the `ubiquity up --sandbox` bootstrap first and rerun the command.

### 4. Apply the core-services Applications in sandbox

Use this only after ArgoCD is installed:

```bash
kubectl get crd applications.argoproj.io

helm template core-services system/core-services \
  --namespace argocd \
  --set global.repoURL="$(git config --get remote.origin.url)" \
  --set global.targetRevision="$(git rev-parse --abbrev-ref HEAD)" \
  | kubectl apply -f -
```

For a purely local sandbox with no ArgoCD repository credentials, `ubiquity up --sandbox` already applies local charts directly where possible. The core-services Application objects are still useful as a rendered contract and as the path you will use in production.

### 5. Sandbox readiness checks

```bash
kubectl -n argocd get applications || true
kubectl get pods -A
kubectl get nodes
kubectl get storageclass
kubectl -n longhorn-system get pods || true
kubectl -n kyverno get pods || true
kubectl -n falco get pods || true
```

Useful waits:

```bash
kubectl -n cert-manager wait --for=condition=Ready pod -l app.kubernetes.io/instance=cert-manager --timeout=180s || true
kubectl -n argocd get applications -o wide || true
```

Access examples, depending on what has finished syncing:

```bash
kubectl -n argocd port-forward svc/argocd-server 8080:443
# open https://localhost:8080
```

For Grafana in the default sandbox ingress pattern, check the current service/ingress first:

```bash
kubectl get ingress -A
kubectl get svc -A | grep -i grafana || true
```

### 6. Sandbox teardown

```bash
./ubiquity-cli down --env sandbox || true
k3d cluster delete ubiquity-dev || true
```

## Production profile on the three ThinkCentre nodes

This path assumes you want the three M700 Tiny machines to run a small HA k3s control plane and workloads. Because there are only three 16 GB nodes, be conservative about optional services and retention settings.

### 1. Prepare firmware and network

On each ThinkCentre:

1. Update BIOS/firmware.
2. Enable virtualization.
3. Enable Wake-on-LAN if you want remote power-on.
4. Configure PXE/network boot first for first install, or be ready to select PXE manually during boot.
5. Disable Secure Boot unless your chosen OS image and boot flow are signed and tested.
6. Confirm the install disk name from a live shell when possible: usually `sda` for SATA SSDs, sometimes `nvme0n1` for NVMe.

Network recommendations:

- Use static DHCP reservations or static IPs.
- Put the bootstrap host and all ThinkCentre nodes on the same L2 network for PXE/WOL simplicity.
- Reserve an address range for LoadBalancer services if using MetalLB.
- Create DNS names before production use, even if they are only internal.

Example addressing:

```text
bootstrap: 10.1.0.10
cp1:       10.1.0.11
cp2:       10.1.0.12
cp3:       10.1.0.13
VIP/LB:    10.1.0.20-10.1.0.40
```

### 2. Create the bare-metal inventory

Edit `metal/inventories/prod.yml` for the ThinkCentre machines. Do not commit real passwords or secrets.

Example Wake-on-LAN/manual-PXE inventory:

```yaml
metal:
  children:
    masters:
      hosts:
        cp1:
          ansible_host: 10.1.0.11
          mac: 'aa:bb:cc:dd:ee:01'
          disk: sda
          network_interface: eno1
          wol: true
        cp2:
          ansible_host: 10.1.0.12
          mac: 'aa:bb:cc:dd:ee:02'
          disk: sda
          network_interface: eno1
          wol: true
        cp3:
          ansible_host: 10.1.0.13
          mac: 'aa:bb:cc:dd:ee:03'
          disk: sda
          network_interface: eno1
          wol: true
    workers:
      hosts: {}
```

If your hardware has a real BMC/IPMI endpoint, use `wol: false` and add `ipmi_addr`, `ipmi_user`, and `ipmi_pass`. Store real secrets outside Git or encrypt them with SOPS.

### 3. Configure SSH access

```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -C ubiquity-bootstrap || true
```

If the nodes already have an OS and SSH enabled:

```bash
ssh-copy-id root@10.1.0.11
ssh-copy-id root@10.1.0.12
ssh-copy-id root@10.1.0.13
ansible -i metal/inventories/prod.yml metal -m ping --key-file ~/.ssh/id_ed25519
```

If you are doing a PXE reinstall, make sure the bootstrap host can serve PXE on the node network and that the nodes boot from network.

### 4. Provision the k3s cluster

Option A: CLI production profile:

```bash
./ubiquity-cli up \
  --env prod \
  --metal-bootstrap-backend ansible \
  --node-lifecycle-backend nico \
  --nico-site thinkcentre-lab
```

Option B: run the metal phases explicitly:

```bash
make -C metal env=prod boot
make -C metal env=prod cluster
```

After cluster provisioning, set kubeconfig:

```bash
export KUBECONFIG="$PWD/metal/kubeconfig.yaml"
kubectl cluster-info
kubectl get nodes -o wide
```

Expected shape:

```text
cp1   Ready
cp2   Ready
cp3   Ready
```

### 5. Install bootstrap/GitOps layer

If you used the full CLI path, this should already be part of the deployment pipeline. If you ran the metal phases manually, install bootstrap:

```bash
make -C bootstrap
kubectl -n argocd get pods
kubectl get crd applications.argoproj.io applicationsets.argoproj.io
```

### 6. Apply core-services for production

Create a local values override for production. Do not commit secrets.

Example `values/core-services-thinkcentre-prod.yaml`:

```yaml
global:
  repoURL: https://github.com/ubiquitycluster/ubiquity.git
  targetRevision: main
  destinationServer: https://kubernetes.default.svc

applications:
  local-path-provisioner:
    enabled: false
  velero:
    enabled: true
    backupBucket: ubiquity-thinkcentre-backups
    backupProvider: aws
    backupRegion: us-east-1
    backupS3Url: ""
```

For a constrained three-node cluster, start with Velero disabled until the object store credentials and restore drill are ready:

```yaml
applications:
  velero:
    enabled: false
```

Render and validate:

```bash
helm lint system/core-services
helm template core-services system/core-services \
  --namespace argocd \
  -f values/core-services-thinkcentre-prod.yaml \
  >/tmp/core-services-prod.yaml

kubectl apply --dry-run=server -f /tmp/core-services-prod.yaml
```

Apply:

```bash
kubectl apply -f /tmp/core-services-prod.yaml
```

Or install it as a Helm release that manages the Application objects:

```bash
helm upgrade --install core-services system/core-services \
  --namespace argocd \
  --create-namespace \
  -f values/core-services-thinkcentre-prod.yaml
```

### 7. Production readiness checks

Application state:

```bash
kubectl -n argocd get applications -o wide
kubectl -n argocd get applications \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.health.status}{"\t"}{.status.sync.status}{"\n"}{end}'
```

Node state:

```bash
kubectl get nodes -o wide
kubectl describe nodes | grep -E 'Name:|Ready|MemoryPressure|DiskPressure|PIDPressure' -A3
```

Core services:

```bash
kubectl -n cert-manager get pods
kubectl -n kube-system get pods | grep -E 'cilium|metrics-server' || true
kubectl -n longhorn-system get pods
kubectl -n kyverno get pods
kubectl -n falco get pods
kubectl -n monitoring get pods
kubectl get storageclass
kubectl get networkpolicy -A
```

Security and behavior proofs:

```bash
test/e2e/core-services-proof.sh --dry-run
test/e2e/network-policy-behavior.sh --dry-run
test/e2e/runtime-security-validation.sh --dry-run
test/e2e/helm-hardening-checks.sh --dry-run
```

Live network policy proof, after choosing a safe namespace:

```bash
test/e2e/network-policy-behavior.sh
```

Backup readiness, only if Velero is enabled:

```bash
kubectl -n velero get pods
velero backup-location get
velero backup create smoke-$(date +%Y%m%d%H%M%S) --include-namespaces default
velero backup get
```

Do not mark backups production-ready until you have restored data and verified the restored application can read it.

### 8. Capacity settings for the M700 Tiny cluster

Recommended initial choices:

- keep local-path-provisioner disabled unless you need single-node scratch storage
- keep Velero disabled until object storage is configured
- reduce log retention in Loki/Grafana if disk pressure appears
- schedule large CI, AI, or HPC workloads elsewhere; this hardware is small
- watch Longhorn replica count and free disk carefully
- avoid overcommitting memory; three 16 GB nodes can become memory constrained quickly

Basic checks:

```bash
kubectl top nodes
kubectl top pods -A --sort-by=memory | tail -30
kubectl -n longhorn-system get volumes.longhorn.io || true
df -h
```

### 9. Day-2 node lifecycle notes

For new production-profile deployments, the default day-2 lifecycle backend is NICo. On the ThinkCentre lab, keep destructive lifecycle operations disabled until you have a dedicated test node and a known-good reinstall path.

Safe dry-run checks:

```bash
UBIQUITY_NICO_MODE=mock ./ubiquity-cli nodes status --backend nico --dry-run
UBIQUITY_NICO_MODE=mock ./ubiquity-cli nodes power cp1 off --backend nico --dry-run --confirm cp1 --reason maintenance-test
UBIQUITY_NICO_MODE=mock ./ubiquity-cli nodes apply-os cp1 --backend nico --dry-run --confirm cp1 --os ubuntu-24.04
```

Do not run destructive node lifecycle scripts on the ThinkCentre nodes unless the target node is drained, backed up, and explicitly dedicated for reinstall testing.

### 10. Rollback and recovery

Remove only the core-services Application objects:

```bash
helm uninstall core-services -n argocd || true
kubectl delete -f /tmp/core-services-prod.yaml || true
```

Stop GitOps sync for a broken app:

```bash
kubectl -n argocd patch application APP_NAME --type merge -p '{"spec":{"syncPolicy":null}}'
```

Recover cluster access:

```bash
export KUBECONFIG="$PWD/metal/kubeconfig.yaml"
kubectl get nodes
kubectl -n argocd get pods
```

If a ThinkCentre node fails:

1. cordon and drain it if Kubernetes still sees it
2. replace disk or hardware
3. reinstall via the same PXE/USB path
4. rejoin it to k3s
5. verify Longhorn replica health before resuming workloads

## Go/no-go checklist

Sandbox is acceptable when:

- `./ubiquity-cli up --sandbox` completes
- `kubectl get nodes` shows Ready nodes
- `test/e2e/core-services-proof.sh --dry-run` prints `core-services-proof-passed`
- ArgoCD Applications render or local charts apply successfully

ThinkCentre production profile is acceptable when:

- all three nodes are Ready
- ArgoCD is running and can read the configured repository
- core-services Applications are Synced/Healthy, or any exceptions are documented
- Longhorn or the selected storage class is healthy
- network policy behavior is proven
- runtime security dry-run and Helm hardening checks pass
- monitoring and alerting are reachable
- backup and restore drill is complete if Velero is enabled
- no secrets are committed in inventory or values files
