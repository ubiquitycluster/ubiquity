# Alerts

## etcdDatabaseHighFragmentationRatio

Example:

```yaml
name: etcdDatabaseHighFragmentationRatio
expr: (last_over_time(etcd_mvcc_db_total_size_in_use_in_bytes[5m]) / last_over_time(etcd_mvcc_db_total_size_in_bytes[5m])) < 0.5
for: 10m
labels:
  severity: warning
annotations:
  description: etcd cluster "{{ $labels.job }}": database size in use on instance {{ $labels.instance }} is {{ $value | humanizePercentage }} of the actual allocated disk space, please run defragmentation (e.g. etcdctl defrag) to retrieve the unused fragmented disk space.
  runbook_url: https://etcd.io/docs/v3.5/op-guide/maintenance/#defragmentation
  summary: etcd database size in use is less than 50% of the actual allocated storage.
```

To fix on a kubespray installation:

- On each affected etcd member, as `root`:

  ```shell
  . /etc/etcd.env
  export ETCDCTL_API
  export ETCDCTL_CERT
  export ETCDCTL_KEY
  export ETCDCTL_CACERT
  export ETCDCTL_ENDPOINTS
  etcdctl defrag
  ```

To fix on a cluster-api installation:

```shell
first_etcd_pod="$(kubectl get pods  -n kube-system --selector=component=etcd -A -o name | head -n 1)"
kubectl exec -n kube-system "$first_etcd_pod" -- \
  etcdctl defrag --cluster \
  --cacert /etc/kubernetes/pki/etcd/ca.crt \
  --key /etc/kubernetes/pki/etcd/server.key \
  --cert /etc/kubernetes/pki/etcd/server.crt
```

## Other alerts

See [kube-prometheus runbooks](https://runbooks.prometheus-operator.dev/).