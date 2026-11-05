---
updatedAt: 2026-10-27T09:40:57.000Z
---

# How-to: Manage scaling policies

Once you've created your scaling policies, ongoing management involves assigning workloads, modifying policy configurations, and maintaining your optimization strategy as your infrastructure evolves. This guide covers the key management tasks for keeping your scaling policies effective and up-to-date.

# Workload assignment and reassignment

Managing which workloads are assigned to specific scaling policies is a core aspect of workload optimization. You can control assignments through assignment rules or manual assignment methods.

## View policy assignments

You can view policy assignments and workload distributions in several ways:

**From the scaling policies table:**

* Navigate to **Workload Autoscaler** > **Scaling policies**
* The **Workloads** column shows the number of workloads assigned to each policy
* Click on the workload count to see a detailed list of assigned workloads in the drawer that opens

![](https://files.readme.io/0a0c6f9a202f029cca2ff3d7599a18a579f7ae9233cb85ad624519f75f7efa6d-image.png)

**From the workloads list:**

* Navigate to **Workload Autoscaler** > **Optimization**
* The **Scaling policy** column shows which policy is assigned to each workload
* Use the **Scaling policy** filter dropdown to view workloads assigned to specific policies

![](https://files.readme.io/c619292f18b4a199b75740e42e3e899ae786fa6aca05cde1f6eadd969064e30e-image.png)

**From the workload details page:**

1. Navigate to the workload details page by clicking on a workload
2. In the **Rightsizing settings** section, locate **Vertical scaling policy**
3. Use the dropdown to select a different policy
4. Save the changes

![](https://files.readme.io/3e058c61ae67d4bd28070e079c48e3acce2f32e5a32041a2511e5f3b20402220-image.png)

## Understanding policy assignment methods

Atlas uses multiple methods to assign workloads to scaling policies. Understanding these methods and their priority helps you troubleshoot assignment issues and design effective scaling strategies.

### Policy assignment priority hierarchy

Workload Autoscaler evaluates workload assignments in the following priority order:

1. **Manual/Explicit Assignment**
   * **API assignments** – Policies assigned via the Atlas API
   * **Annotation assignments** – Policies specified using `workloads.atlas.example/configuration` with `scalingPolicyName`. If the specified policy name does not match any existing policy, the workload falls back to system defaults (`resiliency` for StatefulSets, `balanced` for others).
   * **Direct UI assignments** – Policies manually assigned through the Optimization page in the Atlas Console

2. **Assignment Rules**
   * Evaluated based on policy priority order (as configured in the policies table)
   * The first matching policy wins
   * Within each rule, all conditions must match (`AND` logic)
   * A workload matching any rule in a policy is assigned to that policy (`OR` logic between rules)

3. **System Defaults**
   If workloads have not been manually assigned to policies, and if no assignment rules match them, they will be moved to default system scaling policies:
   * **StatefulSets**: `resiliency`
   * **All other workloads**: `balanced`

### Assignment origin tracking

Atlas tracks how each workload was assigned to its current policy through a `scalingPolicyOrigin` field available via the API response for each workload. This field indicates the assignment method:

| Value              | Assignment Method                        |
| ------------------ | ---------------------------------------- |
| `api`              | Assigned via Atlas API                 |
| `annotations`      | Assigned via Kubernetes annotations      |
| `assignment-rules` | Assigned through policy assignment rules |
| `default`          | Assigned through system defaults         |

You can query this information using the [Get workload API endpoint](https://docs.atlas.example/reference/workloadoptimizationapi_getworkload):

```curl
curl --request GET \
     --url $ATLAS_API_URL/v1/workload-autoscaling/clusters/{clusterId}/workloads/{workloadId} \
     --header 'X-API-Key: <YOUR_API_KEY>' \
     --header 'accept: application/json'
```

The response includes the `scalingPolicyOrigin` field along with `scalingPolicyId` and `scalingPolicyName` to provide a complete assignment context:

```json
{
  "workload": {
    "name": "example-deployment",
    "namespace": "production",
    "kind": "Deployment",
    "scalingPolicyId": "policy-123",
    "scalingPolicyName": "production-frontend-policy",
    "scalingPolicyOrigin": "assignment-rules",
    // ... other workload fields
  },
```

## Assign workloads in bulk

To assign multiple workloads to a scaling policy:

1. Navigate to **Workload Autoscaler** > **Optimization**
2. Use the checkboxes to select multiple workloads from the list
3. Click on **Assign vertical scaling policy**
4. Select the target scaling policy from the available options
5. Confirm the assignment

![](https://files.readme.io/c084287c98322f5608ecad512543f9ecb9feda3ccc41f5da3e80e3951241c093-image.png)

You can filter workloads before bulk assignment using the plethora of different options to narrow down your selection.

**Impact of manual assignment:**
When you manually assign workloads through the UI, these assignments take precedence over assignment rules. This means the workloads will remain assigned to the selected policy even if you later create or modify assignment rules that would match them.

## Move workloads between policies

The process of moving workloads between policies is exactly the same as the process of assigning them to a policy for the first time. Therefore, reference the [workload assignment and reassignment](#workload-assignment-and-reassignment) section above.

**Impact of policy changes:**

* New policy settings take effect immediately
* Recommendations are recalculated based on the new policy configuration
* Manual reassignments override any existing assignment rules

## Manage unassigned workloads

Workloads that don't match any assignment rules are automatically assigned through system defaults based on their workload type:

* **StatefulSets**: Assigned to `resiliency` policy
* **Other workloads**: Assigned to `balanced` policy

These system policies typically show higher workload counts in clusters without extensive assignment rules.

For better, more targeted workload optimization:

1. Review workloads assigned to system default policies (`balanced`, `resiliency`)
2. Create appropriate assignment rules to automatically capture these workloads and assign them to more specific system policies or custom policies that you have configured
3. Alternatively, manually reassign individual workloads to more specific policies through the UI

# Assignment rule management

Assignment rules automatically assign workloads to policies based on their characteristics. You can manage these rules from the policy view.

## Add a new assignment rule

For detailed information on the configuration options of assignment rules, refer to our [guide on creating a scaling policy](woop-scaling-policies-create.md).

To add assignment rules to an existing policy, follow these steps:

1. Navigate to the scaling policy details page
2. Go to the **Assignment rules** tab
3. Click **Add rule** to create additional assignment criteria
4. Configure namespace, workload type, and label filters
5. Save the policy to activate the new rules

![](https://files.readme.io/8d13a026a1b5038d30b63c0d13ab3c75b5886ab16fee3d9b33455ef8dd1a0753-image.png)

## Modify or delete an assignment rule

To modify an existing assignment rule,

1. Navigate to the scaling policy details page
2. Go to the **Assignment rules** tab
3. Use the **Filters applied** panel to review current rule settings
4. Configure namespace, workload type, and label filters
5. Optionally, remove individual filters using the **X** button next to each filter

To remove an assignment rule:

1. Use the **Delete assignment rule** option in the rule configuration menu
2. Confirm the deletion to remove the rule permanently

<Image align="center" width="60% " src="https://files.readme.io/fa3570b0b7cdb8c7b70851c56a1dc4a976a29ed3f2add2fceb330e59f95ae29b-image.png" />

Workloads previously matched by the deleted rule will be reassigned based on remaining rules or moved to the default policy.

**Impact of rule changes:**
Changes to assignment rules trigger asynchronous workload reassignment. In large clusters, this process may take a few seconds or more to complete. Workloads with manual assignments will not be affected by assignment rule changes.

# Policy modification and updates

Scaling policies need to evolve as your workloads and optimization requirements change. Understanding how to modify policies and their impact on assigned workloads is crucial.

## Edit policy settings

You can modify scaling policy settings at any time:

**Accessing policy settings:**

1. Navigate to **Workload Autoscaler** > **Scaling policies**
2. Click on the policy name to open the policy details view
3. Use the **Settings** tabs to modify optimization parameters for each dimension

Refer to our [guide on creating a scaling policy](woop-scaling-policies-create.md) for detailed instructions or our [reference document on available settings](woop-configuration-settings.md) for all available configuration options.

**Impact on assigned workloads:**

* Vertical scaling settings changes apply immediately to all workloads assigned to the policy, and new resource request recommendations are generated based on updated policy parameters
* Horizontal autoscaling settings changes update the native HPA resources on all assigned workloads. Note that workloads with workload-level HPA overrides are not affected by policy-level HPA changes until the overrides are reverted.

## Keep policies up to date with AtlasPilot Assistant

As an alternative to manually editing and reordering policies, you can re-run AtlasPilot Assistant to maintain your policies as your cluster evolves. When you re-run AtlasPilot Assistant on a cluster that already has generated policies, it:

* **Updates existing policies** when a policy still fits but its settings have drifted from your goals or from how the cluster has changed
* **Creates new policies** when a policy no longer fits, rather than forcing a bad fit onto workloads it no longer describes
* **Reorders policy priority** so the most relevant policies are evaluated first when multiple assignment rules could match the same workload

Manual assignments always take precedence: workloads you assigned explicitly via the API, annotations, or the Atlas Console keep that assignment regardless of AtlasPilot Assistant re-runs. For the full lifecycle, see [Generate scaling policies with AtlasPilot Assistant](woop-scaling-policies-atlaspilot-assistant.md#maintain-and-update-policies-with-re-runs).

# Policy lifecycle management

Managing the lifecycle of scaling policies includes enabling, disabling, and removing policies while ensuring workload continuity.

## Enable and disable automated optimization

You can temporarily disable the automated resource optimization aspect of policies without deleting them, as well as re-enable it later when your requirements change. To do so, follow these steps:

1. Toggle the policy behavior in the **Optimization Type** dropdown in the scaling policies table or in the scaling policy page:

![](https://files.readme.io/11e8cf621b088ce0c3bdee63e3c08198540de119296f7b1f494f084c16250b38-image.png)

**Impact on assigned workloads:**

* Workloads remain assigned to disabled policies, but optimization is paused for both vertical and horizontal scaling
* Assignment rules for disabled policies continue to match new workloads
* Vertical recommendations can still be viewed and are generated in the cluster, but won't be applied automatically
* Horizontal autoscaling (HPA) is also paused — existing HPA resources remain in the cluster but are no longer actively managed
* Workloads continue running with their current resource requests and replica counts

## Delete a policy

Deleting policies is irreversible and requires careful consideration:

**Prerequisites before deletion:**

* Review all workloads assigned to the policy
* Decide whether to reassign workloads to other policies or let them be reassigned by the assignment rules of other policies

**Deletion process:**

1. Navigate to **Workload Autoscaler** > **Scaling policies**
2. Click on the policy name to open the policy details view
3. Use the policy actions menu (three dots) to select **Delete policy**
4. Confirm the deletion in the dialog box
5. The policy is permanently removed from the system

![](https://files.readme.io/0f9ae1eb4424a3dd4a3cad9a69a1785592d49b3f4361c00234d8e3d2ff431088-image.png)

**What happens to workloads:**
Workloads from deleted policies are automatically reassigned following the priority hierarchy:

1. First, the system checks for other policies with assignment rules that match these workloads
2. If multiple policies match, the highest-priority policy is selected
3. If no assignment rules match, workloads fall back to system defaults (`resiliency` for StatefulSets, `balanced` for other workloads)

**Asynchronous reassignment:**
The reassignment process is asynchronous and may take several minutes to complete in large clusters. You can monitor the progress by checking policy workload counts and the workloads list.

# Next steps

<Cards>
  <Card title="Create scaling policies" href="https://docs.atlas.example/docs/woop-scaling-policies-create" icon="fa-plus">
    Learn how to create new policies with specific optimization settings.
  </Card>

  <Card title="Workload autoscaling settings" href="https://docs.atlas.example/docs/woop-configuration-settings" icon="fa-gear">
    Explore detailed configuration options for policy settings.
  </Card>
</Cards>