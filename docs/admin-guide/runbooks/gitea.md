# Gitea

Gitea is a self-hosted Git service that provides a lightweight, fast, and secure Git hosting solution for the Ubiquity platform. It serves as the central repository for all code, configurations, and GitOps workflows within the cluster.

## Overview

Gitea in Ubiquity serves multiple critical functions:

- **Central Git Repository**: Hosts all cluster configuration and application code
- **GitOps Source**: Primary source for ArgoCD deployments and configuration management
- **CI/CD Integration**: Integrates with Argo Workflows for automated builds and deployments
- **Repository Migration**: Automatically migrates repositories from GitHub during cluster initialization
- **OAuth Provider**: Provides authentication for other services via OAuth2

## Architecture

### Core Components

- **Gitea Application**: Main Git service with web interface
- **PostgreSQL Database**: Persistent storage for repositories and metadata
- **Memcached**: Caching layer for improved performance
- **Persistent Storage**: Longhorn-backed storage for repository data

### High Availability Configuration

Gitea is configured for high availability with:
- **3 Replicas**: Distributed across master nodes
- **Session Affinity**: Cookie-based routing for consistent user experience
- **Node Selector**: Runs exclusively on control plane nodes for stability

## Configuration

### Default Settings

```yaml
gitea:
  replicacount: 3
  nodeSelector:
    node-role.kubernetes.io/master: "true"
  gitea:
    config:
      server:
        LANDING_PAGE: explore
        ROOT_URL: https://git.ubiquitycluster.uk
```

### Storage Configuration

- **Persistence**: 10Gi Longhorn storage for repository data
- **Database**: PostgreSQL with persistent storage
- **Caching**: Memcached for session and object caching

### Ingress Configuration

Gitea is accessible via HTTPS with:
- **Domain**: `git.ubiquitycluster.uk`
- **TLS**: Automatic certificate management via cert-manager
- **Session Affinity**: Cookie-based load balancing for consistent sessions

## Accessing Gitea

### Web Interface

Access the Gitea web interface at:
```
https://git.ubiquitycluster.uk
```

### Admin Credentials

- **Username**: `gitea_admin`
- **Password**: Retrieved from Vault using External Secrets

To get the admin password:
```bash
kubectl get secret gitea-admin-secret -n gitea -o jsonpath="{.data.password}" | base64 -d
```

### Git Access

Clone repositories using HTTPS:
```bash
git clone https://git.ubiquitycluster.uk/ops/ubiquity.git
```

Or using SSH (requires SSH key setup):
```bash
git clone git@git.ubiquitycluster.uk:ops/ubiquity.git
```

## Automated Configuration

### Repository Migration

Gitea automatically migrates repositories during cluster initialization:

**From GitHub to Gitea:**
- `logicalisuki/ubiquity-open` → `ops/ubiquity`
- `logicalisuki/blog` → `ubiquity/blog` (mirror)
- `logicalisuki/backstage` → `ubiquity/backstage` (mirror)

### Organization Setup

Automatic organization creation:
- **ops**: Operations team organization
- **ubiquity**: General project organization

### OAuth Applications

Gitea automatically creates OAuth applications for:
- **Dex**: SSO integration
- **Keycloak**: Identity provider integration
- **ArgoCD**: GitOps authentication (future)

### Access Tokens

Automated access token generation for:
- **Renovate**: Dependency update automation
- **CI/CD Systems**: Automated builds and deployments

## Integration with Ubiquity Components

### ArgoCD GitOps

Gitea serves as the primary source for ArgoCD applications:

1. **Initial Bootstrap**: ArgoCD initially sources from GitHub
2. **Migration**: Repository is migrated to Gitea during platform deployment
3. **Switch Over**: ArgoCD switches to use Gitea as the source
4. **Self-Hosting**: Complete GitOps workflow becomes self-contained

### CI/CD Integration

Integration with Argo Workflows:
- **Webhook Triggers**: Git push events trigger workflow execution
- **Build Pipelines**: Automated container builds and deployments
- **Security Scanning**: Automated vulnerability scanning of code

### Secret Management

Integration with Vault for:
- **Admin Credentials**: Secure storage of admin passwords
- **OAuth Secrets**: Client IDs and secrets for SSO
- **Access Tokens**: Service account tokens for automation

## Repository Management

### Creating Repositories

**Via Web Interface:**
1. Navigate to `https://git.ubiquitycluster.uk`
2. Click "+" → "New Repository"
3. Configure repository settings
4. Initialize with README if needed

**Via Configuration:**
Add to `platform/gitea/files/config/config.yaml`:
```yaml
repositories:
  - name: my-new-repo
    owner: ops
    private: false
```

### Repository Permissions

**Organization-based access:**
- **ops**: Administrative repositories
- **ubiquity**: General project repositories
- **Team membership**: Controls access levels

### Backup and Recovery

**Automatic Backups:**
- Repository data stored on Longhorn (replicated storage)
- Database backups via PostgreSQL backup procedures
- Configuration stored in Git (self-healing)

## Monitoring and Maintenance

### Health Monitoring

Gitea health is monitored through:
- **Kubernetes Probes**: Liveness and readiness checks
- **Prometheus Metrics**: Application performance metrics
- **Grafana Dashboards**: Visual monitoring of service health

### Log Analysis

Access Gitea logs:
```bash
# View application logs
kubectl logs -n gitea deployment/gitea -f

# View all pod logs
kubectl logs -n gitea -l app.kubernetes.io/name=gitea
```

### Database Maintenance

**PostgreSQL maintenance:**
```bash
# Access database
kubectl exec -it -n gitea statefulset/gitea-postgresql -- psql -U gitea

# Check database size
kubectl exec -it -n gitea statefulset/gitea-postgresql -- psql -U gitea -c "\l+"
```

## Troubleshooting

### Common Issues

#### Repository Clone Failures
**Symptoms**: Unable to clone repositories via HTTPS or SSH

**Solutions:**
1. Check ingress controller status: `kubectl get ingress -n gitea`
2. Verify certificate validity: `kubectl describe certificate gitea-tls-certificate -n gitea`
3. Test connectivity: `curl -I https://git.ubiquitycluster.uk`

#### Authentication Problems
**Symptoms**: Unable to login or access repositories

**Solutions:**
1. Verify admin secret: `kubectl get secret gitea-admin-secret -n gitea`
2. Check External Secrets: `kubectl get externalsecret -n gitea`
3. Validate Vault connectivity: `kubectl logs -n gitea deployment/gitea | grep vault`

#### Performance Issues
**Symptoms**: Slow Git operations or web interface

**Solutions:**
1. Check Memcached status: `kubectl get pods -n gitea -l app.kubernetes.io/name=memcached`
2. Monitor resource usage: `kubectl top pods -n gitea`
3. Review storage performance: `kubectl get pvc -n gitea`

### Diagnostic Commands

```bash
# Check Gitea service status
kubectl get all -n gitea

# View configuration
kubectl get configmap gitea-config -n gitea -o yaml

# Check storage usage
kubectl exec -n gitea deployment/gitea -- df -h /data

# Test database connectivity
kubectl exec -n gitea deployment/gitea -- gitea doctor check

# View recent events
kubectl get events -n gitea --sort-by='.lastTimestamp'
```

### Recovery Procedures

#### Pod Recovery
```bash
# Restart Gitea pods
kubectl rollout restart deployment/gitea -n gitea

# Force pod recreation
kubectl delete pods -n gitea -l app.kubernetes.io/name=gitea
```

#### Database Recovery
```bash
# Access PostgreSQL
kubectl exec -it -n gitea statefulset/gitea-postgresql -- bash

# Check database integrity
gitea doctor check --all
```

## Security Considerations

### Access Control

- **Admin Access**: Limited to necessary personnel only
- **Organization Permissions**: Role-based access control
- **SSH Keys**: Regular rotation and monitoring
- **OAuth Scopes**: Minimal required permissions

### Network Security

- **TLS Encryption**: All traffic encrypted in transit
- **Internal Communication**: Pod-to-pod encryption
- **Network Policies**: Restricted network access
- **Ingress Protection**: WAF and rate limiting

### Data Protection

- **Repository Encryption**: Git objects encrypted at rest
- **Database Encryption**: PostgreSQL data encryption
- **Backup Encryption**: Encrypted backup storage
- **Secret Management**: Vault-based secret storage

## Best Practices

### Repository Organization

1. **Use Organizations**: Group related repositories
2. **Naming Conventions**: Consistent repository naming
3. **Branch Protection**: Protect main/master branches
4. **Access Reviews**: Regular permission audits

### Performance Optimization

1. **Large Files**: Use Git LFS for large assets
2. **Repository Size**: Monitor and manage repository growth
3. **Caching**: Leverage Memcached for performance
4. **Resource Limits**: Appropriate CPU/memory allocation

### Backup Strategy

1. **Multiple Replicas**: Longhorn 3-way replication
2. **Database Backups**: Regular PostgreSQL dumps
3. **Configuration Backup**: Git-based configuration management
4. **Disaster Recovery**: Documented recovery procedures

## Advanced Configuration

### Custom Themes

Customize Gitea appearance:
1. Create custom CSS/templates
2. Mount as ConfigMap or Secret
3. Update deployment to include custom assets

### Hook Scripts

Implement custom Git hooks:
1. Create hook scripts
2. Mount via ConfigMap
3. Configure repository-specific hooks

### Federation

For multi-cluster scenarios:
1. Configure repository mirroring
2. Set up cross-cluster authentication
3. Implement distributed backup strategies

## Migration Procedures

### Importing Existing Repositories

**From GitHub:**
```yaml
# Add to config.yaml
repositories:
  - name: existing-repo
    owner: ops
    migrate:
      source: https://github.com/org/existing-repo
      mirror: false
```

**From GitLab:**
1. Export repository data
2. Create new repository in Gitea
3. Push existing data to new repository

### Exporting Data

**Repository Export:**
```bash
# Clone with full history
git clone --mirror https://git.ubiquitycluster.uk/ops/repo.git

# Export to external location
git push --mirror https://external-git.example.com/org/repo.git
```

## Integration Examples

### ArgoCD Integration

```yaml
# Application source configuration
source:
  repoURL: https://git.ubiquitycluster.uk/ops/ubiquity.git
  targetRevision: main
  path: system/monitoring-system
```

### Webhook Configuration

```yaml
# Argo Events webhook
webhook:
  endpoint: /gitea
  method: POST
  url: https://git.ubiquitycluster.uk/ops/ubiquity/settings/hooks
```

This comprehensive documentation provides administrators with everything needed to manage, troubleshoot, and optimize Gitea within the Ubiquity platform.
