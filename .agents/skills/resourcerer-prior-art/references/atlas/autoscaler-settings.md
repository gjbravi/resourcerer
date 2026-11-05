---
updatedAt: 2026-09-10T08:10:32.000Z
---

# Autoscaler settings

# Cluster CPU limits policy

The cluster CPU limits policy allows you to set a boundary on the total amount of vCPUs available across all worker nodes in your cluster. This policy helps manage cluster scaling and resource allocation.

## How it works

* When enabled, this policy limits the total CPU capacity of the cluster's worker nodes based on the configured maximum CPU limit.
* The cluster will not scale up above the maximum CPU limit, even if there are pending Pods.

<Callout icon="📘" theme="info">
  ### Nodes added outside Atlas

  The maximum CPU limit controls when the autoscaler stops adding nodes. It does not prevent other systems from adding capacity independently — for example, external autoscalers (Karpenter, NAP, Cluster Autoscaler) or nodes connected manually via OMNI.

  If externally-added nodes push total cluster CPU above your configured maximum, the autoscaler treats the limit as already reached and will not provision additional nodes until CPU drops back below it.

  If you need separate limits for different parts of your cluster, use per-node-template constraints rather than the global limit.
</Callout>

## Configuring CPU limits policy

<Callout icon="⚠️" theme="warn">
  ### Minimum CPU limit deprecated

  The cluster-wide minimum CPU limit is deprecated. If you are currently using it, migrate off it by removing the minimum CPU limit and keeping only the maximum CPU limit. The minimum CPU limit is not enforced consistently across Atlas components and will be removed in a future release.

  If you still have a minimum CPU limit configured:

  - **Node downscaler** respects the minimum CPU limit.
  - **Evictor** does not respect the minimum CPU limit.
  - **Rebalancer** does not respect the minimum CPU limit.

  New configurations can no longer set a minimum CPU limit.
</Callout>

You can configure the CPU limits policy through the Atlas Console, the Atlas API, or the <Anchor target="_blank" href="https://registry.terraform.io/providers/atlas/atlas/latest/docs/resources/autoscaler">Atlas Terraform provider</Anchor>.

### The Atlas Console

1. Navigate to ***Cluster*** -> ***Settings*** -> ***Node autoscaling advanced settings***.
2. On the ***Scaling policies*** tab, toggle the ***CPU limit policy*** to *On*.
3. Set the maximum CPU cores.

<Image src="https://files.readme.io/57e3ac0e59ba9496b28da2ddeb3de084a9956d4c4419517c5af4526caa209f10-image.png" align="center" caption="CPU limit policy" />

### API

1. Use the Atlas <Anchor target="_blank" href="https://docs.atlas.example/reference/policiesv2api_updateclusterpolicies">policies API endpoint</Anchor>: `PATCH /cluster-autoscaler/v2/clusters/{clusterId}/policies`.
2. Set the maximum CPU limit:

```json
"clusterLimits": {
    "enabled": true,
    "cpu": {
      "maxCores": 20
    }
```

The new settings will propagate immediately after being applied.

# Scoped mode

**Scoped mode** limits node autoscaling by requiring Pods to have specific tolerations to trigger scale-up, and restricting downscaling to only Atlas-provisioned nodes.

Only Pods that tolerate the `scheduling.atlas.example/scoped-autoscaler` taint will cause the autoscaler to add new nodes, and only nodes with the `provisioner.atlas.example/managed-by=atlas.example` label can be removed during downscaling.

## When to use scoped mode

Scoped mode is useful in scenarios where:

* You want to maintain existing node groups while autoscaling only Atlas-managed infrastructure
* You're gradually migrating from traditional node groups to Atlas node templates
* You need to ensure certain workloads remain on specific, non-Atlas-managed nodes
* You want to prevent the autoscaler from downscaling manually provisioned nodes

## How scoped mode works

When scoped mode is enabled:

* Downscaling operations will only target nodes that have the label `provisioner.atlas.example/managed-by=atlas.example`
* The `default-by-atlas` node template receives an additional taint: `scheduling.atlas.example/scoped-autoscaler:NoSchedule`
* Pods must tolerate this taint to be scheduled on nodes created from the default template
* Custom node templates do not automatically receive this taint and must be configured manually if needed
* Existing Atlas-managed nodes are not automatically tainted when scoped mode is enabled
* Manually provisioned nodes and existing node groups remain untouched by autoscaling operations

This ensures that:

* Only nodes created and managed by Atlas participate in autoscaling
* Existing node groups and manually managed nodes remain stable
* You maintain full control over which parts of your infrastructure are autoscaled

## Enable scoped mode

Scoped mode can only be configured via the Atlas API or Terraform and is not available through the Atlas console.

<Tabs>
  <Tab title="API">
    Use the <Anchor label="policies API endpoint" target="_blank" href="https://docs.atlas.example/reference/policiesv2api_updateclusterpolicies">policies API endpoint</Anchor> to enable scoped mode:

    ```json
    {
      "nodeDownscaler": {
        "enabled": true
      },
      "scopedMode": true
    }
    ```
  </Tab>

  <Tab title="Terraform">
    ```toml
    resource "atlas_autoscaler" "atlas_autoscaler_policy" {
      cluster_id = atlas_eks_cluster.test.id
      
      autoscaler_settings {
        enabled        = true
        is_scoped_mode = true
        
        node_downscaler {
          enabled = true
        }
      }
    }
    ```
  </Tab>
</Tabs>

## Configure Pods for scoped mode

When scoped mode is enabled, Pods that should be scheduled on autoscaler-managed nodes must include the following toleration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-application
spec:
  template:
    spec:
      tolerations:
        - key: "scheduling.atlas.example/scoped-autoscaler"
          operator: "Exists"
          effect: "NoSchedule"
      # Additional Pod configuration
```

Pods without this toleration will not be scheduled on autoscaler-managed nodes and will remain on your existing node groups or manually provisioned nodes. The autoscaler will also not consider Pods lacking this toleration when deciding to scale up the cluster and add nodes for more capacity, even if these Pods are unschedulable.

<Callout icon="📘" theme="info">
  ### Existing workloads

  When enabling scoped mode on a cluster with existing Atlas-managed nodes, those nodes are not automatically tainted. You must update existing Pods with the appropriate toleration before the nodes are replaced or recreated with the scoped-autoscaler taint.
</Callout>

## Identify which nodes are managed by Atlas

You can identify which nodes are managed by the autoscaler by checking for the Atlas management label:

```shell
kubectl get nodes -l provisioner.atlas.example/managed-by=atlas.example
```

Nodes without this label will be excluded from autoscaling operations when running in scoped mode.

## Verify scoped mode operation

To verify that scoped mode is working correctly:

1. **Check autoscaler policy configuration:**

   ```shell
   curl -H "X-API-Key: $ATLAS_API_TOKEN" \
     "$ATLAS_API_URL/cluster-autoscaler/v2/clusters/$CLUSTER_ID/policies"
   ```

   Look for `"scopedMode": true`.

2. **Verify nodes are tainted:**
   ```shell
   kubectl get nodes -l provisioner.atlas.example/managed-by=atlas.example -o json | \
     jq '.items[].spec.taints[] | select(.key=="scheduling.atlas.example/scoped-autoscaler")'
   ```

## Disable scoped mode

When disabling scoped mode, the `scheduling.atlas.example/scoped-autoscaler:NoSchedule` taint is not automatically removed from the `default-by-atlas` node template. You must manually remove the taint after disabling scoped mode.

<Tabs>
  <Tab title="API">
    ```json
    {
      "nodeDownscaler": {
        "enabled": true
      },
      "isScopedMode": false
    }
    ```
  </Tab>

  <Tab title="Terraform">
    ```toml
    resource "atlas_autoscaler" "atlas_autoscaler_policy" {
      cluster_id = atlas_eks_cluster.test.id
      
      autoscaler_settings {
        enabled        = true
        is_scoped_mode = false
        
        node_downscaler {
          enabled = true
        }
      }
    }
    ```
  </Tab>
</Tabs>

After disabling scoped mode, manually remove the taint from your node templates to restore normal scheduling behavior.

<Callout icon="⚠️" theme="warn">
  ### Do not remove the scoped-autoscaler taint manually

  Manually removing the taint on the `default-by-atlas` node template without disabling scoped mode risks improper node downscaling, potentially deleting non-Atlas nodes.
</Callout>

## Combine scoped mode with Evictor

Scoped mode can be used independently for the autoscaler or in combination with Evictor's scoped mode.

* **Autoscaler scoped mode only**: Autoscaler provisions and downscales only Atlas-managed nodes, but Evictor can bin-pack workloads across all nodes
* **Evictor scoped mode only**: Evictor bin-packs only Atlas-managed nodes, but the autoscaler can provision nodes for any workload and downscale any node
* **Both in scoped mode**: Complete isolation where both autoscaler and Evictor only affect Atlas-managed nodes

**For Evictor** (via Helm):

<Callout icon="📘" theme="info">
  ### Configuration-only changes

  This command upgrades all components in the umbrella chart, not just Evictor. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.
</Callout>

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-evictor.dryRun=false \
  --set autoscaler.atlas-evictor.scopedMode=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-evictor.dryRun=false \
  --set autoscaler-anywhere.atlas-evictor.scopedMode=true
```

  </Tab>
</Tabs>

For more information on Evictor's scoped mode and configuration, see <Anchor target="_blank" href="doc:evictor">Evictor documentation</Anchor>.