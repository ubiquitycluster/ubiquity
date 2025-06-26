# SLURM Container

This project containerizes the SLURM HPC job scheduler in a Kubernetes containers orchestrator. You can find an (almost) ready-to-use Kubernetes recipe in ubiquity. This repository contains a YAML file to instantiate the containerized HPC cluster in Kubernetes and Dockerfiles (and their related files) for building images.

### Step 1 : Create your namespace

You have to create a dedicated namespace to welcome your containerized HPC cluster:

```sh
root@admin:~# kubectl create namespace hpc-ubiq
namespace/hpc-ubiq created
```

### Step 2 : Apply the RBAC policy

You must allow read access to pods properties that run in the namespace created in step 1. A rbac.yaml file is supplied and you can apply it:

```sh
root@admin:~# kubectl apply -f rbac.yaml
clusterrole.rbac.authorization.k8s.io/pods-list-hpc-ubiq created
clusterrolebinding.rbac.authorization.k8s.io/pods-list-hpc-ubiq created
```

### Step 3: Create your storage options

You must create a shared storage volume that runs in the namespace created in step 1. There are options for both longhorn and nfs included (nfs maps an external NFS volume, so obviously make sure it's up and accessible):
```sh
root@admin:~# cd storage/nfs
root@admin:~# kubectl apply -f pv-nfs.yaml -f pvc-nfs.yaml
```

### Step 4 : Launch your HPC cluster in k8s

A sample file slurm.yaml is supplied. It instanciates a Slurmctld service with two Slurmd nodes, each with 2 CPUs. You can apply the slurm.yaml file:

```sh
root@admin:~# kubectl apply -f slurm.yaml
service/nodes created
statefulset.apps/hpc-node created
statefulset.apps/control-node created
```

In this file you may want to customize three attributes:

#### 1. The number of Slurmd replicas:

```yaml
spec:
  selector:
    matchLabels:
      app: slurmd
  serviceName: "nodes"
  replicas: 2
```

For that, check the last line ***replicas: 2***.

#### 2. The number of vCPUs for each Slurmd replica:

```yaml
      containers:
      - name: slurmd
        image: nyk0/slurmcontainer
        volumeMounts:
        - mountPath: /run/munge
          name: sock
        - mountPath: /locate
          name: locate
        resources:
          limits:
            cpu: "2"
          requests:
            cpu: "2"
```

You must adapt two parameters ***limits / cpu*** and ***requests / cpu*** to use exactly the correct CPU amount (here, 2 vCPUs).

#### 3. Import your Docker Hub credentials:

To access the SLURM images you may need to customize the section ***ImagPullSecrets*** of each pod:

```yaml
      imagePullSecrets:
      - name: regcred
```

You can find the steps to include your Docker Hub credentials in Kubernetes [here](https://kubernetes.io/fr/docs/tasks/configure-pod-container/pull-image-private-registry/).

### Step 4: Check pods availibility 

List pods in the dedicated namespace:

```sh
root@admin:~# kubectl get pods -n hpc-nico
NAME             READY   STATUS    RESTARTS   AGE
control-node-0   3/3     Running   0          81s
hpc-node-0       2/2     Running   0          81s
hpc-node-1       2/2     Running   0          74s
```

### Step 5: Check SLURM cluster

You have to go in ***slurmctld*** container that belongs to the ***control-node-0*** pod:

```sh
root@admin:~# kubectl exec -n hpc-ubiq -ti control-node-0 -c slurmctld -- /bin/bash
root@control-node-0:/#
```

And now you can su to a common user account included in custom ***slurmctld*** image:

```sh
root@control-node-0:/# su - nico
nico@control-node-0:~$
```

Finally, display the SLURM topology:

```sh
nico@control-node-0:~$ sinfo
PARTITION AVAIL  TIMELIMIT  NODES  STATE NODELIST
COMPUTE*     up   infinite      2   idle hpc-node-[0-1]
```

It's OK, we get two nodes waiting for jobs. You can now run a very simple job with srun command:

```sh
nico@control-node-0:~$ srun -N 2 hostname
hpc-node-0
hpc-node-1
```

### Step 6: Launch MPI job

SLURM images include OpenMPI and a testing code in C; you can find it (and run it from) the test user home directory :

```sh
nico@control-node-0:~$ srun -n 4 ./pi 256
Elapsed time = 0.000006 seconds
Pi is approximately 3.1875000000000000, Error is 0.0459073464102069
```

### Step 7: resize your HPC cluster

We currently have two containerized HPC nodes:

```sh
nico@control-node-0:~$ sinfo
PARTITION AVAIL  TIMELIMIT  NODES  STATE NODELIST
COMPUTE*     up   infinite      2   idle hpc-node-[0-1]
```

They both respond:

```sh
nico@control-node-0:~$ srun -N 2 hostname
hpc-node-0
hpc-node-1
```

Patch the stateful set to move from 2 replicas to 3 replicas:

```sh
root@admin:~# kubectl patch statefulsets hpc-node -n hpc-nico -p '{"spec":{"replicas":3}}'
statefulset.apps/hpc-node patched
```

After few seconds, our containerized cluster scaled-up:

```sh
nico@control-node-0:~$ sinfo
PARTITION AVAIL  TIMELIMIT  NODES  STATE NODELIST
COMPUTE*     up   infinite      3   idle hpc-node-[0-2]
```

We can run a simple job:

```sh
nico@control-node-0:~$ srun -N 3 hostname
hpc-node-0
hpc-node-1
hpc-node-2
```

We can also run an MPI job to check that communication on the scaled containerized HPC cluster work:

```sh
nico@control-node-0:~$ srun -n 6 ./pi 256
Elapsed time = 0.000005 seconds
Pi is approximately 3.0000000000000000, Error is 0.1415926535897931
```

## Know issues

### Issue 1 :

We have this containerized HPC cluster ressources (2 nodes with 2 CPUs each):

```sh
nico@control-node-0:~$ srun -N 2 hostname
hpc-node-0
hpc-node-1 
```

Let's submit a job too large for our current set of ressources:

```sh
nico@control-node-0:~$ srun -N 3 hostname
srun: Requested partition configuration not available now
srun: job 10 queued and waiting for resources 
```

Now, we scale up to 3 replicas:

```sh
root@admin:~# kubectl patch statefulsets hpc-node -n hpc-nico -p '{"spec":{"replicas":3}}'
statefulset.apps/hpc-node patched
```

As soon as the new replica joins our containerized HPC cluster, the pending job fails with these messages:

```sh
srun: job 10 has been allocated resources
srun: error: fwd_tree_thread: can't find address for host hpc-node-2, check slurm.conf
srun: error: Task launch for StepId=10.0 failed on node hpc-node-2: Can't find an address, check slurm.conf
srun: error: Application launch failed: Can't find an address, check slurm.conf
srun: Job step aborted: Waiting up to 32 seconds for job step to finish.
hpc-node-0
hpc-node-1
srun: error: Timed out waiting for job step to complete 
```

If you re-run the command, it works. The reason is that if a pending job relies on ***srun*** and is scheduled on the new coming HPC node, it will fail. You can find more references [here](https://slurm.schedmd.com/faq.html#add_nodes). This feature will be fully supported in the 23.02 version of SLURM. You can find the roadmap [here](https://slurm.schedmd.com/SLUG21/Roadmap.pdf) (slide "Truly Dynamic Nodes").
