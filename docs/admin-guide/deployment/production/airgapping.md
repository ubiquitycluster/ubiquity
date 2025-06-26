# Air gapped Network Deployment

Ubiquity can be deployed in an "air gapped" network environment, provided the follow requirements are met:
1. A local container registry is deployed, or a registry service is available from the infrastructure provider with access from the secure network.
2. At least one "bastion" node (or instance) is available to transfer Ubiquity software from the Internet over to the secure network; the best practice is for this system to be able to access the same container registry that is available to the secure network, and ideally, also the Kubernetes API itself.

## Supported Air Gapped Configurations

There is 1 supported configuration, although Ubiquity can be made to work with many more restrictions beyond this.  In both of the supported cases, the following is true:
1. "Node" can be either a physical compute node, a virtual machine, or a cloud provider instance
2. "Bastion Node" refers to a Linux node that has Docker installed as well as a Git client
3. "Install Node" refers to a Linux node that has a Kubernetes and Helm client installed, with admin-level client certificates for the Kubernetes API; note that for the preferred configuration, this is combined with the "Bastion Node"
4. "Registry" is any OCI-compliant registry - note that if using a cloud infrastructure provider, it is best to use that platform's secure registry service, if any, in order to avoid having to configure "insecure registries" in each node on the secure network side.
5. NGINX Ingress (or LoadBalancer) into Ubiquity and Ubiquity jobs must still be configured and accessible by whatever clients expect to access it, whether in the secure network or not.  This configuration is not covered in this document and is implementation specific.

### Default Configuration (Preferred)

This topology is most resource efficient and simplifies the deployment as much as possible.

![Preferred Secure Network Configuration](secure_env.svg)

Note that in this topology only the "Bastion Node" + "Install Node" environment must be able to access the container registry as well as the Kubernetes API; it does not need to be able to access the Kubernetes workers directly.  This node will be used to pull, tag, and push containers from upstream (Internet) to the secure container registry, and deploy+update Ubiquity using the Kubernetes and Helm clients.

In addition to mirroring containers to the secure registry, the Bastion node in this configuration must be able to initially mirror the [Ubiquity repository in GitHub](https://github.com/logicalisuki/ubiquity) over to itself and then commit back to the local repository. This local repository could clearly be the bastion node in this diagram.

Note the bastion node could quite easily separate its functions into two separate nodes, one for mirroring containers and one for deploying/updating Ubiquity, but this is not recommended as it will require additional configuration to transfer the container images from the mirroring node to the deployment node.

## Mirroring containers

### System Containers

The script [scripts/Ubiquity-pull-system-images](scripts/Ubiquity-pull-system-images) can be used to pull, tag, and optionally push Ubiquity system containers from the upstream registry to the secure registry, and should be run on the Bastion Node for both initial installs and updates.  Run without arguments for usage.  If specifying a push registry and push repository, make sure to populate these values into `Ubiquity.Ubiquity_SYSTEM_REGISTRY` and `Ubiquity.Ubiquity_SYSTEM_REPO_BASE` respectively as helm chart or overrides options for deployment.  The value of `--Ubiquity-version`, to the left of the branch (e.g. starting with the first numeric digit), should be specified as `Ubiquity.IMAGES_VERSION`; the value to the left (e.g. `Ubiquity-master`) should be specified as `Ubiquity.Ubiquity_IMAGES_TAG`.

If not using this script to push (e.g. if using an alternate method to transfer container images from one registry to another), be sure to match the repository base and the version tag with what is configured in the Helm chart or overrides options, as explained above.

### Additional Containers

To determine additional containers to mirror, run the following command in the `Ubiquity-helm` repository:

```grep "image:" values.yaml|awk {'print $2'} && grep  '_IMAGE:' values.yaml|awk '{print $3}'```

The example output will show what containers need to be mirrored, tagged, and pushed to the secure registry, one on each line:

```
$ grep "image:" values.yaml|awk {'print $2'} && grep  '_IMAGE:' values.yaml|awk '{print $3}'
Logicalis/Ubiquity-cache-pull:latest
Logicalis/lxcfs:3.0.3-3
nvidia/k8s-device-plugin:1.11
xilinxatg/xilinx_k8s_fpga_plugin:latest
Ubiquity/k8s-rdma-device:1.0.1
mysql:5.6.41
Logicalis/postfix:3.11_3.4.9-r0
Logicalis/idmapper
memcached:1.5
registry:2
Logicalis/unfs3
alpine:3.8
```

Once mirrored, edit the entries in the `override.yaml` file to the new tags, by searching for the matching tags from the list above.

### Application Containers

Appsync cannot be used in an air gapped configuration, so application containers must be mirrored individually.  A matching application target must be created manually in the Ubiquity system (can be performed as the `root` user), and then those applications can be marked public in the *Administration->Apps* view of the portal.  Note that the desired list of application containers must be obtained from Logicalis.  Each application target should also be explicitly pulled after being created to download the in-container metadata (e.g. AppDef).

## Additional Configuration

1. The Bastion node user must log in with the upstream service account for the Ubiquity system and application containers.
2. The `Ubiquity.imagePullSecret` value must be set to the secure registry's service account (or username/password) since the containers will be pulled from there to be executed.  Note that this is different than the upstream service account, which must be used on the Bastion node.

## Updates

Container mirroring should be repeated when the system is updated, including the Helm chart or overrides configuration.  The updated version of Ubiquity will be deployed automatically by running a Helm upgrade after updating the configuration parameters.  Note that as a best practice, all containers should be mirrored, or at least checked, in case they changed between versions following the Git update of the `Ubiquity` repository.

## Limitations

The following limitations in air gapped environments currently exist:
1. Ubiquity is a multi-architecture platform but the default mirroring scheme will only pull containers for the current architecture.  In most cases, the infrastructure architecture (e.g. `amd64`) will match between the Bastion node and the nodes on the secure network side; if they don't, additional work will be required to create manifests so that multiple architecture containers are available (and pulled correctly) - this is currently not explicitly supported, but can be on request.
2. Appsync cannot be used and should be disabled in the `override.yaml` file before deploying; see above for information on mirroring application containers and creating the appropriate application targets in the system.
3. If using a self-hosted container registry on the secure network, serving over HTTP rather than HTTPS, the Docker daemons on all nodes (including Bastion node) must be configured to support this as an "insecure registry"; it is highly recommended that either you use an infrastructure-provider secure registry, or apply CA-signed certificates (and associated DNS configurations) to the self-hosted registry to avoid this problem.