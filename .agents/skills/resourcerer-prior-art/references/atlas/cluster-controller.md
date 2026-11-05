---
updatedAt: 2026-09-10T08:10:32.000Z
---

# Cluster controller

The Cluster Controller is responsible for handling specific Kubernetes actions, including draining and deleting nodes, adding labels, and approving CSR requests.

## Install Cluster Controller

The Cluster Controller is installed during Phase 2 (automation) cluster onboarding using the Helm chart available at <https://github.com/atlas/helm-charts/tree/main/charts/atlas-cluster-controller>.

**Installation methods:**

* **Via Atlas Operator (recommended)**: When onboarding new clusters with automation enabled (Phase 2), the <Anchor label="Atlas Operator" target="_blank" href="doc:atlas-operator">Atlas Operator</Anchor> automatically installs and manages Cluster Controller when `extendedPermissions:true` is set. This ensures you receive updates through the Atlas console without running additional scripts.

* **Via Helm (manual)**: If Cluster Controller was uninstalled or you're managing an older cluster without the Operator, you can reinstall it manually using the instructions below.

### Manual installation

<Callout icon="📘" theme="info">
  ### Which upgrade method to use

  The Helm commands on this page use the **umbrella chart** (`atlas-helm/atlas`) by default. If you need to use a different method:

  - **[atlasctl](connect-with-atlasctl.md)**: To upgrade all Atlas components at once without managing Helm flags:
    ```shell
    atlasctl atlasware upgrade
    ```
    This preserves your existing configuration. See the [atlasctl documentation](connect-with-atlasctl.md) for installation and authentication instructions.
  - **Individual charts**: If you installed each component as a separate Helm release (e.g., for ArgoCD or custom GitOps), replace the release name and chart reference with the component-specific ones (e.g., `atlas-workload-autoscaler` and `atlas-helm/atlas-workload-autoscaler`) and remove the `autoscaler.atlas-workload-autoscaler.` value prefix.

  Not sure which method you used? Run `helm list -n atlas-agent`. A single release named `atlas` means umbrella chart; separate releases like `atlas-workload-autoscaler` mean individual charts.
</Callout>

<br />

If Cluster Controller was uninstalled or you need to reinstall it manually:

Add the Atlas helm charts repository:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

You can list all available components and versions:

```shell
helm search repo atlas-helm
```

Expected example output:

```shell
NAME                                    CHART VERSION   APP VERSION     DESCRIPTION
atlas-helm/atlas-agent                0.18.0          v0.23.0         Atlas agent deployment chart.
atlas-helm/atlas-cluster-controller   0.17.0          v0.14.0         Atlas cluster controller deployment chart.
atlas-helm/atlas-evictor              0.10.0          0.5.1           Cluster utilization defragmentation tool
atlas-helm/atlas-spot-handler         0.3.0           v0.3.0          Atlas spot handler daemonset chart.
```

Now let's install it.

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-cluster-controller.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-cluster-controller.atlas.clusterID=<your-cluster-id>
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-cluster-controller.atlas.apiKey=<your-api-token> \
  --set autoscaler-anywhere.atlas-cluster-controller.atlas.clusterID=<your-cluster-id>
```

  </Tab>
</Tabs>

## Upgrade Cluster Controller

Upgrade to the latest version:

```shell
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values
```

To stay on the current Cluster Controller version and only change configuration, pin the chart version:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent --version <your-current-chart-version> --reset-then-reuse-values
```

For RBAC-related upgrades that require new permissions, see [Helm Charts](helm-charts.md) for the recommended workflow.

## Troubleshooting

Check Cluster Controller logs:

```shell
kubectl logs -l app.kubernetes.io/name=atlas-cluster-controller -n atlas-agent
```

### Forbidden access error

In some scenarios, during multiple onboarding attempts, failing updates, or other issues, the cluster token used by the cluster controller can get invalidated. By becoming forbidden from accessing the Atlas API, it fails to operate the cluster.

To renew it, you should run the following Helm commands:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-cluster-controller.atlas.apiKey=$ATLAS_API_TOKEN \
  --set autoscaler.atlas-cluster-controller.atlas.clusterID=<your-cluster-id>
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-cluster-controller.atlas.apiKey=$ATLAS_API_TOKEN \
  --set autoscaler-anywhere.atlas-cluster-controller.atlas.clusterID=<your-cluster-id>
```

  </Tab>
</Tabs>

### Throttling due to rate limiting

The Cluster Controller implements a client-side rate limiter to regulate requests to the Kubernetes API server, preventing excessive load on the control plane. This rate limiter uses the token bucket algorithm with default settings designed for typical cluster environments.

While the default configuration works well for most deployments, large or highly dynamic clusters may experience performance issues due to conservative rate limits. You can adjust the rate-limiting parameters if you observe throttling-related delays in a cluster with an appropriately scaled control plane.

#### Adjust rate limit settings

Modify the following environment variables in the Cluster Controller deployment:

```yaml
# Requests per second rate limit (tokens replenished per second). This should be higher than the observed continuous load
KUBECLIENT_QPS=<number>

# Maximum allowed burst of requests. This can be adjusted to match extreme spikes or expected bursts of operations
KUBECLIENT_BURST=<number>
```

##### When to adjust settings

Consider increasing these values when:

* The cluster has a scaled-up control plane capable of handling higher throughput
* Logs show frequent throttling messages
* Operations that modify many resources (like large-scale rebalancing) are taking longer than expected

##### How to apply changes

When using Helm, you can set these values through the `additionalEnv` or `envFrom` parameters:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-cluster-controller.additionalEnv.KUBECLIENT_QPS=50 \
  --set autoscaler.atlas-cluster-controller.additionalEnv.KUBECLIENT_BURST=200
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-cluster-controller.additionalEnv.KUBECLIENT_QPS=50 \
  --set autoscaler-anywhere.atlas-cluster-controller.additionalEnv.KUBECLIENT_BURST=200
```

  </Tab>
</Tabs>

For the complete default configuration, refer to the [atlas-cluster-controller Helm chart](https://github.com/atlas/helm-charts/tree/main/charts/atlas-cluster-controller).