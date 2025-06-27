# Ubiquity Documentation

## Table of Contents

1. [Introduction](#1-introduction)
2. [Prerequisites](#2-prerequisites)
3. [Ubiquity Architecture Overview](#3-ubiquity-architecture-overview)
4. [Installation](#4-installation)
5. [User Guide](#5-user-guide)
6. [Admin Guide](#6-admin-guide)
7. [Developer Guide](#7-developer-guide)
8. [Support](#8-support)
9. [Contributing](#9-contributing)
10. [License](#10-license)

## 1. Introduction
Ubiquity is a platform and framework for managing HPC resources. It does this by creating a cloud-like environment to deploy and manage HPC resources. 
It is used both for deploying and controlling both internal enterprise IT resources and cloud services in a ubiquitous manner. 

The concept of ubiquitous computing was first introduced by Mark Weiser in 1988.
The idea is that computing is everywhere, and it is invisible to the user.
Ubiquity is a platform that allows you to deploy and manage HPC resources in a ubiquitous manner.
It does this by leveraging the concept of GitOps and using modern technologies like Kubernetes and Ansible to deliver and manage platforms.

Ubiquity as a framework is composed of the following main components, identified by folders within this repository:

- metal - A bare-metal bootstrapper that deploys Kubernetes
- cloud - A cloud bootstrapper that deploys Kubernetes in cloud providers
- bootstrap - The second-stage of the bootstrap process, provisioning the Kubernetes pods that our core services run from
- system - The core services that make up the Ubiquity platform
- storage - The storage services that Ubiquity uses to store data
- platform - The platform services that Ubiquity uses to deploy and manage applications
- apps - The applications that Ubiquity can deploy and manage

The remainder of the directories in here effectively are tools and utilities to either stand-up components or run scripts, or document more information about Ubiquity.

There is a special folder called `disabled` that effectively is a mirror of the main folders in Ubiquity, and you simply move a component in to that directory to disable components.

Ubiquity is open-source, extendable and comes with a [professional support](about/support.md) provided by Logicalis.

To get a quick feeling as to what Ubiquity is, take a look at some [screenshots](about/screenshots.md).

If you are interested in deploying, check the [getting started](getting-started.md) section, but read ahead first!

## 2. Prerequisites

To use Ubiquity, you will need:
1. Administrative access, and both Docker and make installed on your system

Either:
2. Hardware to deploy on (optional)
3. Ability to communicate with the hardware using a common network switch environment

Or:
2. Authenticated access to a cloud (optional)
3. Ability to communicate with the cloud provider API from your computer
4. A project with operational limits meeting the requirements described in _Quotas_ subsection.

5. Docker installed on your system
5. ssh-agent running and tracking your SSH key

Thankfully, the setup environment and all tooling is all provided for you by cloning the git
repository for Ubiquity, and running `make tools`. This will spin up a docker container called Opus
(named after the latin for creation) which has all the tools you need to run and install a Ubiquity cluster.

### 2.1 Hardware
Hardware specs for the project are very modest:

### 2.1.1 Sandbox mode
In sandbox mode, you can deploy on your local system. This is to do a functional test of the features available via a Ubiquity environment.

The specs of such a deployment are as follows:

- 1x Lenovo `ThinkPad P16s G1`:
    - CPU: `Intel Core i7-12890T @ 3.4GHz`
    - RAM: `32GB`
    - SSD: `1TB`

### 2.1.2 Production mode (dev/test)
In production (dev/test) mode, you can deploy on a cluster of machines. This is to do a functional test of the features available via a Ubiquity environment only - it is not intended for production use.

This is what Ubiquity was born from and is a bare minimum to get a functional cluster up and running.

The example specifications of such a deployment are as follows:

- 3 × Lenovo `ThinkCentre M700 Tiny`:
    - CPU: `Intel Core i5-6600T @ 2.70GHz`
    - RAM: `16GB`
    - SSD: `500GB`
- Netgear `GS305E` switch:
    - Ports: `5`
    - Speed: `1000Mbps`
- 1x Lenovo `ThinkPad P16s G1`:
    - CPU: `Intel Core i7-12890T @ 3.4GHz`
    - RAM: `32GB`
    - SSD: `1TB`

Yes. It runs on tiny machines. It's not fast, but it works.

Here's an example of what that looks like:

![Hardware](assets/Rig.jpg)

### 2.1.3 Production mode (prod)
In production (prod) mode, you can deploy on a cluster of machines. This is intended for production use.

The example specifications of such a deployment are as follows. These are not intended to be a recommendation, but rather an example of what we have used in the past:

- Control Plane Nodes
  - 3 x Lenovo `ThinkSystem SR650`:
    - CPU: `Intel Xeon Gold 6240 @ 2.60GHz`
    - RAM: `256GB`
    - SSD 1&2: `480GB in RAID 1`
    - SSD 3&4: `480GB in RAID 1 for ETCD` 
    - SSD 5&6: `2TB in RAID 1 for data`

- Worker Nodes
  - 16 x Lenovo `ThinkSystem SR630`:
    - CPU: `Intel Xeon Gold 6240 @ 2.60GHz`
    - RAM: `256GB`
    - SSD: `480GB`

- Switches
  - Ethernet
    - OOB
      - 1 x Lenovo `AS4610` switch:
      - Ports: `48`
      - Speed: `1000Mbps`
    - Internal
      - 1 x Lenovo `SN2010` switch:
      - Ports: `48`
      - Speed: `25Gbps`
  - Infiniband
    - 1 x Mellanox `QM8700` switch:
    - Ports: `32`
    - Speed: `200Gbps`

The equipment should be cabled (logically) as follows:
```
              Ethernet
          ---AS4610(OOB)
         /    /  ||  \ 
        /    / (mgmt) \
       /    /  SN2010  \ 
      /    /  /      \  \    
     /    /  /        \  \  
    /    /  /          \  \     
   |    /  /            \  \        
   |    CPs            Workers
    \   \  \            /
     \   \  \          /  
      \   \  \        /  
       \   \  \      /  
        \      QM8700  
         ----Infiniband
```      
Control planes and worker nodes should be connected to the SN2010 switch, and the SN2010 switch should be connected to the AS4610 switch. The AS4610 switch should be connected to the OOB network, which should be connected to the BMCs of the control plane and worker nodes, and the OOB for the QM8700 switch should be connected to the AS4610 switch so it can be managed by the Control Plane nodes.

You can add more things into this like GPU nodes, Login nodes, storage nodes, etc. but this is the bare minimum to get a functional cluster up and running.

### 2.1 Authentication
To use Ubiquity on a cloud provider, you will need to do the following (dependent on your cloud provider):

#### 2.1.1 Amazon Web Services (AWS)
<!-- markdown-link-check-disable -->
1. Go to [AWS - My Security Credentials](https://console.aws.amazon.com/iam/home?#/security_credentials)
2. Create a new access key.
3. In a terminal, export `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS Access Key and AWS Secret Key:
    ```shell
    export AWS_ACCESS_KEY_ID="an-access-key"
    export AWS_SECRET_ACCESS_KEY="a-secret-key"
    ```
<!-- markdown-link-check-enable -->

Reference: [AWS Provider - Environment Variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#environment-variables)

#### 2.1.2 Google Cloud

1. Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/downloads-interactive)
2. In a terminal, enter : `gcloud auth application-default login`

#### 2.1.3 Microsoft Azure

1. Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
2. In a terminal, enter : `az login`

Reference : [Azure Provider: Authenticating using the Azure CLI](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/azure_cli)

#### 2.1.4 OpenStack / OVH

1. Download your OpenStack Open RC file.
It is project-specific and contains the credentials used
by Terraform to communicate with OpenStack API.
To download, using OpenStack web page go to:
**Project** → **API Access**, then click on **Download OpenStack RC File**
then right-click on **OpenStack RC File (Identity API v3)**, **Save Link as...**,
and save the file.

2. In a terminal located in the same folder as your OpenStack RC file,
source the OpenStack RC file:
    ```
    source *-openrc.sh
    ```
This command will ask for a password, enter your OpenStack password.

### 2.2 Cloud API

Once you are authenticated with your cloud provider, you should be able to
communicate with its API. This section lists for each provider some
instructions to test this.

#### 2.2.1 AWS

1. In a dedicated temporary folder, create a file named `test_aws.tf`
with the following content:
    ```hcl
    provider "aws" {
      region = "us-east-1"
    }

    data "aws_ec2_instance_type" "example" {
      instance_type = "t2.micro"
    }
    ```
2. In a terminal, move to where the file is located, then:
    ```shell
    terraform init
    ```
3. Finally, test terraform communication with AWS:
    ```
    terraform plan
    ```
    If everything is configured properly, terraform will output:
    ``` 
    No changes. Your infrastructure matches the configuration.
    ```
    Otherwise, it will output:
    ```
    Error: error configuring Terraform AWS Provider: no valid credential sources for Terraform AWS Provider found.
    ```
4. You can delete the temporary folder and its content.

#### 2.2.2 Google Cloud

In a terminal, enter:
```
gcloud projects list
```
It should output a table with 3 columns
```
PROJECT_ID NAME PROJECT_NUMBER
```

Take note of the `project_id` of the Google Cloud project you want to use,
you will need it later.

#### 2.2.3 Microsoft Azure

In a terminal, enter:
```
az account show
```
It should output a JSON dictionary similar to this:
```json
{
  "environmentName": "AzureCloud",
  "homeTenantId": "<uuid>",
  "id": "<uuid>",
  "isDefault": true,
  "managedByTenants": [],
  "name": "Pay-As-You-Go",
  "state": "Enabled",
  "tenantId": "<uuid>",
  "user": {
    "name": "user@example.com",
    "type": "user"
  }
}
```

#### 2.2.4 OpenStack / OVH

1. In a dedicated temporary folder, create a file named `test_os.tf`
with the following content:
    ```hcl
    terraform {
      required_providers {
        openstack = {
          source  = "terraform-provider-openstack/openstack"
        }
      }
    }
    data "openstack_identity_auth_scope_v3" "scope" {
      name = "my_scope"
    }
    ```
2. In a terminal, move to where the file is located, then:
    ```shell
    terraform init
    ```
3. Finally, test terraform communication with OpenStack:
    ```
    terraform plan
    ```
    If everything is configured properly, terraform will output:
    ``` 
    No changes. Your infrastructure matches the configuration.
    ```
    Otherwise, it will output:
    ```
    Error: Error creating OpenStack identity client:
    ```
    if the OpenStack cloud API cannot be reached.
4. You can delete the temporary folder and its content.

### 2.3 Quotas

#### 2.3.1 AWS

The default quotas set by Amazon are sufficient to build the Ubiquity
AWS examples. To increase the limits, or request access to special
resources like GPUs or high performance network interface, refer to
[Amazon EC2 service quotas](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html).

#### 2.3.2 Google Cloud

The default quotas set by Google Cloud are sufficient to build the Ubiquity
GCP examples. To increase the limits, or request access to special
resources like GPUs, refer to
[Google Compute Engine Resource quotas](https://cloud.google.com/compute/quotas).

#### 2.3.3 Microsoft Azure

The default quotas set by Microsoft Azure are sufficient to build the Ubiquity
Azure examples. To increase the limits, or request access to special
resources like GPUs or high performance network interface, refer to
[Azure subscription and service limits, quotas, and constraints](https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/azure-subscription-service-limits).

#### 2.3.4 OpenStack

Minimum project requirements:
* 1 floating IP
* 1 security group
* 1 network (see note 1)
* 1 subnet (see note 1)
* 1 router (see note 1)
* 3 volumes
* 3 instances
* 8 VCPUs
* 7 neutron ports
* 12 GB of RAM
* 11 security rules
* 80 GB of volume storage

**Note 1**: Ubiquity supposes that the OpenStack project comes with a network, a subnet and a router already initialised. If any of these components is missing, you will need to create them manually before launching terraform, however Ubiquity will do this during its install.
* [Create and manager networks, JUSUF user documentation](https://apps.fz-juelich.de/jsc/hps/jusuf/cloud/first_steps_cloud.html?highlight=dns#create-and-manage-networks)
* [Create and manage network - UI, OpenStack Documentation](https://docs.openstack.org/horizon/latest/user/create-networks.html)
* [Create and manage network - CLI, OpenStack Documentation](https://docs.openstack.org/ocata/user-guide/cli-create-and-manage-networks.html)

#### 2.3.5 OVH

The default quotas set by OVH are sufficient to build the Ubiquity OVH examples. To increase the limits, or request access to special resources like GPUs, refer to
[OVHcloud - Increasing Public Cloud quotas](https://docs.ovh.com/ie/en/public-cloud/increase-public-cloud-quota/).

### 2.4 ssh-agent

To transfer configuration files, Terraform will connect to your cluster using SSH.
To avoid providing your private key to Terraform directly, you will have to
add it to the authentication agent, ssh-agent.

To learn how to start ssh-agent and add keys, refer to
[ssh-agent - How to configure, forwarding, protocol](https://www.ssh.com/academy/ssh/agent).

**Note 1**: If you own more than one key pair, make sure the private key added to
ssh-agent corresponds to the public key that will be granted access to your cluster
(refer to [public_keys](admin-guide/deployment/cloud/index.md#49-public_keys)).

**Note 2**: If you have no wish to use ssh-agent, you can configure Ubiquity to
generate a key pair specific to your cluster. The public key will be written in
the sudoer `authorized_keys` and Terraform will be able to connect the cluster
using the corresponding private key. For more information,
refer to [generate_ssh_key](admin-guide/deployment/cloud/index.md/#415-generate_ssh_key-optional).

## 3. Ubiquity Architecture Overview

![Ubiquity Architecture](./architecture/overview.md)

## 4. Installation

[Getting Started](getting-started.md)

## 5. User Guide

[User Guide](user-guide/index.md)

## 6. Admin Guide

[Admin Guide](admin-guide/index.md)

## 7. Developer Guide

[Developer Guide](developers/developers.md)
[Releasing Guide](developers/releasing.md)
[Roadmap](developers/roadmap.md)

## 8. Support

[Support](about/support.md)

## 9. Contributing

[Contributing](../CONTRIBUTING.md)

## 10. License

[License](../LICENSE)