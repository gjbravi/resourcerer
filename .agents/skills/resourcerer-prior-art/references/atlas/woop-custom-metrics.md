---
updatedAt: 2026-10-14T07:12:30.000Z
---

# Custom metrics

Custom metrics allow the Workload Autoscaler to scale applications based on signals that Kubernetes does not natively expose. The primary use case is connecting a Prometheus data source so the Workload Autoscaler can collect JVM heap metrics and generate heap-aware memory recommendations for Java workloads.

> 📘 Note
>
> Other features that rely on custom metrics infrastructure — such as PSI (Pressure Stall Information) scaling and startup failure detection — work automatically once the Workload Autoscaler Exporter is installed. They use system-defined data sources and do not require you to connect a Prometheus data source. See [System-defined data sources](#system-defined-data-sources) for details.

## Prerequisites

The following components must be installed in your cluster:

| Component                             | Minimum version | Description                                                                                                                             |
| ------------------------------------- | --------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `atlas-workload-autoscaler-exporter` | `0.85.0`        | Collects custom metrics from your cluster and forwards them to the Workload Autoscaler. Installed by default during cluster onboarding. |
| `atlas-agent`                        | `0.113.0`       | Cluster agent that communicates with the Atlas control plane.                                                                         |

| Workload kind | Custom metrics-based recommendations                                                                                   |
| ------------- | ---------------------------------------------------------------------------------------------------------------------- |
| DaemonSet     | Not supported. For the supported alternative, see [Node-aware DaemonSet sizing](woop-node-aware-daemonset-sizing.md). |

For upgrade instructions, see [Workload Autoscaler configuration](workload-autoscaling-configuration.md).

<Accordion title="Verify installed component versions" icon="fa-circle-check">

Run the following commands to check that the required components are installed and meet the minimum version requirements:

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler-exporter
```

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler
```

```shell
kubectl get deploy -n atlas-agent atlas-agent \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

All pods should be in `Running` status.

</Accordion>

<Accordion title="Install the exporter manually" icon="fa-wrench">

If the exporter was not installed during cluster onboarding, you can install it manually. First, check whether the exporter is already running:

```shell
kubectl get deployment atlas-workload-autoscaler-exporter -n atlas-agent
```

If the deployment exists and is running, no action is needed. If it is not found, install it using one of the methods below.

<Tabs>
  <Tab title="Helm">

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm install atlas-workload-autoscaler-exporter atlas-helm/atlas-workload-autoscaler-exporter -n atlas-agent
```

  </Tab>
  <Tab title="Terraform">

Set `install_workload_autoscaler_exporter = true` in your Atlas Terraform module configuration.

  </Tab>
</Tabs>

</Accordion>

## Configure a data source

A Prometheus data source connects the Workload Autoscaler to a Prometheus server running in your cluster, enabling it to collect application and runtime metrics for scaling decisions.

### Default Exporter

When the exporter is installed, the Workload Autoscaler automatically creates a **Default Exporter** data source. This data source points to the exporter's own Prometheus instance (`http://atlas-workload-autoscaler-exporter-prometheus.atlas-agent.svc:9090`) and provides the system-defined metrics that power PSI scaling and startup failure detection.

For JVM heap scaling, you need a Prometheus data source that scrapes JVM metrics from your Java workloads. If your existing Prometheus server already scrapes JVM metrics, connect it as an additional data source. If you do not have Prometheus scraping JVM metrics, either enable [auto-instrumentation](woop-jvm-optimization.md#auto-instrumentation) in your scaling policy or [scrape your workloads' existing metrics endpoints](woop-jvm-optimization.md#scrape-existing-jvm-metrics-endpoints) — the Workload Autoscaler routes metrics through the Default Exporter automatically in both cases, with no external data source to connect.

JVM optimization can also be enabled on individual workloads using annotations:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    jvm:
      enabled: true
      autoInstrument: true
```

See [JVM workload optimization](woop-jvm-optimization.md#enable-via-annotations) for details, or the [annotations reference](workload-autoscaler-annotations-reference.md) for the full schema.

### Connect a Prometheus data source

1. Navigate to **Workload Autoscaler** > **Custom metrics**
2. Click **Add data source**
3. Enter a **Data source name** (for example, `prometheus-prod-eu`)
4. Enter the **Prometheus server URL** — the in-cluster service URL of your Prometheus server (for example, `http://prometheus.production.svc.cluster.local:9090`)

   ![](https://files.readme.io/872179417477a4c9888e2a1100bb358d05b180e08926240c5ba35a696c769ec9-image.png)
5. Click **Connect**
6. Wait a few minutes for the data source to appear in the cluster

   ![](https://files.readme.io/1d6bf61c02f1d877514d02ab1bd3484e68388347f61319693e0d3b20df44bbb3-image.png)

### Metrics explorer

The Custom metrics page includes a **Metrics explorer** where you can select a JVM workload and time range to inspect the JVM metrics being collected from your data sources. Use this to verify that JVM metrics (heap usage, GC behavior, thread count) are flowing correctly after connecting a data source or enabling auto-instrumentation.

![](https://files.readme.io/9653fb8b8f32f6586d5585ca74d75e43d7cf6b054748fd419acfe42bcd7e4bc7-image.png)

## Verify data source status

After connecting a data source, verify that it is healthy. The Custom metrics page shows the connection status for each data source. You can also check the custom resource status in the cluster:

```shell
kubectl get custommetricsexporterconfigs
```

Expected output when data sources are healthy:

```shell
NAME                      HEALTHY   READY   AGE
atlas-data-source-name   True      1/1     21d
exporter-config           True      3/3     8m51s
```

> 📘 Note
>
> It may take a few minutes for the custom resource to be created in the cluster after connecting a data source. If it does not appear immediately, wait and check again.

## System-defined data sources

The Workload Autoscaler provides built-in data sources that are installed alongside `atlas-workload-autoscaler-exporter` by default. These data sources require no manual configuration and work automatically once the exporter is installed:

* **node-probes** — Provides startup-related metrics used by [startup failure detection](workload-autoscaling-overview.md#startup-failure-detection). No additional setup needed.
* **node-psi** — Provides Pressure Stall Information metrics used by [stall detection](woop-configuration-settings.md#stall-detection). Requires Kubernetes 1.34+. No additional setup needed.

## Troubleshooting

### Data source not appearing

If the custom resource does not appear after connecting a data source:

1. Verify that `atlas-workload-autoscaler-exporter` is running:
   ```shell
   kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler-exporter
   ```
2. Check exporter logs for errors:
   ```shell
   kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler-exporter --tail=50
   ```
3. Wait 2-3 minutes and check again — resource creation is asynchronous

### Data source shows unhealthy

If a data source reports an unhealthy status, inspect the custom resource for details:

```shell
kubectl describe custommetricsexporterconfigs <config-name>
```

The status includes a reason for the unhealthy state. For example, a cluster version incompatibility:

```yaml
Status:
  Datasource Statuses:
    Type:     NodeWorkload
    Message:  Datasource disabled: node workload data source
              not supported on cluster version 1.33.6+k3s1
    Name:     node-psi
    State:    DatasourceDisabled
```

Address the underlying issue indicated in the status message, or contact Atlas support if the error is unclear.

## Next steps

<Cards>
  <Card title="JVM workload optimization" href="https://docs.atlas.example/docs/woop-jvm-optimization" icon="fa-coffee">
    Optimize Java workloads with heap-based scaling and auto-instrumentation.
  </Card>

  <Card title="Scaling policies" href="https://docs.atlas.example/docs/woop-scaling-policies" icon="fa-sliders">
    Create and manage scaling policies that control how the Workload Autoscaler optimizes your workloads.
  </Card>

  <Card title="Annotations reference" href="https://docs.atlas.example/docs/workload-autoscaler-annotations-reference" icon="fa-code">
    Full annotation schema for per-workload JVM settings and all other configuration overrides.
  </Card>
</Cards>

<br />