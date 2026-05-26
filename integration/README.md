# KUTTL Integration Tests

This directory contains integration tests for the Ubiquity cluster
using the [KUTTL](https://kuttl.dev) testing framework.

## Prerequisites

- kubectl
- kubectl-kuttl plugin
- A running Kubernetes cluster (or KinD/K3s cluster)

## Running Tests

From the repository root:

```bash
kubectl kuttl test ./integration/
```

To run with a specific kubeconfig:

```bash
kubectl kuttl test ./integration/ --kubeconfig /path/to/kubeconfig
```

To run a specific test directory:

```bash
kubectl kuttl test ./integration/ --test <test-name>
```

## Directory Structure

```
integration/
├── kuttl-test.yaml    # KUTTL test suite configuration
├── README.md          # This file
└── assertions/        # Shared assertion templates
    └── .gitkeep
```

## Adding Tests

Each test should be in its own subdirectory containing:
- `00-install.yaml` - Resources to create
- `00-assert.yaml` - Assertions to verify the resources
- Optionally `00-errors.yaml` - Error assertions