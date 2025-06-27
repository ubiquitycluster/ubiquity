# Onyxia

Onyxia is a self-service data science platform that provides researchers and data scientists with on-demand access to containerized data science tools, HPC environments, and interactive development environments. In the Ubiquity platform, Onyxia serves as the primary interface for self-service HPC and data science workloads.

## Overview

Onyxia in Ubiquity serves multiple critical functions:

- **Self-Service Data Science Platform**: Provides on-demand access to Jupyter notebooks, RStudio, VS Code, and other data science tools
- **HPC Job Launcher**: Integrates with SLURM and HTCondor for submitting and managing HPC workloads
- **Interactive Computing Environment**: Offers containerized environments with pre-configured software stacks
- **Multi-User Platform**: Supports user isolation and resource quotas through Kubernetes namespaces
- **Catalog Management**: Provides curated collections of data science and HPC applications

## Architecture

### Core Components

- **Onyxia Web UI**: React-based web interface for service management
- **Onyxia API**: Backend service for orchestrating Kubernetes deployments
- **Helm Chart Catalogs**: Collections of pre-configured applications and services
- **Service Discovery**: Integration with Kubernetes for service management
- **Authentication Integration**: OAuth2/OIDC integration with Keycloak

### High Availability Configuration

Onyxia is configured for high availability with:
- **Single Replica API**: Lightweight API server on master nodes
- **Node Selector**: Runs exclusively on control plane nodes for stability
- **Session Affinity**: Cookie-based routing for consistent user experience
- **Ingress Load Balancing**: NGINX-based load balancing with SSL termination

## Configuration

### Default Settings

```yaml
onyxia:
  serviceAccount:
    clusterAdmin: true
  ui:
    image:
      name: inseefrlab/onyxia-web
      version: 2.13.53
    nodeSelector:
      node-role.kubernetes.io/master: "true"
    env:
      KEYCLOAK_REALM: ubiquity
      KEYCLOAK_CLIENT_ID: ubiquity-client
      THEME_ID: ultraviolet
      HEADER_ORGANIZATION: Ubiquity
      HEADER_USECASE_DESCRIPTION: HPCLab
```

### Ingress Configuration

Onyxia is accessible via HTTPS with:
- **Domain**: `datalab.ubiquitycluster.uk`
- **TLS**: Automatic certificate management via cert-manager
- **Session Affinity**: Cookie-based load balancing for consistent sessions
- **CORS**: Enabled for cross-origin requests

### Authentication Configuration

- **Provider**: Keycloak OAuth2/OIDC
- **Realm**: `ubiquity`
- **Client ID**: `ubiquity-client`
- **JWT Token**: Used for API authentication
- **User Namespace Isolation**: Automatic namespace creation per user

## Accessing Onyxia

### Web Interface

Access the Onyxia web interface at:
```
https://datalab.ubiquitycluster.uk
```

### Authentication Flow

1. **Initial Access**: Navigate to Onyxia URL
2. **Keycloak Redirect**: Automatic redirect to Keycloak for authentication
3. **User Credentials**: Login with Ubiquity cluster credentials
4. **Token Exchange**: OAuth2 token exchange for API access
5. **Dashboard Access**: Access to personalized Onyxia dashboard

### User Dashboard Features

- **Service Catalog**: Browse available data science tools and HPC applications
- **Running Services**: Monitor and manage active deployments
- **File Browser**: Access to persistent storage and shared filesystems
- **Configuration Management**: Personal settings and preferences
- **Resource Monitoring**: View resource usage and quotas

## Service Catalogs

### Available Catalogs

**Logicalis Datascience Catalog:**
- **Repository**: `https://cjcshadowsan.github.io/helm-charts-datascience`
- **Maintainer**: chris.coates@uk.logicalis.com
- **Status**: Production
- **Content**: Custom data science tools and HPC applications

**InseeFrLab Datascience Catalog:**
- **Repository**: `https://inseefrlab.github.io/helm-charts-datascience`
- **Maintainer**: innovation@insee.fr
- **Status**: Production
- **Content**: Comprehensive data science and analytics tools

**InseeFrLab Interactive Services:**
- **Repository**: `https://inseefrlab.github.io/helm-charts-interactive-services`
- **Maintainer**: innovation@insee.fr
- **Status**: Production
- **Content**: Interactive development environments and tools

### Common Applications

**Data Science Tools:**
- Jupyter Notebooks (Python, R, Scala)
- RStudio Server
- VS Code Server
- Apache Zeppelin

**Analytics Platforms:**
- Apache Spark
- Apache Flink
- Dask
- Ray

**Machine Learning:**
- MLflow
- Kubeflow
- TensorFlow Serving
- PyTorch

**Databases:**
- PostgreSQL
- MongoDB
- Redis
- InfluxDB

**HPC Tools:**
- SLURM Job Submission Interface
- HTCondor Submission Portal
- Parallel Computing Environments
- GPU-accelerated Computing

## User Namespace Management

### Automatic Namespace Creation

Onyxia automatically creates isolated namespaces for users:

**Namespace Pattern:**
- **Individual Users**: `user-{username}`
- **Group Projects**: `project-{groupname}`
- **Username Prefix**: `oidc-` for OpenID Connect users

### Resource Isolation

**Features:**
- **CPU/Memory Quotas**: Configurable per user/group
- **Storage Quotas**: Persistent volume claim limits
- **Network Policies**: Isolation between user workspaces
- **Pod Security**: Security contexts and policies

### Quota Configuration

Default quotas can be configured per region:
```yaml
quotas:
  enabled: true
  allowUserModification: false
  default:
    requests.storage: 1Gi
    count/pods: "10"
```

## Integration with Ubiquity Components

### Keycloak Authentication

**Single Sign-On Integration:**
- Unified authentication across Ubiquity services
- Role-based access control (RBAC)
- Group membership management
- Multi-factor authentication support

### Vault Integration

**Secret Management:**
- Automatic injection of secrets into user environments
- Database credentials and API keys
- Secure storage of user configurations
- Integration with service authentication

### Storage Integration

**Persistent Storage:**
- Longhorn-backed persistent volumes
- NFS shared storage integration
- S3-compatible object storage (future)
- User home directories and shared project spaces

### HPC Integration

**SLURM Integration:**
- Direct job submission from Onyxia services
- Resource allocation and scheduling
- Job monitoring and management
- Interactive and batch job support

**HTCondor Integration:**
- High-throughput computing workflows
- Container-based job execution
- Distributed computing across cluster nodes

## Service Deployment and Management

### Launching Services

**Via Web Interface:**
1. Browse service catalog
2. Select desired application/tool
3. Configure service parameters
4. Deploy to personal namespace
5. Access via generated URL

**Service Configuration Options:**
- Resource requests (CPU, memory, GPU)
- Storage requirements and persistence
- Environment variables and secrets
- Network configuration and ingress
- Security contexts and policies

### Service Lifecycle Management

**Operations:**
- **Start/Stop**: Pause and resume services
- **Scale**: Adjust resource allocation
- **Update**: Upgrade to new versions
- **Delete**: Clean up unused services
- **Clone**: Duplicate service configurations

### Persistent Data

**Storage Options:**
- **Personal Storage**: User-specific persistent volumes
- **Shared Storage**: Project-based shared filesystems
- **Temporary Storage**: Ephemeral storage for compute jobs
- **External Storage**: Integration with external data sources

## Monitoring and Administration

### Service Monitoring

**User Monitoring:**
- Resource usage dashboards
- Service health status
- Performance metrics
- Cost tracking (resource consumption)

**Administrator Monitoring:**
- Platform-wide resource utilization
- User activity and service deployments
- Catalog usage statistics
- Performance and capacity planning

### Log Management

**User Logs:**
```bash
# Access service logs via Onyxia UI
# Or via kubectl
kubectl logs -n user-{username} {service-pod-name}
```

**Platform Logs:**
```bash
# Onyxia API logs
kubectl logs -n onyxia deployment/onyxia-api

# Web UI logs (nginx)
kubectl logs -n onyxia deployment/onyxia-ui
```

## Troubleshooting

### Common Issues

#### Service Deployment Failures
**Symptoms**: Services fail to start or remain in pending state

**Solutions:**
1. Check resource quotas: `kubectl describe quota -n user-{username}`
2. Verify image pull permissions: `kubectl describe pod -n user-{username} {pod-name}`
3. Validate storage availability: `kubectl get pvc -n user-{username}`
4. Review security policies: `kubectl describe psp -n user-{username}`

#### Authentication Problems
**Symptoms**: Unable to login or access services

**Solutions:**
1. Verify Keycloak connectivity: `curl -I https://keycloak.ubiquitycluster.uk/auth`
2. Check OAuth client configuration: Review Keycloak admin console
3. Validate token exchange: Check browser developer tools for API errors
4. Verify user permissions: Check Keycloak user roles and groups

#### Resource Exhaustion
**Symptoms**: Cannot deploy new services or services are terminated

**Solutions:**
1. Check cluster resources: `kubectl top nodes`
2. Review user quotas: `kubectl get resourcequota -n user-{username}`
3. Clean up unused services: Delete stopped services via Onyxia UI
4. Monitor storage usage: `kubectl get pvc -n user-{username}`

### Diagnostic Commands

```bash
# Check Onyxia components
kubectl get all -n onyxia

# View API configuration
kubectl get configmap onyxia-api-config -n onyxia -o yaml

# Check user namespaces
kubectl get namespaces | grep "user-\|project-"

# Monitor resource usage
kubectl top pods -n onyxia
kubectl describe node {node-name}

# Check ingress status
kubectl get ingress -n onyxia
kubectl describe certificate onyxia-tls-certificate -n onyxia
```

### Recovery Procedures

#### Service Recovery
```bash
# Restart Onyxia API
kubectl rollout restart deployment/onyxia-api -n onyxia

# Restart Web UI
kubectl rollout restart deployment/onyxia-ui -n onyxia

# Clean up stuck user services
kubectl delete pods --field-selector=status.phase=Failed -n user-{username}
```

#### Configuration Recovery
```bash
# Verify Keycloak integration
kubectl exec -n onyxia deployment/onyxia-api -- curl -s https://keycloak.ubiquitycluster.uk/auth/realms/ubiquity/.well-known/openid_configuration

# Test catalog connectivity
kubectl exec -n onyxia deployment/onyxia-api -- curl -s https://inseefrlab.github.io/helm-charts-datascience/index.yaml
```

## Security Considerations

### Access Control

- **Authentication**: Keycloak-based OAuth2/OIDC
- **Authorization**: Kubernetes RBAC integration
- **Namespace Isolation**: Strict user/group separation
- **Network Policies**: Controlled inter-service communication

### Container Security

- **Image Scanning**: Vulnerability scanning for deployed images
- **Security Contexts**: Non-root containers and security policies
- **Resource Limits**: Prevent resource exhaustion attacks
- **Secrets Management**: Vault integration for sensitive data

### Data Protection

- **Encryption at Rest**: Longhorn storage encryption
- **Encryption in Transit**: TLS for all communications
- **Data Isolation**: User-specific persistent volumes
- **Backup Protection**: Encrypted backup storage

## Best Practices

### Service Management

1. **Resource Planning**: Set appropriate CPU/memory requests
2. **Storage Management**: Clean up unused persistent volumes
3. **Service Cleanup**: Regularly remove stopped services
4. **Configuration Backup**: Export service configurations

### Performance Optimization

1. **Resource Requests**: Set realistic resource requirements
2. **Node Affinity**: Use node selectors for workload placement
3. **Persistent Storage**: Use appropriate storage classes
4. **Network Optimization**: Minimize cross-node communication

### User Training

1. **Platform Orientation**: Provide user onboarding documentation
2. **Service Catalog**: Maintain updated service descriptions
3. **Best Practices**: Share resource management guidelines
4. **Support Channels**: Establish clear support procedures

## Advanced Configuration

### Custom Service Catalogs

**Adding Custom Catalogs:**
1. Create Helm repository
2. Update Onyxia configuration
3. Configure catalog metadata
4. Test service deployments

**Catalog Structure:**
```yaml
catalogs:
  - id: custom-catalog
    name: Custom Catalog
    description: Organization-specific tools
    maintainer: admin@organization.com
    location: https://charts.organization.com
    status: PROD
    type: helm
```

### Integration Extensions

**Custom Authentication:**
- LDAP/Active Directory integration
- SAML federation
- Multi-realm support

**Storage Backends:**
- S3-compatible object storage
- External NFS mounts
- Distributed filesystems

**Compute Integration:**
- GPU resource scheduling
- External compute clusters
- Hybrid cloud resources

## Migration and Backup

### User Data Migration

**Export User Services:**
```bash
# Export user configurations
kubectl get all -n user-{username} -o yaml > user-backup.yaml

# Backup persistent data
kubectl exec -n user-{username} {pod-name} -- tar czf - /home/user | gzip > user-data.tar.gz
```

**Import User Services:**
```bash
# Restore user configurations
kubectl apply -f user-backup.yaml

# Restore persistent data
kubectl exec -n user-{username} {pod-name} -- tar xzf - -C /home/user < user-data.tar.gz
```

### Platform Migration

**Configuration Backup:**
- Helm values and configurations
- Keycloak realm export
- Custom catalog definitions
- User quota configurations

**Data Backup:**
- Persistent volume snapshots
- User workspace backups
- Service configuration exports
- Access control policies

## Integration Examples

### Jupyter Notebook with SLURM

**Service Configuration:**
```yaml
jupyter:
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
  persistence:
    enabled: true
    size: "10Gi"
  environment:
    SLURM_ENDPOINT: "slurmctld.hpc-ubiq.svc.cluster.local"
```

### RStudio with Shared Storage

**Service Configuration:**
```yaml
rstudio:
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
  persistence:
    enabled: true
    size: "5Gi"
  sharedStorage:
    nfs:
      server: "nfs.ubiquitycluster.uk"
      path: "/shared/projects"
```

### Spark Cluster

**Service Configuration:**
```yaml
spark:
  master:
    resources:
      requests:
        cpu: "1"
        memory: "2Gi"
  worker:
    replicas: 3
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"
```

This comprehensive documentation provides administrators and users with everything needed to deploy, manage, and troubleshoot Onyxia within the Ubiquity platform, enabling self-service data science and HPC capabilities.