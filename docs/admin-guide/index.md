# Admin Guide

## Acronyms
- **HPC** - High-Performance Computing
- **MPI** - Message Passing Interface
- **GPU** - Graphics Processing Unit
- **CPU** - Central Processing Unit
- **FPGA** - Field-Programmable Gate Array
- **TPU** - Tensor Processing Unit
- **CUDA** - Compute Unified Device Architecture
- **OpenCL** - Open Computing Language
- **GPGPU** - General-Purpose Computing on Graphics Processing Units
- **FFT** - Fast Fourier Transform
- **LINPACK** - Linear Algebra Package
- **PDE** - Partial Differential Equation
- **CFD** - Computational Fluid Dynamics
- **IO** - Input/Output
- **RAM** - Random Access Memory
- **ROM** - Read-Only Memory
- **SSD** - Solid-State Drive
- **HDFS** - Hadoop Distributed File System
- **NFS** - Network File System
- **HDF5** - Hierarchical Data Format version 5
- **SSH** - Secure Shell
- **LAN** - Local Area Network
- **WAN** - Wide Area Network
- **SAN** - Storage Area Network
- **NUMA** - Non-Uniform Memory Access
- **SMP** - Symmetric Multiprocessing
- **HT** - Hyper-Threading
- **OS** - Operating System
- **BIOS** - Basic Input/Output System
- **UEFI** - Unified Extensible Firmware Interface
- **PBS** - Portable Batch System
- **SLURM** - Simple Linux Utility for Resource Management
- **TORQUE** - Terascale Open-source Resource and QUEue manager
- **PBS** - Portable Batch System
- **DRAM** - Dynamic Random Access Memory
- **ECC** - Error-Correcting Code
- **RAID** - Redundant Array of Independent Disks
- **SSD** - Solid State Drive
- **NVMe** - Non-Volatile Memory Express
- **IB** - InfiniBand
- **NIC** - Network Interface Card
- **FP64** - Double Precision Floating Point
- **FP32** - Single Precision Floating Point
- **FP16** - Half Precision Floating Point
- **FP8/FP4** - AI Floating point
- **AI** - Artificial Intelligence
- **ML** - Machine Learning
- **DL** - Deep Learning
- **VLSI** - Very-Large-Scale Integration
- **ASIC** - Application-Specific Integrated Circuit
- **CSI** - Container Storage Interface
- **ASU** - Lenovo Advanced Systems Utility
- **XCC** - XClarity Controller
- **BMO** - Bare Metal Operator

## Introduction

### Document Format
The documentation for Ubiquity is in markdown format - Embedded within the git repository and as such can be served as github pages or equivalent.

## System Introduction
Ubiquity is deployed as a high-availability control plane, using keepalived for a floating IP address.

The kubernetes service underpinning it then manages all other services as Kubernetes PODS. This includes:
- **DNS**
- **Ansible via AWX**
- **KeyCloak**
- **Monitoring**
- **Workload Managers (where necessary)**
- **Vault**
- **Onyxia (self-serve Kubernetes provisioning layer)**
- **ArgoCD**
- **BareMetalOperator**
- ..and so on.

All services are broken down into functions, the functions of which are within subdirectories:
- **Bootstrap** - For installing ArgoCD and its [app-of-apps](https://argo-cd.readthedocs.io/en/stable/operator-manual/cluster-bootstrapping/) functionality
- **System** - For system-underpinning services core to the cluster
- **Platform** - For user-accessible platforms
- **Apps** - For end-user applications
  
## Administrator users
All access to Ubiquity platforms is provided via a toolbox container called OPUS (also the tool used to deploy the cluster) - No administrative tooling is generally installed on the system unless site-specfics dictate that it's required. Therefore there are no specific users to mention other than the user who is entitled to spin up OPUS which is controlled by an AzureCR key.

Access directly to nodes can be achieved using SSH-key only - The root passwords on every node are locked. The key is created on installation and is an ED25519 key.

### The NFS Storage Subsystem
NFS is available within Ubiquity - It can be provided using an exisitng NFS share and provisioned out using the NFS CSI - There are also playbooks to provision an NFS server as a 3-way installation using Pacemaker (if you do not have an NFS share).

### The Lustre Storage Subsystem
Lustre is available within Ubiquity - It can be provided using an existing Lustre filesystem and provisioned out using the Lustre CSI or passing through the node as a directory.

### The Storage Scale Storage Subsystem
TBC

## IP Numbering Scheme
Generally a default install of a cluster follows the following convention, using an RFC1918-compliant 10.0.0.0/8 network.

Networking within Kubernetes uses the Kubernetes flannel network as a software-defined networking scheme, with IP addresses that are auto-allocated and can continually change - These never route out of the network and generate lots of firewall rules, operating via NFTables (as it's faster than IPTables).

Site-specific configurations will be in the site-specific folder.

### Variable Subnet Definitions:

- 10.0.0.x/22 - Vlan102 - OOB/IPMI
- 10.46.0.0/16 - Kubernetes flannel network
- 10.48.0.0/16 - Kubernetes service network
- 10.1.0.x/22 - Vlan103 - MGMT
- 10.8.0.x/22 - N/A - InfiniBand IPoIB

### IP Address Assignment
The general configuration is as follows:
- <OOB>.1-3 - OOB for control plane nodes 1-3
- <OOB>.<last subnet in OOB range>.251-253 - OOB interface on control plane nodes to control OOB devices
- <MGMT>.1-3 - Control plane nodes 1-3
- <IB>.1-3 - IB for control plane nodes 1-3
- <OOB>.11-?? - OOB for compute nodes
- <MGMT>.11-?? - MGMT for compute nodes
- <IB>.11-?? - IB for compute nodes

Site-specfic configurations will be in the site-specfic folder

### Ethernet port Allocation
Good practice would be to label all ports and shut all unused ports for physical security - This means that the port label on switch configs matches your running cluster and is merely a case of good housekeeping to keep on top of.

#### Ethernet Port connections on physical nodes
This should be addressed via port labelling on switches which should say nodename.portnum

## Initial System Build

### Hardware state
#### BIOS Settings
BIOS settings can be retrieved using the ASU tool within the XCC. You can SSH to the XCC address and use credentials from the BareMetalHost definition for the node to gain access.

##### Retrieve BIOS Settings from ASU
To retrieve BIOS Settings from ASU, you can use SSH to SSH to the OOB address for the XCC which then presents you with a `system>` prompt.

At this prompt a question mark at any point followed by enter will show you context and syntax help for any command.

From there, run the `asu show` command, this dumps out all current configuration for the node.

##### Clone BIOS Settings from One Node to Another (post-install addition to SMG)
Using the same process for XCC, use ASU to dump out all the configuration, then:

- Sed replace the '=' with a space
- Sed insert the first words on each line of output with a 'asu set '
- In the XCC terminal, paste.

This could be automated using a very simple container.

##### Hardware RAID Card configuration
RAID configuration can be set or shown via the XCC using ASU, or via the web interface for the node in question on the same IP address.

### Build sequence for HAWK systems from management node.
#### Install basic image
Images are created and deployed via the Bare Metal Operator. The BMO uses openstack ironic to deploy bare metal images, pre-built in qcow2 format.

These images are created using the diskimage-creator tool (please see training slides on usage, however these will be copied into the administration folder under disk image build, then uploaded/copied into the ironic HTTP server (in the `/shared/html/images` folder), and then referenced in the BareMetalHost definition/manifest.

## System management

### Access methods
The general access method for administrators is via OPUS, however if you have access to the ED25519 key and the kubeconfig.yaml then you can access from anywhere that has SSH access and network access to the in-band/out-of-band addresses, such as a bastion node. This simplifies access, and a bastion node can have all the required tooling necessary to run OPUS as a container to continue to avoid installing specific tools on a node (which becomes a security risk).

You can do this via SSH-tunnelling accordingly and defining inside your kubeconfig the `proxy:` setting to allow your kubernetes client to talk to the main cluster control plane.

### Test/Dev
You can test/dev changes to kubernetes environments via using the "sandbox" mode that is defined inside the deployment folder . This means that you can deploy thea copy of your environment settings as a k3d (or kubernetes in docker) image which duplicates your kubernetes environment inside a docker container that you can tear down once wfinished.

## MAINTENANCE PROCEDURES

### Getting Support
For community support, please visit our [GitHub repository](https://github.com/ubiquitycluster/ubiquity) and open an issue.

### General Maintenance Procedures
#### Power Up Procedures
To power up a Ubiquity cluster, the following steps should be taken:
- **Power Up switches**
- **Power Up Storage shelves and wait to settle**
- **Power Up Control Planes** - Once this comes online, all functions will be attempted to be restored to previous known state, including power status.
- **Power Up Storage Nodes (if not controlled by BareMetalHost)** - This _may_ occur irrespective of manual intervention in the event of an unscheduled power outage.
- **Power Up Compute Nodes** - Again, controlled by BareMetalHost
- **Confirm Service Restoration** - Again, Kubernetes will endeavour to restore service if they are Kubernetes Pods - Else manual steps to restore services such as Workload Managers etc.
  
#### Power Down Procedures
- **Start with workers**:

- list worker nodes in kubetctl or k9s  - `kubectl get nodes`
- cordon (drain+unschedule) worker nodes - `kubectl cordon <node>`
- wait for worker nodes to be clean
- BareMetalHost definition online to false (See below) << only use ipmitool if absolutely necessary (from opus)
 
##### Control plane:
Make sure no workers are running

ssh to each control plane node and issue init 0
 

To check status of workers

```
kubectl get nodes | egrep -v '(control-plane|etcd|master|^NAME)'| awk -F '.' '{ print $1}' | while read line; doecho $line; kubectl -n metal-nodes get bmh $line -o custom-columns=ONLINE:.spec.online;echo ""; done
```
 
##### Cordon workers
 
```
kubectl get nodes | egrep -v '(control-plane|etcd|master|^NAME)'| awk '{ print $1 }' | while read line; dokubectl cordon $line; done
```
 
Get bmh_host kubectl get bmh -n metal-nodes...

##### Get bmh_host
```
kubectl get bmh -n metal-nodes
```
##### poweroff
```
kubectl -n metal-nodes annotate bmh <bmh_host> --overwrite reboot.metal3.io/poweroff=""
```
##### reboot
```
kubectl -n metal-nodes annotate bmh <bmh_host> --overwrite reboot.metal3.io='true'
```
##### Remove Poweroff flag: Add a dash at the end of the annotation for poweroff
```
kubectl -n metal-nodes annotate bmh <bmh_host> --overwrite reboot.metal3.io/poweroff-
```

##### Powering on a Compute Node
A node defined inside the BareMetalHost definition as long as the online: true state is defined will automatically be powered on. This is the "default state" for a BareMetalHost.

##### Marking a Compute Node Down for maintenance
See [##### poweroff] for details

##### Resuming a Compute Node
See [##### Remove Poweroff flag: Add a dash at the end of the annotation for poweroff] for details
            
##### Viewing Downed Nodes
Using OPUS, you can see node status by:

```
kubectl get nodes
```
Or by running k9s and going to the `:nodes` view.

##### Replacing a Compute Node System
To replace a compute node, simply change the MAC address inside the BareMetalHost definition, and configure the BMC with the same address (and password). If necessary, clone the BIOS settings as per [##### Clone BIOS Settings from One Node to Another (post-install addition to SMG)]

## Business as Usual Management

### Unbanning login node IPs (fail2ban)
Dependent on system configuration specifics, however if fail2ban is installed please follow these steps to unban an IP:
<fail2ban remove commands>

### Editing slurm.conf
There are 3 modes of operation for Ubiquity that can be chosen:
- Baremetal - Effectively has monitoring and general services on Kubernetes, but all workload manager functions and compute functions are baremetal.
- Hybrid - Workload manager control daemons are inside Kubernetes and can communicate with baremetal workers.
- Native - Workload manager control daemons are inside Kubernetes and worker nodes are PODS within Kubernetes.

To edit in baremetal mode:
Edit the ansible playbook in ubiq-playbooks and edit the vars for it, located in `ubiq-playbooks/workload-managers/<workload manager>/vars/main.yml` - Then call the workload-managers role within AWX on the nodes required.

To edit in Hybrid or Native mode:
Edit the values.yaml inside platform/hpc-ubiq/slurm/base/values.yaml - Add required configs and git commit, git push.

Wait for undrain-nodes-hook to complete.

### Work Load Management
#### Slurm Accounting
To inspect slurm accounting, login to a slurm node that has slurm control rights and run sacctmgr. 
In addition, a slurm-exporter is included in all installation modes and can be access by adding a slurm dashboard for monitoring purposes.

### Monitoring
Monitoring for Ubiquity is provided by Prometheus, and visualised with Grafana. There are alerts that can be configured using AlertManager.
To login to grafana, using a web browser go to [grafana][https://grafana.<dns-naming-convention>]
You can also log into the prometheus instance, however all configuration for prometheus is inside git and we don't provision an ingress for prometheus for this reason.
For more detail on prometheus and grafana, please see [administration/tutorials/kube-prometheus-stack.md](administration/tutorials/kube-prometheus-stack.md)

For logging purposes, Loki is used - which aggregates logging information to Grafana to be able to be searched interactively. For more details please see [administration/tutorials/loki-promtail.md](administration/tutorials/loki-promtail.md)

### Customising Monitoring
All configuration for monitoring is inside git and controlled by values.yaml for monitoring-system. Edit the values.yaml for the monitoring platform, git commit, git push and wait for ArgoCD to implement monitoring changes.

### Monitoring users          
Monitoring within Ubiquity defines a default admin user, of which you can get the admin user password from a secret within Kubernetes. There is also a helpful convenience script called `grafana-admin-password.sh` which when ran from OPUS gives you the admin password.

This can be changed inside the values.yaml - As explained in [### Customising Monitoring].

You can add extra users either inside values.yaml or you can attach external authentication methods such as KeyCloak.

## NFS/Lustre Storage Administration
### NFS Quota
### NFS Faults
### Lustre Quota
### Lustre Faults
### Lustre Performance Tracking
### Disk Failure

### Pacemaker/Corosync troubleshooting
pcs status etc

## Log files

## Cluster Shell
To run commands om multiple nodes, please see [run-commands-on-multiple-nodes](administration/tutorials/run-commands-on-multiple-nodes.md)
    
## Serial Console
Attach via XCC or IPMITool
    
## Compiling Software, Adding Modules
Adding a new Environment Module

## Load balancing users across the login nodes
DNS round robin?

## Updating InfiniBand HCA firmware
### Using mlxfw

## InfiniBand Troubleshooting

## LDAP (openLDAP-based)
Ubiquity can run a 3-way openLDAP instance that can import LDIFs and manage appropriately.

## Compute Node Frequency Control & Turbo
This can be controlled via scripts that are present inside the ubiq-playbooks directory inside the HPCTuning role.

## Additional Package Installation

## Firewalld configuration

## Temperature monitor
Ubiquity can shutdown on thermal events using the IPMI exporter within Grafana

## Backups and DR
For more information, please see [administration/tutorials/backup-cluster.md](administration/tutorials/backup-cluster.md)
