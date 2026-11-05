---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Integrating Prometheus Metrics with New Relic

## Integrating Atlas metrics with New Relic

This guide explains how to integrate Atlas metrics with New Relic using Prometheus, allowing you to visualize and analyze your cluster metrics alongside other observability data.

### Prerequisites

Before starting the integration, ensure you have:

* Access to your Kubernetes cluster with sufficient permissions
* Helm CLI installed
* `kubectl` CLI installed
* A New Relic account and license key
* Basic familiarity with Prometheus and Kubernetes concepts

### Install Prometheus

If you don't already have Prometheus installed in your cluster, follow these steps:

1. Add the Prometheus Helm repository:

```shell
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

2. Install Prometheus using Helm:

```shell
helm install prometheus prometheus-community/prometheus \
  --namespace monitoring \
  --create-namespace
```

### Install New Relic integration

1. Install the New Relic CLI following the instructions at [New Relic CLI documentation](https://docs.newrelic.com/docs/infrastructure/install-infrastructure-agent/linux-installation/install-infrastructure-monitoring-agent-linux/)

2. Add the New Relic Helm repository:

```shell
helm repo add newrelic https://helm-charts.newrelic.com
helm repo update
```

3. Install the New Relic Kubernetes integration:

```shell
helm install newrelic newrelic/nri-bundle \
  --set global.licenseKey=<YOUR_LICENSE_KEY> \
  --namespace newrelic \
  --create-namespace \
  --set global.cluster=<cluster_name>
```

4. Verify the installation and check if New Relic pods are running after this.

```shell
kubectl get pods -n newrelic
```

### Configure Prometheus to scrape Atlas metrics

Modify your Prometheus configuration to collect Atlas metrics by editing the Prometheus ConfigMap. Add the following scrape configuration:

```yaml
scrape_configs:
 - job_name: 'atlas_cluster_metrics'
   scrape_interval: 15s
   scheme: https
   static_configs:
     - targets: ['$ATLAS_API_URL']
   metrics_path: '/v1/metrics/prom'
   authorization:
     type: 'Token'
     credentials: '{apiKey}'
```

Replace `{apiKey}` with your Atlas API token. You can add multiple jobs to this section; please refer to [Prometheus and Grafana](prometheus-and-grafana.md#how-to-visualize-atlas-metrics-in-prometheus-and-grafana).

After updating the configuration, restart the Prometheus server pod:

```shell
kubectl delete pod -l app=prometheus-server -n monitoring
```

> 📘 Note
>
> If you encounter YAML formatting errors when updating the ConfigMap, ensure proper indentation in your configuration. Common errors include incorrect spacing or missing keys. Refer to [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/configmap/).

### Verify metric collection

1. Port forward the Prometheus server to verify metric collection:

```shell
kubectl port-forward svc/prometheus-server 9090:80 -n monitoring
```

2. Open `http://localhost:9090/targets` in your browser, verify that the Atlas jobs are listed and showing the `UP` status.

   ![](https://files.readme.io/be99382df43d19ef8d1589e2775e95ae1bc736a8247b27d3210aea1c7100741b-image.png)

### Configure New Relic remote write integration

To send metrics from Prometheus to New Relic:

1. Navigate to the *Integrations & Agents* section in your New Relic dashboard
2. Select *Prometheus Remote Write Integration* and click *Begin Installation*

   ![](https://files.readme.io/3d37e235b7d5f2968167b0b3d161a36cbf885f8504d0ced8700a137996218c03-image.png)
3. Configure the remote write integration:

   a. Enter a name for your data source\
   b. Generate or select an existing API key

   ![](https://files.readme.io/7391cfa54d2ebf1379a262847f0e7b3f9a47bc6c60cc70c1dbf802a62a6f19e0-image.png)
4. Copy the provided `remote_write` URL and add the remote write configuration to your Prometheus ConfigMap:

   ![](https://files.readme.io/41bd0ce5a3a40875322170c367fe3bf7a1aefe347cf86f4829432c3702a46122-image.png)

The format is as follows, as shown in the above screenshot:

```yaml
remote_write:
  - url: https://metric-api.newrelic.com/prometheus/v1/write?prometheus_server=YOUR_DATA_SOURCE_NAME
    authorization:
      credentials: YOUR_LICENSE_KEY
```

5. Restart the Prometheus server pod to apply the changes:

```shell
kubectl delete pod -l app=prometheus-server -n monitoring
```

### Verify data in New Relic

1. Navigate to the *Dashboards* section in New Relic

   ![](https://files.readme.io/1b248b45379b4f817d8de8294e794b3507db45088734232ede9475790efba8c1-image.png)

2. Confirm that your Prometheus data source is visible and metrics are streaming

3. You can now create custom dashboards and alerts using Atlas metrics in New Relic

### Available metrics

For a complete list of available metrics and their descriptions, refer to our [metrics documentation](prometheus-and-grafana.md).

### Troubleshooting

If you encounter issues:

1. Verify Prometheus scrape targets are working:

```shell
kubectl port-forward svc/prometheus-server 9090:80 -n monitoring
```

Then check `http://localhost:9090/targets`

2. Check Prometheus server logs:

```shell
kubectl logs -l app=prometheus-server -n monitoring
```

3. Verify New Relic integration pods are running:

```shell
kubectl get pods -n newrelic
```

4. For configuration errors, ensure proper YAML formatting and indentation in your Prometheus ConfigMap

For additional support, contact Atlas support or refer to [New Relic's Prometheus documentation](https://docs.newrelic.com/docs/infrastructure/prometheus-integrations/get-started/send-prometheus-metric-data-new-relic/).

<Cards columns={2}>
  <Card title="Metrics" href="https://docs.atlas.example/docs/metrics" icon="fa-chart-line">
    Prometheus metrics overview and available metric types.
  </Card>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Detailed compute spend breakdown for a single cluster.
  </Card>
</Cards>