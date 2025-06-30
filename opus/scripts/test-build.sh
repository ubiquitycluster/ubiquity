#!/bin/bash

set -euo pipefail

# Configuration for testing
REGISTRY="${REGISTRY:-ghcr.io}"
NAMESPACE="${NAMESPACE:-ubiquity}"
IMAGE_NAME="${IMAGE_NAME:-opus}"
ANSIBLE_VERSION="${ANSIBLE_VERSION:-latest}"

echo "Testing Opus build process..."
echo "Registry: ${REGISTRY}"
echo "Namespace: ${NAMESPACE}"
echo "Image: ${IMAGE_NAME}"
echo "Ansible Version: ${ANSIBLE_VERSION}"

# Test building just the base image
echo "Building base image..."
make build ANSIBLE="${ANSIBLE_VERSION}"

# Check if the image was built successfully
if docker images localhost/opus:${ANSIBLE_VERSION} --format "table {{.Repository}}:{{.Tag}}" | grep -q "localhost/opus:${ANSIBLE_VERSION}"; then
    echo "✅ Base image built successfully: localhost/opus:${ANSIBLE_VERSION}"
    
    # Test tagging for ghcr.io (but don't push without authentication)
    REMOTE_TAG="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}:latest-test"
    echo "Tagging for ${REMOTE_TAG}..."
    docker tag "localhost/opus:${ANSIBLE_VERSION}" "${REMOTE_TAG}"
    
    echo "✅ Image tagged successfully: ${REMOTE_TAG}"
    echo "Image is ready for push to ${REGISTRY}"
    
    # Show image info
    echo "Image details:"
    docker inspect "${REMOTE_TAG}" --format='{{.Config.Labels}}' | grep -o '"[^"]*":"[^"]*"' | head -5
    
else
    echo "❌ Base image build failed"
    exit 1
fi

echo "Test completed successfully!"
