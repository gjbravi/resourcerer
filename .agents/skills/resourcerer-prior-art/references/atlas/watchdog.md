---
updatedAt: 2026-08-11T15:43:16.000Z
---

# Watchdog

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

# What is Watchdog?

**Watchdog** is a fallback mechanism built for Atlas-managed Kubernetes clusters to ensure **critical workload availability during Atlas outages**. Watchdog mitigates the risk of Atlas outages by automatically provisioning **emergency compute capacity** during such events.

# Why use Watchdog?

In rare cases when Atlas is temporarily unavailable due to network issues, platform downtime, or misconfigurations, workloads might remain pending. Watchdog ensures your applications remain available by:

* Detecting Atlas outages
* Watching pending pods requiring scheduling
* Provisioning fallback nodes through cloud provider APIs
* Enabling workload recovery independently of the Atlas SaaS control plane

Watchdog is specifically designed to mitigate **application availability risks**, not cost-optimization concerns.

# How it works

When Atlas cannot provision nodes (during an outage), Watchdog:

* Bypasses placement preferences like node selectors, affinity rules, and taints/tolerations
* Selects **On-Demand** instances only (no Spot or GPU)
* Uses fallback node pools created ahead of time by Atlas

Pods scheduled this way are prioritized for **availability**, not cost.

## Key behaviors

* **Only affects new pods created during an outage** – running pods are unaffected.
* **Works only with x86 nodes** – workloads must have x86-compatible images.
* **No GPU or Spot nodes supported**.
* **Fallback node pools** are automatically created by Atlas after Watchdog is enabled.

# Installation

> 🚧 Limited Availability Feature
>
> This feature is currently available through feature flags. Contact us to enable access for your organization.

## Prerequisites

* **Namespace**: Watchdog must be installed in the `atlas-agent` namespace. Custom namespaces are not supported. The standard Helm chart injects required GCP environment variables (`GOOGLE_PROJECT`, `GOOGLE_CLUSTER_LOCATION`, `GOOGLE_CLUSTER_NAME`) into the Watchdog pod; deploying to a different namespace with custom manifests can result in these variables being missing or misconfigured.
* **Workload Identity** must be configured for the `atlas-watchdog` Service Account in the `atlas-agent` namespace.
* **Required GCP IAM Permissions**:
  * `compute.instanceGroupManagers.get`
  * `compute.machineTypes.get`
  * `container.clusters.get`
  * `container.clusters.update`
  * `container.operations.get`

## Enable fallback node pools

Contact **Atlas support** to enable Watchdog for your cluster or organization. This will automatically create fallback node pools in your GKE environment for Watchdog to use.

## API Key

Watchdog requires a Atlas API key. The key from `atlas-cluster-controller` Secret can be reused, if desired.

You can also use a [Atlas Service Account](service-accounts.md) key with `Viewer` role for the Watchdog API key.

<Callout icon="📘" theme="info">
  ### Note

  If you see an 403 error in the Watchdog logs with the current API key, you might need to generate a new API key. The Watchdog API key requires a new permission which was added in July 2025.
</Callout>

## Install using Helm

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts

helm upgrade --install atlas-watchdog atlas-helm/atlas-watchdog \
  -n atlas-agent \
  --set atlas.apiKeySecretRef=atlas-cluster-controller \
  --set atlas.clusterID=<cluster-id> \
  --set atlas.organizationID=<organization-id> \
  --set gcp.project=<gcp-project-id> \
  --set gcp.clusterName=<gke-cluster-name> \
  --set gcp.location=<gke-cluster-location>
```

## Simulating an outage

To verify that the Watchdog is working properly, you can simulate an outage on your side by scaling down the `atlas-agent` Deployment to zero pods. After a few minutes, Atlas will mark the cluster as unhealthy, Watchdog will see a failing health check, and start provisioning emergency compute capacity for pending pods.

1. Scale down the `atlas-agent` Deployment to zero pods:

   ```shell
   kubectl scale -n atlas-agent deployment atlas-agent --replicas 0
   ```
2. Create a pending pod, which will verify that the Watchdog strips node constraints from the pod and provisions a node for it:
   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     labels:
       app: watchdog-test
     name: watchdog-test
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: watchdog-test
     template:
       metadata:
         labels:
           app: watchdog-test
       spec:
         nodeSelector:
           watchdog: testing
         containers:
         - image: nginx
           name: nginx
           resources:
             requests:
               cpu: 1
   EOF
   ```
3. Wait around 5–7 minutes, till the Watchdog detects the failing health check and starts working. You should see the following log line in the Watchdog:<br />`starting autoscaling`<br />The pending pod should be recreated, stripped from the node selector, and a new node should be created.

# Webhook Notifications

Watchdog supports sending HTTP Webhook Notifications when it starts and stops working. You can provide a Go template for the payload to support different receivers (Slack, Discord, etc.).

Add the following values to the Watchdog Helm chart to set up a webhook notification:

```yaml
webhooks:
  - url: https://hooks.slack.com/services/...
    method: POST
    headers:
      x-header-name: header-value
    payloadTemplate: |
      {{- $e := .Event -}}
      {
        "attachments": [
          {
          {{- if eq $e.Type "StartAutoscaling" }}
            "color": "#a63636",
            "title": "Watchdog started to autoscale",
            "text": "Watchdog detected Atlas is not available and started working.",
          {{- else if eq $e.Type "StopAutoscaling" }}
            "color": "#36a64f",
            "title": "Watchdog stopped to autoscale",
            "text": "Watchdog detected that Atlas is working again.",
          {{- else }}
            "color": "#cccccc",
            "title": "{{ $e.Type }}",
          {{- end }}
            "fields": [
              {
                "title": "Time",
                "value": "{{ $e.Timestamp.Format "2006-01-02 15:04:05 MST" }}",
                "short": true
              },
              {
                "title": "Reason",
                "value": "{{ $e.Reason }}",
                "short": false
              },
              {
                "title": "Cluster",
                "value": "<your-cluster>",
                "short": false
              }
            ]
          }
        ]
      }
```

# Limitations

* **Compatible only with Atlas Node Autoscaling** — disable any other cluster autoscaler (e.g. GKE cluster autoscaler, Karpenter, Cluster Autoscaler) on clusters managed by Atlas.
* Only supports GKE clusters (with fallback node pool creation).
* No GPU or Spot node provisioning.
* Only provisions **On-Demand** compute.
* Ignores all workload placement preferences (labels, affinities, taints).
* Supports **x86** nodes only (workloads and container images must be compatible).

## Opt-out labels

Pods can be excluded from Watchdog by setting the `watchdog.atlas.example/ignore: "true"` label on a pod or its namespace. Pods with this label will remain pending until either existing capacity becomes available or the Atlas autoscaler resumes activity.

# Troubleshooting

## Verify Watchdog deployment

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-watchdog
```

## Check logs

```shell
kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-watchdog
```

## Common issues

### "Requested entity was not found" when listing fallback node pools

```text
Error: checking permissions: health check failed: unable to get fallback pools:
failed to list node pools: rpc error: code = Unknown desc = Requested entity was not found.
```

This error occurs when Watchdog cannot resolve the GKE cluster to retrieve its fallback node pools. The most common cause is deploying Watchdog to a **custom namespace** instead of the required `atlas-agent` namespace.

The standard Helm chart injects `GOOGLE_PROJECT`, `GOOGLE_CLUSTER_LOCATION`, and `GOOGLE_CLUSTER_NAME` environment variables into the Watchdog pod. When deploying with custom manifests to a different namespace, these variables may be missing or misconfigured, causing the GKE API call to fail.

**To resolve:**

Redeploy Watchdog to the `atlas-agent` namespace using the standard Helm install command in the [Installation](#install-using-helm) section.

### Other issues

* **No fallback node pool present** – check if Atlas support has enabled Watchdog.
* **Permissions error** – verify GCP IAM roles and Workload Identity binding are set up properly.
* **Pending pods not handled** – ensure Watchdog is running, check the logs.

For additional help, contact Atlas support or visit our [community Slack channel](https://atlas-community.slack.com/).