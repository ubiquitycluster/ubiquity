# NVIDIA Infra Controller OS images

Status: experimental/preview. This reference describes how Ubiquity documents Operating System image intent for NVIDIA Infra Controller (NICo). It is not a certification claim for any vendor image, driver stack, GPU, firmware, or server platform.

## Definition

In NICo vocabulary, an Operating System is the bootable image, version, configuration, and install intent assigned to a Machine. Ubiquity should treat OS image records as immutable release artifacts. Publish a new image record for any material change to the base OS, kernel, drivers, CUDA components, OFED/RDMA stack, bootstrap configuration, partitioning, or security posture.

NICo multi-OS boot is experimental/preview. Sites may maintain more than one approved Operating System image for different Machine pools, but each OS/version/architecture/boot-mode/driver-stack combination needs its own catalog record, checksum, rollback image, and local approval evidence. Do not switch a production Machine between OS families without a change record and a NICo Task that records the intended image ID.

## Required metadata

Every approved OS image entry should have:

- Image ID: stable name used by operators and automation.
- Version: human-readable version or date.
- Source artifact: object storage path, registry reference, or image service ID.
- Checksum: SHA-256 or stronger digest from the image build pipeline.
- Architecture: for example `amd64`.
- Boot mode: UEFI, secure boot expectation, PXE/iPXE notes.
- Disk layout: root disk selection, partitioning, wipe policy, and RAID assumptions.
- Kubernetes compatibility: supported kubelet/container runtime versions when applicable.
- GPU stack: whether NVIDIA driver, GPU Operator prerequisites, CUDA libraries, MIG policy, or DCGM components are included or expected later.
- Network stack: NIC firmware assumptions, OFED/RDMA/IPoIB expectations, and kernel module requirements.
- Security posture: hardening profile, default users, SSH policy, and update source.
- Secret policy: confirmation that no credentials or private keys are embedded. No secrets should be present in the image artifact or catalog entry.
- External secret references: where required, use Vault plus External Secrets Operator or an approved equivalent for runtime credentials; never embed them in the image.
- Rollback image: previous known-good image ID when one exists.

## Image promotion states

Use site-local promotion language rather than certification language:

- `draft`: built but not yet installed by NICo in the target site.
- `candidate`: installed on a test Machine and under validation.
- `approved`: approved by the site owner for specified Machine pools.
- `deprecated`: still present for rollback but not for new installs.
- `blocked`: must not be installed due to a known issue.

Do not use terms such as certified, validated by NVIDIA, or supported unless the repository contains an explicit external support statement. A Ubiquity smoke test only proves local behavior observed during that test.

## Example catalog entry

```yaml
id: ubuntu-22.04-gpu-2026-06-01
state: candidate
architecture: amd64
bootMode: uefi
source: s3://example-os-images/ubuntu-22.04-gpu-2026-06-01.raw.gz
sha256: replace-with-build-pipeline-digest
kubernetes:
  kubelet: 1.30.x
  containerRuntime: containerd
nvidia:
  gpuDriver: provided-by-gpu-operator
  cuda: provided-by-workload-image
  dcgm: provided-by-gpu-operator
network:
  rdma: site-specific
security:
  secretsEmbedded: false
  sshPasswordLogin: false
rollback: ubuntu-22.04-gpu-2026-05-15
notes: Experimental/preview image pending site-local validation.
```

## Approval checklist

1. Verify the checksum from the build system.
2. Confirm no secrets are embedded in the image.
3. Confirm the image source is reachable from the provisioning network.
4. Install on one non-production Machine through NICo.
5. Confirm NICo Task success, Machine state, Instance state, and boot logs.
6. Confirm Kubernetes node readiness where the image is for cluster nodes.
7. Confirm GPU and RDMA evidence where applicable.
8. Record the Task ID, image ID, Machine ID, and validation results.
9. Move the image to `approved` only for the Machine pools actually tested by the site.

## Failure handling

If an install fails, mark the image `blocked` for the affected pool while the issue is investigated. Preserve Task logs, console output, BMC event logs, and image build metadata. Do not repeatedly reinstall production Machines with an unproven image.


## Live proof and approval boundary

Live proof means evidence observed from the target cluster or a gated smoke-test script: controller status, API reachability, workload behavior, restore-drill readability, or service-specific smoke markers. Render, lint, dry-run, or object existence prove intent only.

These docs and scripts do not claim the system is NVIDIA approved or NVIDIA certified. The platform is not NVIDIA approved and not NVIDIA certified by repository evidence alone. Treat any NVIDIA approval evidence, support statement, or certification letter as an external artifact that must be attached to the deployment record before using approved/certified wording. Without that approval evidence, Ubiquity can claim only local validation results and live proof observed during the run.
