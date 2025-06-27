# BareMetalHost Provisioning Error Resolution

## Overview

This troubleshooting guide addresses bare metal hosts that enter a 'provisioned registration error' state after Ubiquity cluster updates. This typically occurs when the image URL references become invalid due to IP address changes during the update process.

## Problem Description

After updating a Ubiquity cluster, bare metal hosts may show as "Provisioned" in the status but actually be in an error state. The BareMetalHost resource will display an error message similar to:

```yaml
status:
  errorCount: 1
  errorMessage: 'Host adoption failed: Error while attempting to adopt node 7a8d8aa7-e39d-48ec-98c1-ed05eacc354f:
    Validation of image href http://10.10.10.10/images/ubiquity-node-rocky9.qcow2 failed,
    reason: Got HTTP code 404 instead of 200 in response to HEAD request..'
  errorType: provisioned registration error
```

## Root Cause

This issue occurs when:
- The Ubiquity cluster's internal networking configuration changes during updates
- Image URLs contain hardcoded IP addresses instead of stable DNS names
- The httpd-http service serving images becomes unreachable at the old endpoint

## Resolution Steps

Follow these steps to resolve the provisioning error and restore proper BareMetalHost functionality:

### Prerequisites

- Access to the Ubiquity cluster kubectl command line
- Administrative privileges on the cluster
- Knowledge of the affected BareMetalHost resource name

> **Note**: In the examples below, we update `master-2` as the BareMetalHost name. Replace this with your actual BareMetalHost resource name.

### Step 1: Configure Cluster Access

Ensure you have proper access to the affected Ubiquity cluster:

```bash
# Verify cluster access
kubectl get nodes

# Confirm BareMetalHost resources
kubectl get bmh -A
```

### Step 2: Start kubectl proxy

Start the Kubernetes API proxy to enable direct API access:

```bash
kubectl proxy &
```

### Step 3: Pause BareMetalHost Reconciliation

Temporarily pause the Bare Metal Operator reconciliation for the affected host:

```bash
kubectl patch bmh master-2 --type=merge --patch '{"metadata":{"annotations":{"baremetalhost.metal3.io/paused": "true"}}}'
```

### Step 4: Create Payload Files

Create the necessary payload files with corrected image URLs that use the stable httpd-http service endpoint.

**Create `status_payload.json`:**

```json
{
   "status": {
      "errorCount": 0,
      "errorMessage": "",
      "provisioning": {
         "image": {
            "checksum": "http://httpd-http/images/ubiquity-node-rocky9.qcow2.md5sum",
            "url": "http://httpd-http/images/ubiquity-node-rocky9.qcow2"
         },
         "state": "provisioned"
      }
   }
}
```

**Create `spec_payload.json`:**

```json
{
   "spec": {
      "image": {
         "checksum": "http://httpd-http/images/ubiquity-node-rocky9.qcow2.md5sum",
         "url": "http://httpd-http/images/ubiquity-node-rocky9.qcow2"
      }
   }
}
```

> **Image Reference Note**: The example above uses `ubiquity-node-rocky9.qcow2`, which is the standard Ubiquity Rocky Linux 9 image. Replace this with your actual image name if different. Common Ubiquity images include:
> - `ubiquity-node-rocky9.qcow2` - Standard Ubiquity node image
> - `ubiquity-hpc-rocky9.qcow2` - HPC-optimized image with InfiniBand support
> - `ubiquity-gpu-rocky9.qcow2` - GPU-enabled compute image

### Step 5: Validate Payload Files

Verify that the payload files are valid JSON:

```bash
cat status_payload.json | jq
cat spec_payload.json | jq
```

The command output should display the JSON structure without errors, confirming valid syntax.

### Step 6: Apply Status Patch

Update the BareMetalHost status using the Kubernetes API:

```bash
curl -k -v -XPATCH \
  -H "Accept: application/json" \
  -H "Content-Type: application/merge-patch+json" \
  --data-binary "@status_payload.json" \
  127.0.0.1:8001/apis/metal3.io/v1alpha1/namespaces/default/baremetalhosts/master-2/status
```

### Step 7: Apply Spec Patch

Update the BareMetalHost specification:

```bash
kubectl patch bmh master-2 --type=merge --patch "$(cat spec_payload.json)"
```

### Step 8: Resume Reconciliation

Re-enable the Bare Metal Operator reconciliation:

```bash
kubectl patch bmh master-2 --type=merge --patch '{"metadata":{"annotations":{"baremetalhost.metal3.io/paused":null}}}'
```

### Step 9: Clean Up

Stop the kubectl proxy process:

```bash
# Stop the background kubectl proxy
pkill -f "kubectl proxy"

# Or bring the background process to foreground and stop with Ctrl+C
fg
```

## Verification

After completing the resolution steps, verify that the issue has been resolved:

```bash
# Check BareMetalHost status
kubectl get bmh master-2 -o yaml

# Verify the error count is 0 and errorMessage is empty
kubectl get bmh master-2 -o jsonpath='{.status.errorCount}'
kubectl get bmh master-2 -o jsonpath='{.status.errorMessage}'

# Confirm the image URLs are updated
kubectl get bmh master-2 -o jsonpath='{.spec.image.url}'
```

## Prevention

To prevent this issue in future Ubiquity cluster updates:

1. **Use Stable DNS Names**: Always configure image URLs with stable DNS service names like `httpd-http` instead of IP addresses
2. **Image Management**: Use Ubiquity's image building tools in `tools/disk-image/mkimage/` to create standardized images
3. **Update Procedures**: Follow Ubiquity's recommended update procedures that preserve service endpoint stability

## Related Documentation

- [Operating System Images for Ubiquity](../runbooks/osimages.md) - Comprehensive guide for building and managing Ubiquity images
- [Metal3 API Reference](../../reference/metal3/api.md) - BareMetalHost resource specification
- [Bare Metal Operator](../../reference/metal3/bmh_live_iso.md) - Advanced BareMetalHost configuration options
