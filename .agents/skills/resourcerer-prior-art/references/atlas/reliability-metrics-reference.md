---
updatedAt: 2026-09-02T10:37:30.000Z
---

# Reference

atlasctl flags, auto-sizing behavior, and what gets deployed in your cluster when you enable reliability metrics.

This page documents the `atlasctl` flags available for reliability metrics, explains how `atlasctl` automatically sizes components for your cluster, and describes what gets provisioned in the cluster during installation. For an overview of the feature and step-by-step installation instructions, see [Reliability metrics](reliability-metrics).

***

## atlasctl cluster connect flags

These flags are used with `atlasctl cluster connect` to install and configure reliability metrics.

| Flag                    | Description                                                                                                                       |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `--reliability-metrics` | Enable reliability metrics during cluster connect. Required to install the feature.                                               |
| `--non-interactive`     | Skip all interactive prompts. Use this flag in CI/CD pipelines where there is no terminal.                                        |
| `--cluster-name <name>` | Set the display name for the cluster being connected.                                                                             |
| `--dry-run`             | Print what would be installed without making any changes to the cluster. Useful for reviewing the install plan before committing. |

These flags can be combined with flags for other Atlas features in a single command:

```bash
atlasctl cluster connect \
  --non-interactive \
  --cluster-name production-us-east \
  --reliability-metrics \
  --cluster-optimization \
  --workload-autoscaler
```

***

## atlasware upgrade flags

Use `atlasctl atlasware upgrade` to upgrade reliability metrics and all other Atlas components in your cluster.

| Flag                | Description                                        |
| ------------------- | -------------------------------------------------- |
| `--non-interactive` | Skip confirmation prompts. Use in CI/CD pipelines. |

```bash
atlasctl atlasware upgrade --non-interactive
```

Upgrading with `atlasware` preserves your existing configuration and applies any new data migrations automatically when the collector restarts.

***

## How atlasctl sizes components

When you run `atlasctl cluster connect --reliability-metrics`, `atlasctl` inspects your cluster before installing anything and automatically determines the right resource allocation for two components: the collector agent that runs on each node, and the metrics database.

You do not need to specify sizing manually. The selections below describe what `atlasctl` chooses and why.

### Collector agent sizing

The collector agent runs as a DaemonSet (one instance per node) and observes HTTP and gRPC traffic on that node. `atlasctl` counts the maximum number of containers running on any single node and selects a sizing profile accordingly.

| Max containers on any node | Selected profile | CPU request | CPU limit | Memory request | Memory limit |
| -------------------------- | ---------------- | ----------- | --------- | -------------- | ------------ |
| Up to 5                    | `small`          | 50m         | 200m      | 128 MiB        | 256 MiB      |
| Up to 15                   | `medium`         | 100m        | 400m      | 256 MiB        | 512 MiB      |
| Up to 30                   | `large`          | 200m        | 800m      | 512 MiB        | 1 GiB        |
| More than 30               | `xlarge`         | 400m        | 1600m     | 1 GiB          | 2 GiB        |

These profiles reflect expected traffic volume per node at each workload density level. A node running many containers typically serves more concurrent requests and needs more resources to observe them accurately.

### Metrics database sizing

`atlasctl` also provisions a ClickHouse time-series database in your cluster for temporary metric storage. It sizes the database by summing your cluster's total allocatable CPU across all nodes.

| Total cluster allocatable CPU | CPU request | Memory request | Memory limit |
| ----------------------------- | ----------- | -------------- | ------------ |
| Less than 100 cores           | 500m        | 1 GiB          | 2 GiB        |
| Less than 500 cores           | 1 core      | 2 GiB          | 4 GiB        |
| Less than 1,000 cores         | 2 cores     | 4 GiB          | 8 GiB        |
| Less than 5,000 cores         | 2 cores     | 8 GiB          | 16 GiB       |
| Less than 10,000 cores        | 4 cores     | 16 GiB         | 32 GiB       |
| 10,000 cores or more          | 4 cores     | 32 GiB         | 64 GiB       |

The default persistent volume size for the metrics database is **10 GiB**. Data collected in the cluster is retained for up to 7 days before being purged. It is exported to Atlas continuously and only needs to be stored locally for a short window.

### ClickHouse operator detection

`atlasctl` checks whether a ClickHouse operator (Altinity) is already running in your cluster before installing. If one is found, `atlasctl` skips the operator installation and reuses the existing one. This prevents conflicts on clusters that already use ClickHouse for other workloads.

If your cluster has restricted Custom Resource Definition permissions and the operator cannot be installed automatically, `atlasctl` will surface an error. In this case, install the Altinity ClickHouse Operator CRDs manually first, then re-run `atlasctl cluster connect --reliability-metrics`.

***

## What gets installed in your cluster

When reliability metrics is enabled, `atlasctl` installs the following components in the `atlas-agent` namespace:

| Component           | Type                         | Role                                                                                            |
| ------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------- |
| Collector agent     | DaemonSet (one pod per node) | Observes HTTP and gRPC network traffic and sends measurements to the metrics database           |
| Metrics database    | StatefulSet (ClickHouse)     | Stores and aggregates traffic measurements temporarily before they are exported to Atlas      |
| Exporter            | Deployment                   | Reads aggregated data from the metrics database and streams it to Atlas continuously          |
| ClickHouse Operator | Deployment (cluster-scoped)  | Manages the lifecycle of the ClickHouse StatefulSet. Skipped if an operator is already present. |

All components are managed by `atlasctl` and `atlasctl atlasware upgrade`. You do not need to interact with them directly under normal operation.

***

## Default monitored ports

The collector observes traffic on these ports by default:

| Port | Common use                   |
| ---- | ---------------------------- |
| 8080 | HTTP application traffic     |
| 8443 | HTTPS application traffic    |
| 8090 | Common alternative HTTP port |
| 6379 | Redis                        |

> **Services on other ports**
>
> Traffic on ports not in this list is not collected. When installing with Helm, you can add ports by setting `autoscaler.atlas-sentinel.agent.reliabilityMetrics.obi.openPorts` (a comma-separated list). For the full set of Helm values, see [ArtifactHUB](https://artifacthub.io/packages/helm/atlas/atlas-sentinel/).

***

## Helm charts

Installation via `atlasctl` is the recommended method. It takes care of many manual pitfalls and sets up the sizing profiles along with the requested resources properly.

To install the reliability-metrics stack with the Atlas Helm chart, including for GitOps workflows, first review the generic Helm instructions at [Helm charts](helm-charts.md).

Because `atlasctl` is not sizing the components for you, you will need to size the resources properly on your own, see:

* [Collector agent sizing](#collector-agent-sizing)
* [Metrics database sizing](#metrics-database-sizing)

### Prerequisites

If you don't use the ClickHouse operator in your cluster, you need to install the Altinity ClickHouse CRDs:

```bash
kubectl apply -f https://raw.githubusercontent.com/Altinity/clickhouse-operator/refs/tags/release-0.27.0/deploy/operator/parts/crd.yaml
```

If you already use the ClickHouse operator configured to watch all namespaces (including `atlas-agent`), disable the
ClickHouse operator installed by this chart by setting:

```bash
--set autoscaler.atlas-sentinel.reliabilityMetrics.operator.enabled=false
```

### Installation

```bash
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update

helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<apiKey>" \
  --set global.atlas.provider="<provider>" \
  --set tags.readonly=true \
  --set autoscaler.atlas-sentinel.enabled=true \
  --set autoscaler.atlas-sentinel.reliabilityMetrics.enabled=true \
  --set autoscaler.atlas-sentinel.reliabilityMetrics.operator.enabled=true \
  --set autoscaler.atlas-sentinel.agent.reliabilityMetrics.obi.sizingProfile="<sizingProfile>" \
  --set autoscaler.atlas-sentinel.reliabilityMetrics.clickhouse.resources.requests.cpu="<cpu>" \
  --set autoscaler.atlas-sentinel.reliabilityMetrics.clickhouse.resources.requests.memory="<memory>" \
  --set autoscaler.atlas-sentinel.reliabilityMetrics.clickhouse.resources.limits.memory="<memory>"
```

> 📘 Note
>
> We run a Helm lookup check for the ClickHouse operator in your cluster. This lookup doesn't work well with `helm template`. If you want to run `helm template` in your setup, consider adding:
>
> ```bash
> --set autoscaler.atlas-sentinel.reliabilityMetrics.clickhouse.skipCRDValidation=true
> ```

For all available configuration options, see [ArtifactHUB](https://artifacthub.io/packages/helm/atlas/atlas-sentinel/).

### Common Helm values

The values below are the ones most Helm and GitOps users need beyond the install command. Paths are relative to the `atlas-sentinel` chart, so prefix each with `autoscaler.atlas-sentinel.` when setting it on the umbrella chart.

| Value                                                    | Default               | Description                                                                                                                                                                                                                                                                                                                 |
| -------------------------------------------------------- | --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `agent.reliabilityMetrics.obi.openPorts`                 | `8080,8443,8090,6379` | Comma-separated list of ports the collector instruments. Add ports your services listen on. Each additional port raises collector memory use (roughly 27 MiB per instrumented process).                                                                                                                                     |
| `agent.reliabilityMetrics.obi.sizingProfile`             | `medium`              | Collector sizing profile: `small`, `medium`, `large`, `xlarge`, or `custom`. See [collector agent sizing](#collector-agent-sizing). Set to `custom` to size manually with `agent.reliabilityMetrics.obi.resources`.                                                                                                         |
| `agent.reliabilityMetrics.obi.exclude`                   | `[]`                  | Workloads to skip, reducing memory use. Each entry matches on `k8s_namespace`, `exe_path`, `open_ports`, `container_name`, `k8s_pod_labels`, `k8s_pod_annotations`, or `cmd_args` (glob patterns supported). Built-in exclusions (`kube-system`, `cert-manager`, and others) apply unless you set `excludeDefaults: false`. |
| `reliabilityMetrics.clickhouse.persistence.size`         | `100Gi`               | Size of the ClickHouse persistent volume (chart default; `atlasctl` sizes this differently — see [metrics database sizing](#metrics-database-sizing)).                                                                                                                                                                       |
| `reliabilityMetrics.clickhouse.persistence.storageClass` | unset                 | Storage class for the ClickHouse volume. Uses the cluster's default storage class when unset.                                                                                                                                                                                                                               |

Exclusions are easier to set from a values file than from `--set`:

```yaml
autoscaler:
  atlas-sentinel:
    agent:
      reliabilityMetrics:
        obi:
          exclude:
            - k8s_namespace: "monitoring"
            - exe_path: "*nginx*"
```

### Ensure successful installation

```bash
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller --tail 100 | grep "component=clickhouse_exporter"
```

Look for a line similar to the one below, which indicates that data was successfully sent to Atlas:

```
time=2026-07-29T14:32:58.736Z level=INFO source=worker.go:205 msg="cursor committed: last_exported_time=2026-07-29 14:31:21 +0000 UTC last_exported_pk=16025012183535387539" component=clickhouse_exporter table=reliability_metrics_db
```

It can take up to 3 minutes for the data to start flowing to Atlas.
If you don't see this log, ensure you have at least one observable process in your cluster.
To test the installation, rerun it with `--set autoscaler.atlas-sentinel.trafficGenerator.enabled=true`.
It will create a `traffic-gen` namespace running some arbitrary traffic generators.

<Cards columns={2}>
  <Card title="Reliability metrics" href="https://docs.atlas.example/docs/reliability-metrics" icon="fa-signal">
    Overview of reliability metrics, installation, and console navigation.
  </Card>
  <Card title="Configuring Sentinel" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Configure Sentinel for resource monitoring and security scanning.
  </Card>
</Cards>