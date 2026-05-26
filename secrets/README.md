# Secrets

This directory contains encrypted secrets using SOPS + Age.

## Setup
1. Generate an Age key: `age-keygen -o ~/.config/sops/age/keys.txt`
2. Export the public key: `age-keygen -y ~/.config/sops/age/keys.txt`
3. Replace the placeholder key in .sops.yaml with your public key

## Usage
- Edit a secret: `./scripts/sops-edit secrets/mysecret.enc.yaml`
- Decrypt: `./scripts/sops-decrypt secrets/mysecret.enc.yaml`
- Decrypt to kubectl: `./scripts/sops-decrypt secrets/mysecret.enc.yaml | kubectl apply -f -`