# Loki-Promtail

## Prerequisites on loki promtail
You need to provide a storage provider for this component. Proceed with the following prerequisite description to use the Velero Building Block out of the box.

## Recommended setup
A recommended resource overview is listed in the table below.

CPU/vCPU	Memory
none	512MiB + 256 per node
No further activities need to be carried out in advance.

## Adding Loki-Promtail to your cluster
Add the directory loki to your system directory within your repository. 

## Required configuration
No confguration is required.

## What logs are collected?
The building block consists of two parts: Loki and Promtail. Promtail runs as a DaemonSet on each node in the cluster and collects logs. Loki is running as a central instance in the cluster and stores the logs it receives from Promtail. When looking into the logs, you usually only interact with Loki.
The building block collects stdout and stderr of all pods in the cluster. In addition to the pods deployed by the SysEleven provided building blocks, this also includes the logs of all application pods deployed by the user.
In addition to the pod logs the systemd journals of all nodes in the cluster are collected.
By default the logs are stored by Loki for one week. The retention period can be adjusted in values.yaml.

## How can I access the logs?
The easiest way to access the logs is Grafana provided with the kube-stack-prometheus module. We automatically create a Loki data source for you in it, so you can skip this step in the linked upstream documentation. In Grafana, you can use the explore feature to take a look at your logs and refine your query.
For something more permanent, you can add a log panel to one of your dashboards.

If you prefer the command line you can use logcli.

## Monitoring

### Additional Alertrules
None

### Additional Grafana dashboards
Loki Top 10 producer
An overview of the top 10 kubernetes namespaces that produce logs
b
### Loki & Promtail
An overview of performance metrics from Loki and Promtail

### Scale loki volume
If you need to scale the persistent volume Loki stores your data on, perform the following steps:

# Delete StatefulSet (the PVC will remain)
kubectl delete sts loki

# Patch the PVC to e.g. 10Gi
kubectl patch pvc storage-loki-0 -p '{"spec":{"resources":{"requests":{"storage":"10Gi"}}}}' --type=merge
Then, adapt your adapt values.yaml, e.g. values-loki-stage.yaml for the same size as the patch above.

To deploy the StatefulSet again, push your changes and merge them to the default branch - the CI will then deploy it again.

Scale Setup
This building block consists of multiple components. Each of the components can and must be scaled individually.

Scaling Loki
Also see Scaling with Loki for upstream documentation.

Scaling replicas is not supported with this building block. See loki-distributed for a possible solution.
Requests/limits for CPU/memory can be adjusted
Scaling Promtail
Runs as one DaemonSet on each node, so no further replica scaling is needed
Requests/limits for CPU/memory can be adjusted
Release-Notes
Please find more infos on release notes and new features Release notes Loki-Promtail