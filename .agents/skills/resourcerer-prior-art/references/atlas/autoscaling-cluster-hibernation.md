---
updatedAt: 2026-09-18T08:48:48.000Z
---

# Cluster hibernation

Cluster hibernation allows you to optimize costs by temporarily scaling your cluster to zero nodes while preserving the control plane and cluster state. This feature is ideal for non-production environments, development clusters, or any cluster that doesn't need to run 24/7.

You can manage cluster hibernation through both the Atlas console, API, and Terraform, with options for manual hibernation and scheduled automation.

# How hibernation works

When you hibernate a cluster:

1. The cluster enters a "Hibernating" state
2. All nodes are systematically removed from the cluster
3. The control plane remains active, but no workloads run
4. Costs are minimized to just the control plane charges from your cloud provider

When you resume a hibernated cluster:

1. Resume nodes are created to run essential system components
2. Atlas's autoscaler is reactivated
3. Workloads are scheduled according to their requirements
4. The cluster returns to its normal operational state

Resume relies on pod priority to ensure Atlas components will be scheduled on the resume nodes. By default our components have the `system-cluster-critical` priority class.

The resume node needs to run critical Atlas components for the cluster to transition into the connected state. To ensure these components are scheduled during the resume process, Atlas may apply a `CriticalAddonsOnly` taint to the resume node. Whether the taint is applied depends on when your cluster was onboarded and whether its Atlas components include the corresponding toleration:

**Clusters onboarded after May 25, 2026:** The `CriticalAddonsOnly` taint is applied to the resume node. This also applies to clusters onboarded earlier whose Atlas components have been upgraded to include the `CriticalAddonsOnly` toleration.

**Clusters onboarded before May 25, 2026:** If your Atlas components do not have the `CriticalAddonsOnly` toleration, the resume node is created without any additional taint and no validation is performed. This is the legacy behavior. Even in this case, you are encouraged to add the `CriticalAddonsOnly` toleration to your Atlas components so that Atlas can apply the taint to the resume node during resumption. Without it, any pending workloads scheduled onto the resume node can compete with Atlas components for resources. If too many pending pods are scheduled ahead of them, the Atlas components required for a successful resumption may fail to be scheduled, causing the resume process to fail.

Whenever the `CriticalAddonsOnly` taint will be applied to the resume node — whether because the cluster was onboarded after May 25, 2026 or because its Atlas components have been upgraded to include the toleration — Atlas validates during hibernation that all known critical components required for the resume node to join the cluster in a `Ready` state (such as CNI, DNS, admission webhook controllers, and important daemonsets) carry the corresponding toleration. The validation results are recorded in the audit log as part of the hibernation job.

# Prerequisites

> 🚧 Important
>
> Current hibernation is not designed for large clusters. If your cluster is above 1000 nodes, cluster hibernation might fail. If you want to hibernate a larger cluster, first reduce the number of nodes in it.

Before using cluster hibernation, ensure your environment meets these requirements:

## Resume node size

Ensure the resume node size is enough to host Atlas agent, cluster-controller. If you use pod mutations and Workload Autoscaler, ensure also the pod-mutator and workload-autoscaler can fit on the resume node.
Keep in mind that also some other cluster critical components might start on the resume node and take capacity.

If you see that resume is failing and the Atlas components are not scheduled on the resume node, increase the node size.

## Cloud-specific requirements

### EKS clusters

EKS clusters require IAM permissions for hibernation operations.

**Clusters onboarded after October 2025:** The required IAM permissions are automatically included in the default permission set. No additional setup is required.

**Clusters onboarded before October 2025:** You need to update your IAM permissions using one of these methods:

* **Re-onboard via script:** Run the Phase 2 onboarding script again to update your Atlas IAM role with hibernation permissions.
* **Upgrade Terraform provider:** If you manage your cluster with Terraform, upgrade to [version 7.74.0](https://github.com/atlas/terraform-provider-atlas/releases/tag/v7.74.0) or later and re-apply your configuration.

#### EKS clusters using aws-auth ConfigMap authentication

EKS clusters where node IAM roles are granted cluster access via the `aws-auth` ConfigMap require additional configuration for cluster hibernation to work correctly.

**Background:** When EKS uses the `aws-auth` ConfigMap for node authentication, AWS automatically removes an IAM role from the ConfigMap when the last managed node group using that role is deleted. During cluster resume, Atlas creates a temporary managed node group. When this node group is later deleted, AWS removes its IAM role from `aws-auth`, which can break authentication for other Atlas-managed nodes still using that role.

**Requirement:** You must set `nodeGroupArn` in your **default** Node Configuration. Despite its name, this field specifies a dedicated IAM role ARN (not a node group ARN) that Atlas will use exclusively for resume node groups. By using a dedicated role separate from your regular Atlas node instance profile, the removal of this role from `aws-auth` after the resume node group is deleted won't affect other Atlas-managed nodes.

> 🚧 When is this required?
>
> This configuration is required when your node instance profile's IAM role is granted cluster access via the `aws-auth` ConfigMap. This applies to:
>
> * Clusters using `CONFIG_MAP` authentication mode
> * Clusters using `API_AND_CONFIG_MAP` authentication mode **where the node IAM role is configured in `aws-auth`**
>
> This is **not required** when:
>
> * Clusters use only `API` authentication mode (EKS access entries)
> * Clusters use `API_AND_CONFIG_MAP` mode but the node IAM role is configured via EKS access entries (not `aws-auth`)
>
> EKS does not remove entries from the access entries API when node groups are deleted, so clusters using access entries for node authentication are not affected.

**Error message:** If this configuration is missing, you'll see the following error when attempting to resume a cluster:

```
initiating cluster resume: rpc error: code = Unknown desc = validating resume capability: persistent error: cluster authentication mode is CONFIG_MAP, node configuration must specify nodeGroupArn for cluster resume
```

**Check your cluster's authentication mode:**

```bash
aws eks describe-cluster --name <cluster-name> --query 'cluster.accessConfig.authenticationMode'
```

**Configure via Terraform:**

```hcl
resource "atlas_node_configuration" "default" {
  name       = "default"
  cluster_id = atlas_eks_cluster.my_cluster.id
  # ... other configuration ...
  
  eks {
    instance_profile_arn = aws_iam_instance_profile.atlas_node_profile.arn
    node_group_arn       = "arn:aws:iam::123456789012:role/atlas-dedicated-node-role"
    # ... other EKS settings ...
  }
}
```

**Configure via API:**

Set the `nodeGroupArn` field in the EKS section when [creating](https://docs.atlas.example/reference/nodeconfigurationapi_createconfiguration) or [updating](https://docs.atlas.example/reference/nodeconfigurationapi_updateconfiguration) a node configuration.

> 📘 Note
>
> The `nodeGroupArn` field is not available in the Atlas console UI. You must configure it via Terraform or API.

## Prerequisites verification checklist

Before testing hibernation, verify:

* [ ] Atlas components have the `CriticalAddonsOnly` toleration — upgrade to current Helm chart versions if needed

<Accordion title="Minimum chart versions with CriticalAddonsOnly toleration">

| Component | Chart version | appVersion |
|---|---|---|
| atlas-agent | 0.139.0 | v0.113.0 |
| atlas-cluster-controller | 0.90.0 | v0.61.1 |
| atlas-pod-mutator | 0.6.0 | v0.2.6 |
| atlas-pod-pinner | 1.12.0 | v1.9.5 |
| atlas-workload-autoscaler | 0.1.190 | v0.97.0 |

</Accordion>

* [ ] Any other workload required during cluster recovery also has the `CriticalAddonsOnly` toleration. This includes the critical components Atlas validates during hibernation — such as CNI, DNS, admission webhook controllers, and important daemonsets — which are needed for the first resume node to join the cluster in a `Ready` state. If a webhook configuration flagged by this validation does not block your critical components, you can exclude it from the check (see [Missing CriticalAddonsOnly toleration](#missing-criticaladdonsonly-toleration)):
  ```yaml
  tolerations:
    - key: CriticalAddonsOnly
      operator: Exists
  ```
* [ ] Atlas cluster role has hibernation permissions
* [ ] Resume node size is big enough to schedule cluster critical and Atlas components
* [ ] (EKS with aws-auth ConfigMap only) `nodeGroupArn` is configured in the default Node Configuration

# Managing hibernation schedules in the console

The Atlas console provides a comprehensive interface for managing hibernation schedules across your organization. You can create, view, and manage hibernation schedules that automatically hibernate and resume clusters based on defined schedules.

## Accessing hibernation schedules

You can access hibernation schedules in two ways:

**Organization-level access:**

1. Navigate to the **Optimization** section in the Atlas console
2. Select **Cluster hibernation** from the menu

![](https://files.readme.io/09b83f4c72430d9b08bc8b5e04b17b0563311ed080bfc04b97da311fd4dc08d4-image.png)

**Cluster-level access:**

1. Navigate to any cluster dashboard
2. Click on **Hibernation schedules** to open the hibernation schedules interface

![](https://files.readme.io/f27c6c33f8b65f0d8e6e7c8185efe4dfc4f20679bb5139045193911d88647aec-image.png)

## Creating hibernation schedules

You can create hibernation schedules from either the organization-level hibernation page or directly from any cluster dashboard. Both entry points lead to the same schedule creation workflow.

The console provides a two-step process for creating hibernation schedules:

1. **Create schedule**: Define the hibernation schedule parameters
2. **Assign clusters**: Select which clusters the schedule applies to

When creating a schedule from a cluster dashboard, that cluster will be automatically pre-selected in the assignment step.

### Step 1: Create a schedule

When creating a new hibernation schedule, you can configure:

**Schedule name**: A descriptive name for your hibernation schedule (e.g., "Spot drift", "Weekday work hours")

**Schedule mode**: Choose what you'd like to automate:

* **Hibernation and resumption**: Full automation of both hibernation and resumption
* **Only hibernation**: Automatically hibernate clusters (manual resumption required)
* **Only resumption**: Automatically resume hibernated clusters

**Time zone**: Select your preferred time zone to align schedule times with your local time. The interface automatically detects your location and suggests the appropriate time zone.

**Hibernation period**: Define when and how often to hibernate clusters:

* **Standard**: Define intervals via a user-friendly interface
* **Crontab**: Use cron expressions for custom scheduling

The console displays a preview of your schedule (e.g., "Hibernation scheduled for every Friday at 18:00").

**Node configuration parameters**: For resumption, specify the node configuration in JSON format. This allows you to define the desired configuration for nodes that will be used when your cluster resumes.

<Image align="center" width="75% " src="https://files.readme.io/4fe61467e5645c19fabf311d19c1ffad27dd5d0cbec4c27bd91add92ad0b6a9f-image.png" />

### Step 2: Assign clusters

After creating the schedule, assign it to specific clusters:

![](https://files.readme.io/be92e8040f11bb26c66b1e491ef91910faa69d6d1974ef9904c4bde6fd0ad6cd-image.png)

> 📘 Additional setup for EKS clusters
>
> EKS clusters onboarded before October 2025 may need updated IAM permissions. If the console displays a permissions notification, re-run the onboarding script or upgrade your Terraform provider to v7.74.0+.

## Managing existing schedules

You can view and manage hibernation schedules from both the organization-level hibernation page and individual cluster dashboards.

**From Optimization > Cluster hibernation**, the schedules list shows all your configured schedules. You can edit them and assign or unassign clusters at any time.

To make changes to an existing schedule, click on it in the list to open the settings drawer:

![](https://files.readme.io/49a44eca55921bcb007cddcee1eefa6aee74b1985ce8b752d63f92c1b45565c0-image.png)

A schedule can be deleted through the three-dot (⋮) menu:

![](https://files.readme.io/7707ed09e42d1f2ff69f924e9ac7067d9837824f0ee349609905441ae64f3cb0-image.png)

**From cluster dashboards**, you'll see a **Hibernation schedules** card showing the number of schedules assigned to that cluster (e.g., "2"). Clicking this opens a drawer where you can:

* **View assigned schedules**: See all schedules affecting the current cluster (e.g., "default\_schedule", "Workday")
* **Toggle schedule status**: Enable or disable individual schedules for this cluster
* **Create new schedules**: Start the schedule creation process with the current cluster pre-selected

![](https://files.readme.io/6de22a8a7eaf2dcfaed84da5f0592635b12bb40c5146a829fbeac1d727427373-image.png)

# API-based hibernation schedule management

You can also create and manage hibernation schedules programmatically using the Atlas API. This provides the same scheduling capabilities as the console interface but allows for automation and integration with your existing infrastructure as code workflows.

## Creating hibernation schedules via API

You can create hibernation schedules using the <Anchor label="Atlas Hibernation Schedules API" target="_blank" href="https://docs.atlas.example/reference/hibernationschedulesapi_createhibernationschedule">Atlas Hibernation Schedules API</Anchor>:

```shell
curl --request POST \
     --url $ATLAS_API_URL/cluster-autoscaler/v1beta/organizations/<organizationId>/hibernation-schedules \
     --header 'accept: application/json' \
     --header 'content-type: application/json' \
     --header 'X-API-Key: <your-api-key>' \
     --data '
{
  "enabled": true,
  "name": "Development Environment Schedule",
  "pauseConfig": {
    "enabled": true,
    "schedule": {
      "cronExpression": "0 18 * * MON-FRI"
    }
  },
  "resumeConfig": {
    "enabled": true,
    "schedule": {
      "cronExpression": "0 8 * * MON-FRI"
    },
    "jobConfig": {
      "nodeConfig": {
        "instanceType": "m5.large",
        "spotConfig": {
          "spot": true
        },
        "zone": "us-east-1a"
      }
    }
  },
  "clusterAssignments": {
    "items": [
      {
        "clusterId": "<your-cluster-id>"
      }
    ]
  }
}
'
```

# Manual hibernation in the console

In addition to scheduled hibernation, you can manually hibernate and resume individual clusters. Manual hibernation can be initiated from either the cluster dashboard or the cluster list view.

## Hibernating a cluster manually

You can manually hibernate a cluster from two locations:

**From cluster dashboard:**

1. Navigate to your cluster dashboard
2. Locate the **Hibernate cluster** option in the cluster actions menu (top right)
3. Click **Hibernate cluster** to initiate the hibernation process

![](https://files.readme.io/77e3d5520d3f54cf6a02bce4833876d809f2178803a130ec45e9e4045ab2dc8a-image.png)

**From cluster list:**

1. Navigate to the cluster list view
2. Locate the cluster you want to hibernate
3. Click **Hibernate cluster** for that cluster to initiate the hibernation process

![](https://files.readme.io/d793b6825dbf5f06d335716020367681eed3f23846f2674cffaa2cc86b770d1d-image.png)

Both methods will open a hibernation confirmation dialog where you can:

* **Review hibernation details**: Understand what will happen during hibernation
* **Confirm hibernation**: Click **Hibernate now** to proceed
* **Cancel**: Exit without hibernating if you change your mind

<Image align="center" width="65% " src="https://files.readme.io/740b2a89e893430045c2961b02b75346a25dd7e8bdc17fd28be05cb781b4aa52-image.png" />

### Cluster status during hibernation

When a cluster is hibernated, its status will change to reflect the hibernation state:

* **Connected** → **Hibernating** (during the hibernation process)
* **Hibernating** → **Hibernated** (when hibernation is complete)

The hibernation status is displayed throughout the console, including in the cluster list and cluster dashboard views.

## Resuming a hibernated cluster manually

To resume a hibernated cluster:

1. Navigate to the hibernated cluster (from either the cluster dashboard or the cluster list)
2. Click **Resume cluster** in the available cluster actions
3. Configure the resumption parameters in the dialog:
   * **Instance type**: Specify the node instance type for resumption
   * **Node configuration**: Additional node template configuration (optional)

The resumption process provides different configuration options depending on your setup:

**When you have an existing node configuration:**

* The system will display your current node configuration
* You can modify the configuration using the JSON editor
* Default values are pre-populated based on your cluster's previous configuration

<Image align="center" width="60% " src="https://files.readme.io/b10b5cf27c0716e5b00baeef394d486b55c0b7b30d37e002bcb4627313824799-image.png" />

**When you don't have node configuration:**

* You'll be prompted to specify the node configuration for resumption
* The system provides a template you can customize, or choose to import a configuration from a hibernation schedule
* You can define instance type, region, availability zones, and other node parameters

<Image align="center" width="60% " src="https://files.readme.io/ebedcb6746306fc6a956ef6ddb4161aff819d56a29cfea20b0b81196851ac8b8-image.png" />

### Managing resumption nodes

After a cluster resumes, the temporary nodes created during resumption remain active in your cluster. Atlas does not automatically remove or replace these nodes.

If you selected larger instance types during resumption to ensure successful startup, you have two options to optimize your cluster after resumption:

**Option 1: Manual rebalancing**

1. Navigate to **Rebalancer** in the Atlas console
2. Generate a rebalancing plan to replace resumption nodes with more cost-efficient instances
3. Execute the plan to complete the optimization

**Option 2: Manual node removal**

1. Remove the resumption nodes manually from your cluster
2. Atlas's autoscaler will automatically provision optimized nodes based on your workload requirements

> 📘 Note
>
> If you initially chose larger instance types to ensure reliable resumption, plan to rebalance your cluster afterward to optimize costs.

### Resumption status changes

During the resumption process, the cluster status will progress through:

* **Hibernated** → **Resuming** (during the resumption process)
* **Resuming** → **Connected** (when resumption is complete and cluster is active)

The resumption process creates temporary nodes to run essential system components, then hands over control to Atlas's autoscaler for normal operations.

# Manual hibernation via API

You can also manage cluster hibernation programmatically using the Atlas API.

## Hibernating a cluster via API

You can hibernate a cluster using the [Atlas Hibernate API](https://docs.atlas.example/reference/hibernationjobsapi_createhibernationjob):

```shell
curl --request POST \
     --url $ATLAS_API_URL/cluster-autoscaler/v1/organizations/<your-organization-id>/clusters/<your-cluster-id>/hibernation-jobs \
     --header 'X-API-Key: <your-api-key>' \
     --header 'accept: application/json' \
     --header 'content-type: application/json' \
     --data '{"type":"PAUSE"}'
```

Replace the `<placeholders>` in the path with the actual IDs of your cluster and organization.

The cluster will transition to a "Hibernating" state, and the following actions will occur:

1. The autoscaler will be disabled to prevent new nodes from being created
2. All nodes will be systematically removed from the cluster in stages
   * First, regular workload nodes are removed
   * Temporary nodes keep essential services running until all other nodes are removed
   * Finally, temporary nodes are removed as well
3. The control plane will remain active, but no workloads will run

> 📘 Note
>
> During hibernation, you will still incur charges for the cloud provider's control plane and any persistent volumes that remain.

## Resuming a hibernated cluster via API

To resume a hibernated cluster, use the [Atlas Resume API](https://docs.atlas.example/reference/hibernationjobsapi_resumecluster):

```curl
curl --request POST \
     --url $ATLAS_API_URL/cluster-autoscaler/v1/organizations/<your-organization-id>/clusters/<your-cluster-id>/hibernation-jobs:resumeCluster \
     --header 'X-API-Key: <your-api-key>' \
     --header 'accept: application/json'
```

It requires an existing hibernation schedule that is assigned to the cluster being resumed. The schedule's node configuration is used as the base for the resumption node.

The node specified in the schedule's node configuration has to be sufficient to run the cluster's essential components. The instance type requirements depend on your specific workloads and cloud service provider.

The resumption process works as follows:

1. Temporary nodes are created using your specified instance type, and the rest of the node template configuration, if provided
2. These nodes run essential system components, including Atlas agents
3. The autoscaler is reactivated and takes over normal cluster scaling operations
4. The cluster returns to its "Ready" state

After resumption completes, the temporary nodes remain in your cluster. Atlas does not automatically remove or replace these nodes. To optimize your cluster after resumption, you can either use the rebalancing feature to replace these nodes with more cost-efficient instances or manually remove them to allow the autoscaler to provision optimized nodes based on your workload requirements.

### Troubleshooting Resume Operations

#### Insufficient Resources

If the resumption process fails, it's typically because:

1. The specified `instanceType` cannot be provisioned in your cloud environment
2. The instance type doesn't have enough resources to run all cluster-critical components

In these scenarios, manual user intervention is required to resolve the issue and restore the cluster to a running state.

To restore the cluster to a running state, you have two options:

**Option 1:** Try resuming again with a different instance type in the node configuration:

```shell
curl --request POST \
     --url $ATLAS_API_URL/cluster-autoscaler/v1/organizations/<your-organization-id>/clusters/<your-cluster-id>/hibernation-jobs:resumeCluster \
     --header 'X-API-Key: <your-api-key>' \
     --header 'accept: application/json'
```

**Option 2:** Add nodes manually using your cloud provider

1. Create nodes large enough to run the critical components
2. Ensure the nodes are properly joined to the cluster

Once you successfully provision nodes with sufficient capacity to schedule all cluster-critical components (either through the resume API or manually), Atlas will detect this and automatically complete the cluster resumption process.

> 🚧 Important
>
> Choosing an appropriate instance type is crucial for successful cluster resumption. Requirements vary by cluster type and size - while some clusters may operate with smaller instances (2 vCPU, 8GB RAM), other environments might require larger instances (16 vCPU, 64GB RAM). When in doubt, we recommend erring on the side of larger instances. After successful resumption, you can rebalance the cluster or manually remove these nodes to allow Atlas to provision more cost-efficient alternatives based on your workload requirements.

> 📘 Node Configuration parameters used in temporary nodes
>
> The following parameters from the **default** node configuration are being used for temporary node creation during the resume process:
>
> * Security groups
> * Subnets
> * IMDS hop limit
> * IMDSv1/v2
> * Volume size and type, IOPS, throughput, and KMS encryption
> * SSH key pair
> * Instance tags

#### Missing CriticalAddonsOnly toleration

When the `CriticalAddonsOnly` taint is applied to the resume node (see [How hibernation works](#how-hibernation-works)), only workloads that carry the matching toleration can be scheduled on it. If a critical component required for the resume node to join the cluster in a `Ready` state is missing the CriticalAddonsOnly toleration, it will not be scheduled and the resumption process can fail.

Atlas validates critical components during hibernation (see [How hibernation works](#how-hibernation-works)); the results are recorded in [Audit logging](#audit-logging). If resumption fails, review the validation results from the preceding hibernation job in the audit log to identify which components were flagged as missing the toleration, then add the toleration to those components:

```yaml
tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
```

Atlas validates admission webhook configurations for the `CriticalAddonsOnly` toleration as part of hibernation. This covers both `ValidatingWebhookConfiguration` and `MutatingWebhookConfiguration` objects, in particular webhooks with `failurePolicy: Fail`, which can block the critical components from starting on the resume node.

If Atlas flags a webhook that doesn't actually block your critical components, for example one your team operates that acts on workloads outside the critical path, you can exclude it from this validation by adding the `autoscaling.atlas.example/ignore-during-hibernation` annotation to the webhook configuration:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: my-webhook
  annotations:
    autoscaling.atlas.example/ignore-during-hibernation: "true"
```

The annotation removes the webhook from validation only. It does not change how the webhook serves traffic during resumption, so exclude a webhook only after confirming it cannot block the components the resume node needs to reach a `Ready` state.

> 🚧 Important
>
> Atlas only validates the critical components it is aware of. If resumption still fails after addressing the flagged components, review your cluster's other workloads for anything required for the resume node to reach a `Ready` state and add the `CriticalAddonsOnly` toleration to those as well.

#### Admission webhook deadlock (Gatekeeper and Kyverno)

Clusters that run two admission policy engines — most commonly OPA Gatekeeper and Kyverno — can fail to resume when both engines' webhooks use `failurePolicy: Fail`. With this setting, the API server rejects any pod creation when the webhook's controller is not running. If Gatekeeper's webhook blocks Kyverno's pods and Kyverno's webhook blocks Gatekeeper's pods, neither controller can start on the resume node, and the resume process cannot complete.

> 🚧 Important
>
> The deadlock described here is one specific case of a broader problem: during hibernation, Atlas validates admission webhook controllers (among other critical components) for the `CriticalAddonsOnly` toleration. See [Missing CriticalAddonsOnly toleration](#missing-criticaladdonsonly-toleration) for the validation itself and for how to opt a webhook out of it when the flag is not accurate.

You can recognize this scenario by the following symptoms:

* The cluster resume appears stuck: the resume node joins the cluster in a `Ready` state, but Atlas components do not start on it.
* Describing the pods for either controller (with `kubectl describe pod -n <namespace> <pod-name>`) shows the other controller's admission webhook rejecting them, with an event of the form `failed calling webhook "<webhook-name>" ...` referencing the other engine.

To resolve the deadlock, apply the following changes, then resume the cluster again:

**1. Add the `CriticalAddonsOnly` toleration to both controllers**

The Gatekeeper and Kyverno admission controller deployments (typically `gatekeeper-controller-manager` and `kyverno-admission-controller` in recent releases) must tolerate the `CriticalAddonsOnly` taint so they can be scheduled on the resume node:

```yaml
tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
```

See [Missing CriticalAddonsOnly toleration](#missing-criticaladdonsonly-toleration) for details on how Atlas applies and validates this taint.

**2. Allow `kube-system` components to run unconditionally**

Both webhooks must leave the `kube-system` namespace alone so that node-critical components (such as the CNI, DNS, and kube-proxy) can start before any admission controller is running. There are two ways to configure this, in order of preference.

**Option A — exclude only the workloads the node needs to become ready (recommended)**. Use an `objectSelector` on each webhook to target just the specific components required for the node to join the cluster — for example, on EKS, the CNI, DNS, and node networking agent. This keeps admission control active for everything else in `kube-system`.

```yaml
# snippet from a (Validating|Mutating)WebhookConfiguration object
webhooks:
  - name: <webhook-name>
    # ...
    objectSelector:
      matchExpressions:
        - key: k8s-app
          operator: NotIn
          values:
            - aws-node      # EKS node agent
            - kube-proxy    # node networking
            - kube-dns      # EKS CoreDNS label
```

The exact label keys and values depend on your cluster's setup — check the labels on the actual pods in `kube-system` before writing the selector (`kubectl get pods -n kube-system --show-labels`).

**Option B — exclude the whole `kube-system` namespace (fallback)**. Use a `namespaceSelector` to skip the namespace entirely. This is the simplest and most broadly compatible approach, but it removes admission control from every pod in `kube-system`, including any custom pods you may have added there.

```yaml
# snippet from a (Validating|Mutating)WebhookConfiguration object
webhooks:
  - name: <webhook-name>
    # ...
    namespaceSelector:
      matchExpressions:
        - key: kubernetes.io/metadata.name
          operator: NotIn
          values:
            - kube-system
```

> 🚧 Important
>
> **Option B weakens policy enforcement for a highly privileged namespace.** Use it only if you cannot reliably identify the workloads Option A needs, or as a temporary fix while you build the per-workload list.

The exact path you use to set these fields depends on how each engine was installed. For the Kyverno Helm chart, the equivalent snippet goes under `config.webhooks` in `values.yaml`. For Gatekeeper, apply the `namespaceSelector` or `objectSelector` directly to the `MutatingWebhookConfiguration` and `ValidatingWebhookConfiguration` objects. For any other installation method, consult the engine's installation documentation for the correct field path.

> 📘 Note
>
> Do not assume a default exclusion already exists. Older Kyverno releases configured exclusions via `resourceFilters` in the `kyverno` ConfigMap rather than `namespaceSelector`, and default webhook configurations differ between Helm charts, raw manifests, and managed add-ons. Verify each webhook's selector directly with `kubectl get validatingwebhookconfiguration <name> -o yaml` (and the same for `mutatingwebhookconfiguration`) before relying on it.

**3. Break the circular dependency between the two webhooks**

As long as each engine's webhook intercepts the other engine's pods, the deadlock can still occur. Untie it by excluding one engine from the other's validation. Gatekeeper is commonly installed as an EKS add-on while Kyverno is installed manually, so the typical setup is to exclude the `gatekeeper-system` namespace from Kyverno's webhook, using the same `namespaceSelector` pattern shown above for `kube-system`. Gatekeeper then continues to validate Kyverno's pods as before, but Gatekeeper's pods no longer require Kyverno to be running in order to start.

> 📘 Note
>
> This step changes how the webhook behaves: it narrows what Kyverno intercepts. Atlas may also offer a way to exclude a webhook configuration from hibernation validation without changing the webhook's coverage — see [Missing CriticalAddonsOnly toleration](#missing-criticaladdonsonly-toleration) for details. That option removes the warning but does not prevent the webhook from blocking the other engine's pods if they are still in its scope, so the `namespaceSelector` exclusion above remains the more thorough fix for the deadlock itself.

**4. Right-size the resume node**

Both admission controllers add their own resource requests on top of Atlas components and cluster-critical workloads. If the resume node does not have enough capacity for everything that must start on it, resume can fail with unschedulable pods — see [Insufficient resources](#insufficient-resources) and [Resume node size](#resume-node-size). Check the combined resource requests of your Gatekeeper and Kyverno controllers against the instance type you plan to use for resume, and adjust if needed.

After resolving the webhook configuration, the startup sequence on the resume node is the following, which you can observe while resume progresses:

1. `kube-system` components excluded from both webhooks start, and the node becomes `Ready`.
2. The Gatekeeper controller starts (no longer blocked by Kyverno's webhook).
3. The Kyverno controller starts (Gatekeeper's webhook accepts its pods).
4. The Atlas components in the `atlas-agent` namespace are admitted.
5. The resume process completes, and the Atlas autoscaler takes over.

# Audit logging

Atlas generates [audit log](audit-log.md) events for all hibernation and resumption activities, providing visibility into cluster state changes and helping with compliance and troubleshooting. These events can be integrated with your monitoring and alerting systems.

## Hibernation events

The following audit events are generated during hibernation operations:

* **Cluster hibernated**: When the cluster successfully enters the hibernated state
* **Cluster hibernation failed**: If the hibernation process encounters errors
* **Critical components validated**: During hibernation, Atlas validates that all known critical components required for the resume node to join the cluster in a `Ready` state — such as CNI, DNS, admission webhook controllers, and important daemonsets — carry the `CriticalAddonsOnly` toleration. The results are recorded in the audit log, flagging any components that are missing it.

> 🚧 Important
>
> Atlas only validates the critical components it is aware of. Components it does not yet recognize are not checked, so review your cluster's workloads to confirm nothing else required for the resume node to join in a `Ready` state is missing the toleration.

## Resumption events

Resumption operations generate these audit events:

* **Cluster resumed**: When the cluster successfully returns to the active state
* **Cluster resumption failed**: If the resumption process encounters errors

# API reference

For detailed API specifications, see our reference documentation:

* [Create Hibernation Schedule API](https://docs.atlas.example/reference/hibernationschedulesapi_createhibernationschedule)
* [Create Hibernation Job API](https://docs.atlas.example/reference/hibernationjobsapi_createhibernationjob)
* [Resume Cluster API](https://docs.atlas.example/reference/hibernationjobsapi_resumecluster)

# Comparison with legacy pause feature

> 📘 Note
>
> Atlas previously offered a different mechanism for pausing clusters. The new hibernation feature provides improved reliability and better cost optimization. For information on the legacy pause feature, see [Pausing a cluster](pausing-a-cluster.md).

Key differences between hibernation and the legacy pause feature:

| Feature                              | Cluster hibernation                           | Legacy pause                                        |
| ------------------------------------ | --------------------------------------------- | --------------------------------------------------- |
| Implementation                       | API-driven with automated resumption          | Uses Kubernetes CronJobs                            |
| Setup complexity                     | Simple API calls with one-time configuration  | Requires manual creation and scheduling of CronJobs |
| Cloud service provider (CSP) support | All major cloud providers                     | Limited provider support                            |
| Console interface                    | Full UI support for scheduling and management | No console interface                                |
| Audit logging                        | Complete audit trail for all operations       | Limited logging capabilities                        |