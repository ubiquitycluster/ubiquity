# AIStore evaluation for NVIDIA AI platform profiles

Source repositories:

- NVIDIA/aistore
- NVIDIA/ais-k8s

decision: adopt AIStore as the preferred NVIDIA-maintained AI data-plane option when the workload needs high-throughput object storage, dataset caching, model/checkpoint artifact serving, S3-compatible access, sharded archive access, or data movement close to GPU workers.

Compatibility with prior evaluation wording: this decision still means enable only after readiness evidence, persistence, capacity, and operational ownership are proven for the target cluster.

Boundary: AIStore replaces Longhorn for AI dataset/cache paths when those paths are object/data-plane oriented. AIStore is not a generic PVC replacement and must not be described as a drop-in substitute for Kubernetes ReadWriteMany/POSIX shared filesystems or every application PVC.

## Why AIStore is a better AI storage option than Longhorn for these paths

Longhorn is useful generic replicated block storage for ordinary Kubernetes PVCs, but it is not NVIDIA-maintained and it is not designed as the primary high-throughput AI dataset/cache layer. For the AI workload platform, the stronger NVIDIA source-backed option is AIStore:

- NVIDIA/aistore describes AIStore as high-performance, scalable storage for AI workloads.
- AIStore provides object storage and an HTTP/S3-compatible API that can serve unmodified S3 clients.
- AIStore includes PyTorch integration, Python SDK support, and batch object retrieval patterns such as Get-Batch.
- AIStore supports local and remote buckets, prefetch, cache/eviction, ETL/data transforms, mirroring, erasure coding, rebalance, resilver, and observability.
- NVIDIA/ais-k8s is the production Kubernetes deployment toolkit with the AIS operator, Helm charts, Ansible playbooks, and monitoring guidance.

Therefore:

- Use AIStore for model artifacts, checkpoints, training datasets, inference datasets, sharded archives, remote bucket acceleration, and GPU-adjacent cache/object access.
- Retain Longhorn only for generic platform PVCs and non-performance-critical app state until a stronger POSIX/RWX shared filesystem choice is selected.
- Do not use Longhorn readiness as evidence that the NVIDIA AI data plane is ready.

## What AIStore does not replace

AIStore is not a universal shared filesystem replacement.

It does not automatically replace:

- Generic ReadWriteOnce application PVCs.
- ReadWriteMany/POSIX shared filesystem requirements.
- Stateful service PVCs that expect block-volume semantics.
- The PV provider required to back AIS target disks.

NVIDIA/ais-k8s documents that AIS target pods use PVCs bound to existing local PV or other existing PV objects. The storage volumes documentation states that target PVs must include ReadWriteOnce, matching storage class names, adequate capacity, and a preconfigured filesystem with XFS recommended. Local PVs should include node affinity. The chart also documents a hostPath option, but hostPath has security implications and must be explicitly selected.

## Recommended Ubiquity policy

1. Keep Longhorn available as a generic platform storage class.
2. Stop treating Longhorn as the preferred AI dataset/cache data plane.
3. Add AIStore as the preferred optional production AI data-plane candidate.
4. Require explicit readiness evidence before claiming AIStore is ready.
5. Keep the replacement claim scoped: AIStore replaces Longhorn for AI dataset/cache paths, not for generic PVCs.

## AIStore readiness evidence

AIStore readiness is reported separately from core GPU/NIM/KAI readiness because AIStore is an optional AI dataset/cache/object path, not a generic PVC replacement and not required for every AI workload.

Before Ubiquity can mark AIStore ready, collect all of the following:

- AIS operator deployment is available.
- AIS CRDs are established.
- AIS cluster custom resource is ready.
- AIS proxy pods are ready.
- AIS target pods are ready.
- AIS target PVCs are bound to the expected PVs or explicit hostPath mode is documented and approved. Ubiquity records this with the `aistore-target-storage-proven` ConfigMap after the operator validates the backing path.
- A bucket create/list/write/read/delete smoke test passes. Ubiquity records this with the `aistore-bucket-smoke-test-passed` ConfigMap.
- A GPU workload can read a representative model or dataset artifact through the AIS/S3-compatible endpoint. Ubiquity records this with the `aistore-gpu-artifact-read-passed` ConfigMap.
- Prometheus/metrics or AIS CLI health output is available for operational troubleshooting. Ubiquity records this with the `aistore-metrics-proven` ConfigMap after metrics or health output has been verified.

## Implementation notes for Ubiquity

- `ai-production` may include AIStore as an evaluated optional component.
- `gpu-basic` should not enable AIStore by default.
- Sandbox render validation can prove the wrapper chart renders, but it cannot prove production storage performance or persistence.
- Production deployment needs capacity planning for AIS targets, disk failure domains, network paths from GPU workers to AIS proxies/targets, bucket lifecycle policy, authentication, and observability.
- If an AI workload needs POSIX/RWX semantics, evaluate a dedicated filesystem path separately rather than mislabeling AIStore or Longhorn as a universal solution.
