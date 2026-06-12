# Ubiquity CLI Reference

## Overview
The `ubiquity` CLI is the primary entry point for managing Ubiquity clusters.

## Installation
```
make cli
sudo make install
```

## Commands

### init
Bootstrap Ubiquity configuration. Creates ~/.ubiquity/ with skeleton config.
```
ubiquity init
```

### configure
Interactive configuration wizard for cluster settings.
```
ubiquity configure --domain mycluster.example.com
ubiquity configure --interactive
```

### up
Deploy the full cluster stack (6 phases).
```
ubiquity up --sandbox
ubiquity up --env prod
ubiquity up --skip-security
```

### down
Tear down cluster and cloud resources.
```
ubiquity down
```

### status
Show provisioning state and phase progress.
```
ubiquity status
ubiquity status --plain
```

### logs
Read provisioning logs from state.
```
ubiquity logs
ubiquity logs bootstrap
```

### retry
Retry a failed provisioning phase.
```
ubiquity retry metal
ubiquity retry bootstrap
```

### ai-platform
Inspect a declarative NVIDIA-backed AI workload platform profile.

This command prints the selected profile, capabilities, source repositories, chart repositories, replacement decisions, bare-metal orchestration alternatives, and fail-closed readiness policy. It does not claim NVIDIA approval or certification.

```
ubiquity ai-platform --profile sandbox
ubiquity ai-platform --profile gpu-basic
ubiquity ai-platform --profile gpu-rdma
ubiquity ai-platform --profile gpu-mig
ubiquity ai-platform --profile ai-production
```

Use `ai-production` when reviewing the full target platform: GPU Operator, NVIDIA Network Operator, NIM Operator, KAI Scheduler, Stallscope GPU workload telemetry, AI workload tenancy, telemetry, and validation.

Stallscope integration is delivered as `platform/stallscope`. It runs on GPU nodes, writes Prometheus textfile metrics for node-exporter, and alerts when workloads are classified as SLOW or FAIL_RISK from GPU, host, RDMA, PFC, and network evidence. Production deployments should pin a scanned internal Stallscope image or retain the chart's pinned upstream archive commit.

### nodes
Operate NVIDIA Infra Controller-backed bare-metal node lifecycle. Commands default to dry-run/mock safe output unless live NICo configuration is provided with `UBIQUITY_NICO_MODE=live`, `UBIQUITY_NICO_BASE_URL`, `UBIQUITY_NICO_ORG`, and credentials.

```
ubiquity nodes list [-o table|json]
ubiquity nodes status <name> [-o json]
ubiquity nodes os list
ubiquity nodes os apply <image>
ubiquity nodes os apply --inventory examples/node-inventory/nico-prod.yaml
ubiquity nodes add <name> --os-image <image>
ubiquity nodes add <name> --inventory examples/node-inventory/nico-prod.yaml
ubiquity nodes remove <name> --confirm <name> --drain-confirmed
ubiquity nodes reinstall <name> --os-image <image> --confirm <name> --drain-confirmed
ubiquity nodes power <name> on|off|reset --confirm <name> --drain-confirmed --reason <why>
ubiquity nodes task <task-id>
```

Live power operations resolve the target through NICo/Kubernetes status first, enforce confirmation/drain/quorum/storage/AIStore safety gates, and then submit a NICo machine power task. `reset` and `off` never fall back to BMO or shell snippets by default.

### test
Run the test suite.
```
ubiquity test
ubiquity test --integration
ubiquity test --sandbox-deploy
```

`ubiquity test --sandbox-deploy` runs NVIDIA AI sandbox deploy render validation. It dependency-builds and renders the sandbox-safe NVIDIA AI Helm charts with CRDs, cleans generated Helm artifacts afterward, and does not require NVIDIA devices. It proves manifest/render correctness, not GPU runtime, RDMA, NIM model serving, or production scheduling readiness.

### version
Print version information.
```
ubiquity version
ubiquity version --json
```

### info
Show cluster information summary.
```
ubiquity info
```

### health
Check cluster health. Focused readiness flags fail closed and return a non-zero exit when required live evidence is missing.
```
ubiquity health
ubiquity health --ai
ubiquity health --aistore
ubiquity health --nico
```

Flags:
- `--ai`: run only core NVIDIA AI platform checks for GPU Operator, device plugin, DCGM metrics, RDMA evidence, NIM smoke evidence, and KAI Scheduler evidence.
- `--aistore`: run only NVIDIA AIStore data-plane checks. AIStore is evaluated as an optional AI dataset/cache/object path and not a generic PVC replacement.
- `--nico`: run only NVIDIA Infra Controller readiness checks.
