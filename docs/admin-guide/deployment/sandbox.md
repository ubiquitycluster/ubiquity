# Ubiquity Sandbox Deployment

This guide describes how to deploy Ubiquity in a sandbox mode on a single machine.
This is the fastest way to get Ubiquity up and running, but it is not suitable for production.
It is recommended to use [cloud](cloud/index.md) or [on-premises](on-prem.md) deployment for production.

## Prerequisites

- at least 16GB RAM to run all containers in k3d
- Docker v1.13+

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
make configure-sandbox
```

`make configure-sandbox` will ask you a few questions using a configure script with specialised arguments and generate `.env` file with your answers.

This in sandbox mode, you can leave all the default values as it configures pretty much everything for you so please press enter to all questions.

For more information, please see [configuration guide](configure.md).

## Deploying Ubiquity

Prior to deploying Ubiquity, you should now push all of your changes to your chosen git repository that you configured during `make configure-sandbox`.

```bash
# deploy
git push origin
```

Then you can trigger a deployment of Ubiquity by running:

```bash
# start Ubiquity environment
make sandbox

# verify
kubectl get pods -A
k9s
```

Ubiquity will start a K3d cluster with all the required containers and services, and get ArgoCD to deploy the Ubiquity environment - reading your configuration from the git repository you configured during `make configure-sandbox` and pushed.
## Administrating Ubiquity

To administrate Ubiquity, go look at the [admin-guide](../index.md).

To add users, go look at the [user accounts](../administration/user-accounts.md) section.

## Accessing Ubiquity

To access Ubiquity, go look at the [user-guide](../../user-guide/index.md).

## Logs/Monitoring

Ubiquity is monitored by using Prometheus and Grafana. You can access Grafana at `https://grafana.127-0.0.1.nip.io` (default credentials you can get by running `scripts/grafana-admin-password`).
Logs emitted by the containers are collected and saved inside Loki. You can access them via Grafana located at `https://grafana.127-0-0-1.nip.io` (default credentials you can get by running `scripts/grafana-admin-password`).

## Maintenance

Please see the [admin-guide](../admin-guide/index.md) for more information.

## Known issues

When ubiquity is launched for the first time, it takes a little while to apply all configs.
It means that you may need to wait few minutes until these applications are setup. This is normally a 10-15 minute process for sandbox mode.

## Keycloak

Keycloak is an Identity and Access Management software bundled with Ubiquity. it is used to authenticate users and manage their permissions.

To find the keycloak admin account run:

```bash
./scripts/keycloak-admin-password
```

Login to the admin interface at [https://keycloak.127-0-0-1.nip.io/auth/admin](https://keycloak.127-0-0-1.nip.io/auth/admin) and create ubiquity users

## Integration with SLURM

SLURM integration already exists - In sandbox mode, the concept of node as pod exists. In the `hpc-ubiq` space a slurm cluster should already exist with 1x compute replica. You can check this by using K9s and attaching a shell to the `hpc-ubiq/slurmctld` instance. You can then run `sinfo` to see the cluster status.

You can get this instance to be accessible by setting up a port-forward rule accordingly. For example:

```bash
kubectl port-forward -n hpc-ubiq svc/slurmctld 2222:22
```

### Updating Ubiquity

Ubiquity can be updated by simply backing up your .env file, git pulling the latest changes, and then running `make sandbox` again.