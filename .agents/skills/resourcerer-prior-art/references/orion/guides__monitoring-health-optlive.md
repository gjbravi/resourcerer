# Monitoring the health of Orion

> Monitor Orion health with status metrics, logs, and component alerts


## Health and status metrics

Orion components expose health and status metrics for monitoring Orion. Health and status metrics differ from the workload metrics Orion needs to create request and limit recommendations. See the [Security FAQ][sec-faq] for details on workload metrics.

Orion installs from a single `orion` Helm chart and runs as separate components in your cluster. Three components deploy as their own Kubernetes Deployment:

- `orion-cluster-agent` — the cluster telemetry agent that processes workloads and generates recommendations
- `orion-forwarder` — forwards cluster metrics to Orion
- `orion-applier` — applies recommendations to workloads

One optional component deploys as a DaemonSet:

- `orion-node-agent` — collects GPU metrics on GPU nodes. Disabled by default; see [Node Agent configuration][node-agent-config].

| `orion-cluster-agent` metrics                  | Description                                                                  |
| ------------------------------------------- | ---------------------------------------------------------------------------- |
| `orion_workloads_failed_processed_total`       | Total number of workloads for which recommendations failed to generate       |
| `orion_workloads_inactive_deletion_total`      | Total number of inactive workloads deleted from Orion                   |
| `orion_workloads_inactive_failed_total`        | Total number of inactive workloads that could not be deleted from Orion |
| `orion_workloads_processed_namespace`          | Number of workloads processed per namespace                                  |
| `orion_workloads_processed_namespace_resource` | Number of workloads processed per namespace per resource                     |
| `orion_workloads_processed_resource`           | Number of workloads processed per resource                                   |
| `orion_workloads_processed_total`              | Number of workloads processed                                                |

| `orion-applier` metrics           | Description                                                                              |
| -------------------------------------- | ---------------------------------------------------------------------------------------- |
| `orion_applier_api_disconnections_total`  | Total number of times `orion-applier` lost connection to the Orion API         |
| `orion_applier_patches_expired_total`     | Total number of Orion-generated patches expired                                     |
| `orion_applier_patches_failed_total`      | Total number of Orion-generated patches failed to apply                             |
| `orion_applier_patches_processed_total`   | Total number of Orion-generated patches                                             |
| `orion_applier_patches_rolled_back_total` | Total number of Orion-generated patches rolled back to the previous resource values |

Beyond these Orion metrics, `orion-cluster-agent` and `orion-forwarder` expose standard controller-runtime and Prometheus-agent metrics, such as `workqueue_depth` and the `prometheus_*` series. See [Alerting on component health](#alerting-on-component-health).

You can scrape these metrics in OpenMetrics format from the following component pods and endpoints:

| Component             | Pod selector labels                                                              | Port | Endpoint   |
| --------------------- | -------------------------------------------------------------------------------- | ---- | ---------- |
| `orion-cluster-agent`    | `app.kubernetes.io/name: orion` and `app.kubernetes.io/component: cluster-agent`    | 8080 | `/metrics` |
| `orion-applier`  | `app.kubernetes.io/name: orion` and `app.kubernetes.io/component: applier`  | 8080 | `/metrics` |
| `orion-forwarder` | `app.kubernetes.io/name: orion` and `app.kubernetes.io/component: forwarder` | 9090 | `/metrics` |

Output and errors from each component are available in its pod logs. The Troubleshooting topic has steps for viewing the [`orion-cluster-agent` logs][trbl-agent-depl-logs] and [`orion-applier` logs][trbl-applier-depl-logs], plus the [warnings and errors][warnings-errors] to watch for.

## Alerting on component health

For a baseline set of alerts, start with two metrics per component — one for each scope:

- **SaaS** — visibility into the Orion cloud connection
- **Local** — visibility into in-cluster component behavior

| Component             | Metric                                      | What it signals                                                         |
| --------------------- | ------------------------------------------- | ----------------------------------------------------------------------- |
| `orion-cluster-agent`    | `workqueue_depth{controller="workload"}`    | Growth shows workload reconciliation is falling behind                  |
| `orion-applier`  | `orion_applier_api_disconnections_total`       | An increase shows a lost connection to the Orion Platform          |
| `orion-applier`  | `orion_applier_patches_failed_total`           | An increase shows recommendation patches are failing to apply           |
| `orion-forwarder` | `prometheus_remote_storage_samples_pending` | Growth shows metrics are not being forwarded to the Orion Platform |
| `orion-forwarder` | `prometheus_agent_samples_appended_total`   | Steady growth shows targets are being scraped successfully              |

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Recommended alert pattern</div>


Alert on rate-of-increase (not raw counter value) for applier `_total` counters. Alert on sustained growth (rather than instantaneous spikes) or high flat values for `workqueue_depth` and `prometheus_remote_storage_samples_pending`. Alert on flat or declining `prometheus_agent_samples_appended_total`, since that indicates scraping has stalled.
</div>


For a webhook notification each time the Applier applies a recommendation, see [Configure applier alerts][applier-alerts].

[applier-alerts]: /docs/guides/applier-alerts/
[sec-faq]: /docs/reference/faq-security/#metrics
[warnings-errors]: /docs/reference/troubleshooting/#errors-warnings
[trbl-agent-depl-logs]: /docs/reference/troubleshooting/#trbl-agent-depl-logs
[trbl-applier-depl-logs]: /docs/reference/troubleshooting/#trbl-applier-depl-logs
[node-agent-config]: /docs/installation/node-agent-config/
