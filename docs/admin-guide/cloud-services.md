# Ubiquity cloud services

This guide documents Ubiquity-native cloud primitives inspired by mature Kubernetes cloud platforms while preserving Ubiquity's NICo-first and NVIDIA-focused architecture.

## CRD and operator provenance

Ubiquity renders cloud resources only against explicit operator/CRD contracts. Rendering a CR is not readiness proof; the matching operator must be installed, version-compatible, healthy, and reconciling.

| Capability | Rendered API/kind | Expected operator family | Readiness boundary |
| --- | --- | --- | --- |
| VM disks | `cdi.kubevirt.io/v1beta1` `DataVolume`, core `PersistentVolumeClaim` | KubeVirt CDI | DataVolume/PVC bound and import/clone succeeded |
| VM instances | `kubevirt.io/v1` `VirtualMachine` | KubeVirt | VM conditions ready, guest agent optional checks complete |
| Object buckets | `objectbucket.io/v1alpha1` `ObjectBucketClaim` | object bucket controller / objectstorage-controller | claim bound and connection secret present |
| Postgres | `postgresql.cnpg.io/v1` `Cluster` | CloudNativePG | instances ready and service/secret present |
| Redis | `databases.spotahome.com/v1` `RedisFailover` | Redis Operator | Redis and Sentinel pods ready |
| Kafka | `kafka.strimzi.io/v1beta2` `Kafka` | Strimzi | Kafka cluster ready condition true |
| Registry | `goharbor.io/v1alpha1` `Project` | Harbor operator or compatible registry controller | project reconciled and registry endpoint reachable |
| Tenant Kubernetes | `cluster.x-k8s.io/v1beta1` `Cluster`, `TenantControlPlane` intent | Cluster API and Kamaji-compatible control-plane provider | kubeconfig secret present and workload API responds |
| Backup | `k8up.io/v1` `Schedule` | K8up | latest backup/check/prune succeeded |
| Snapshot | `snapshot.storage.k8s.io/v1` `VolumeSnapshotClass` | snapshot-controller and Longhorn CSI | snapshot creation and restore drill passed |

Cloud resource definitions must preserve this table when API versions change.

## VM disks

Standalone VM disks let operators create reusable KubeVirt/CDI storage before a VM exists. This avoids tying disk lifecycle to a single VM manifest.

Supported disk sources:

- `blank`: renders a PersistentVolumeClaim for a new empty disk.
- `http`: renders a CDI DataVolume that imports an image from an operator-approved URL.
- `pvc`: renders a CDI DataVolume that clones an existing PVC, for example from a golden image namespace.

Examples:

```sh
ubiquity cloud render vm-disk --name data-disk --namespace tenant-a --size 500Gi --source blank
ubiquity cloud render vm-disk --name ubuntu-base --namespace tenant-a --source http --source-url https://images.example/ubuntu.qcow2
ubiquity cloud render vm-disk --name clone-a --namespace tenant-a --source pvc --source-pvc ubuntu-base --source-pvc-namespace golden-images
```

Rendering or applying a disk is not proof that a VM booted from it. CDI import/clone readiness must be checked in-cluster before VM readiness is claimed.

## Tenant VPCs

Tenant VPC rendering creates a namespace, ResourceQuota, Multus `NetworkAttachmentDefinition`, and deny-by-default `NetworkPolicy` pair. This complements the existing Ubiquity AI tenancy and NVIDIA networking work; it does not replace cluster bootstrap, NICo, Cilium, or the NVIDIA Network Operator.

```sh
ubiquity cloud render vpc --name tenant-a --cidr 10.60.0.0/24 --gateway 10.60.0.1 --bridge br-tenant-a --gpu-quota 8
```

Each tenant VPC starts closed and only allows same-tenant namespace traffic until additional reviewed policies are added.

## Tenant Kubernetes clusters

Tenant Kubernetes cluster rendering expresses Cluster API/Kamaji-style workload cluster intent while preserving the Ubiquity decision that NICo remains primary for physical node lifecycle. These manifests are workload-cluster intent, not bootstrap replacement and not proof that a tenant cluster is ready.

```sh
ubiquity cloud render tenant-cluster --name tenant-a-dev --namespace tenant-a --kubernetes-version v1.31.4 --control-plane-class kamaji --node-pool-class nico-managed-workers --worker-replicas 3
```

## Managed services

The managed service catalog renders operator-backed service CRs without installing or replacing their operators. Supported first-slice services are `bucket`, `postgres`, `redis`, `kafka`, `registry`, `mariadb`, `mongodb`, `nats`, `rabbitmq`, `clickhouse`, `opensearch`, `qdrant`, `openbao`, `http-cache`, and `tcp-balancer`. Production readiness still depends on the matching operator CRD/controller being installed and healthy.

```sh
ubiquity cloud render service --name datasets --namespace tenant-a --service-type bucket --service-storage-class object-store
ubiquity cloud render service --name pg-ai --namespace tenant-a --service-type postgres --service-size 200Gi --service-replicas 3
```

## Backup, snapshots, and resource presets

Platform ops policies render K8up backup schedules, retained Longhorn-compatible `VolumeSnapshotClass` resources, and resource preset ConfigMaps for tenant self-service defaults. Rendering these resources is not restore proof; operators must validate backup repository access, snapshot creation, and restore drills in-cluster.

```sh
ubiquity cloud render backup-policy --name tenant-a-daily --namespace tenant-a --backup-schedule "0 2 * * *" --backup-retention 30d --backup-repository-secret tenant-a-backup-repo --snapshot-class longhorn-snapshots --preset-name gpu-medium --preset-cpu 16 --preset-memory 128Gi --preset-gpu 1
```

## Production readiness gates

Before an operator can claim a rendered cloud primitive is production-ready, run the applicable gates in order:

1. Helm schema validation and template render for the chart and values in use.
2. Kubernetes server-side dry-run against a cluster that has the required CRDs installed.
3. Apply through the intended GitOps or operator workflow.
4. Confirm controller reconciliation using status conditions rather than object existence.
5. Run workload smoke tests: VM boot, network path, database connection, Kafka topic, bucket write/read, or registry push/pull as applicable.
6. Run backup and restore drill for any persistent service before marking it durable.
7. Record provenance: chart version, CRD version, controller image, rendered manifest hash, and validation timestamp.

Failure at any gate must fail closed; do not infer readiness from render/apply success alone.

## Governance policy bundle

`ubiquity cloud render governance` emits a cross-cutting tenant bundle for RBAC, admission, GitOps lifecycle, observability alerts, cost allocation labels, Gateway API, external DNS intent, VPN egress policy, expandable retained storage, and upgrade/rollback policy metadata. The bundle is intentionally policy intent: controllers such as Kyverno, Argo CD, Prometheus Operator, OpenCost, Gateway API, external-dns, and Longhorn must be installed and reconciled before operators claim readiness.

## Operator install-plan bundle

`ubiquity cloud render operator-bundles` emits the controller ownership/provenance contract for cloud CRDs. Each bundle records the controller family, owned CRD, install namespace, upstream source, pinned-by-platform version policy, and expected air-gap artifact location. This does not fetch or install third-party controllers by itself; it gives GitOps and CI a reviewable installation contract before CRD-backed cloud resources are applied.
