# Configure

The configure script is located in `scripts/configure` and can be run directly.

The script makes changes not only to the `.env` file, but also to multiple files within the git repository.

If you are not happy with the changes, you can revert them with `git checkout .`.

At any point you can type ? and be presented with help for any question. It also tells you what the default (or previously set) value is.

The questions asked by the script are:
- `Select text editor` - This is the text editor that will be used to inspect the configuration at the end of the configure process. It is recommended to use `nano` if you are not familiar with `vim`. This is used at the end of the configure process to inspect the configuration.

- `What is the domain name of your Ubiquity deployment?` - This is the domain name that will be used to access Ubiquity. It is recommended to use a domain name that resolves to the machine where Ubiquity is deployed. If you are deploying Ubiquity on your laptop or on a cluster where you don't have DNS upstream access configured, you can use `nip.io`. If you are deploying Ubiquity on a server, you can use a domain name that resolves to the cluster. Wildcard DNS records are supported, just make sure that the * is removed from the domain name. For example, if you want to use `*.example.com`, you should enter `example.com` here. This domain name will be used to generate TLS certificates for Ubiquity. Note that if you are using `nip.io`, you will get a warning about the certificate being invalid. This is expected and you can safely ignore it. If you do use nip.io, you will need to use `https://<your service>.<your-ip>.nip.io` to access Ubiquity. For example, if you are using nip.io and your IP is `10.212.84.200`, you will need to use `https://<your service>.10-212-84-200.nip.io` to access Ubiquity. If you are using a domain name that resolves to the machine where Ubiquity is deployed, you can use `https://<your service>.<your domain name>` to access Ubiquity. For example, if you are using `example.com`, you will need to use `https://<your service>.example.com` to access Ubiquity.

Note: If you use `nip.io` you are expected to be running non-production workloads. If that is the case, then your cert_provider should be set to `pebble-issuer` and your cert_provider_pebble_issuer_url should be set to `https://pebble-issuer:8080/cluster-ca`. If you are running production workloads, then you should use a real domain name and a real certificate provider which in our case should be `letsencrypt-prod`. You will need to set your cert_provider to `letsencrypt-prod` and your cert_provider_letsencrypt_prod_email to your email address.

- `Enter seed repo` - This is the original git repository that the cluster is defined at. This is used to pull the cluster definition from. It is also used to push the cluster definition to and so must have a service account with read access. This forms the basis of the cluster definition and is the source of truth for the cluster. It is also used as a core of the deployment, leveraging GitOps to deploy and maintain cluster services. Please make sure that if you choose to store sensitive information that this repository is private!

- `Enter seed repo username` - This is the username that has permission to access the original git repository that the cluster is defined at in the question previously. This is the service account that has read access to the repository.

- `Enter seed repo password` - This is the password that has permission to access the original git repository that the cluster is defined at in the question previously. This is the service account that has read access to the repository.

- `Enter DNS server` - This is a DNS server for DNS records to be resolved. This is used to resolve upstream DNS for the cluster. This can be self-contained however it is recommended to reduce "blast radius" by using an upstream DNS server that you can cache records from locally on-cluster. Please make sure that this server is reachable!

- `Enter NTP server` - This is a NTP server for network-time. This is used to ensure that all nodes are in sync with each other. This can be self-contained however it is recommended to reduce "blast radius" by using an upstream NTP server that you can serve the time locally on-cluster from. Please make sure that this server is reachable!

- `Please define OS (Rocky or Ubuntu)` - This is an OS flavour. Currently Rocky or Ubuntu. Define version next!

- `Please define OS version for Rocky or Ubuntu:` - This is an OS version. Supported versions are Rocky (8.7/8.8/8.9/9.1) or Ubuntu (22.04/23.04). Define flavour first!

- `Enter time zone` - This is the timezone for your cluster specified in TZ format. For example `Europe/London` or `America/New_York`.

- `Enter cluster name` - This is a name for your cluster. This is used to identify your cluster in Ubiquity. This is also used to identify your cluster in the cluster definition repository.

- `Enter cluster domain` - this is a domain for your cluster. This domain is used to identify hostnames for your cluster. This is also used to identify your cluster in the cluster definition repository.

- `Enter cluster network CIDR` - this is a network CIDR for your Kubernetes cluster. This is the address range that all of your pods will be provisioned on. This is normally a private address range.

- `Enter cluster service CIDR` - this is a network CIDR for the services you want to run on your cluster. This is the address range that services provision themselves on. Services aren't often presented directly to the user and normally go through an Ingress Controller.

- `Enable OFED` - this is if you want to enable OFED on your cluster or not. OFED is a high-performance network stack that is used for RDMA and other high-performance networking. This is used for high-performance workloads such as HPC and AI/ML. This is not required for all workloads.

- `Enter OFED version` - this is the OFED version you want to install on your cluster. This is the version of the OFED stack that will be installed on your cluster. This is used for high-performance workloads such as HPC and AI/ML. This is not required for all workloads.

- `Enter external interface` - this is the externally-facing interface for your HA cluster control plane. This can be an interface that is exposed to the internet or a private interface that is exposed to a load balancer. Used if you want to separate internal and external traffic.

- `Enter internal interface` - this is the internally-facing interface for your HA cluster control plane. This can be an interface that is exposed to the internet or a private interface that is exposed to a load balancer. Used if you want to separate internal and external traffic.

- `Enter keepalived interface` - this is the keepalived interface for your HA cluster control plane. This is the physical interface that will be used to provision a highly-available keepalived virtual IP on that floats between all 3 control plane masters. 

- `Enter keepalived CIDR` - this is the keepalived CIDR notation the keepalived address exists on. This is to define an address space and mask that the keepalived virtual IP will float on. This is normally in the format of `X.X.X.X/XX`.

- `Enter keepalived VIP` - this is the keepalived virtual IP address for your HA cluster control plane. This address will be the address you will talk to your Kubernetes API on and will float between all 3 control plane masters. This is normally in the format of `X.X.X.X`.

- `Enter MetalLB external IP range` - this is the MetalLB external IP range for your cluster that will be used for services. Future services will be able to assign from a pool of external addresses only.

- `Enter MetalLB internal IP range` - this is the MetalLB internal IP range for your cluster. Future services will be able to assign from a pool of internal addresses only. Currently this feature is disabled.

- `Enter k3s version` - this is the k3s version for your cluster. 

- `Enter k3s encryption secret` - this is the k3s encryption secret for your cluster. This is auto-generated but you can get it now here. This is used to encrypt all traffic between nodes in the cluster.

- `Enter cert provider` - this is the certificate provider for the cluster. This can be pebble-issuer which is an internal-only provider, or letsencrypt-prod which is a production provider. This is used to generate TLS certificates for Ubiquity.

- `Enter internal ipv4 address of bootstrapper` - this is the internal address of the bootstrap node. This is the address that sits on a bootstrap node and provisions on-premise environments. This is the address that is defined as the default gateway for the cluster during provisioning.

- `Enter internal ipv4 network address space` - this is the internal address space used for the cluster. This is the address space that all of your pods will be provisioned on. This is normally a private address space.

- `Enter internal ipv4 network gateway` - this is the internal network gateway address. This is the address that is defined as the default gateway for the cluster.

- `Enter external ipv4 address of bootstrapper` - this is the external address of the bootstrap node. This is the address that sits on a bootstrap node and in some cases delivers NAT to the remaining nodes.

- `Enter external ipv4 network address space` - this is the external address space used for the cluster.

- `Enter external ipv4 network gateway` - this is the external network gateway address. This is the address that sits on a bootstrap node and provisions on-premise environments.

- `Enter your username for the container registry:` - this is your username for accessing container images - Please refer to the documentation or GitHub repository for access credentials!

- `Enter your registration key:` - this is the special token to allow you to log onto the container registry. Without this you will not be able to pull the images required to run Ubiquity.

- `Enter your dockerhub username:` - this is your dockerhub username. This is used to prevent rate-limiting on dockerhub.

- `Enter your password:` - this is your dockerhub password. This is used to prevent rate-limiting on dockerhub.


The most typical aspects for configuration are:

- Configuring [identity providers](../identities/summary.md). Ubiquity supports a range of OIDC and SAML based IdPs.
- Configuring [storage providers](../storage/summary.md). Ubiquity supports a range of storage providers.
- Configuring [platform providers](../providers/summary.md). Ubiquity supports a range of platform providers.
- [Adding catalogue items](../providers/adding-a-catalogue-app.md) for sharing amongst Ubiquity users as a self-serve function.