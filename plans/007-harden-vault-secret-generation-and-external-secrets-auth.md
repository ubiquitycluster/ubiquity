# Plan 007: Harden Vault secret generation and External Secrets authentication

> **Executor instructions**: Remove steady-state root-token dependence and in-cluster runtime builds. Do not print secret values. Add chart/document tests that inspect references only.
>
> **Drift check (run first)**: `git diff --stat efd46ed..HEAD -- system/vault/templates system/vault/files/generate-secrets platform/external-secrets/templates pkg/cloud`

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: HIGH
- **Depends on**: none
- **Category**: security/supply-chain
- **Planned at**: commit `efd46ed`, 2026-06-15
- **Status**: DONE (implemented and verified)

## Why this matters

The current Vault chart runs `go get`/`go run` inside the cluster while holding a Vault root token, and External Secrets uses the stored root token as a cluster-wide credential. This is too much privilege and too much mutable supply chain at runtime.

## Current state

- `system/vault/templates/generate-secrets-job.yaml:18-53` defines a CronJob using `golang:1.17-alpine`, `VAULT_TOKEN` from `vault-unseal-keys/vault-root`, and `go get . && go run .`.
- `system/vault/templates/generate-secrets-source.yaml:18-24` mounts source code from a ConfigMap.
- `system/vault/templates/cr.yaml:124-134` stores the root token by default.
- `platform/external-secrets/templates/clustersecretstore.yaml:18-31` uses `vault-unseal-keys` key `vault-root`.

## Commands you will need

| Purpose | Command | Expected on success |
|---|---|---|
| Security contract tests | `go test ./pkg/cloud -run 'Vault|ExternalSecrets|Security' -count=1` | exits 0 |
| Render Vault | `helm template vault system/vault --namespace vault >/tmp/vault.yaml` | exits 0 and contains no `go get .` / `go run .` |
| Render External Secrets | `helm template external-secrets platform/external-secrets --namespace external-secrets >/tmp/external-secrets.yaml` | exits 0 and does not reference `vault-root` |

## Scope

**In scope**:
- `system/vault/templates/generate-secrets-job.yaml`
- `system/vault/templates/cr.yaml`
- `platform/external-secrets/templates/clustersecretstore.yaml`
- security contract tests under `pkg/cloud`

**Out of scope**:
- Decoding, printing, or rotating actual live secrets.
- Building/publishing a container image in this branch.

## Steps

1. Add tests that fail if rendered charts contain `go get .`, `go run .`, `golang:1.17-alpine`, `vault-root` in External Secrets auth, or `storeRootToken: true` as a steady-state default.
2. Change the generate-secrets job to use a pinned configurable image and command, defaulting to a project-owned image reference with a non-root token Secret/key name.
3. Remove the ConfigMap source-code mount from the job if the pinned image contains the binary.
4. Change Vault CR default to avoid storing the root token unless explicitly opted in.
5. Change External Secrets to reference a scoped token Secret/key such as `vault-external-secrets-token` / `token`.
6. Document that live clusters must migrate and rotate any previously stored root token.

## Done criteria

- [ ] Runtime `go get`/`go run` removed from Vault chart.
- [ ] External Secrets no longer references `vault-root`.
- [ ] Vault root-token storage is not enabled by default.
- [ ] Focused tests and Helm renders pass.

## STOP conditions

Stop if the current Vault operator requires `storeRootToken: true` for bootstrap and no scoped token can be referenced without live operator changes; document the operator constraint and propose a two-phase migration.
