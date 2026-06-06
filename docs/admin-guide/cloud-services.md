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
