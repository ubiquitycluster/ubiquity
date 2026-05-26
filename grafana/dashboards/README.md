# Grafana Dashboards

This directory contains Grafana dashboards as code.
Dashboards are in JSON format and can be imported into Grafana.

## Usage

Import via Grafana UI or use kubectl:
​```
kubectl create configmap ubiquity-dashboard \
  --from-file=grafana/dashboards/cluster.json \
  -n monitoring
​```