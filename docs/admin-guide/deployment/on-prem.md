# On-Premises Production Deployment
This guide describes how to deploy Ubiquity on-premises in a production environment.

## Prerequisites

- at least 4 nodes:
    - 3x Control-Plane nodes
    - 1x Compute Node
- An appropriate network switch with VLAN support
    - 1x VLAN for management
    - 1x VLAN for storage
    - 1x VLAN for HPC
    - 1x VLAN for Kubernetes
    - 1x VLAN for OOB
- A DNS server (your laptop can be used for this and then pivoted to a dedicated DNS server provisioned by Ubiquity)
- A NTP server (your laptop can be used for this and then pivoted to a dedicated NTP server provisioned by Ubiquity)
- A bootstrap node (your laptop can be used for this)

## Prepare environment

```bash
# clone repo
git clone https://github.com/logicalisuki/ubiquity.git
cd ubiquity
git submodule update --init --recursive
```
## Configuring Ubiquity

```bash
# jump into opus environment
sudo make tools
# configure
make configure
```

`make configure` will ask you a few questions using a configure script and generate `.env` file with your answers.

For more information, please see [configuration guide](configure.md).

## Deploying Ubiquity

Prior to deploying Ubiquity, you should now push all of your changes to your chosen git repository that you configured during `make configure`.

```bash
# deploy
git push origin
```

Then you can trigger a deployment of Ubiquity by running:

```bash
# start Ubiquity environment
make

# verify
kubectl get pods -A
k9s
```

Ubiquity will: 
- Bootstrap a PXE environment
- IPMI network boot the 3x control-plane nodes
- Install Ubuntu 20.04 on the 3x control-plane nodes
- Install k3s on the 3x control-plane nodes
- Install MetalLB on the 3x control-plane nodes
- Install Keepalived on the 3x control-plane nodes
- Install Longhorn on the 3x control-plane nodes
- Install ArgoCD on the 3x control-plane nodes
- Get ArgoCD to provision the remaining environment components including DNS and NTP
- Install the BareMetal Operator on the 3x control-plane nodes

ArgoCD reads your configuration from the git repository you configured during `make configure-sandbox` and pushed.

Once deployed, you can pivot to the provisioned DNS and NTP servers on the 3x control plane nodes by running:
    
```bash
make pivot
```

Which will take your NTP and DNS server settings from the .env file and configure your control plane nodes to use them.

## Administrating Ubiquity

To administrate Ubiquity, go look at the [admin-guide](../index.md).

To add users, go look at the [user accounts](../administration/user-accounts.md) section.

## Accessing Ubiquity

To access Ubiquity, go look at the [user-guide](../../user-guide/index.md).

## Logs/Monitoring

Ubiquity is monitored by using Prometheus and Grafana. You can access Grafana at `https://grafana.<your domain>.nip.io` (default credentials you can get by running `scripts/grafana-admin-password`).
Logs emitted by the containers are collected and saved inside Loki. You can access them via Grafana located at `https://grafana.<your domain>.io` (default credentials you can get by running `scripts/grafana-admin-password`).

## Maintenance

Please see the [admin-guide](../index.md) for more information.

## Known issues

When ubiquity is launched for the first time, it takes a little while to apply all configs.
It means that you may need to wait few minutes until these applications are setup. This is normally a 10-15 minute process for sandbox mode.

## Keycloak

Keycloak is an Identity and Access Management software bundled with Ubiquity. it is used to authenticate users and manage their permissions.

To find the keycloak admin account run:

```bash
./scripts/keycloak-admin-password
```

Login to the admin interface at `keycloak.<your domain>/auth/admin` and create ubiquity users. See the [user accounts](../administration/user-accounts.md) section.

## Integration with SLURM

SLURM integration already exists - In production mode, the concept of node as pod exists but can be pivoted to bare-metal if required. In the `hpc-ubiq` space a slurm cluster should already exist with 1x compute replica. You can check this by using K9s and attaching a shell to the `hpc-ubiq/slurmctld` instance. You can then run `sinfo` to see the cluster status.

You can get this instance to be accessible by setting up a port-forward rule accordingly. For example:

```bash
kubectl port-forward -n hpc-ubiq svc/slurmctld 2222:22
```

If you wish to pivot to bare-metal, see the [bare-metal](../providers/bare-metal.md) provider documentation.

### Updating Ubiquity

Ubiquity can be updated by simply backing up your .env file, git pulling the latest changes, and then committing this back to your upstream git repository. ArgoCD will then automatically update your environment.