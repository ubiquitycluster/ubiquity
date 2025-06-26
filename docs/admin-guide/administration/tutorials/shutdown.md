# How to shutdown a Ubiquity cluster

## Situation

### Task
This article provides instructions for safely shutting down a Ubiquity cluster provisioned.


## Requirements
- A Ubiquity cluster

## Background
If you have a need to shut down the infrastructure running a Ubiquity cluster (datacentre maintenance, migration, etc.) this guide will provide steps in the proper order to ensure a safe cluster shutdown.

Please ensure you complete an etcd backup before continuing this process. A guide regarding the backup and restore process can be found here.

## Solution
N.B. If you have nodes that share worker, control plane, or etcd roles, postpone the docker stop and shutdown operations until worker or control plane containers have been stopped.

## Turning off any bare-metal services
If you deployed your cluster using bare-metal services (i.e. not hybrid pod or k8s-native), then you are responsible for disabling and re-enabling those services. You can do that manually by shutting those respective services down cleanly.

### Bare-Metal Service: Slurm
Shut down Slurm following best practices:

- Place a maintenance reservation on the cluster in question, to prevent any new jobs running. Change dates/times and duration where necessary for your outage window. The duration is in minutes:
```
scontrol create reservation starttime=2023-09-15T10:30:00 duration=5760 user=root flags=maint,ignore_jobs nodes=ALL
```

Confirm the reservation is in-place:
```
scontrol show reservations
```
Ensure no jobs are running:

```
squeue
```

Then proceed to shut down the service accordingly:

- Shut down slurmctld on both nodes:
```
systemctl stop slurmctld
```

Wait for a minute for any writes from slurmctld to have finished to the slurmdbd instance.

- Shut down slurmdbd on both nodes:
```
systemctl stop slurmdbd
```

Again, wait for a minute for any writes to the MySQL database to be finished

- Then shutdown MySQL on both nodes:
```
systemctl stop mysqld
```

- Confirm that mysqld is off:

```
systemctl status mysqld
```

At which point you are ok to shutdown slurmd on the remaining compute nodes (but this step is not essential)

```
systemctl stop slurmd
```

### Bare-Metal Service: HTCondor

### Draining Worker Nodes
For all worker nodes, prior to stopping the containers, run:

```
kubectl get nodes
```

To identify the desired node, then run:
```
kubectl drain <node name>
```

This will safely evict any pods, and you can proceed with the following steps to a shutdown.

### Shutting down the workers nodes
For each worker node:

- ssh into the worker node
- stop kubelet and kube-proxy by running `sudo docker stop kubelet kube-proxy`
- stop docker by running `sudo service docker stop` or `sudo systemctl stop docker`
- shutdown the system `sudo shutdown now`

### Shutting down the control plane nodes
For each control plane node:

- ssh into the control plane node
- stop kubelet and kube-proxy by running sudo docker stop kubelet kube-proxy
- stop kube-scheduler and kube-controller-manager by running sudo docker stop kube-scheduler kube-controller-manager
- stop kube-apiserver by running sudo docker stop kube-apiserver
- stop docker by running sudo service docker stop or sudo systemctl stop docker
- shutdown the system sudo shutdown now

### Shutting down the etcd nodes
For each etcd node:

- ssh into the etcd node
- stop kubelet and kube-proxy by running sudo docker stop kubelet kube-proxy
- stop etcd by running sudo docker stop etcd
- stop docker by running sudo service docker stop or sudo systemctl stop docker
- shutdown the system sudo shutdown now

### Shutting down storage
Shut down any persistent storage devices that you might have in your datacenter (such as NAS storage devices) if applicable. It iss important that you do this after shutting everything else down to prevent data loss/corruption for containers requiring persistency.

N.B. If you are running a cluster that was not deployed through RKE then the order of the process is still the same, however the commands may vary. For instance, some distributions run kubelet and other control plane items as a service on the node rather than in docker. Check documentation for the specific Kubernetes distribution for information as to how to stop these services.

### Starting a Kubernetes cluster up after shutdown
Kubernetes is good about recovering from a cluster shutdown and requires little intervention, though there is a specific order in which things should be powered back on to minimize errors.

Power on any storage devices if applicable.

Check with your storage vendor on how to properly power on you storage devices and verify that they are ready.

For each etcd node:
- Power on the system/start the instance.
- Log into the system via ssh.
- Ensure docker has started sudo service docker status or sudo systemctl status docker
- Ensure etcd and kubelet’s status shows Up in Docker sudo docker ps

For each control plane node:
- Power on the system/start the instance.
- Log into the system via ssh.
- Ensure docker has started sudo service docker status or sudo systemctl status docker
- Ensure kube-apiserver, kube-scheduler, kube-controller-manager, and kubelet’s status shows Up in Docker sudo docker ps

For each worker node:
- Power on the system/start the instance.
- Log into the system via ssh.
- Ensure docker has started sudo service docker status or sudo systemctl status docker
- Ensure kubelet’s status shows Up in Docker sudo docker ps
- Log into K9s (or use kubectl) and check your various projects to ensure workloads have started as expected. This may take a few minutes depending on the number of workloads and your server capacity.