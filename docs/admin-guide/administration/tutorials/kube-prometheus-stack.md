# Kube-Prometheus
Kube-Prometheus is the CNCF monitoring system and time series database. It comes with an Alertmanager and Grafana as a WebUI component to visualise collected metrics.

## Recommended setup
Before you deploy the Kube-Prometheus, you should consider the following recommendations.

cpu (vCPU)	Memory
6 CPU + 0.5 per node	6320MiB + 50MiB per node
No further activities need to be carried out in advance.

## Adding Kube Prometheus Stack to your cluster
Add the directory monitoring-system to your master ubiquity repository. This directory contains the building block for the Kube-Prometheus stack.

## Configuration
### Required configuration
You have to set grafana.adminPassword. If you don’t, the Grafana admin password changes on each CI run.

Configure it in values.yaml:

```yaml
grafana:
  adminPassword: highly-secure-production-password
```

### Configuring alertmanager
When adding a receiver, you need to copy the null receiver config into your own cluster configuration as well.
If you customise the configuration of a BB with values.yaml files, you have to be careful that you cannot simply extend lists. Helm cannot merge lists. Therefore, the existing list plus the new entry must be added to the customised configuration.

```yaml
alertmanager:
  config:
    receivers:
      - name: "null"  # Add this to your config as well
      - name: myotherreceiver
        webhook_configs:
          - send_resolved: true
            url: https://myurl
```

With kube-prometheus-stack we already deploy an Alertmanager for you. In combination with Prometheus and the default rules a lot of base metrics are monitored and alerts for them are created when something goes wrong.
In the default settings alerts are only visible in the webinterface of the alertmanager. Most of the time it is desirable to send those alerts to your operations team, your on-call engineer or someone else. To achieve that you can configure the alertmanager in values.yaml. There you can use the alertmanager.config.receivers setting to set all available options supported by alertmanager.

Normally no alert should be triggered, but there is an exception! In the default configuration the Prometheus operator creates a watchdog alarm which is always triggered. This alarm can be used to check if your monitoring is working. If it stops triggering, either Prometheus or the alert manager is not working as expected.
You can set up an external alerting provider (or a webhook hosted by you) to notify you when the alert stops triggering.

### Forward Alertmanager to localhost
```bash
kubectl port-forward -n monitoring-system alertmanager-kube-prometheus-stack-alertmanager-0 9093 
```

### Send test alerts
You can send test alerts to an Alertmanager instance to test alerting.

```bash
kubectl port-forward -n monitoring-system alertmanager-kube-prometheus-stack-alertmanager-0 9093 &

curl -H "Content-Type: application/json" -d '[{"labels":{"alertname":"myalert"}}]' localhost:9093/api/v1/alerts
```

### Adding alert rules
If you want to configure additional alerts, you need to add PrometheusRules resources. The example below will generate an alert if the ServiceMonitor with the name your-service-monitor-name has less than one target up.

Deploy those resources together with your application in e.g. a Helm chart so that they are bundled together and nicely versioned.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  labels:
    prometheus: kube-prometheus-stack-prometheus
    role: alert-rules
  name: your-application-name
spec:
  groups:
    - name: "your-application-name.rules"
      rules:
        - alert: PodDown
          for: 1m
          expr: sum(up{job="your-service-monitor-name"}) < 1 or absent(up{job="your-service-monitor-name"})
          annotations:
            message: The deployment has less than 1 pod running.
```

### Adding Grafana dashboards
If you want to deploy additional Grafana dashboards, we recommend adding a ConfigMap or Secret with the label grafana_dashboard=1.
The ConfigMap or Secret does not have to be in the same namespace as Grafana and can be deployed together with your application or service.

Although you can create dashboards in the Grafana user interface, they are not persistent in the default configuration. You can enable persistence, but then you also have to set the number of replicas to 1, which has the disadvantage of losing availability.
We strongly recommend not to do this or to set up an external database for storing dashboards as described in the Grafana documentation.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    grafana_dashboard: "1"
  name: new-dashboard-configmap
data:
  new-dashboard.json: |-
    {
      "id": null,
      "uid": "cLV5GDCkz",
      "title": "New dashboard",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
```

For more possibilities to deploy Grafana dashboards, have a look at the upstream helm chart repo.

### Make Grafana available via an ingress
Requirements:
- Ingress Controller, for example with the SysEleven Building Block
- Optional: cert-manager and external-dns

#### Configuring Ingress
Add to the matching values file:

```yaml
grafana:
  ingress:
    enabled: true
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-production"
    hosts:
      - grafana.example.com
    tls:
      - secretName: grafana.example.com-tls
        hosts:
          - grafana.example.com
  grafana.ini:
    server:
      root_url: https://grafana.example.com
```

## Monitoring
The building blocks (in this case kube-prometheus-stack itself) come with a set of predefined alert rules and Grafana dashboards. Alert rules and dashboards are synchronized by kubernetes-mixin.
This includes basic monitoring of the local Kubernetes cluster itself (e.g. resource limits/requirements, pod crash loops, API errors, ...).
The Ubiquity community may add additional alertrules and dashboards to each component.

## Additional alertrules
As an example, alert rules have been created in the directory kube-prometheus-stack-extension/templates/alerts.
Add your own alert rules to the existing files or create your own files in the directory with your alert rules.

Each alert rule should include a meaningful description as an annotation.

## Additional Grafana dashboards

### Alerts
An overview of firing prometheus alerts

### Cluster Capacity
An overview of the capacity of the local kubernetes cluster

## Scale prometheus persistent volumes
```bash
# Set replicas to 0 and PVC template to new value
kubectl edit prometheuses

# Patch PVC (e.g. 100Gi)
kubectl patch pvc prometheus-kube-prometheus-stack-prometheus-db-prometheus-kube-prometheus-stack-prometheus-0 --namespace syseleven-managed-kube-prometheus-stack -p '{"spec":{"resources":{"requests":{"storage":"100Gi"}}}}' --type=merge
kubectl patch pvc prometheus-kube-prometheus-stack-prometheus-db-prometheus-kube-prometheus-stack-prometheus-1 --namespace syseleven-managed-kube-prometheus-stack -p '{"spec":{"resources":{"requests":{"storage":"100Gi"}}}}' --type=merge

# Verify pending resize
kubectl describe pvc prometheus-kube-prometheus-stack-prometheus-db-prometheus-kube-prometheus-stack-prometheus-0

# Scale replicas back
kubectl edit prometheuses

# Commit the changes to the values*.yaml files.
Scaling setup
This building block consists of multiple components. Each of the components can and must be scaled individually.
```

## Scaling prometheus

### prometheus-operator
Usually should only be run with replicas=1
Requests/limits for CPU/memory can be adjusted

### prometheus-node-exporter
Runs as DaemonSet on each node, so no further replica scaling needed
Requests/limits for CPU/memory can be adjusted

### kube-state-metrics
Usually should only be run with replicas=1
Requests/limits for CPU/memory can be adjusted

### prometheus
Also see Prometheus High Availability for upstream documentation
Replicas can be increased, though each replica will be a dedicated prometheus that scrapes everything
Requests/limits for CPU/memory can be adjusted

### Scaling alertmanager
Also see Alertmanager High Availability for upstream documentation.

Replicas can be increased to achieve higher availability
New replicas will automatically join the alertmanager cluster
Requests/limits for CPU/memory can be adjusted

### Scaling grafana
Also see Set up Grafana for high availability for upstream documentation.

## Replicas
Dashboards - when increasing replicas for grafana it is important to think about where you want the dashboards/user configuration to come from. By default, we run with replicas=2 but do not save dashboards locally. So after respawn the dashboards are gone if they are not automated by supplying them as ConfigMap.

Sessions - when increasing replicas for grafana it is important to think about where you want to store user sessions. We configuresticky sessions by default, so each user is mapped to a specific replica until the replica is not available anymore.

Requests/limits for CPU/memory can be adjusted

## Release-Notes
Please find more infos on release notes and new features Release notes Prometheus-Stack

## Known Issues
For Kubernetes <= 1.24, when upgrading  there will be errors in the diff stage of the CI pipeline. This is due to a bug in the helm chart. The upgrade will still work, but the CI pipeline will fail. This will be fixed in the next release.