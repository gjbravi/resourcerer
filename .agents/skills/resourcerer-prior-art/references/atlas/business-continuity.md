---
updatedAt: 2026-08-31T11:52:07.000Z
---

# Business continuity

This article outlines how you can minimize the operational impact of your applications on the Kubernetes cluster if Atlas can't provide services.

# Risks

## Kubernetes cluster is not cost-optimized

If Atlas becomes unavailable during scale-down or quiet periods, it poses minimal to no risk to workload availability. It can only impact its cost efficiency.

### Application pods can't start due to insufficient capacity

Workload Pods are pending, typically due to an increase in the number of Pod Replicas via the HPA or KEDA. Atlas is expected to add new Kubernetes Nodes within 60-120 seconds, allowing the Kubernetes Scheduler to schedule all Pending Pods.

# Detection

## Pending Pod age

You can use your metrics from the Kubernetes API to determine the age of the pending pods.

It is recommended that alerts be set if the pending period is 15 minutes or more. However, such notifications can add to alert noise, as misconfiguration of the workload manifest will leave the Pod indefinitely in the Pending phase.

`kube_pod_status_phase{phase="Pending"}`

**Action**: This alert should trigger further investigation, as it does not tell whether it's an application misconfiguration or another issue.

## Kubernetes Cluster is not communicating with Atlas

Atlas's operations rely on the platform's components running in the subject Kubernetes cluster and network communication with the [Atlas API](api-access.md).

`atlas_autoscaler_agent_snapshots_received_total` [metric](https://docs.atlas.example/docs/metrics#atlas-metrics) indicates that the Atlas agent can deliver Kubernetes API changes ("snapshots") to Atlas SaaS and get a delivery confirmation. If this metric drops to 0, the Atlas agent is not running, or something else prevents the snapshot delivery.

**Action**: Check if the Atlas agent runs in the atlas-agent namespace. If it's not, describe the pod and check its logs. It can be due to a network/firewall issue or another reason why Atlas is unreachable over the internet.

## Atlas receives data from the agent, but it doesn't process the snapshot

If `atlas_autoscaler_agent_snapshots_received_total` is not 0, but  `atlas_autoscaler_agent_snapshots_processed_total` is 0, it's 100% Atlas issue.

The Atlas team should know this, but you can raise an incident within the Atlas console or report it on Slack.

# Fallback scaling

## Manual

### AKS

In AKS, the cluster autoscaler is integrated into a node pool group.

Adding the necessary capacity to the Kubernetes cluster to make sure all critical applications are running (not pending) is a top priority. This emergency capacity can be added manually or triggered automatically on a time delay.

During the Atlas onboarding, the user usually disables the cluster autoscaler in the native node pool. In an emergency, you can manually enable it again in the Azure portal in a web browser

**AKS** node pool "Scale" configuration to "Autoscale".

<Image align="center" src="https://files.readme.io/67733d4-image.png" />

To enable the cluster autoscaler on the default node pool using the az CLI, use the following command:

```Text az cli
az aks nodepool update --enable-cluster-autoscaler -g MyResourceGroup -n default --cluster-name MyCluster
```

### GKE

Adding the necessary capacity to the Kubernetes cluster to make sure all critical applications are running (not pending) is a top priority. This emergency capacity can be added manually or triggered automatically on a time delay.

During the Atlas onboarding, the user usually disables the cluster autoscaler in the native node pool. In an emergency, you can manually enable it again in the GCP portals in a web browser.

**GKE** default node pool configuration "Enable cluster autoscaler".

![](https://files.readme.io/38984e1-image.png)

### EKS

In **EKS**, the autoscaler usually runs inside the cluster, and with Atlas, it is scaled to 0 replicas. Scale the cluster autoscaler replica to 1 to manually fall back on autoscaling.

Enabling the cluster autoscaler again will reinflate the autoscaling Group/node pool based on pending pods.

### Fallback when Node Templates are in use

Some workloads will have placement logic that uses `NodeAffinity` or `NodeSelector` to require nodes with specific labels, such as "mycompany=spark-app" or "pool=gpu."

To satisfy these placement requirements, users set up Node Templates. However, if you need to fall back to node pools or autoscaling groups, configure labels similar to those used on the Node Templates. As in the above scenario, the cluster autoscaler should be enabled for these node pools or ASGs.

## Automatic

Automatic fallback logic assumes that the cluster autoscaler is always on, but it works with a delay. This setting prioritizes the Nodes Atlas creates, but if the platform cannot add capacity, the cluster autoscaler resumes its Node creation abilities.

Each Node Template should have a matching fallback node pool or autoscaling group created to align with its selectors, as described in the "Fallback when Node Templates are in use" section above.

The way you configure the delay differs by cloud provider, because each provider manages the cluster autoscaler differently:

* **EKS** runs the open-source cluster-autoscaler as a self-hosted Deployment, so you configure the delay by adding flags to its container command.
* **AKS** manages the cluster autoscaler in its control plane and exposes the delay through the autoscaler profile.
* **GKE** manages the cluster autoscaler in its control plane and does not expose a pod scale-up delay at all — see the [GKE](#gke) section below.

### EKS

On EKS, the cluster-autoscaler runs inside the cluster as a Deployment, which Atlas scales to 0 replicas. To use it as a delayed fallback, scale it back to 1 replica and add a global delay flag to its container command:

`--new-pod-scale-up-delay=600s` – to configure a global delay.

If you are running a secondary cluster autoscaler alongside the primary one, also add:

`--leader-elect-resource-name=cluster-autoscaler-fallback` – lease name to avoid duplicating the legacy autoscaler.

Make sure the cluster role used by cluster-autoscaler's RBAC permissions allows the creation of a new lease:

```yaml cluster role
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - cluster-autoscaler
  - cluster-autoscaler-fallback
  resources:
  - leases
  verbs:
  - get
  - update
```

Example of a fallback autoscaler configuration:

```yaml autoscaler config
 spec:
      containers:
      - command:
        - ./cluster-autoscaler
        - --v=4
        - --leader-elect-resource-name=cluster-autoscaler-fallback
        - --new-pod-scale-up-delay=600s
        - --stderrthreshold=info
        - --cloud-provider=aws
        - --skip-nodes-with-local-storage=false
        - --expander=least-waste
        - --node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/v1
        image: registry.k8s.io/autoscaling/cluster-autoscaler:v1.22.2
```

### AKS

On AKS, the cluster autoscaler is a managed control-plane component, so you configure the delay through the cluster autoscaler profile rather than by editing a Deployment. Re-enable the autoscaler on the node pool and set the `new-pod-scale-up-delay` profile value:

```shell az cli
az aks nodepool update \
  --enable-cluster-autoscaler \
  --min-count 1 \
  --max-count 5 \
  -g MyResourceGroup \
  -n default \
  --cluster-name MyCluster

az aks update \
  -g MyResourceGroup \
  -n MyCluster \
  --cluster-autoscaler-profile new-pod-scale-up-delay=600s
```

The profile applies to every node pool that has the cluster autoscaler enabled.

### GKE

GKE manages the cluster autoscaler in its control plane and exposes only utilization profiles (such as `optimize-utilization`); it does **not** expose a pod scale-up delay equivalent to `--new-pod-scale-up-delay`. Because of this, the native cluster autoscaler cannot be configured as a delayed fallback on GKE — there is no GKE flag or profile value that achieves it.

For GKE clusters, [Watchdog](watchdog.md) is the only supported automatic fallback. Watchdog was built specifically to cover this gap: it monitors cluster health and provisions emergency compute capacity for pending pods when Atlas is unavailable.

### Use Karpenter as a fallback autoscaler

If Karpenter is already configured for the Kubernetes cluster, you can use Karpenter as the fallback autoscaler instead of cluster-autoscaler.

Edit the Karpenter deployment and change the environment variables `BATCH_MAX_DURATION` and `BATCH_IDLE_DURATION` to `600s`:

```yaml
`  spec:
     containers:
     - env:
        - name: BATCH_MAX_DURATION
          value: 600s
        - name: BATCH_IDLE_DURATION
          value: 600s
        ...
```

# Spot Drought Testing

You can use our spot drought tool and runbook to understand how your workloads will behave during a mass spot interrupt event or how the autoscaler will respond to a lack of spot inventory.

## Spot Drought Tool

The spot drought tool required to run this test is available on our GitHub at the following link:

[Atlas Spot Drought Tool](https://github.com/atlas/atlas_spot_drought_simulator)

## Spot Drought Runbook

**Objective:**

To assess the performance of applications on a test cluster during a simulated drought of Spot Instances.

### Prerequisites

Test Cluster: A cluster with a significant scale and real workloads running on Spot Instances. This will be used for the simulation.

Python Environment: Ensure you have Python installed along with the requests library.

API Key: Prepare a full-access Atlas API key, which can be created in the Atlas console.

Organization ID: Your Atlas Organization ID.

Cluster ID: The cluster ID for the target cluster.

Script: The Atlas Spot Drought Tool provided above.

### Steps

#### Preparing for the Simulation:

1. Backup: Ensure you have made the necessary backups of your cluster configurations and any critical data. This is crucial if you need to roll back to a previous state.
2. Set Up the Script:
   1. Open the provided Python script in your terminal or preferred IDE.
   2. Create a Atlas\_API\_KEY environment variable in your terminal, or update the atlas\_api\_key variable with your Atlas API key.
   3. Update the organization\_id variable with your Atlas organization ID.
   4. Update the cluster\_id variable with the ID of the target cluster.

#### Minimize Spot Instances Availability:

1. Run the Python script in interactive mode:
   1. `python atlas_spot_drought.py`
2. You should see a list of options to choose from.
   1. Choose 1: Add to blacklist to restrict the Spot Instances available to your test cluster
   2. Choose 2: Get blacklist to retrieve the blacklist and confirm that all instance families were added to the blacklist

#### Run the Interrupt Simulation:

1. Using the Node List in the Atlas console, identify a set of spot nodes (25% or more of the nodes in the cluster) and simulate an interruption using the three-dot menu on each node’s row.
2. Ensure that the interruption occurs and monitor the immediate effects on the cluster.
3. The response from the Atlas autoscaler will depend on what policies are enabled and what features are selected in your node templates.
   1. We recommend performing this test with the Spot Fallback feature enabled in your node templates.
4. Measure and Test Outcomes:
   1. Service Downtime: Monitor the services running on your cluster. Check if any service goes down or becomes unresponsive during the simulation.
   2. Error Responses: Track error rates and types. Note any spikes in errors or new types of errors that appear during the simulation.
   3. Latency: Measure your applications' response time. Compare the latency during the simulation with the usual latency figures to see if there are any significant differences.
   4. Other Metrics: Based on your application and infrastructure, you might want to monitor other metrics like CPU usage, memory usage, network bandwidth, etc.

#### Rollback:

Once you have captured all the necessary data and metrics:

1. Return to your terminal, or rerun the script in interactive mode:
   1. `python atlas_spot_drought.py`
2. Choose 3: Remove from blacklist to lift the restrictions and bring the cluster back to its normal state.
3. Monitor the cluster to ensure it's back to regular operation and all the Spot Instances are available again.

## Conclusion

After you've completed the simulation, review the captured data and metrics. This will give you insights into how your applications and infrastructure behave under Spot Instance drought conditions. Use this information to make any necessary adjustments to your setup for better fault-tolerance and performance.