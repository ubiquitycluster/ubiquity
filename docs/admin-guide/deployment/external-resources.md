# External resources

!!! info

    These resources are optional for a local dry-run, but production clusters need
    explicit ownership for DNS, load balancer ingress, state storage, and backup
    destinations. Dry-run/local proof validates manifests and workflows only; it
    does not prove public DNS, certificate issuance, or offsite recovery.

Ubiquity keeps external dependencies small. The required production-facing
resources are listed below.

| Provider | Resource | Purpose | Owner |
| -------- | -------- | ------- | ----- |
| Terraform Cloud or another Terraform backend | Workspace/state backend | Stores infrastructure state for external DNS/tunnel resources | Platform operations |
| DNS provider such as Cloudflare | external DNS zone | DNS records and DNS-01 challenges for trusted certificates | Platform operations |
| Cloudflare Tunnel, hardware load balancer, or equivalent ingress path | load balancer / tunnel endpoint | Provides controlled external access to selected services | Network/platform operations |
| S3-compatible object storage or enterprise backup platform | backup target | Stores encrypted backup and restore artifacts outside the cluster | Backup owner |

## Create credentials

Create credentials before the first live production run. Store them in the
approved password manager, Vault, or sealed secret workflow. Do not commit raw
credential values to Git.

### Terraform workspace

Terraform is stateful and needs a backend. Terraform Cloud is one option, but any
backend that meets your retention and access-control requirements is acceptable.

1. Create a workspace named `ubiquity-external` or your environment-specific
   equivalent.
2. Change the execution mode to `Local` when Terraform must reach private cluster
   endpoints from the operator workstation.
3. Record the workspace owner, recovery contact, and state-retention policy.
4. If you use another backend, update `external/versions.tf` and record the
   backend migration procedure.

## DNS and certificate provider

For Cloudflare, prefer a scoped API token over a global API key. The token should
be limited to the production account/zone and only the permissions required by
Terraform and cert-manager.

Recommended Cloudflare API token permissions:

- Zone:Read for the managed zone
- DNS:Edit for the managed zone
- Account Settings:Read only if a tunnel workflow requires it
- Tunnel edit permissions only for the account that owns the tunnel

Record credential ownership and rotation expectations:

- Owner: platform operations
- Storage: approved password manager, Vault, or sealed secret workflow
- Rotation: at least annually and immediately after operator departure or
  suspected exposure
- Evidence: record token ID/name and rotation date, never the token value

## Load balancer and ingress path

Choose one ingress pattern and document it in the environment change record:

- Cloudflare Tunnel for environments that cannot expose direct inbound ports
- hardware or virtual load balancer for datacenter ingress
- direct port-forwarding only for non-production or lab environments

Production validation must include live DNS resolution, certificate issuance,
and an application reachability test through the selected load balancer. Helm
rendering or local dry-run output is not sufficient evidence of external access.

## Backup target

Off-cluster backup is required for production. Use S3-compatible object storage,
enterprise backup storage, or another approved encrypted destination. Define:

- bucket/container/project name
- retention period
- encryption owner/key policy
- restore-test cadence
- operator responsible for access review

## Alternatives

- Terraform Cloud: any supported Terraform backend
- Cloudflare DNS: any DNS provider supported by your ExternalDNS/cert-manager
  workflow, or manual DNS setup for small lab clusters
- Cloudflare Tunnel: hardware load balancer, HAProxy/WireGuard jump host, or
  other approved ingress architecture
- S3-compatible backup: Backblaze B2, MinIO, AWS S3/Glacier, enterprise object
  storage, or a backup appliance that supports restore testing
