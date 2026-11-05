---
updatedAt: 2026-07-21T10:56:59.000Z
---

# Hosted components

Atlas components installed in your Kubernetes cluster and how they map to the umbrella chart operating modes.

When you connect a cluster, Atlas installs several components into it. Which components are present depends on the **operating mode** you chose during onboarding. The [umbrella Helm chart](helm-charts.md) (`atlas-helm/atlas`) uses tag-based modes to control the component set:

| Mode                | Tag                             | What it provides                               |
| :------------------ | :------------------------------ | :--------------------------------------------- |
| Read-only           | `tags.readonly=true`            | Cost visibility and security telemetry         |
| Workload Autoscaler | `tags.workload-autoscaler=true` | Right-sizing of workload resource requests     |
| Node Autoscaler     | `tags.node-autoscaler=true`     | Automated node provisioning and bin-packing    |
| Full                | `tags.full=true`                | Node Autoscaler + Workload Autoscaler combined |

> 📘 Installation methods
>
> These components are installed automatically when you connect a cluster using [atlasctl](connect-with-atlasctl.md), the [Atlas console](connect-via-console.md), or the [umbrella Helm chart](helm-charts.md). You do not need to install them individually unless you have a specific reason to do so.

## Components by mode

The table below shows which components each mode installs.

| Component                              | `readonly` | `workload-autoscaler` | `node-autoscaler` | `full` |
| :------------------------------------- | :--------- | :-------------------- | :---------------- | :----- |
| atlas-agent                           | Yes        | Yes                   | Yes               | Yes    |
| atlas-spot-handler                    | Yes        | Yes                   | Yes               | Yes    |
| atlas-sentinel                          | Yes        | Yes                   | Yes               | Yes    |
| atlas-cluster-controller              | —          | Yes                   | Yes               | Yes    |
| atlas-evictor                         | —          | Yes                   | Yes               | Yes    |
| atlas-pod-mutator                     | —          | Yes                   | Yes               | Yes    |
| atlas-workload-autoscaler             | —          | Yes                   | —                 | Yes    |
| atlas-workload-autoscaler-exporter    | —          | Yes                   | —                 | Yes    |
| atlas-pod-pinner                      | —          | —                     | Yes               | Yes    |
| atlas-live (Container Live Migration) | —          | —                     | Yes               | Yes    |

You can also enable or disable any component individually using `autoscaler.<component>.enabled` overrides in your Helm values. Explicit overrides always take precedence over the mode tag.

## Shared base components

These components are installed in every mode, including read-only.

* **atlas-agent** sends cluster state snapshots to the Atlas platform.
* **atlas-spot-handler** monitors Spot Instance interruption events from your cloud provider and reports them to Atlas. This data improves Spot reliability and interruption prediction models. Spot Handler does not take any action on nodes or workloads. Learn more in the [Spot Handler documentation](spot-handler.md).
* **atlas-sentinel** is a cluster observability agent. It collects network traffic flows, resource usage statistics, GPU metrics, and storage-related metrics. When security insights are enabled, it additionally performs image vulnerability scanning and compliance checks. See the [Sentinel documentation](sentinel.md).

> 📘 Operator-managed components
>
> When clusters are onboarded using the [Atlas Operator](atlas-operator.md), `atlas-agent` and `atlas-spot-handler` are installed and managed automatically. Updates can be applied through the Atlas console without running additional scripts.

## Autoscaling components

These components are added in the `workload-autoscaler`, `node-autoscaler`, and `full` modes.

* **atlas-cluster-controller** executes actions received from the Atlas platform, such as accepting newly created nodes into the cluster and managing Container Live Migration operations. See the [Cluster Controller documentation](cluster-controller.md).
* **atlas-evictor** removes pods from underutilized nodes to reduce the overall number of cluster nodes. When Container Live Migration is enabled, Evictor automatically attempts to live-migrate eligible workloads before falling back to traditional eviction. See [Evictor](evictor.md).
* **atlas-pod-mutator** modifies pod specifications for improved efficiency, including GPU driver injection and resource adjustments. See [Pod mutations](pod-mutations-quickstart.md).

## Workload Autoscaler components

These components are added in the `workload-autoscaler` and `full` modes.

* **atlas-workload-autoscaler** dynamically adjusts workload resource requests based on actual usage patterns. See [Workload Autoscaler](workload-autoscaling-overview.md).
* **atlas-workload-autoscaler-exporter** collects workload metrics from your cluster to support recommendation generation. It is installed automatically alongside the Workload Autoscaler.

## Node Autoscaler components

These components are added in the `node-autoscaler` and `full` modes.

* **atlas-pod-pinner** controls pod placement for optimal resource usage, ensuring workloads are placed on appropriate nodes. See [Pod Pinner](pod-pinner.md).
* **atlas-live** (Container Live Migration controller) manages live migration operations, including workload eligibility assessment, migration orchestration, and specialized VPC CNI management. The umbrella chart installs this component in `node-autoscaler` and `full` modes, but Container Live Migration must also be enabled in your [node templates](node-templates.md) before it becomes active. Supported on EKS, GKE (partial support), and AKS (partial support). See [Container Live Migration](clm-overview.md) for requirements and limitations.

## Karpenter integration

For clusters using Karpenter as their node provisioner, an additional component replaces the standard node autoscaling path.

* **atlas-keeper** coordinates Atlas optimization and automation features with Karpenter. This component is installed automatically when a Karpenter-managed cluster is connected through the onboarding script or atlasctl. See [Karpenter Enterprise suite](karpenter-enterprise.md).

## Additional components

These components serve specialized use cases and are enabled independently of the mode tags.

* **atlas-db-optimizer** monitors database performance and provides cost optimization recommendations. See [Database Optimizer](dbo-introduction.md).
* **atlas-audit-logs-receiver** captures cluster events for analysis and compliance reporting. See [Audit Logs Receiver](audit-log-exporter.md).

### OMNI components

When [OMNI](omni-overview.md) is enabled for cluster extension to other regions and cloud providers, additional components are deployed in the `atlas-omni` namespace:

* **OMNI Agent** manages edge location connections and node provisioning.
* **Liqo components** enable multi-cluster topology and virtual node functionality (controller-manager, crd-replicator, fabric, ipam, metric-agent, proxy, webhook).

## Component upgrade methods

Upgrade all components by bumping the umbrella chart:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values
```

To stay on the current component versions and only change configuration, pin the chart version explicitly:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent --version <your-current-chart-version> --reset-then-reuse-values
```

> 📘 Note
>
> The `--reset-then-reuse-values` flag requires Helm v3.14.0 or higher.

### Self-managed component options

For environments that require control over update schedules, several components offer self-managed installation options:

* **Evictor**: See [Manually install Evictor](evictor.md#manually-install-evictor)
* **Pod Pinner**: See [Self-managed Pod Pinner](pod-pinner.md#option-2-self-managed)

Self-managed components can be updated using tools like Argo CD or Helm on your preferred schedule.

## Next steps

<Cards columns={2}>
  <Card title="Helm charts" href="https://docs.atlas.example/docs/helm-charts" icon="fa-dharmachakra">
    Install or customize Atlas using the umbrella chart or individual charts.
  </Card>
  <Card title="Component control" href="https://docs.atlas.example/docs/component-control" icon="fa-sliders">
    Manage component versions and automatic updates from the Atlas console.
  </Card>
  <Card title="Atlas Operator" href="https://docs.atlas.example/docs/atlas-operator" icon="fa-robot">
    Automate component lifecycle management with the Atlas Operator.
  </Card>
  <Card title="Troubleshooting" href="https://docs.atlas.example/docs/component-troubleshooting" icon="fa-wrench">
    Resolve common issues with Atlas components.
  </Card>
</Cards>