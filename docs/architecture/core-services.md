# Core services architecture

Ubiquity deploys core cluster services through small, reviewable component charts and a GitOps orchestration layer instead of one monolithic chart. This is better than a monolithic chart for this repository because reviewers can lint, template, test, version, and reason about each component boundary independently while still rendering one auditable set of ArgoCD `Application` objects for cluster bootstrap.

## Included capabilities

`system/core-services` renders GitOps applications for these local component charts when Ubiquity already owns the wrapper:

- `cert-manager`
- `cilium`
- `external-secrets`
- `longhorn`
- `network-policies`
- `kyverno`
- `kyverno-policies`
- `falco`
- `monitoring-system`
- `ingress-nginx`

It also renders public upstream chart applications for components that do not yet need a local wrapper:

- `metrics-server`
- `node-feature-discovery`
- `node-problem-detector`
- `snapshot-controller`
- `velero`
- `vertical-pod-autoscaler`
- `kubescape`
- `local-path-provisioner`

Helm components use `CreateNamespace=true` in ArgoCD sync options so namespace creation is explicit and reviewable.

## Excluded capabilities

The chart intentionally excludes any GitOps controller other than ArgoCD and any private or proprietary repository dependency. Provider-specific storage integrations are not enabled by this bundle unless they are already represented as local, reviewable Ubiquity charts. Longhorn remains the default general-purpose persistent storage path, and local-path remains optional for sandbox or single-node use.

## Backup safety

Velero is disabled by default. Enabling it without `applications.velero.backupBucket` fails closed with `velero.backupBucket is required`. Rendered backup objects prove intent only; production readiness still requires live backup and restore proof.

## Validation

```sh
helm lint system/core-services
helm template core-services system/core-services --namespace argocd
helm unittest system/core-services
test/e2e/core-services-proof.sh --dry-run
```

The dry-run proof checks that expected capability names render, no excluded GitOps manifests render, and backup configuration gates fail closed.
