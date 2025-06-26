Bare metal hosts in ‘provisioned registration error’ state after update
After update of a management or managed cluster created using the Container Cloud release earlier than 2.6.0, a bare metal host state is Provisioned in the Container Cloud web UI while having the error state in logs with the following message:

status:
  errorCount: 1
  errorMessage: 'Host adoption failed: Error while attempting to adopt node  7a8d8aa7-e39d-48ec-98c1-ed05eacc354f:
    Validation of image href http://10.10.10.10/images/stub_image.qcow2 failed,
    reason: Got HTTP code 404 instead of 200 in response to HEAD request..'
  errorType: provisioned registration error
The issue is caused by the image URL pointing to an unavailable resource due to the URI IP change during update. To apply the issue resolution, update URLs for the bare metal host status and spec with the correct values that use a stable DNS record as a host.

To apply the issue resolution:

Note

In the commands below, we update master-2 as an example. Replace it with the corresponding value to fit your deployment.

Exit Lens.

In a new terminal, configure access to the affected cluster.

Start kube-proxy:

kubectl proxy &
Pause the reconcile:

kubectl patch bmh master-2 --type=merge --patch '{"metadata":{"annotations":{"baremetalhost.metal3.io/paused": "true"}}}'
Create the payload data with the following content:

For status_payload.json:

{
   "status": {
      "errorCount": 0,
      "errorMessage": "",
      "provisioning": {
         "image": {
            "checksum": "http://httpd-http/images/stub_image.qcow2.md5sum",
            "url": "http://httpd-http/images/stub_image.qcow2"
         },
         "state": "provisioned"
      }
   }
}
For status_payload.json:

{
   "spec": {
      "image": {
         "checksum": "http://httpd-http/images/stub_image.qcow2.md5sum",
         "url": "http://httpd-http/images/stub_image.qcow2"
      }
   }
}
Verify that the payload data is valid:

cat status_payload.json | jq
cat spec_payload.json | jq
The system response must contain the data added in the previous step.

Patch the bare metal host status with payload:

curl -k -v -XPATCH -H "Accept: application/json" -H "Content-Type: application/merge-patch+json" --data-binary "@status_payload.json" 127.0.0.1:8001/apis/metal3.io/v1alpha1/namespaces/default/baremetalhosts/master-2/status
Patch the bare metal host spec with payload:

kubectl patch bmh master-2 --type=merge --patch "$(cat spec_payload.json)"
Resume the reconcile:

kubectl patch bmh master-2 --type=merge --patch '{"metadata":{"annotations":{"baremetalhost.metal3.io/paused":null}}}'
Close the terminal to quit kube-proxy and resume Lens.
