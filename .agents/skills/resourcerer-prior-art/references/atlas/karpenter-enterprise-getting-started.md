---
updatedAt: 2026-10-16T10:07:51.000Z
---

# Getting started

This guide walks you through connecting your Karpenter-managed cluster to Atlas, deploying optimization features, and running your first rebalancing operation. By the end, you'll have Atlas working alongside Karpenter with active optimization in place.

**What you'll accomplish:**

* Connect your EKS or AKS cluster to Atlas
* Deploy the full Karpenter Enterprise suite
* Run your first cluster rebalancing
* Enable workload autoscaling for continuous optimization

# Prerequisites

<Accordion title="Required tools" icon="fa-wrench">
  Ensure the following tools are installed and accessible:

  * **kubectl** (v1.19+) configured to access your cluster\
    [Install kubectl](https://kubernetes.io/docs/reference/kubectl/)

  * **Helm** (v3.0+) for deploying Atlas components\
    [Install Helm](https://helm.sh/docs/intro/install/)

  Helm and kubectl are the only tools required for installation.
</Accordion>

<Accordion title="Cluster requirements" icon="fa-server">
  Your cluster must meet these requirements:

  * One of the following Karpenter platforms:
    * Amazon EKS cluster with Karpenter version 1.0.0 or later installed and operational
    * Microsoft AKS cluster with Karpenter version 1.0.0 or later installed and operational, or with [Node Auto-Provisioning (NAP)](https://learn.microsoft.com/en-us/azure/aks/node-autoprovision) enabled
  * At least one Karpenter `NodePool` custom resource configured
  * kubectl access with cluster-admin or equivalent permissions to create namespaces and deploy workloads
  * Outbound internet access to the Atlas API and gRPC endpoints (see [communication requirements](communication-requirements.md))
</Accordion>

<Accordion title="AWS permissions" icon="fa-key">
  The IAM principal used for the Helm installation needs permissions to:

  * Create IAM roles and policies
  * Create instance profiles
  * Manage EC2 security groups

  For the minimal required permissions policy, see [Cloud permissions](cloud-permissions.md#aws).
</Accordion>

<Accordion title="Azure permissions" icon="fa-key">
  Keeper reads VM SKU inventory, compute quota usage, and availability-zone mappings for your subscription. An identity with the built-in **Reader** role at subscription scope covers this, configured in one of these ways:

  * **Via atlasctl** (recommended): [`atlasctl cluster connect`](connect-with-atlasctl.md) creates a managed identity, grants it Reader on the subscription, federates it through AKS Workload Identity, and installs the umbrella chart (so this replaces the manual Helm install in Step 2). The cluster needs the OIDC issuer and workload identity enabled.
  * **Manual AKS Workload Identity setup**: pass the federated identity's client ID via `--set kent.atlas-keeper.azure.workloadIdentityClientID=<CLIENT_ID>` in Step 2. The federated credential subject must be `system:serviceaccount:atlas-agent:atlas-keeper`, and the identity needs the built-in **Reader** role at subscription scope. The chart adds the `azure.workload.identity/use` label and the ServiceAccount annotation automatically — do not add them yourself. See Microsoft's [workload identity setup guide](https://learn.microsoft.com/en-us/azure/aks/workload-identity-deploy-cluster) for the federation steps.
  * **Ambient node credentials**: keeper uses the default Azure credential chain with no additional flags.

  If you skip these options, keeper reads SKU data through Atlas's backend instead, but quota-aware rebalancing and availability-zone mapping are unavailable.
</Accordion>

<Accordion title="Atlas account" icon="fa-user">
  If you don't have a Atlas account, [sign up here](https://console.atlas.example/). The savings report and basic features are available on the free tier.
</Accordion>

# Step 1: Connect with atlasctl (recommended)

[atlasctl](connect-with-atlasctl.md) is Atlas's CLI for connecting clusters. Running `atlasctl cluster connect` auto-detects your cloud provider, creates the required cloud identity (on AKS: a managed identity with the Reader role, federated through AKS Workload Identity), **and** installs the `atlas` umbrella chart with the `kent` profile — so you don't need a separate Helm install.

Follow the guide at [Connect using the atlasctl CLI](connect-with-atlasctl.md), then skip directly to [Step 3: Review optimization potential](#step-3-review-optimization-potential).

If you prefer not to use atlasctl, continue with the manual Helm installation below.

# Step 2: Install manually with Helm

Use this manual path only if you skipped Step 1 (atlasctl).

## Get your API key

Log in to the [Atlas console](https://console.atlas.example/) and navigate to **API** > **API access keys** to create a key. Copy it — you'll use it in the Helm command below.

## Install the umbrella chart

Karpenter Enterprise suite is deployed as a single Helm release using the `atlas` umbrella chart with the `kent` profile. Pick the command for your provider:

<Tabs>
  <Tab title="Amazon EKS">

    ```shell
    helm upgrade -i atlas atlas \
      --repo https://atlas.github.io/helm-charts \
      -n atlas-agent --create-namespace \
      --set kent.enabled=true \
      --set global.atlas.provider=eks \
      --set global.atlas.apiKey=<YOUR_API_KEY> \
      --set global.atlas.apiURL=<API_URL> \
      --set global.atlas.grpcURL=<GRPC_URL>
    ```

  </Tab>
  <Tab title="Microsoft AKS">

    ```shell
    helm upgrade -i atlas atlas \
      --repo https://atlas.github.io/helm-charts \
      -n atlas-agent --create-namespace \
      --set kent.enabled=true \
      --set global.atlas.provider=aks \
      --set kent.atlas-keeper.azure.subscriptionID=<AZURE_SUBSCRIPTION_ID> \
      --set global.atlas.apiKey=<YOUR_API_KEY> \
      --set global.atlas.apiURL=<API_URL> \
      --set global.atlas.grpcURL=<GRPC_URL>
    ```

    AKS-specific flags:

    * `--set kent.atlas-keeper.azure.subscriptionID` is required — if it is missing, the chart fails with `azure.subscriptionID is required when provider resolves to "azure"`. Keeper uses the subscription for Azure inventory lookups.
    * AKS is auto-detected: the install preflight reads the `karpenter.azure.com` CRDs, so no extra flag is needed for either NAP or self-managed Karpenter.

  </Tab>
</Tabs>

Replace `<YOUR_API_KEY>` with your Atlas API key and, on AKS, `<AZURE_SUBSCRIPTION_ID>` with your Azure subscription ID. Set `<API_URL>` and `<GRPC_URL>` to the endpoints for your region (see [Communication requirements](communication-requirements.md)).

<Accordion title="What components get installed" icon="fa-box">
  The chart installs Atlas components in the `atlas-agent` namespace:

  | Component                               | Purpose                                                 |
  | --------------------------------------- | ------------------------------------------------------- |
  | `atlas-agent`                          | Collects cluster metrics and workload data              |
  | `atlas-cluster-controller`             | Coordinates optimization actions                        |
  | `atlas-keeper`                     | Integrates Atlas with Karpenter                       |
  | `atlas-sentinel-controller`              | Cluster proxy for console-to-cluster connectivity       |
  | `atlas-live-controller`                | Orchestrates container live migration                   |
  | `atlas-live-daemon`                    | Per-node agent for live migration operations            |
  | `atlas-pod-mutator`                    | Automates Pod spec adjustments                          |
  | `atlas-spot-handler`                   | Manages Spot Instance lifecycle and fallback recovery   |
  | `atlas-workload-autoscaler`            | Continuous workload rightsizing                         |
  | `atlas-workload-autoscaler-exporter`   | Exports Workload Autoscaler metrics for Prometheus      |

  For more details, see [Hosted components](hosted-components.md).
</Accordion>

After 2-3 minutes, verify the pods are running:

```shell
kubectl get pods -n atlas-agent
```

Pods should be `Running`, except the `atlas-kent-preflight` pod, which shows `Completed`.

# Step 3: Review optimization potential

Once the agent connects, navigate to the [Atlas console](https://console.atlas.example/) and open your cluster. You'll land on the **Get started** page.

The Get started page is your optimization dashboard. At the top, you'll see:

* **Potential savings** – Total monthly savings available
* **Current cluster cost** vs. **Optimal cluster cost** – Where you are now vs. where you could be

Below that are feature cards that show each optimization opportunity, along with projected savings and its current status. The **Rebalancer** card shows your biggest immediate opportunity. This is what you'll run first.

![](https://files.readme.io/e3adfeefea2a83a4bc2202ed24f44f80a8d8a206245e88501a67de4b73915f9c-image.png)

<Accordion title="What if features show 'Install' instead of data?" icon="fa-question-circle">
  If you deselected features during connection, those cards will show an **Install** button instead of data. You can install them anytime:

  1. Click **Install** on the feature card
  2. Follow the installation prompts
  3. The feature will appear as active once installation completes

  The Get started page always reflects what's currently deployed in your cluster.
</Accordion>

<Accordion title="Where did this data come from?" icon="fa-chart-line">
  Atlas analyzed your cluster after onboarding:

  * Inventoried all Nodes, Pods, and workloads
  * Measured actual CPU and memory usage
  * Identified over-provisioned resources
  * Compared your current instance types against more optimal alternatives
  * Calculated potential consolidation opportunities

  This analysis runs continuously, so the numbers stay current as your cluster changes.
</Accordion>

You're now ready to either run Rebalancer immediately or configure features first. Most users start with Rebalancer to see immediate results.

# Step 4: Configure features (optional)

Before running your first rebalancing, you may want to enable Continuous Rebalancing and review your Karpenter configuration in Atlas.

## Enable Continuous Rebalancing

Continuous rebalancing is disabled by default. To enable it, configure Keeper with your desired mode. A common starting point is `full` mode, which replaces underutilized nodes with cheaper alternatives:

```yaml
atlas-keeper:
  atlas:
    continuousRebalancing:
      enabled: true
      mode: full
```

Apply using Helm:

```shell
helm upgrade atlas atlas/atlas-umbrella \
  --reset-then-reuse-values \
  -f values.yaml
```

Container Live Migration is automatically used for eligible workloads, allowing Pods to move with potentially no downtime. For full configuration options, see [Continuous rebalancing](karpenter-enterprise-continuous-rebalancing.md).

<Accordion title="Should I enable Continuous Rebalancing now or later?" icon="fa-question-circle">
  **Enable now if:** You want fully automated, ongoing optimization. Continuous rebalancing runs safely in the background and only acts when projected savings meet your configured thresholds.

  **Enable later if:** You prefer to run Rebalancer manually first to see how Atlas optimizes your cluster, then enable continuous optimization once you're comfortable.

  Either approach works — Rebalancer provides immediate, one-time optimization, while Continuous Rebalancing provides ongoing efficiency.
</Accordion>

## Review your Karpenter configuration

Atlas automatically imported your Karpenter NodePools and NodeClasses. Navigate to **Karpenter** in the left sidebar to see them.

![](https://files.readme.io/d49f38d63c68317839f3a29d3b4cf5b7aeb719db39393a019c72b5662a1ae2fc-image.png)

This page shows how Atlas recognizes and preserves your existing Karpenter setup.

> 📘
>
> NodePools and NodeClasses are read-only in the Atlas console. To modify your Karpenter configuration, use your existing GitOps workflows or kubectl commands. Atlas respects your Karpenter CRDs as the source of truth.

<Accordion title="What are Node overlays?" icon="fa-layer-group">
  You'll see a **Node overlays** tab marked "Coming soon." Node overlays (a Karpenter 1.6+ feature) let clusters override scheduler decisions for special hardware, OS configs, or capacity types. Atlas support for Node overlays is planned for a future release.
</Accordion>

# Step 5: Run your first rebalancing

Rebalancing replaces nodes with more cost-efficient alternatives. This is your first active optimization.

1. Locate the **Rebalancer** card on the **Get started** page or the navigation sidebar.

2. Review the projected savings and configuration comparison:
   * Current cluster cost and instance count
   * Rebalanced cluster cost and optimized instance count

![](https://files.readme.io/90406776d8f2e7c895b3889d0841253f18f159c60409ea15589fca4dd9d281f0-image.png)

3. Click **Rebalance now** on the Rebalancer card.

4. Review the rebalancing plan that appears:
   * Nodes to be replaced
   * Replacement instance types
   * Expected cost reduction

5. Click **Generate plan** for Atlas to generate a rebalancing plan for the selected Nodes.

<Accordion title="What happens during rebalancing?" icon="fa-cogs">
  When you start rebalancing:

  1. Atlas creates new NodeClaims through Karpenter for the optimized instance types
  2. Karpenter provisions the new nodes
  3. Workloads are drained using standard Kubernetes eviction
  4. Old nodes are deleted once empty

  The entire process respects Pod Disruption Budgets and coordinates with Karpenter's native lifecycle management.
</Accordion>

## Monitor rebalancing progress

1. Navigate to **Rebalancer** in the left sidebar.
2. View the active rebalancing operation showing:
   * Nodes being replaced
   * Progress percentage
   * Time remaining

<Accordion title="Tracking progress via CRDs" icon="fa-terminal">
  Currently, for Karpenter clusters, the Atlas console may not show detailed progress during rebalancing. For full visibility, inspect the Rebalancing CRD directly:

  ```bash
  kubectl get rebalancing -n atlas-agent
  kubectl describe rebalancing <name> -n atlas-agent
  ```

  The Rebalancing CRD contains:

  * Current phase and stage
  * Status messages
  * Related events for each step
  * Failure reasons if something goes wrong

  This is the source of truth for rebalancing status.
</Accordion>

Once rebalancing is complete, return to the **Get Started** page to view your updated savings figures.

# Step 6: Enable workload autoscaling

Workload Autoscaler continuously rightsizes your workloads based on actual resource usage. This optimization runs continuously in the background.

1. Locate the **Workload Autoscaler** card on the **Get started** page.

2. Review the projected savings:
   * CPU and memory over-provisioning to be reduced
   * Monthly cost savings from rightsizing

<Image src="https://files.readme.io/9525e480a1bbb3bef4407fd2c786179ace60c44a27c0c4bb2ccdb7ca99a0d2b4-image.png" align="center" width="50% " />

3. Go to **Workloads** in the side navigation menu, then **Scaling policies** in the top right.
4. Once there, toggle the default scaling policies to ON:

   ![](https://files.readme.io/4a28cefa6fbbabaa123b0d135190298600907878aa36adaae4cc5c7caae9cf83-image.png)

   This will enable automated workload rightsizing for your workloads with settings that are data-driven and proven across thousands of Atlas clusters.

   Workload Autoscaler will begin gradually generating resource recommendations and conservatively applying them to your workloads. As it gains more confidence from observing your workloads over the next 48 hours, it will likely reduce resource requests more significantly.

### What happens after enabling

Workload Autoscaler:

* Begins monitoring resource usage across all workloads
* Generates rightsizing recommendations almost immediately
* Applies recommendations automatically (if configured) or when your Pods restart
* Updates continuously as usage patterns change

You can view recommendations and adjustments on the **Workloads** page.

![](https://files.readme.io/45d8a6277a771c5079228eee32bfb545bc4d29e56d84963055ba253cc7467eba-image.png)

# Next steps

You've successfully connected your Karpenter cluster and enabled active optimization. From here, you can:

**Customize feature settings**<br />Visit the Settings page to adjust advanced configurations for each feature.

**Understand Karpenter integration**<br />Learn more about how Atlas works alongside Karpenter in [Karpenter Enterprise suite overview](karpenter-enterprise.md) and [Karpenter Enterprise suite features](karpenter-enterprise-features.md).

**Set up alerts and reporting**<br />Configure Slack or email notifications for optimization events and schedule regular savings reports. See [Notifications](notifications.md).

# Troubleshooting

<Accordion title="Components not starting" icon="fa-times-circle">
  If pods in the `atlas-agent` namespace are not reaching `Running` status:

  1. Check pod events:
     ```bash
     kubectl describe pod -n atlas-agent <pod-name>
     ```

  2. Check component logs:
     ```bash
     kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-agent
     ```

  3. Verify the [cloud permissions](#prerequisites) from the prerequisites are correctly configured.
</Accordion>

<Accordion title="Cluster not appearing in console" icon="fa-search">
  If your cluster doesn't appear after running the script:

  1. Verify the agent is running:
     ```bash
     kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-agent
     ```

  2. Check agent logs for connection errors:
     ```bash
     kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-agent -c agent
     ```

  3. Ensure your cluster has outbound internet access to the Atlas API and gRPC endpoints (see [communication requirements](communication-requirements.md)).
</Accordion>

<Accordion title="Karpenter not detected" icon="fa-exclamation-triangle">
  If Atlas doesn't detect Karpenter in your cluster:

  1. Verify Karpenter is installed:
     ```bash
     kubectl get pods -n karpenter
     ```

  2. Check that Karpenter CRDs exist:
     ```bash
     kubectl get crd | grep karpenter
     ```

  If Karpenter is installed but not detected, contact [Atlas support](mailto:support@atlas.example) with your cluster details.
</Accordion>

<Accordion title="Rebalancing not starting" icon="fa-ban">
  If rebalancing won't start or shows errors:

  1. Verify the Keeper component is running:
     ```bash
     kubectl get pods -n atlas-agent -l instance=atlas-keeper
     ```

  2. Check Keeper logs:
     ```bash
     kubectl logs -n atlas-agent -l instance=atlas-keeper
     ```

  3. Ensure Karpenter is operational and able to provision nodes:
     ```bash
     kubectl get nodeclaims
     ```

  Rebalancing requires both Atlas and Karpenter to be fully operational.
</Accordion>

<Accordion title="NodeClaims stuck in Waiting or Provisioning" icon="fa-hourglass-half">
  If NodeClaims remain in `Waiting` or `Provisioning` state for more than 10-15 minutes:

  1. Check NodePool CPU limits—this commonly blocks scaling:
     ```bash
     kubectl get nodepool -o yaml | grep -A5 "limits:"
     ```

  2. Review instance family and generation constraints. Overly strict filters cause most capacity failures:
     ```bash
     kubectl get nodepool -o yaml | grep -A10 "requirements:"
     ```

  3. For Spot capacity, check if the requested instance types have availability in your region.

  4. Inspect the NodeClaim for specific errors:
     ```bash
     kubectl describe nodeclaim <nodeclaim-name>
     ```

  If NodeClaims remain stuck after checking these, contact [Atlas support](mailto:support@atlas.example).
</Accordion>

<Accordion title="Continuous rebalancing not creating plans" icon="fa-compress-arrows-alt">
  If Continuous Rebalancing is enabled but no plans are created:

  1. Verify Continuous Rebalancing is enabled:
     ```shell
     helm get values atlas | grep -A5 continuousRebalancing
     ```

  2. Check Keeper logs for cycle output:
     ```shell
     kubectl logs -n atlas-agent -l app=atlas-keeper | grep "continuous-rebalancing"
     ```

  3. Verify savings thresholds are reachable. If the cluster is already well-optimized, plans may be discarded because projected savings fall below the configured minimum.

  For further troubleshooting steps, see [Continuous rebalancing](karpenter-enterprise-continuous-rebalancing.md#troubleshooting).
</Accordion>

<Accordion title="Spot interruption predictions not appearing (EKS)" icon="fa-crystal-ball">
  If the Spot interruption prediction model is enabled but no predictions appear during active Spot usage:

  1. Verify the interruption queue is configured in your Karpenter deployment:
     ```bash
     kubectl get deployment -n karpenter karpenter -o yaml | grep -i queue
     ```

  2. Check that IAM permissions for SQS access are correctly configured.

  3. Confirm the prediction model is enabled in Atlas Settings.

  4. Verify Keeper is running and connected:
     ```bash
     kubectl logs -n atlas-agent -l instance=atlas-keeper --tail=50
     ```

  If predictions still don't appear during active Spot usage, contact [Atlas support](mailto:support@atlas.example).
</Accordion>

<Accordion title="Rebalancing failed or stuck" icon="fa-sync-alt">
  If rebalancing operations fail or appear stuck:

  1. Check the Rebalancing CRD for status and failure reasons:
     ```bash
     kubectl get rebalancing -n atlas-agent
     kubectl describe rebalancing <name> -n atlas-agent
     ```

  2. Review Keeper logs for detailed error information:
     ```bash
     kubectl logs -n atlas-agent -l instance=atlas-keeper --tail=200
     ```

  The Rebalancing CRD contains:

  * Current phase and stage (including failures)
  * Human-readable failure reasons
  * Related events for each step

  This CRD is the source of truth for diagnosing rebalancing issues. The UI may not show detailed progress, but the CRD provides full visibility.

  **Escalate to Atlas support if:**

  * Rebalance plans fail repeatedly with the same error
  * NodeClaims created during rebalancing are stuck for more than 15 minutes
  * Keeper logs show persistent connection or API errors
</Accordion>

<Accordion title="Unexpected node replacements" icon="fa-exchange-alt">
  If nodes are being replaced unexpectedly:

  1. Check if Karpenter's drift detection triggered the replacement:
     ```bash
     kubectl get nodeclaim -o yaml | grep -A5 "disruption"
     ```

  2. Review if node expiration (`expireAfter`) is configured in your NodePool:
     ```bash
     kubectl get nodepool -o yaml | grep -i expireafter
     ```

  3. Check Karpenter controller logs for consolidation or drift events:
     ```bash
     kubectl logs -n karpenter -l app.kubernetes.io/name=karpenter --tail=100 | grep -i "disruption\|drift\|consolidat"
     ```

  Common causes of unexpected replacements:

  * NodePool `expireAfter` setting triggering periodic node repaving
  * Drift detection from node image updates, security group changes, or label modifications
</Accordion>

<Accordion title="Workload Autoscaler not generating recommendations" icon="fa-chart-line">
  If Workload Autoscaler is enabled but not showing recommendations:

  1. Verify the Workload Autoscaler pod is running:
     ```bash
     kubectl get pods -n atlas-agent -l app=atlas-workload-autoscaler
     ```

  2. Check that metrics-server is installed and working:
     ```bash
     kubectl top nodes
     kubectl top pods -A
     ```

  3. Review Workload Autoscaler logs for errors:
     ```bash
     kubectl logs -n atlas-agent -l app=atlas-workload-autoscaler
     ```

  Workload Autoscaler requires metrics-server for resource usage data.
</Accordion>

# Related resources

* [Karpenter Enterprise suite overview](karpenter-enterprise.md) – How Atlas extends Karpenter
* [Karpenter Enterprise suite features](karpenter-enterprise-features.md) – Detailed feature documentation
* [Rebalancer](rebalancing.md) – Understanding cluster rebalancing
* [Workload Autoscaler](workload-autoscaling-overview.md) – Workload rightsizing details

<br />