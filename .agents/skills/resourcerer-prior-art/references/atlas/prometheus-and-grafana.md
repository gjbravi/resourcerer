---
updatedAt: 2026-09-22T08:46:05.000Z
---

# Prometheus and Grafana

Learn how to scrape Atlas metrics and visualize them in Grafana.

## Cluster metrics integration

Atlas delivers detailed metrics on your cluster utilization so that you can better understand your cloud infrastructure and ultimately reduce its cost. All metrics are scrapable, so you can scrape the data using Prometheus API and visualize it in Grafana or another tool of your choice.

As a result, you can draw on the cluster utilization and cost stats and include them effortlessly in your team's wider cloud monitoring and reporting efforts.

This guide outlines the metrics available in Atlas and describes the process of exporting them to Prometheus and Grafana step by step.

### How to visualize Atlas metrics in Prometheus and Grafana

#### Why use Atlas with Prometheus and Grafana

The combination of Prometheus and Grafana has become a common choice for DevOps and CloudOps teams, and this is for a reason.

The first provides a powerful querying language and gathers rich metrics, while the latter transforms these into meaningful visualizations. Both Prometheus and Grafana are compatible with most data source types.

#### How to connect Atlas with Prometheus and Grafana

#### Create your Atlas API key

Enter your cluster in the Atlas platform, click the API tab in the top menu, and generate a one-time token.

You must specify your key name and choose between read-only or full access. Then, copy and paste it into the respective place in the above code and execute.

You can also use this key to access the Atlas API in tools like Swagger UI.

#### Call the Atlas API

Open your Prometheus scraper config in your favorite tool and add scraping for Atlas metrics. Our API provides separate endpoints for scraping the cluster, node, workload, and allocation group-level metrics. Here is an example of how to configure scraping all of them:

> 📘 Note
>
> Replace `$ATLAS_API_URL` with your regional API endpoint (for example, `api.atlas.example` for US). See [API access](api-access.md) for the full list of regional endpoints.

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
 - job_name: 'atlas_node_template_metrics'
   scrape_interval: 1m
   scheme: https
   static_configs:
     - targets: ['$ATLAS_API_URL']
   metrics_path: '/v1/metrics/nodes'
   authorization:
     type: 'Token'
     credentials: '{apiKey}'
 - job_name: 'atlas_workload_metrics'
   scrape_interval: 15s
   scheme: https
   static_configs:
     - targets: ['$ATLAS_API_URL']
   metrics_path: '/v1/metrics/workloads'
   authorization:
     type: 'Token'
     credentials: '{apiKey}'
 - job_name: 'atlas_allocation_group_metrics'
   scrape_interval: 15s
   scheme: https
   static_configs:
     - targets: ['$ATLAS_API_URL']
   metrics_path: '/v1/metrics/allocation-groups'
   authorization:
     type: 'Token'
     credentials: '{apiKey}'
 - job_name: 'atlas_workload_network_metrics'
   scrape_interval: 15s
   scheme: https
   static_configs:
     - targets: ['$ATLAS_API_URL']
   metrics_path: '/v1/metrics/workload-network'
   authorization:
     type: 'Token'
     credentials: '{apiKey}'
```

Please replace **\{apiKey}** with the token created in step 1.

To limit the amount of returned data, the Node, Workload, Allocation group, and Workload network metric endpoints support filtering by `clusterID` (see [API reference](https://docs.atlas.example/reference/reportmetricsapi_getworkloadprommetrics)).

#### Specify your data source in Grafana

Open Grafana, head to the Configuration tab, and click on Data Sources.

When you select the Add data source option, you'll see a list of all supported data sources. From here, choose Prometheus and insert all required details, including HTTP, Auth, and more.

![](https://files.readme.io/988967d-1_grafana_data_source.png "1_grafana_data_source.png")

After you specify your data source, you can go to Explore, select your data source by name, and start typing the metric name for auto-completion.

![](https://files.readme.io/791b322-2_grafana_metric_name.png "2_grafana_metric_name.png")

![](https://files.readme.io/8b3d856-3_grafana_metric_name_view.png "3_grafana_metric_name_view.png")

#### Create a dashboard in Grafana

Click on the Dashboards tab in Grafana's main menu and select the Browse option. That's where you'll see the button to start a new dashboard. Give it a meaningful name and set the main options.

For more information, refer to [Grafana's documentation](https://grafana.com/docs/grafana/latest/dashboards/) or check this [list of best practices for creating dashboards](https://grafana.com/docs/grafana/latest/best-practices/best-practices-for-creating-dashboards/).

You can also start by importing the Example dashboard JSON provided below.

#### Add and format your metrics

Now, it's time to start populating your dashboard with data.

Add a new panel and scroll down to its bottom to ensure that the data source is set to Prometheus. Then, start typing the name of the required metric in the metric browser box, and it will appear on the screen.

Common metrics include the requested vs. provisioned CPUs and memory, and the monthly cost of your cluster. You can also expand the metrics presented in your dashboard by importing data from JSON files.

![](https://files.readme.io/4dbf12d-4_grafana_panel.png)

Use the panel on the right to specify your stat title, legend, visualization styles, and other values to help you ensure the report makes the most sense to your team.

You can then expand your dashboard with additional features, including [annotations](https://grafana.com/docs/grafana/latest/dashboards/annotations/) and [alerts](https://grafana.com/docs/grafana/latest/alerting/).

#### Example Grafana dashboard

Here's an example dashboard displaying Atlas data.

![](https://files.readme.io/07e825b-Screenshot_2024-04-26_at_16.18.48.png)

You can get the [code here](https://docs.atlas.example/atlas/docs/atlas-metrics-grafana-dashboard-example-json).

### Atlas metrics

> 📘 Note
>
> Label `atlas_node_type` is deprecated, so instead of it, please use `atlas_node_lifecycle`.

| Name                                                | Type    | Description                                                                                                                   | Action                                        |
| :-------------------------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| `atlas_autoscaler_agent_snapshots_received_total`  | Counter | The Atlas Autoscaler agent snapshots received total.                                                                        | Check if the Agent is running in the cluster. |
| `atlas_autoscaler_agent_snapshots_processed_total` | Counter | The Atlas Autoscaler agent snapshots processed total.                                                                       | Contact Atlas support.                      |
| `atlas_cluster_total_cost_hourly`                  | Gauge   | Cluster total hourly cost.                                                                                                    |                                               |
| `atlas_cluster_compute_cost_hourly`                | Gauge   | Cluster compute cost. Has a `lifecycle` dimension which can be summed up to a total cost: `[on_demand, spot_fallback, spot]`. |                                               |
| `atlas_cluster_total_cost_per_cpu_hourly`          | Gauge   | Normalized cost per CPU.                                                                                                      |                                               |
| `atlas_cluster_compute_cost_per_cpu_hourly`        | Gauge   | Normalized cost per CPU. Has a `lifecycle` dimension, similar to `atlas_cluster_compute_cost_hourly`.                        |                                               |
| `atlas_cluster_allocatable_cpu_cores`              | Gauge   | Cluster allocatable CPU cores.                                                                                                |                                               |
| `atlas_cluster_allocatable_memory_bytes`           | Gauge   | Cluster allocatable memory.                                                                                                   |                                               |
| `atlas_cluster_provisioned_cpu_cores`              | Gauge   | Cluster provisioned CPU cores.                                                                                                |                                               |
| `atlas_cluster_provisioned_memory_bytes`           | Gauge   | Cluster provisioned memory.                                                                                                   |                                               |
| `atlas_cluster_requests_cpu_cores`                 | Gauge   | Cluster requested CPU cores.                                                                                                  |                                               |
| `atlas_cluster_used_cpu_cores`                     | Gauge   | Cluster used CPU cores.                                                                                                       |                                               |
| `atlas_cluster_used_memory_bytes`                  | Gauge   | Cluster used memory.                                                                                                          |                                               |
| `atlas_cluster_requests_memory_bytes`              | Gauge   | Cluster requested memory.                                                                                                     |                                               |
| `atlas_cluster_node_count`                         | Gauge   | Cluster nodes count.                                                                                                          |                                               |
| `atlas_cluster_pods_count`                         | Gauge   | Cluster pods count.                                                                                                           |                                               |
| `atlas_cluster_unschedulable_pods_count`           | Gauge   | Cluster unschedulable pods count.                                                                                             |                                               |
| `atlas_evictor_node_target_count`                  | Gauge   | Atlas Evictor targeted nodes count.                                                                                         |                                               |
| `atlas_evictor_pod_target_count`                   | Gauge   | Atlas Evictor targeted pods count.                                                                                          |                                               |
| `atlas_cluster_provisioned_storage_bytes`          | Gauge   | Cluster-provisioned storage. Currently available only for GCP.                                                                |                                               |
| `atlas_cluster_requests_storage_bytes`             | Gauge   | Cluster requested storage. Currently available only for GCP.                                                                  |                                               |
| `atlas_cluster_storage_cost_hourly`                | Gauge   | Cluster storage hourly cost. Currently available only for GCP.                                                                |                                               |

#### Query examples

**Cost per cluster:**

```javascript PromQL
sum(atlas_cluster_total_cost_hourly{}) by (atlas_cluster)
```

**Compute the cost of Spot Instances of a specific cluster:**

```javascript PromQL
atlas_cluster_compute_cost_hourly{atlas_cluster="$cluster", lifecycle="spot"}
```

**Received snapshots count:**

```javascript PromQL
sum(increase(atlas_autoscaler_agent_snapshots_received_total{atlas_cluster="$cluster"}[5m]))
```

**Alert on missing snapshots:**

```javascript PromQL
absent_over_time(atlas_autoscaler_agent_snapshots_received_total{atlas_cluster="$cluster"}[5m])
```

**Total cluster cost over the last 7 days:**

```javascript PromQL
sum_over_time(atlas_cluster_total_cost_hourly{atlas_cluster="$cluster"}[7d:1m]) / 60
```

**Get `atlas_node_lifecycle`(`on_demand`, `spot`, `spot_fallback`) of running nodes in the cluster:**

```javascript PromQL
sum(atlas_cluster_node_count{atlas_cluster="$cluster"}) by (atlas_node_lifecycle)
```

**Get CPU cores provisioned for `spot_fallback` nodes:**

```javascript PromQL
atlas_cluster_provisioned_cpu_cores{atlas_node_lifecycle="spot_fallback"}
```

> 📘 Note
>
> Replace `$cluster` with existing `atlas_cluster` label value.

### Node metrics

Node metrics have common labels: `cluster_id, node_id, node_name, node_template, managed_by`

| Name                                        | Type  | Description                                               | Note                                                                                                           |
| ------------------------------------------- | ----- | --------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| atlas\_node\_spot\_interruption            | Gauge | Nodes that have been interrupted over the last 60 seconds | will have `managed_by="atlas"` label only when Atlas manages the node, otherwise, this label will be absent |
| atlas\_node\_requested\_cpu\_cores         | Gauge | Requested CPU cores by node                               | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot]                                               |
| atlas\_node\_requested\_ram\_bytes         | Gauge | Requested RAM bytes by node                               | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_provisioned\_cpu\_cores       | Gauge | Provisioned CPU cores by node                             | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_provisioned\_ram\_bytes       | Gauge | Provisioned RAM bytes by node                             | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_used\_cpu\_cores              | Gauge | Utilized CPU cores by node                                | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_used\_ram\_bytes              | Gauge | Utilized RAM bytes by node                                | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_cost\_cpu\_hourly             | Gauge | Hourly cost of CPU by node                                | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_cost\_ram\_gib\_hourly        | Gauge | Hourly cost of RAM by node                                | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_overprovisioned\_cpu\_percent | Gauge | The overprovisioned percentage of CPU by node             | Has a `lifecycle` dimension                                                                                    |
| atlas\_node\_overprovisioned\_ram\_percent | Gauge | Overprovisioned percent of RAM by node                    | Has a `lifecycle` dimension                                                                                    |

Possible `managed_by` values are: `atlas`, `karpenter`, `eks`, `gke`, `aks`, `kops`, `openshift`

#### Query example

Get average `on_demand` CPU overprovisioned percentage per node provider:

```jsx PromQL
avg(atlas_node_overprovisioned_cpu_percent{lifecycle="on_demand"}) by (managed_by)
```

Get CPU cores provisioned for the default Atlas node template and aggregated by lifecycle:

```jsx PromQL
sum(atlas_node_provisioned_cpu_cores{node_template="default-by-atlas"}) by (lifecycle)
```

Get a count of node spot interruptions aggregated by node templates:

```jsx PromQL
sum(atlas_node_spot_interruption{}) by (node_template)
```

### Workload metrics

By default, when using the `v1/metrics/workloads` [endpoint](https://docs.atlas.example/reference/reportmetricsapi_getworkloadprommetrics), it returns common labels such as `cluster_id, namespace, workload_name, workload_type`. However, you can also specify additional labels to be returned by setting workload labels in the `additionalLabels` query parameter.

> 📘 Note
>
> There is a special `workload_optimization_enabled` label, which can be specified in the `additionalLabels` query parameter. This will return a label indicating whether <Anchor label="optimization" target="_blank" href="doc:workload-autoscaling-overview">optimization</Anchor> is enabled for a particular workload.

| Name                                                   | Type  | Description                                | Note                                                             |
| ------------------------------------------------------ | ----- | ------------------------------------------ | ---------------------------------------------------------------- |
| atlas\_workload\_cost\_hourly                         | Gauge | Workload cost per hour                     | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_workload\_pod\_count                           | Gauge | Number of pods                             | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_workload\_requested\_cpu\_cores                | Gauge | Requested CPU cores                        | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_workload\_requested\_memory\_bytes             | Gauge | Requested memory bytes                     | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_workload\_persistent\_volume\_cost\_hourly     | Gauge | Workload persistent volume cost per hour   |                                                                  |
| atlas\_workload\_requested\_persistent\_volume\_bytes | Gauge | Workload requested persistent volume bytes |                                                                  |

#### Query example

Get workload costs running on `spot` instances for a specific cluster:

```jsx
atlas_workload_cost_hourly{cluster_id="$clusterId", lifecycle="spot"}
```

Get average workload costs running on `on_demand` instances over 1h aggregated by namespace for specific cluster:

```jsx
sum(avg_over_time(atlas_workload_cost_hourly{cluster_id="$clusterId", lifecycle="on_demand"}[1h])) by (namespace)
```

### Allocation group metrics

| Name                                                | Type  | Description            | Note                                                             |
| --------------------------------------------------- | ----- | ---------------------- | ---------------------------------------------------------------- |
| atlas\_allocation\_group\_compute\_cost\_hourly    | Gauge | Compute cost per hour  | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_allocation\_group\_cpu\_cost\_hourly        | Gauge | CPU cost per hour      | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_allocation\_group\_memory\_cost\_hourly     | Gauge | Memory cost per hour   | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_allocation\_group\_requested\_cpu\_cores    | Gauge | Requested CPU cores    | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |
| atlas\_allocation\_group\_requested\_memory\_bytes | Gauge | Requested memory bytes | Has a `lifecycle` dimension: \[on\_demand, spot\_fallback, spot] |

### Workload network metrics

| Name                                           | Type  | Description                        | Note |
| ---------------------------------------------- | ----- | ---------------------------------- | ---- |
| atlas\_workload\_network\_bytes               | Gauge | Workload network bytes             |      |
| atlas\_workload\_to\_workload\_traffic\_bytes | Gauge | Workload to workload traffic bytes |      |

<Cards columns={2}>
  <Card title="Integrating metrics with New Relic" href="https://docs.atlas.example/docs/observability-metrics-integrating-prometheus-metrics-with-new-relic" icon="fa-chart-mixed">
    Export Atlas metrics to New Relic for unified observability.
  </Card>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Detailed compute spend breakdown for a single cluster.
  </Card>
</Cards>