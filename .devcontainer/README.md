# Devcontainer

This directory contains a VS Code devcontainer configuration for Ubiquity development.

## Usage

1. Install Docker and VS Code
2. Install the "Dev Containers" extension (ms-vscode-remote.remote-containers)
3. Open the ubiquity directory in VS Code
4. Click "Reopen in Container" when prompted
5. Wait for the container to build (first time takes a few minutes)
6. The CLI is already built at `ubiquity-cli`

## What's included

- Go 1.24
- Helm, kubectl
- Docker-in-Docker
- pre-commit with all hooks
- shellcheck, yamllint
- goreleaser, sops, govulncheck
- VS Code extensions for Go, YAML, Docker, shellcheck