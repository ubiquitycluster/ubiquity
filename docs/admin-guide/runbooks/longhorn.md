# Longhorn

Longhorn is a distributed block storage system designed for Kubernetes that provides persistent storage for containerized workloads in Ubiquity clusters. It creates replicated block storage across multiple nodes to ensure high availability and data persistence.

## Overview

Longhorn provides distributed storage capabilities for Ubiquity by:

- Creating replicated block storage volumes across cluster nodes
- Enabling persistent storage for stateful applications
- Providing snapshot and backup capabilities
- Offering a web-based management interface
- Integrating with Kubernetes Storage Classes and Persistent Volume Claims

## Architecture

Longhorn consists of several key components:

### Manager Pods
- **longhorn-manager**: Runs on each node, manages volumes and handles orchestration
- **longhorn-driver**: CSI driver components for Kubernetes integration
- **longhorn-ui**: Web interface for management and monitoring

### Storage Components
- **Longhorn Engine**: Handles volume operations and replication
- **Replica Instances**: Store actual data blocks across multiple nodes
- **Recovery Backend**: Handles backup and restore operations

## Configuration

### Default Settings

The Longhorn system in Ubiquity is configured with these defaults:

```yaml
defaultSettings:
  defaultReplicaCount: 3
  disableSchedulingOnCordonedNode: true
  nodeDownPodDeletionPolicy: delete-both-statefulset-and-deployment-pod
  replicaAutoBalance: best-effort
  replicaSoftAntiAffinity: false
  storageMinimalAvailablePercentage: 10
  taintToleration: StorageNode=true:PreferNoSchedule
```

### Storage Classes

Longhorn provides the default storage class with:
- **Default replica count**: 3 replicas across different nodes
- **File system**: ext4
- **Replica auto-balance**: best-effort for even distribution

### Node Configuration

During installation, nodes are prepared with:
- NFS client tools for NFS support
- Dedicated `/var/lib/longhorn` partition (40-60% of data volume)
- Proper kernel modules loaded

## Accessing Longhorn UI

The Longhorn web interface is available at:
```
https://longhorn.ubiquitycluster.uk
```

Features include:
- Volume management and monitoring
- Node and disk management
- Backup and snapshot operations
- Performance metrics and health status

## Common Operations

### Creating Persistent Volumes

1. **Using Storage Class** (Recommended):
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: my-namespace
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: longhorn
  resources:
    requests:
      storage: 10Gi
```

2. **Direct Volume Creation**:
- Use Longhorn UI to create volumes manually
- Attach to nodes as needed

### Volume Operations

- **Snapshots**: Create point-in-time snapshots via UI or CLI
- **Backups**: Configure S3-compatible backup targets
- **Volume Expansion**: Resize volumes online through PVC expansion
- **Volume Migration**: Move volumes between nodes for maintenance

### Monitoring Integration

Longhorn integrates with Prometheus monitoring:
- ServiceMonitor automatically configured
- Metrics endpoint: `:9500/metrics`
- Grafana dashboards available for visualization

## Troubleshooting

### Common Issues

#### Volume Mount Failures
**Symptoms**: Pods stuck in ContainerCreating with volume mount errors

**Solutions**:
1. Check node disk space: `df -h /var/lib/longhorn`
2. Verify longhorn-manager pods are running: `kubectl get pods -n longhorn-system`
3. Check volume status in Longhorn UI

#### Replica Scheduling Problems
**Symptoms**: Volumes showing degraded state or insufficient replicas

**Solutions**:
1. Verify node labels and taints: `kubectl get nodes --show-labels`
2. Check disk space on storage nodes
3. Review replica anti-affinity settings in Longhorn UI

#### Performance Issues
**Symptoms**: Slow I/O operations or high latency

**Solutions**:
1. Monitor disk I/O on storage nodes
2. Check network connectivity between nodes
3. Review replica placement and consider rebalancing
4. Verify storage node resources (CPU/Memory)

### Diagnostic Commands

```bash
# Check Longhorn system status
kubectl get pods -n longhorn-system

# View Longhorn manager logs
kubectl logs -n longhorn-system -l app=longhorn-manager

# Check volume status
kubectl get pv,pvc -A

# View storage class configuration
kubectl get storageclass longhorn -o yaml

# Check node storage capacity
kubectl get nodes -o custom-columns=NAME:.metadata.name,CAPACITY:.status.capacity.storage
```

### Recovery Procedures

#### Node Failure Recovery
1. Longhorn automatically handles single node failures
2. Replicas on failed node will be rebuilt on healthy nodes
3. Monitor rebuild progress in Longhorn UI

#### Data Recovery from Backup
1. Configure backup target in Longhorn settings
2. Create regular volume backups
3. Restore from backup when needed via UI

## Maintenance

### Regular Tasks

1. **Monitor Storage Usage**: Keep storage utilization below 85%
2. **Check Replica Health**: Ensure all volumes have healthy replicas
3. **Backup Critical Volumes**: Schedule regular backups to external storage
4. **Node Maintenance**: Properly drain nodes before maintenance

### Storage Expansion

To add storage capacity:
1. Add new nodes with storage to the cluster
2. Label nodes appropriately for Longhorn scheduling
3. Longhorn will automatically discover and use new storage

### Single Node Adjustments

For single-node clusters, modify replica settings:
```yaml
# In system/longhorn-system/values.yaml
persistence:
  defaultClassReplicaCount: 1
```

## Integration with Ubiquity Components

### HPC Workloads
- Provides persistent storage for Slurm job data
- Supports shared storage scenarios through ReadWriteMany access modes
- Integrates with NFS for traditional HPC workflows

### Backup Integration
- Works with Velero for cluster-wide backup/restore
- Supports volume snapshots for application-consistent backups
- Integrates with external backup targets (S3, NFS)

### Monitoring Integration
- Metrics exposed to Prometheus
- Grafana dashboards for storage monitoring
- Alert rules for storage capacity and health

## Security Considerations

- Volume encryption at rest (when supported by underlying storage)
- Network traffic encryption between replicas
- RBAC integration for access control
- Secure backup target configuration

## Performance Tuning

### Optimization Tips
1. **Use local SSDs** for better performance
2. **Configure appropriate replica count** based on availability requirements
3. **Monitor and tune network performance** between storage nodes
4. **Use volume locality** settings for performance-critical workloads

### Resource Requirements
- **Minimum**: 2 CPU cores, 4GB RAM per storage node
- **Recommended**: 4+ CPU cores, 8GB+ RAM for production workloads
- **Storage**: Dedicated storage devices or partitions preferred
