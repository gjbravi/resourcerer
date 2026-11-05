---
updatedAt: 2026-07-15T14:05:09.000Z
---

# Spot Handler

Spot Handler monitors Spot Instance interruption events from your cloud provider and reports them to the Atlas platform.

## What Spot Handler does

Spot Handler monitors Spot Instance interruption signals from your cloud provider and sends this data to Atlas. This data helps improve Atlas's Spot reliability and interruption prediction models.

**Important:** Spot Handler does not take any action on nodes or workloads. It does not:

* Drain nodes when receiving interruption signals
* Initiate pod rescheduling
* Perform graceful shutdowns

## How Spot interruptions are handled

When a Spot Instance is interrupted by your cloud provider:

1. **Spot Handler** detects the interruption signal and reports it to Atlas
2. The cloud provider terminates the node according to their policies (e.g., GCP provides 30 seconds before termination)
3. Kubernetes handles pod termination using standard procedures
4. Atlas's Autoscaler provisions replacement capacity

For proactive interruption mitigation before nodes are interrupted, enable these Autoscaler features:

* **<Anchor label="Spot reliability" target="_blank" href="doc:spot#spot-instance-reliability">Spot reliability</Anchor>**: Targets instance types less likely to be interrupted
* **<Anchor label="Interruption prediction model" target="_blank" href="doc:spot#interruption-prediction-model">Interruption prediction model</Anchor>**: Gracefully drains and rebalances nodes predicted for interruption

## Install Spot Handler

Spot Handler is installed automatically during Phase 1 cluster onboarding using the Helm chart available at <Anchor label="https://github.com/atlas/helm-charts/tree/main/charts/atlas-spot-handler" target="_blank" href="https://github.com/atlas/helm-charts/tree/main/charts/atlas-spot-handler"><https://github.com/atlas/helm-charts/tree/main/charts/atlas-spot-handler></Anchor>. This ensures that Atlas can begin collecting Spot interruption data immediately after connecting your cluster, improving Spot reliability predictions from the start.

If Spot Handler was uninstalled or you need to reinstall it manually, you can do so using the following commands.

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

<Helm_v4_error />

1. Add the Atlas Helm charts repository:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

2. Install Spot Handler:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set global.atlas.apiKey=<your-api-token> \
  --set global.atlas.provider=<eks|gke|aks>
```

## Upgrade Spot Handler

### For Operator-managed installations

If Spot Handler is managed by the Atlas Operator, upgrades are handled automatically through the Atlas console:

1. Navigate to **Manage Org** >**Component Control** in the Console
2. Select **Spot Handler**
3. Click **Update** when a new version is available

No additional steps are required.

### For manual (Helm) installations

To upgrade Spot Handler to the latest version manually:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas --reset-then-reuse-values -n atlas-agent
```

## Improve Spot interruption handling

To minimize the impact of Spot interruptions:

1. **Design for high availability**: Use multiple replicas and distribute pods across failure domains
2. **Enable Spot handling mechanisms**: Configure <Anchor label="Spot reliability" target="_blank" href="doc:spot#/spot-instance-reliability">Spot reliability</Anchor> and <Anchor label="interruption prediction" target="_blank" href="doc:spot#/interruption-prediction-model">interruption prediction</Anchor> in your Node templates
3. **Configure Pod Disruption Budgets**: Set appropriate PDBs for critical workloads
4. **Use preStop hooks**: Configure <Anchor label="lifecycle hooks" target="_blank" href="https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/">lifecycle hooks</Anchor> in your containers for graceful shutdown

For more information, see <Anchor label="Spot Instances" target="_blank" href="doc:spot">Spot Instances</Anchor>.

## Troubleshooting

Check Spot Handler logs:

```shell
kubectl logs -l name=spot-handler -n atlas-agent
```

<br />