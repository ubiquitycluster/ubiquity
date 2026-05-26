# ADR-011: Use SOPS for secrets management

**Status:** Accepted

**Date:** 2026-05-26

## Context

Need to manage secrets (API tokens, passwords, TLS certs) in a GitOps-compatible way.
Previously used Vault (complex to operate) and Sealed Secrets (Kubernetes-specific).
The upstream homelab project uses SOPS with Age keys.

## Decision

Use SOPS with Age keys for encrypting secrets at rest. Encrypted files are committed to Git;
decryption happens at deploy time or via ArgoCD's SOPS plugin.

## Consequences

- Positive: Works offline (no dependency on Vault cluster)
- Positive: Encrypted files are standard YAML — Git diff shows encrypted content has changed
- Positive: Age keys are simpler than PGP/GPG
- Negative: Key management is manual (Age private key must be distributed)
- Neutral: Coexists with Vault (can use both for different use cases)