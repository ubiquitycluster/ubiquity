# Kustomize and Helm relationship

Ubiquity uses Helm and Kustomize for different boundaries rather than treating them as interchangeable deployment systems.

## Boundary

- Helm is the component packaging boundary. System and platform components that expose reusable configuration should have a `Chart.yaml`, `values.yaml`, lint/template coverage, and Helm unittest coverage.
- Kustomize is the composition and environment-specific patches boundary. It is appropriate when a directory stitches together already-rendered resources, applies site overlays, or carries legacy workload layouts that are not yet a reusable chart.
- ArgoCD may sync either boundary, but the review question is different: Helm changes review values and chart templates; Kustomize changes review resource composition and patch behavior.

## `platform/hpc-ubiq`

`platform/hpc-ubiq` remains Kustomize-oriented because it carries HPC workload composition for Slurm, OpenPBS, HTCondor, storage examples, and site overlays. Those manifests are closer to environment-specific patches and workload examples than to a single reusable component chart.

This is an explicit exception to the general platform convention. New reusable services should default to Helm component packaging. Existing HPC overlays can remain Kustomize until they are split into stable components with clear values contracts and tests.

## Validation expectations

- Helm components: run `helm lint`, `helm template`, dependency checks, and Helm unittest coverage.
- Kustomize overlays: run `kubectl kustomize <overlay>` or `kustomize build <overlay>` and, where possible, Kubernetes server-side dry-run.
- Both paths require live proof before readiness claims. A rendered Helm chart or Kustomize build is intent, not service readiness.

See also `docs/reference/helm-charts.md` for the generated chart inventory and Kustomize roots under `platform/hpc-ubiq`.
