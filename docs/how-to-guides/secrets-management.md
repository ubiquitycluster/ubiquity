# Secrets Management with SOPS

## Overview
SOPS (Secrets OPerationS) encrypts secrets using Age keys.
Encrypted files are committed to Git; decryption happens at deploy time.

## Setup
1. Install age: `apt install age` or `brew install age`
2. Install sops: `go install github.com/getsops/sops/v3/cmd/sops@latest`
3. Generate an Age key: `age-keygen -o ~/.config/sops/age/keys.txt`
4. Export the public key: `age-keygen -y ~/.config/sops/age/keys.txt`
5. Add the public key to .sops.yaml

## Usage
Edit a secret:
```
./scripts/sops-edit secrets/mysecret.enc.yaml
```

Decrypt a secret:
```
./scripts/sops-decrypt secrets/mysecret.enc.yaml
```

Decrypt and pipe to kubectl:
```
./scripts/sops-decrypt secrets/mysecret.enc.yaml | kubectl apply -f -
```

## Integration with External Secrets Operator
SOPS-encrypted files can be used with the External Secrets Operator
via the `sops` provider (or by pre-decrypting with a GitOps tool like
ArgoCD's SOPS plugin).