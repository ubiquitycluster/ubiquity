# Cilium + Hubble in Ubiquity

Cilium is an alternative CNI that replaces k3s's built-in flannel networking.
Hubble adds network observability (service maps, flows, metrics).

## When to use Cilium

- You need **network policies** beyond what k3s provides
- You want **Hubble observability** (network flow visualization)
- You prefer **eBPF-based** networking over flannel's overlay
- You want Cilium's **L2 LoadBalancer IP announcement** instead of MetalLB

## Migration from flannel + MetalLB

### Step 1: Configure k3s to disable flannel

Edit `metal/group_vars/metal.yml` and add:

```yaml
k3s_server_config:
  flannel-backend: none
  disable-kube-proxy: true
  disable-network-policy: true
  secrets-encryption: true
  disable:
    - local-storage
    - servicelb
    - traefik
```

This tells k3s to not start flannel, kube-proxy, or its network policy controller —
Cilium handles all of that.

### Step 2: Install Cilium via Helm

```bash
helm repo add cilium https://helm.cilium.io
helm repo update

helm install cilium cilium/cilium \
  --namespace kube-system \
  --values - <<EOF
operator:
  replicas: 1
kubeProxyReplacement: true
l2announcements:
  enabled: true
k8sServiceHost: 127.0.0.1
k8sServicePort: 6444
hubble:
  relay:
    enabled: true
  ui:
    enabled: true
EOF
```

The `k8sServiceHost: 127.0.0.1` and `k8sServicePort: 6444` are k3s-specific —
k3s exposes its API server on localhost:6444.

### Step 3: Apply Cilium resources

```bash
helm template system/cilium/ | kubectl apply -f -
```

This creates:
- `CiliumLoadBalancerIPPool` — IP range for LoadBalancer services
- `CiliumL2AnnouncementPolicy` — announces LoadBalancer IPs via ARP
- ConfigMap with install notes for reference

### Step 4: Remove MetalLB

```bash
kubectl delete namespace metallb-system
git rm system/metallb-system/
```

### Step 5: Access Hubble UI

```bash
kubectl port-forward -n kube-system svc/hubble-ui 12000:80
```

Then open http://localhost:12000

Or configure an Ingress via the values in `system/cilium/values.yaml`:
```yaml
hubble:
  ui:
    ingress:
      enabled: true
      host: hubble.yourdomain.com
```

## Verification

```bash
# Check Cilium status
kubectl -n kube-system exec daemonset/cilium -- cilium status

# Check Hubble relay
kubectl -n kube-system get pod -l k8s-app=hubble-relay

# Check CiliumLoadBalancerIPPool
kubectl get CiliumLoadBalancerIPPool

# Check L2 announcement
kubectl get CiliumL2AnnouncementPolicy
```

## Rollback

To go back to flannel + MetalLB:
1. Delete Cilium: `helm delete cilium -n kube-system`
2. Remove the k3s config changes from metal/group_vars/metal.yml
3. Reinstall MetalLB: `helm template system/metallb-system/ | kubectl apply -f -`