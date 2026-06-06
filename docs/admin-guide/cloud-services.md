# Ubiquity cloud services

This guide documents Ubiquity-native cloud primitives inspired by mature Kubernetes cloud platforms while preserving Ubiquity's NICo-first and NVIDIA-focused architecture.

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

The managed service catalog renders operator-backed service CRs without installing or replacing their operators. Supported first-slice services are `bucket`, `postgres`, `redis`, `kafka`, and `registry`. Production readiness still depends on the matching operator CRD/controller being installed and healthy.

```sh
ubiquity cloud render service --name datasets --namespace tenant-a --service-type bucket --service-storage-class object-store
ubiquity cloud render service --name pg-ai --namespace tenant-a --service-type postgres --service-size 200Gi --service-replicas 3
```
