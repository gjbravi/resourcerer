---
updatedAt: 2026-06-30T12:10:47.000Z
---

# Cluster hibernation (Legacy)

> 🚧 Legacy Feature
>
> This page describes a legacy method for pausing clusters using CronJobs. While this approach continues to be supported for existing implementations, we recommend using our improved <Anchor label="Cluster Hibernation" target="_blank" href="doc:autoscaling-cluster-hibernation">Cluster Hibernation</Anchor> feature for new implementations.

Atlas provides a set of CronJobs that can be used to pause and resume Kubernetes clusters on a defined schedule. When executed, Atlas components will continue to run on a defined single node while the rest of the cluster capacity will be removed. Once the cluster is set to resume, Atlas will use standard Autoscaler capabilities to provide the most cost-efficient nodes to run pending pods.

# How It Works

The pausing mechanism consists of two main CronJobs.

## Hibernate-pause Job

* Disables the Unscheduled Pod Policy (to prevent growing the cluster)
* Prepares a Hibernation node (node that will stay hosting essential components)
* Marks essential Deployments with Hibernation toleration
* Deletes all other nodes (only the hibernation node will stay running)

## Hibernate-resume Job

* Reenables the Unscheduled Pod Policy to allow the cluster to expand to the needed size

# Installation

## Install via `kubectl`

Run this command to install the Hibernate CronJobs:

```shell
kubectl apply -f https://raw.githubusercontent.com/atlas/hibernate/main/deploy.yaml
```

### Configure API Key

Create an API token with Full Access permissions and encode it in `base64`:

```shell
echo -n "your-atlas-api-key" | base64
```

Use this value to update the Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: atlas-hibernate
  namespace: atlas-agent
type: Opaque
data:
  API_KEY: >-
    Atlas-API-KEY-REPLACE-ME-WITH-ABOVE==
```

Alternatively, use this one-liner for convenience:

```shell
kubectl get secret atlas-hibernate -n atlas-agent -o json | jq --arg API_KEY "$(echo -n your-atlas-api-key | base64)" '.data["API_KEY"]=$API_KEY' | kubectl apply -f -
```

## Install with Helm

Add Atlas helm charts repository:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

Install the hibernate component (update `cloud` and `apiKey` variables):

```shell
helm upgrade -i atlas-hibernate atlas-helm/atlas-hibernate -n atlas-agent --set cloud=<AKS|EKS|GKE> --set apiKey=<Atlas-API-KEY-REPLACE-ME-WITH-BASE64_ENCODE>
```

# Configuration

## Scheduling CronJobs

Update the hibernate-pause and hibernate-resume CronJob schedules according to your business needs. The schedules follow the [Kubernetes CronJob syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax).

**Default examples:**

```yaml
# Update hibernate-pause schedule according to business needs
pauseCronSchedule: "0 22 * * 1-5"

# Update hibernate-resume schedule according to business needs
resumeCronSchedule: "0 7 * * 1-5"
```

Beginning with Kubernetes v1.25 and later versions, you can define a time zone for a CronJob by setting a valid time zone name to `.spec.timeZone`. For example, setting `.spec.timeZone: "Etc/UTC"` will interpret the schedule according to Coordinated Universal Time (UTC). You can find a list of valid time zones in the [tz database time zones list](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).

## Advanced Configuration Options

The pause and resume jobs support several configuration options through environment variables:

### Override default hibernate-node size

Set the `HIBERNATE_NODE` environment variable to override the default node sizing selections. Make sure the size selected is appropriate for your cloud provider.

### Specify namespaces to preserve

Set the `NAMESPACES_TO_KEEP` environment variable to override the default preserved namespaces (defaults to "opa,istio"). The namespaces listed here will not be affected by the hibernation process.

### Protect nodes marked as removal-disabled

Set `PROTECT_EVICTION_DISABLED` to "true" to prevent the removal of nodes that have the `autoscaling.atlas.example/removal-disabled="true"` label. This is useful for preserving nodes running critical or stateful workloads.

### Set API region

If your cluster is not in the US region, provide your regional API URL via the environment variable:

```yaml
API_URL: "<your-regional-api-url>"
```

See [API access](api-access.md) for regional endpoint URLs.

# Advanced Use Cases

## Selective Namespace Hibernation

If you need to pause and resume specific namespaces rather than the entire cluster, you can:

1. Set the `NAMESPACES_TO_KEEP` environment variable to specify which namespaces should be preserved during hibernation
2. All other namespaces will be affected by the hibernation process

Example configuration:

```yaml
env:
- name: NAMESPACES_TO_KEEP
  value: "kube-system,monitoring,logging,istio-system"
```

## Time-sharing GPU Resources

For clusters with GPU resources that you want to time-share among different teams or applications:

1. Configure the hibernate CronJobs to run on a schedule that matches your time-sharing requirements
2. Set `NAMESPACES_TO_KEEP` to include the namespaces that should always have access to the resources
3. Configure separate hibernate CronJobs for different time slots if needed

For more fine-grained control over scaling specific workloads (rather than the entire cluster), you can combine this approach with [kube-downscaler](https://codeberg.org/hjacobs/kube-downscaler). When using `kube-downscaler`:

1. Configure it to scale specific deployments to zero pods during off-hours
2. Enable Atlas's Evictor in aggressive mode to ensure the cluster shrinks after pods are removed
3. This approach provides more granular control than completely hibernating the cluster

# Upgrading

To upgrade the hibernate component to the latest version, run:

```shell
helm repo update atlas-helm
helm upgrade atlas-hibernate atlas-helm/atlas-hibernate --reset-then-reuse-values -n atlas-agent
```

# Troubleshooting

## Check the logs

To view the logs of the hibernate jobs:

```shell
kubectl logs -l app=atlas-hibernate-pause -n atlas-agent
kubectl logs -l app=atlas-hibernate-resume -n atlas-agent
```

## Common issues

1. **Jobs not running at scheduled time**: Check the CronJob schedules and ensure they are properly defined according to your timezone
2. **Nodes not being removed**: Verify that there are no pods preventing node draining, or check if nodes have the `autoscaling.atlas.example/removal-disabled` label
3. **Cluster not resuming properly**: Check if the Unscheduled Pod Policy is enabled in the Atlas console

For persistent issues, contact Atlas support.