---
updatedAt: 2026-09-02T10:32:13.000Z
---

# Reliability metrics

Monitor request rate, error rate, and latency for every service in your cluster, with no changes to your application code.

Reliability metrics gives you continuous visibility into the health of services running in your Kubernetes cluster. Once installed, it collects request rate, error rate, and latency data for all your workloads automatically. No changes to your application code, no sidecars, and no instrumentation libraries required.

The data appears in the Atlas console under **Cost Report → Reliability**, alongside your cost data, so you can see which workloads are expensive *and* unreliable at the same time.

## How reliability metrics works

A collector agent runs on each node in your cluster and observes HTTP and gRPC network traffic at the network level. The agent inspects request and response traffic without intercepting it. Your application code and behavior are not affected.

The collector ignores health-check and infrastructure traffic automatically, so routes like `/healthz`, `/readyz`, `/metrics`, and `/actuator/*` do not inflate your request rate or error rate figures.

Measurements are aggregated within the cluster and exported to Atlas, where they appear in the Reliability views within a few minutes.

> **Storage in your cluster**
>
> Reliability metrics provisions a [ClickHouse](https://clickhouse.com/) time-series database inside your cluster for temporary metric storage. By default, 10 GiB of persistent storage is allocated. `atlasctl` sizes both the collector agent and this database automatically based on your cluster. See [how atlasctl sizes components](reliability-metrics-reference#how-atlasctl-sizes-components) for details.

## Supported protocols

| Protocol                 | Status                                                                   |
| ------------------------ | ------------------------------------------------------------------------ |
| HTTP/1.x                 | Available (request rate, error rate, and latency)                        |
| HTTP/2                   | Available (request rate, error rate, and latency)                        |
| gRPC                     | Available (request rate, error rate, and latency; gRPC runs over HTTP/2) |
| Database queries         | Collected but not yet shown in the console. Coming in a future release.  |
| Message queue operations | Collected but not yet shown in the console. Coming in a future release.  |

## Understanding the metrics

### The RED method

Reliability metrics follows the RED method: **Rate** (how many requests per second a service handles), **Errors** (how many are failing), and **Duration** (how long they take to complete). These three signals together are enough to detect most service health problems and are the standard starting point for reliability monitoring.

### Metrics available in the console

| Metric       | Unit  | What it tells you                                                                           |
| ------------ | ----- | ------------------------------------------------------------------------------------------- |
| Request Rate | req/s | How many requests per second the workload is receiving                                      |
| Error Rate   | req/s | How many requests per second are failing                                                    |
| Error Rate   | %     | What share of all requests are failing                                                      |
| Latency P50  | ms    | The response time that 50% of requests complete within (the typical experience)             |
| Latency P95  | ms    | The response time that 95% of requests complete within (what most users experience at load) |
| Latency P99  | ms    | The response time that 99% of requests complete within (the worst tail-end performance)     |
| Availability | %     | Derived as `100 minus error rate %`. Shown in the daily details table.                      |

### Unhealthy workloads

A workload is classified as **unhealthy** when its error rate exceeds 5% over the selected time range. The Reliability tab shows a summary count of how many workloads in your cluster meet this threshold.

### Data freshness

Metrics appear in the console within approximately 2–3 minutes of the corresponding requests occurring. The most recent 1–2 minutes of data may not yet be reflected at any given time.

## Navigating the Reliability tab

### Cluster Reliability view

Navigate to **Cost Report → Reliability** to see your cluster's overall health. This view shows:

* **Summary cards**: current request rate, error rate, and P95 latency across the entire cluster, each with a trend indicator comparing the current period to the previous one
* **Traffic chart**: time series of request rate and error rate
* **Latency chart**: time series of P50, P95, and P99 latency
* **Daily details table**: one row per day in the selected range, with availability %, request rate, error rate %, P50, P95, and P99 latency

If reliability metrics is not yet installed, this view shows an **Enable reliability metrics** button instead.

### Workload Reliability view

Navigate to **Cost Report → Workloads → \[workload name] → Reliability** to see the same charts and table scoped to a single workload.

### Workloads list with reliability columns

The **Reliability** tab in the workloads list shows all workloads that have collected data. Columns include request rate, error rate, P95 latency, and cost, letting you identify workloads that are both expensive and unhealthy in one view.

### Time range options

The Reliability views use shorter preset ranges than the general cost report, suited for operational monitoring:

* Last 15 minutes
* Last 30 minutes
* Last hour
* Last 6 hours
* Last 24 hours

## Ask the Atlas agent

If your organization uses the Atlas agent, you can ask natural language questions about workload reliability directly in the console. Examples:

* "What is the error rate of the payments service in production?"
* "Is the checkout deployment experiencing high latency?"
* "Show me the request rate for the api-gateway over the last hour."

The agent queries the same data shown in the Reliability views and returns a plain-language answer.

***

## Enable reliability metrics

The recommended way to enable reliability metrics is with `atlasctl`, which auto-sizes the collector and metrics database for your cluster. If you prefer to manage the installation yourself with Helm, see [Helm charts](reliability-metrics-reference#helm-charts) in the reference.

### Prerequisites

Before enabling reliability metrics, confirm:

* You have `kubectl` configured for the target cluster with cluster administrator permissions
* You have a Atlas API token
* Your cluster has a default storage class (required for the metrics database)

You can install reliability metrics with `atlasctl` (recommended) or directly with the Atlas Helm chart. The steps below cover the `atlasctl` path. For Helm, including GitOps, see [Helm charts](reliability-metrics-reference#helm-charts) in the reference.

### Install and authenticate atlasctl

Install atlasctl following the [atlasctl installation guide](connect-with-atlasctl.md).

Once authenticated, enable reliability metrics on your cluster.

### Run the install command

Run `atlasctl cluster connect` with the `--reliability-metrics` flag. This command handles the full installation: it inspects your cluster's workload density, sizes the collector agent and metrics database accordingly, and waits until it confirms data is flowing before finishing.

**Interactive (prompts you to select features):**

```bash
atlasctl cluster connect
```

When the **Feature Selection** prompt appears, choose **Reliability Metrics** from the list.

**Non-interactive (for scripts and CI):**

```bash
atlasctl cluster connect \
  --non-interactive \
  --reliability-metrics
```

**Combined with other Atlas features:**

```bash
atlasctl cluster connect \
  --non-interactive \
  --cluster-name my-cluster \
  --reliability-metrics \
  --cluster-optimization \
  --workload-autoscaler
```

**Preview what will be installed without making changes:**

```bash
atlasctl cluster connect --dry-run --reliability-metrics
```

After the install finishes, `atlasctl` confirms the full pipeline is running. If it times out, it prints a verification command you can run to check the status manually.

### Available atlasctl flags for reliability metrics

Use these flags with `atlasctl cluster connect` to control the installation:

| Flag                    | Description                                                |
| ----------------------- | ---------------------------------------------------------- |
| `--reliability-metrics` | Enable reliability metrics. Required.                      |
| `--non-interactive`     | Skip all interactive prompts. Use this in CI/CD pipelines. |
| `--cluster-name <name>` | Set the name for the cluster being connected.              |
| `--dry-run`             | Show what would be installed without making any changes.   |

### Enable via the console (for already-connected clusters)

If your cluster is already connected to Atlas, you can enable reliability metrics without running `atlasctl cluster connect` again:

1. Navigate to **Cost Report → Reliability** for the cluster.
2. Click **Enable reliability metrics**.
3. A setup dialog opens with your API token and cluster ID pre-filled.
4. Copy and run the `atlasctl` command shown in the dialog.

***

## Upgrade reliability metrics

To upgrade all Atlas components in your cluster, including reliability metrics:

```bash
atlasctl atlasware upgrade
```

This checks for newer versions of all installed components and upgrades them, preserving your existing configuration. Any new data migrations run automatically when the collector restarts. Use `--non-interactive` in CI pipelines:

```bash
atlasctl atlasware upgrade --non-interactive
```

***

## Limitations

**Default ports only.** The collector monitors traffic on ports **8080, 8443, 8090, and 6379** by default. When installing with Helm, you can add ports by setting `autoscaler.atlas-sentinel.agent.reliabilityMetrics.obi.openPorts` (a comma-separated list). See [Helm charts](reliability-metrics-reference#helm-charts) in the reference.

**Server-side metrics only.** The collector measures each request from the receiving service's perspective. This intentionally avoids double-counting when the client and server of a request are both running in the same cluster. Outbound call metrics from client services are not shown separately.

**Database and messaging metrics not yet in the console.** Database query duration and message queue operation metrics are collected by the agent but are not yet visible in the Atlas console. They will appear in a future release.

**Data freshness.** The most recent approximately 2 minutes of data are not yet available in the console at any given time. This is by design to ensure accuracy before data is displayed.

**Latency percentile accuracy.** Latency percentiles (P50, P95, P99) use linear interpolation within measurement buckets. For workloads with unusual latency distributions (such as strongly bimodal response times), percentile values near histogram bucket boundaries may be approximate.

**ClickHouse operator requirement.** Installing reliability metrics provisions a ClickHouse database in your cluster using the Altinity ClickHouse Operator. If your cluster restricts Custom Resource Definition permissions, you may need to install the operator CRDs manually before running the installer. `atlasctl` detects an existing operator automatically and skips that step if one is already running.

***

## Troubleshooting

### The Reliability tab shows "Enable reliability metrics"

Reliability metrics is not installed on this cluster. Click the button to open the setup dialog and follow the installation steps.

### The Reliability tab shows "No data" after installation

The collector may still be starting up. Wait 2–5 minutes after installation and refresh the page.

If the issue persists, confirm that your services are running and receiving traffic on ports 8080, 8443, 8090, or 6379. Traffic on other ports is not collected.

### Data appears but stops updating

Check that the collector DaemonSet is healthy on all nodes:

```bash
kubectl get daemonset atlas-sentinel-agent -n atlas-agent
```

The `READY` count should match the `DESIRED` count. If nodes are missing, check the pod logs for errors:

```bash
kubectl logs -n atlas-agent -l app=atlas-sentinel-agent --tail=50
```

### A service is not appearing in the Reliability workloads list

The service is likely listening on a port outside the default set (8080, 8443, 8090, 6379). When installing with Helm, you can add ports by setting `autoscaler.atlas-sentinel.agent.reliabilityMetrics.obi.openPorts` (a comma-separated list). See [Helm charts](reliability-metrics-reference#helm-charts) in the reference.

<Cards columns={2}>
  <Card title="Reference" href="https://docs.atlas.example/docs/reliability-metrics-reference" icon="fa-book">
    Metric definitions, sizing details, and port configuration
  </Card>
  <Card title="Cluster score" href="https://docs.atlas.example/docs/cluster-score" icon="fa-gauge-high">
    Comprehensive rating of your cluster's optimization and efficiency
  </Card>
</Cards>