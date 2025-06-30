# Opus GitHub Container Registry (GHCR) Setup

This document explains how to build and push Opus images to GitHub Container Registry (ghcr.io) both locally and via GitHub Actions.

## Overview

The Opus project now supports building and pushing container images to GitHub Container Registry (ghcr.io) instead of Azure Container Registry. This provides:

- ✅ Free public container hosting via GitHub
- ✅ Integrated with GitHub repository permissions
- ✅ Automated builds via GitHub Actions
- ✅ Security scanning with Trivy
- ✅ Multi-architecture support

## Repository Structure

```
opus/
├── Dockerfiles/          # Docker build files
│   ├── builder           # Base builder image (Rocky Linux 9)
│   ├── Dockerfile        # Base Opus image
│   └── Dockerfile-*      # Flavour-specific images
├── scripts/
│   ├── buildout.sh       # Multi-flavour build script for GHCR
│   └── test-build.sh     # Test script for local builds
└── Makefile              # Build automation
```

## Available Image Flavours

The following image variants are built and published:

| Flavour | Tag | Description |
|---------|-----|-------------|
| Base | `latest` | Core Ansible installation |
| Tools | `latest-tools` | Base + additional tooling |
| AWS | `latest-aws` | Base + AWS CLI and tools |
| AWS K8s | `latest-awsk8s` | Base + AWS + Kubernetes tools |
| AWS Helm | `latest-awshelm3.10` | Base + AWS + Helm 3.10 |
| Opus All | `latest-opus-all-helm3.10` | All tools + Helm 3.10 |

## Local Development

### Prerequisites

- Docker installed and running
- Make installed
- GitHub CLI (`gh`) for authentication (optional)

### Building Images Locally

1. **Build base image only:**
```bash
cd opus
make build
```

2. **Build specific flavour:**
```bash
cd opus
make build FLAVOUR=aws
make build FLAVOUR=tools
make build FLAVOUR=awshelm HELM=3.10
```

3. **Build all flavours for GHCR:**
```bash
cd opus
# Set environment variables (optional)
export REGISTRY=ghcr.io
export NAMESPACE=ubiquity
export IMAGE_NAME=opus

# Run the buildout script
./scripts/buildout.sh
```

### Authentication for Local Push

To push images locally, authenticate with GitHub Container Registry:

```bash
# Option 1: Using GitHub CLI
gh auth token | docker login ghcr.io -u USERNAME --password-stdin

# Option 2: Using Personal Access Token
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Option 3: Using Classic Personal Access Token
echo $CR_PAT | docker login ghcr.io -u USERNAME --password-stdin
```

**Required token permissions:**
- `write:packages` - To push images
- `read:packages` - To pull images
- `delete:packages` - To delete images (optional)

## GitHub Actions Automation

### Workflow File

The GitHub Actions workflow is located at `.github/workflows/opus-build.yml` and automatically:

- Builds all image flavours in parallel
- Pushes to `ghcr.io/ubiquity/opus:*`
- Runs security scans with Trivy
- Tests images after build
- Tags images with commit SHA for traceability

### Triggering Builds

**Automatic triggers:**
- Push to `main` or `develop` branches with changes in `opus/` directory
- Pull requests to `main` with changes in `opus/` directory

**Manual trigger:**
```bash
# Via GitHub CLI
gh workflow run "Build and Push Opus Images"

# Via GitHub web interface
# Go to Actions tab -> "Build and Push Opus Images" -> "Run workflow"
```

### Environment Variables

The workflow uses these environment variables:

| Variable | Value | Description |
|----------|-------|-------------|
| `REGISTRY` | `ghcr.io` | Container registry URL |
| `IMAGE_NAME` | `opus` | Base image name |
| `GITHUB_TOKEN` | Auto-provided | GitHub authentication token |

## Image Usage

### Pulling Images

```bash
# Pull latest base image
docker pull ghcr.io/ubiquity/opus:latest

# Pull specific flavour
docker pull ghcr.io/ubiquity/opus:latest-aws
docker pull ghcr.io/ubiquity/opus:latest-tools

# Pull with commit SHA (for reproducibility)
docker pull ghcr.io/ubiquity/opus:latest-abc1234
```

### Using in Docker Compose

```yaml
version: '3.8'
services:
  ansible:
    image: ghcr.io/ubiquity/opus:latest-aws
    volumes:
      - ./playbooks:/ansible/playbooks
    environment:
      - AWS_REGION=us-west-2
```

### Using in Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: opus-runner
spec:
  replicas: 1
  selector:
    matchLabels:
      app: opus-runner
  template:
    metadata:
      labels:
        app: opus-runner
    spec:
      containers:
      - name: opus
        image: ghcr.io/ubiquity/opus:latest-awsk8s
        command: ["ansible-playbook"]
        args: ["/ansible/playbooks/deploy.yml"]
```

## Security

### Image Scanning

All images are automatically scanned for vulnerabilities using Trivy:
- Scans run on every push to `main`
- Results are uploaded to GitHub Security tab
- SARIF format for integration with GitHub Advanced Security

### Access Control

Images inherit repository permissions:
- **Public repositories**: Images are publicly readable
- **Private repositories**: Images require authentication
- **Organization repositories**: Follow organization settings

### Best Practices

1. **Use specific tags** rather than `latest` in production
2. **Regular updates** - Images are rebuilt on code changes
3. **Monitor security alerts** in the GitHub Security tab
4. **Use commit SHA tags** for absolute reproducibility

## Troubleshooting

### Common Issues

**Build fails in GitHub Actions:**
- Check the Actions tab for detailed logs
- Verify Dockerfile syntax
- Ensure all required packages are available

**Authentication fails:**
- Check token permissions (`write:packages`)
- Verify token is not expired
- Use `docker logout ghcr.io` then re-authenticate

**Image not found:**
- Verify image name and tag
- Check if repository/organization is correct
- Ensure you have pull permissions

### Debug Commands

```bash
# Check local images
docker images | grep opus

# Test image functionality
docker run --rm ghcr.io/ubiquity/opus:latest ansible --version

# Check image labels
docker inspect ghcr.io/ubiquity/opus:latest --format='{{.Config.Labels}}'

# View build logs
gh run list --workflow="Build and Push Opus Images"
gh run view <run-id> --log
```

## Migration from Azure CR

If migrating from Azure Container Registry:

1. **Update CI/CD pipelines** to use new image URLs
2. **Update documentation** with new registry paths  
3. **Verify all automation** uses GitHub authentication
4. **Update deployment scripts** with new image names

**Old format:** `ubiquity.azurecr.io/opus:latest`  
**New format:** `ghcr.io/ubiquity/opus:latest`

## Contributing

When making changes to Opus:

1. Create a feature branch
2. Modify Dockerfiles or scripts as needed
3. Test locally using `make build`
4. Create a pull request
5. Automated builds will test your changes
6. Merge to `main` triggers production builds

For questions or issues, please open a GitHub issue in the repository.
