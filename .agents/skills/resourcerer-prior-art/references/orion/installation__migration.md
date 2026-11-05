# Migrate from v2 to v3

> Migrate from separate Agent and Applier Helm charts to a unified Orion chart


With v3, the separate `orion-agent` and `orion-applier` Helm charts are combined into a single `orion` chart.

## Prerequisites

Before you begin, ensure you have:

- [The Orion CLI installed](installation__install-adv.md#set-up-the-orion-cli)
- Your current v2 values (see [Convert your values](#convert-values))
- `clusterName` and your API credentials (`authorization.clientID` and `authorization.clientSecret`)

## Webhook considerations {#webhook-considerations}

The v3 chart's `MutatingWebhookConfiguration` reuses the name of the v2 Applier webhook. This means:

- **Helm-managed installs**: If a v2 Applier is still installed when you run `helm install orion`, Helm's release-ownership check will refuse to install.
- **GitOps users (ArgoCD, Flux, kubectl apply)**: There is no automatic conflict detection. You **must** fully remove the v2 Applier before applying the v3 manifests. Failing to do so will result in two conflicting `MutatingWebhookConfiguration` resources.

## CRD considerations {#crd-considerations}

When you uninstall the v2 Agent, the `optimize.orion.example` CustomResourceDefinitions (CRDs) — and all associated custom resources, such as optimization configurations — remain on your cluster. This is expected, as `helm uninstall` doesn't remove CRDs. **Don't delete them manually.** Deleting a CRD removes every associated custom resource from the cluster.

In step 4 of the migration, you'll use Helm to transfer ownership of the CRDs to the v3 release.

**GitOps users (ArgoCD, Flux, kubectl apply)**: No action is needed. There is no release-ownership check when applying rendered manifests; the existing CRDs are updated in place.

## Migration steps

<div class="alert alert-warning" role="alert">


There will be a brief period of downtime during the migration. Plan accordingly.
</div>


### 1. Convert your values {#convert-values}

Convert your v2 values into a single v3 `values.yaml`. At least one of `--agent-values` and `--applier-values` is required — omit `--applier-values` if the Applier isn't installed.

```sh
orion migrate helm-values \
  --agent-values ./agent-values.yaml \
  --applier-values ./applier-values.yaml \
  --output ./values.yaml
```

The command writes the converted values to the output file and reports any keys that require manual attention.

Review the output file and address any warnings before proceeding.

<div class="alert alert-info" role="alert">


If you don't have values files (for example, if you set your values using `--set`), export the configuration for each component using `helm get values`. If you don't have the Applier installed, skip its command. Don't use `-a`/`--all`, as this includes default values:

```sh
helm get values orion-agent -n orion-system -o yaml > agent-values.yaml
helm get values orion-applier -n orion-system -o yaml > applier-values.yaml
```

If `applier-values.yaml` contains only `null`, you did not customize that component, so you can ignore it.

</div>


### 2. Confirm required values {#required-values}

Your `values.yaml` must include `clusterName` and your API credentials. Values exported with `helm get values` already contain them. Add them if you started from a fresh file. At minimum:

```yaml
clusterName: my-cluster
authorization:
  clientID: CLIENT_ID
  clientSecret: CLIENT_SECRET
```

### 3. Uninstall the Applier and Agent

Remove the Applier (if it's installed), then the Agent:

```sh
helm uninstall orion-applier -n orion-system
helm uninstall orion-agent -n orion-system
```

### 4. Install the v3 unified chart {#complete-the-migration}

The `--take-ownership` flag (Helm 3.17 or later) lets Helm take ownership of the [CRDs left in place by the v2 Agent](#crd-considerations):

```sh
helm install orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  --create-namespace \
  --take-ownership \
  --values values.yaml
```

This installs the Agent, Forwarder, and Applier in a single chart, replacing the separate v2 installations.

If your Helm version doesn't support `--take-ownership`, label and annotate the existing CRDs to match the v3 release before installing:

```sh
for crd in workloadoptimizers optimizationconfigurations clusteroptimizationconfigurations; do
  kubectl label crd ${crd}.optimize.orion.example \
    app.kubernetes.io/managed-by=Helm --overwrite
  kubectl annotate crd ${crd}.optimize.orion.example \
    meta.helm.sh/release-name=orion \
    meta.helm.sh/release-namespace=orion-system --overwrite
done
```

## If migration fails

Migration issues are typically environment-specific and may require manual intervention. If you run into problems, [contact Orion support](https://www.orioncorp.io/company/contact-us/).

## Values reference {#values-mapping}

The `orion migrate helm-values` command handles these conversions automatically. These tables are a reference for what changed between v2 and v3.

### Renamed values

| v2 key | v3 key |
|---------|---------|
| `authSecret` | `authSecretName` |
| `envSecret` | `envSecretName` |
| `commonMetaLabels` | `commonLabels` |
| `manageAuthSecret` | `createAuthSecret` |
| `rbac.additional` | `clusterAgent.rbac.additionalRules` |
| `rbac.*` (all other keys) | `clusterAgent.rbac.*` |
| `serviceAccount.*` | `clusterAgent.serviceAccount.*` |
| `workload.allowNamespaces` | `clusterAgent.allowNamespaces` |
| `workload.denyNamespaces` | `clusterAgent.denyNamespaces` |
| `workload.inactiveWorkloadsGC` | `clusterAgent.garbageCollectInactiveWorkloads` |
| `prom.extraScrapeConfigs` | `forwarder.additionalScrapeConfigs` |
| `workloadResourceTypes` | Merged into `workloadResourceKinds` |

### Removed values

| v2 key | Action in v3 |
|---------|---------------|
| `clusterDefaultConfig` | Deprecated, but still accepted as a migration aid — prefer a [ClusterOptimizationConfiguration resource](configure__optimization-configurations.md) |
| `featureGates.optimizeCronJobs` | Removed — an enabled gate is converted to the equivalent `workloadResourceKinds` entry |
| `featureGates.optimizeSparkOperator` | Removed — an enabled gate is converted to the equivalent `workloadResourceKinds` entry |
| `featureGates.optimizeGithubARC` | Removed — an enabled gate is converted to the equivalent `workloadResourceKinds` entry |
| `patchTargetTypes` | Removed — use `workloadResourceKinds` |
| `enableOptimizationConfigurationCRDs` | Always-on |
| `extraEnvVars` | Removed — was a deprecated alias; use `additionalEnvVars` |

## Resource name changes

The following Kubernetes resource names changed in v3. Update any references to these names in monitoring, alerts, or scripts.

### Deployments

| v2 name | v3 name |
|----------|----------|
| `orion-agent-workload-controller` | `orion-cluster-agent` |
| `orion-agent-metrics-forwarder` | `orion-forwarder` |
| `{release}-orion-applier` (where `{release}` is your Helm release name) | `orion-applier` |

### Secrets

| v2 name | v3 name |
|----------|----------|
| `orion-agent-auth` | `orion-auth` |

## CI/CD automation updates

Update any CI/CD pipelines that reference the old chart URLs or release name:

| | v2 | v3 |
|---|---|---|
| **Agent chart URL** | `oci://registry.orion.example/library/orion-agent` | `oci://registry.orion.example/library/orion` |
| **Applier chart URL** | `oci://registry.orion.example/library/orion-applier` | Not needed — Applier is bundled |
| **Release name** | `orion-agent` | `orion` |
