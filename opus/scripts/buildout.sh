#!/bin/bash

set -euo pipefail

# Color codes for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REGISTRY="${REGISTRY:-ghcr.io}"
NAMESPACE="${NAMESPACE:-ubiquity}"
IMAGE_NAME="${IMAGE_NAME:-opus}"
ANSIBLE_VERSION="${ANSIBLE_VERSION:-latest}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're authenticated to the registry
check_registry_auth() {
    log_info "Checking registry authentication for ${REGISTRY}..."
    
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        log_info "Using GITHUB_TOKEN for authentication"
        echo "${GITHUB_TOKEN}" | docker login "${REGISTRY}" -u "${GITHUB_ACTOR:-$(whoami)}" --password-stdin
    elif [ -n "${CR_PAT:-}" ]; then
        log_info "Using CR_PAT for authentication"
        echo "${CR_PAT}" | docker login "${REGISTRY}" -u "${GITHUB_ACTOR:-$(whoami)}" --password-stdin
    else
        log_warning "No GITHUB_TOKEN or CR_PAT found. Assuming already logged in to ${REGISTRY}"
        
        # Test if we can access the registry
        if ! docker pull hello-world >/dev/null 2>&1; then
            log_error "Docker daemon not accessible or not logged in to registry"
            log_info "To authenticate, run:"
            log_info "  echo \$GITHUB_TOKEN | docker login ${REGISTRY} -u USERNAME --password-stdin"
            exit 1
        fi
    fi
}

# Build and tag a specific flavour
build_and_tag() {
    local flavour="$1"
    local helm_version="${2:-}"
    local tag_suffix="$3"
    local local_tag=""
    local remote_tag="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:${tag_suffix}"
    
    log_info "Building ${flavour} flavour..."
    
    # Determine the build command and resulting tag
    if [ "$flavour" = "base" ]; then
        log_info "Building base image with Ansible ${ANSIBLE_VERSION}"
        make build ANSIBLE="${ANSIBLE_VERSION}"
        local_tag="localhost/opus:${ANSIBLE_VERSION}"
    elif [ -n "$helm_version" ]; then
        log_info "Building ${flavour} with Helm ${helm_version}"
        make build FLAVOUR="$flavour" HELM="$helm_version" ANSIBLE="${ANSIBLE_VERSION}"
        local_tag="localhost/opus:${ANSIBLE_VERSION}-${flavour}-all-helm${helm_version}"
    else
        log_info "Building ${flavour} flavour"
        make build FLAVOUR="$flavour" ANSIBLE="${ANSIBLE_VERSION}"
        local_tag="localhost/opus:${ANSIBLE_VERSION}-${flavour}"
    fi
    
    # Verify the local image was created
    if ! docker images "${local_tag}" --format "table {{.Repository}}:{{.Tag}}" | grep -q "${local_tag}"; then
        log_error "Local image ${local_tag} was not created successfully"
        return 1
    fi
    
    log_success "Local image created: ${local_tag}"
    
    # Tag for registry
    log_info "Tagging ${local_tag} -> ${remote_tag}"
    docker tag "${local_tag}" "${remote_tag}"
    
    # Push to registry
    log_info "Pushing ${remote_tag}"
    if docker push "${remote_tag}"; then
        log_success "Successfully pushed ${remote_tag}"
    else
        log_error "Failed to push ${remote_tag}"
        return 1
    fi
    
    # Show image size
    local size=$(docker images "${remote_tag}" --format "{{.Size}}" | head -1)
    log_info "Image size: ${size}"
    
    return 0
}

# Main build process
main() {
    log_info "Starting multi-flavour build for ${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}"
    log_info "Ansible version: ${ANSIBLE_VERSION}"
    log_info "Working directory: $(pwd)"
    
    # Check prerequisites
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v make >/dev/null 2>&1; then
        log_error "Make is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "Makefile" ] || [ ! -d "Dockerfiles" ]; then
        log_error "Please run this script from the opus directory"
        log_info "Expected structure: Makefile, Dockerfiles/, scripts/"
        exit 1
    fi
    
    # Check registry authentication
    check_registry_auth
    
    # Build all flavours
    local failed_builds=()
    
    log_info "Building base image..."
    if build_and_tag "base" "" "latest"; then
        log_success "Base build completed"
    else
        failed_builds+=("base")
        log_error "Base build failed"
    fi
    
    log_info "Building tools flavour..."
    if build_and_tag "tools" "" "latest-tools"; then
        log_success "Tools build completed"
    else
        failed_builds+=("tools")
        log_error "Tools build failed"
    fi
    
    log_info "Building AWS flavour..."
    if build_and_tag "aws" "" "latest-aws"; then
        log_success "AWS build completed"
    else
        failed_builds+=("aws")
        log_error "AWS build failed"
    fi
    
    log_info "Building AWS K8s flavour..."
    if build_and_tag "awsk8s" "" "latest-awsk8s"; then
        log_success "AWS K8s build completed"
    else
        failed_builds+=("awsk8s")
        log_error "AWS K8s build failed"
    fi
    
    log_info "Building AWS Helm flavour..."
    if build_and_tag "awshelm" "3.10" "latest-awshelm3.10"; then
        log_success "AWS Helm build completed"
    else
        failed_builds+=("awshelm")
        log_error "AWS Helm build failed"
    fi
    
    log_info "Building Opus all-in-one with Helm..."
    if build_and_tag "opus" "3.10" "latest-opus-all-helm3.10"; then
        log_success "Opus all-in-one build completed"
    else
        failed_builds+=("opus")
        log_error "Opus all-in-one build failed"
    fi
    
    # Summary
    echo
    log_info "Build Summary:"
    if [ ${#failed_builds[@]} -eq 0 ]; then
        log_success "All builds completed successfully!"
        log_info "Images pushed to ${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:"
        log_info "  ✓ latest"
        log_info "  ✓ latest-tools"
        log_info "  ✓ latest-aws"
        log_info "  ✓ latest-awsk8s"
        log_info "  ✓ latest-awshelm3.10"
        log_info "  ✓ latest-opus-all-helm3.10"
    else
        log_error "Some builds failed:"
        for build in "${failed_builds[@]}"; do
            log_error "  ✗ ${build}"
        done
        exit 1
    fi
}

# Handle script interruption
cleanup() {
    log_warning "Build process interrupted"
    exit 130
}

trap cleanup INT TERM

# Run main function
main "$@"