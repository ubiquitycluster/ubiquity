# CNI Support for MPI over InfiniBand

The following document is a HOWTO for how to enable Ubiquity support for using MPI over InfiniBand. Note that the referenced components are managed by their respective third party maintainers.

## Howto enable IPoIB in Ubiquity

Assumptions:
- working Ubiquity K3s cluster with a functional primary CNI
- required drivers already installed for InfiniBand host adapter - Installed via Ubiquity, using NVIDIA Netowork Operator
- an existing IPoIB interface is up and active on the hosts - Installed via Ubiquity. using Ansible
- this has only been tested with NVIDIA InfiniBand cards and Mellanox OFED - Other network card providers instructions may (will) differ.

### Required 3rd party tools
- [Multus](https://github.com/k8snetworkplumbingwg/multus-cni) project will be utilized to add the secondary IpoIB interface
- [ipoib-cni](https://github.com/Mellanox/ipoib-cni) from NVIDIA/Mellanox
- [Whereabouts](https://github.com/k8snetworkplumbingwg/whereabouts) for ipam

1. Install Multus by applying [multus-daemonset.yaml](https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml). 
```
kubectl apply -f multus-daemonset.yml
```
2. Deploy the [ipoib-cni](https://github.com/Mellanox/ipoib-cni/blob/master/images/ipoib-cni-daemonset.yaml).
```
kubectl apply -f ipoib-cni-daemonset.yaml
```
3. Install the whereabouts ipam.
```
git clone https://github.com/k8snetworkplumbingwg/whereabouts && cd whereabouts
kubectl apply \
    -f doc/crds/daemonset-install.yaml \
    -f doc/crds/whereabouts.cni.cncf.io_ippools.yaml \
    -f doc/crds/whereabouts.cni.cncf.io_overlappingrangeipreservations.yaml
```
4. Define the network for IPoIB with a NetworkAttachmentDefinition. Please modify ip range and master interface as required.

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: ubiquity-ipoib
spec:
  config: '{
  "cniVersion": "0.3.1",
  "type": "ipoib",
  "name": "ubiquity-ipoib",
  "master": "ib0",
  "ipam": {
    "type": "whereabouts",
    "range": "192.168.0.0/24"
  }
}'
```

5. Validate that containers can launch and attach this secondary network by running the following pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: ipoib-test
  annotations:
    k8s.v1.cni.cncf.io/networks: ubiquity-ipoib
spec:
  restartPolicy: OnFailure
  containers:
  - image: 
    name: ipoib-test
    securityContext:
      capabilities:
        add: [ "IPC_LOCK" ]
    resources:
      limits:
        .com/rdma: 1
    command:
    - sh
    - -c
    - |
      ls -l /dev/infiniband /sys/class/infiniband /sys/class/net
      sleep 1000000
```

### Rootless applications and Memlock rlimit when Infiniband

When using rootless applications, which is default with appdef V2, init executions do not posses 
sufficient privileges to set Memlock rlimit to unlimited, leading to failure of Infiniband usage.

Two possibilities are available to bypass this issue:

1. Cluster administrator need to force unlimited Memlock rlimites at docker / containerd level.
   * For docker, append `--default-ulimit memlock=-1:-1` inside docker service file, at `ExecStart=/usr/bin/dockerd` line, then reload systemctl and restart docker service.
   * For containerd, edit service file, and add under `[SERVICE]` line `LimitMEMLOCK=infinity`, then reload systemctl and restart containerd service.
2. Or allow apps to run init as root, before dropping to unprivileged user. To do so, set `JARVICE_APP_ALLOW_ROOT_INIT` value to `true` in `override.yaml`. This will unlock root usage for appdef V2 apps, which will set Memlock rlimit to unlimited before MPI execution.
