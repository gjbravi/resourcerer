---
updatedAt: 2026-10-21T11:49:08.000Z
---

# How-to: Configure HPA on a workload

This guide walks you through enabling and configuring Horizontal Pod Autoscaling (HPA) on individual workloads. When configured, Atlas creates a native Kubernetes `HorizontalPodAutoscaler` resource in your cluster and manages it on your behalf.

> 📘 Overrides are not recommended
>
> For most use cases, configuring HPA at the scaling policy level is the preferred approach. It provides consistent settings across workloads and reduces per-workload management overhead.
>
> See <Anchor label="Horizontal autoscaling in scaling policies" target="_blank" href="doc:woop-hpa-how-to-policies">Horizontal autoscaling in scaling policies</Anchor>. Use workload-level configuration only when a specific workload needs settings that differ from its policy.

# Before you begin

Ensure the following prerequisites are met:

* The following minimum component versions are installed in your cluster:

| Component                  | Version          |
| :------------------------- | :--------------- |
| atlas-workload-autoscaler | v0.44.4 or later |
| atlas-agent               | v0.60.0 or later |

* The workload is a supported type: Deployment, StatefulSet, ReplicaSet, or Rollout (Argo)
* If the workload already has a native HPA that you want Atlas to manage, it must not be owned by a third-party controller such as KEDA and must use only basic resource targets (CPU or memory utilization). See [Eligibility](#eligibility) for full requirements.

# Configure via the Atlas console

## Enable HPA on a workload

1. Navigate to **Workload Autoscaler** > **Optimization**
2. Select the workload you want to configure
3. In the workload page, select the **Horizontal autoscaling** tab

If the workload does not yet have HPA configured (via a native Kubernetes HPA object), you will see one of two states:

* If HPA is configured in the workload's scaling policy, the workload inherits those settings and displays them here, allowing you to override them for this workload. Note that until HPA is enabled, these settings remain passive.

  <Image align="center" caption="Workload page with the Horizontal autoscaling tab open" src="https://files.readme.io/257e22707c2ec7f472bc5666f7e4e3c76a9311a16120c836805323085c9c0cfa-image.png" />

  <br />

* If the policy does not have HPA configured, you will see a **Missing HPA configuration** message. The options for adding HPA to a workload in the Console vary:
  * **Create HPA object** (workload-level override) is always available for an individual workload that supports HPA
  * **Configure in policy** (recommended) is only available for workloads that are not assigned to a system policy but to a custom policy that does not yet have HPA configuration defined.

4. To create a workload-level configuration, click **Create HPA object**

   ![](https://files.readme.io/f00b27e83d518c278609e69f420496e9a2256ba6a5992e6b3e612d10427cd4c6-image.png)

> 📘 Note
>
> Creating a workload-level HPA is a workload-level override. The workload's HPA settings are no longer inherited from its scaling policy. See <Anchor label="HPA in scaling policies" target="_blank" href="doc:woop-hpa-how-to-policies">HPA in scaling policies</Anchor> for details on how overrides work.

## Configure pod count

Under **Pod count**, set the minimum and maximum number of replicas the autoscaler can manage. The autoscaler will not scale below the minimum or above the maximum.

![](https://files.readme.io/bc8f3f2c7815919ce29dc2145c7a75b7fdfac8216b1312a06447facc329e01c4-image.png)

## Configure triggers

Under **Triggers**, configure the resource utilization metrics that drive scaling decisions. You can add CPU utilization, memory utilization, or both as triggers.

For each trigger, set the target utilization percentage. For example, setting CPU utilization to 70% means the autoscaler will attempt to keep average CPU utilization across all pods at approximately 70% by adding or removing replicas.

![](https://files.readme.io/8998c4bd30acd1d2dc11795b896ad3dd17750323b752328fd0fb8df7bd15f002-image.png)

## Configure autoscaling behavior (optional)

Under **Autoscaling behavior**, you can fine-tune how the autoscaler responds to metric changes. This section supports separate configuration for scale-up and scale-down directions.

**Stabilization windows** control how long the autoscaler waits before acting on new scaling recommendations, preventing rapid fluctuations (flapping):

| Field                        | Description                                                                        |
| :--------------------------- | :--------------------------------------------------------------------------------- |
| **Scale Up Stabilization**   | Seconds to wait before scaling up. Set to `0` for immediate response.              |
| **Scale Down Stabilization** | Seconds to wait before scaling down. `300` (5 minutes) is a common starting point. |

**Scaling policies** control the rate at which replicas are added or removed. You can define one or more policies per direction:

| Field           | Description                                                                                         |
| :-------------- | :-------------------------------------------------------------------------------------------------- |
| **Policy Type** | The unit for the rate limit: `Pods` (absolute count) or `Percent` (percentage of current replicas). |
| **Value**       | Maximum amount of change allowed per period.                                                        |
| **Period**      | Time window in seconds for the rate limit.                                                          |

When multiple policies are defined for the same direction, Kubernetes uses the policy that allows the most change by default. To control this selection behavior via annotations, see the `selectPolicy` field in the <Anchor label="Annotations reference" target="_blank" href="doc:workload-autoscaler-annotations-reference#behaviordirectionselectpolicy">Annotations reference</Anchor>.

These settings are optional. If not configured, Kubernetes defaults apply.

## Save your configuration

Click **Save changes** to apply your configuration. Workload Autoscaler will create a native `HorizontalPodAutoscaler` resource in your cluster matching this configuration.

# Configure via annotations

You can configure HPA using the `workloads.atlas.example/configuration` annotation on your workload manifest. This creates the same native HPA as the console method.

> 📘 Console or API recommended
>
> Annotations provide a declarative option for HPA configuration, but the Atlas console or API is generally easier to manage and validate. Annotations do not surface configuration errors until the workload is applied, whereas the console provides immediate feedback. Consider using annotations only when you need GitOps-style configuration or have specific workflow requirements.

## Basic configuration

The minimum required annotation configures the optimization mode, enables native HPA, sets replica bounds, and defines at least one scaling trigger:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  annotations:
    workloads.atlas.example/configuration: |
      horizontal:
        optimization: true
        useNative: true
        minReplicas: 2
        maxReplicas: 10
        metrics:
          - type: resource
            resource:
              name: cpu
              target:
                type: Utilization
                averageUtilization: 80
```

<Accordion title="Verify the HPA was created" icon="fa-terminal">
  After applying the annotation to your manifest, confirm that a native HPA resource exists and reflects your configuration:

  ```bash
  # Check that the HPA was created
  kubectl get hpa -n <namespace>

  # Inspect details — confirm target, min, and max replicas
  kubectl describe hpa <workload-name> -n <namespace>

  # Confirm the deployment scaled to at least minReplicas
  kubectl get <workload-kind> <workload-name> -n <namespace>
  ```

  Note that you can apply the annotation to any workload, but only eligible workloads will have the HPA resource created. If your workload is ineligible (for example, a CronJob or a workload with an incompatible existing HPA), the Atlas console displays a misconfiguration error explaining why the HPA was not created.
</Accordion>

## Configuration with scaling behavior

To add custom scaling behavior, include the `behavior` block. This example configures a conservative scale-down (10% of replicas per minute, with a 5-minute stabilization window) and a more aggressive scale-up (up to 4 pods per minute):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  annotations:
    workloads.atlas.example/configuration: |
      horizontal:
        optimization: true
        useNative: true
        minReplicas: 2
        maxReplicas: 20
        metrics:
          - type: resource
            resource:
              name: cpu
              target:
                type: Utilization
                averageUtilization: 80
        behavior:
          scaleUp:
            stabilizationWindowSeconds: 60
            selectPolicy: Max
            policies:
              - type: Pods
                value: 4
                periodSeconds: 60
          scaleDown:
            stabilizationWindowSeconds: 300
            selectPolicy: Min
            policies:
              - type: Percent
                value: 10
                periodSeconds: 60
```

<Accordion title="Verify the behavior configuration" icon="fa-terminal">
  Confirm the behavior spec was applied to the HPA:

  ```bash
  # Check the behavior section in the HPA description
  kubectl describe hpa <workload-name> -n <namespace> | grep -A 20 "Behavior:"

  # Or inspect the full YAML
  kubectl get hpa <workload-name> -n <namespace> -o yaml | grep -A 30 "behavior:"
  ```

  To observe the behavior in action, generate load against your workload and watch the HPA react:

  ```bash
  # Watch HPA scaling decisions in real time
  kubectl get hpa -n <namespace> --watch
  ```

  Scale-up should happen according to your scale-up policies.
</Accordion>

## Multiple triggers

You can configure both CPU and memory triggers. The autoscaler evaluates all triggers and uses the one that results in the highest replica count:

```yaml
annotations:
  workloads.atlas.example/configuration: |
    horizontal:
      optimization: true
      useNative: true
      minReplicas: 2
      maxReplicas: 15
      metrics:
        - type: resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 80
        - type: resource
          resource:
            name: memory
            target:
              type: Utilization
              averageUtilization: 80
```

<Accordion title="Verify both metrics are active" icon="fa-terminal">
  Confirm both metrics appear in the HPA:

  ```bash
  kubectl describe hpa <workload-name> -n <namespace> | grep -A 5 "Metrics:"
  ```

  Both CPU and memory targets should be listed.
</Accordion>

## Combining HPA and vertical autoscaling

To use both HPA and vertical autoscaling together, include both blocks in the same annotation. Atlas automatically coordinates between the two dimensions to prevent conflicts:

```yaml
annotations:
  workloads.atlas.example/configuration: |
   scalingPolicyName: balanced
   vertical:
      optimization: on
    horizontal:
      optimization: true
      useNative: true
      minReplicas: 2
      maxReplicas: 10
      metrics:
        - type: resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 80
```

For details on how vertical and horizontal scaling interact, see <Anchor label="Vertical & horizontal workload autoscaling" target="_blank" href="doc:workload-autoscaler-hpa-vpa">Vertical & horizontal workload autoscaling</Anchor>.

For the complete field reference, see the <Anchor label="horizontal section in the Annotations reference" target="_blank" href="doc:workload-autoscaler-annotations-reference#horizontal">horizontal section in the Annotations reference</Anchor>.

# Take ownership of an existing native HPA

If your workload already has a native Kubernetes HPA that is not managed by a third-party controller, Atlas can take ownership of it. This replaces the existing HPA's configuration with your Atlas-managed settings.

## Eligibility

Not all existing HPAs can be taken over. Atlas can only take ownership of HPAs that meet the following criteria:

* The HPA is not managed by a third-party controller (such as KEDA)
* The HPA uses only CPU or memory utilization metrics at the workload level
* The HPA does not use external metrics, pod metrics, or object metrics
* The HPA does not use container-level resource targets

If a workload's HPA uses unsupported features, the **Horizontal autoscaling** tab displays **Cannot transfer HPA ownership**, explaining that the workload's HPA uses features not currently supported by scaling policies. These workloads retain their existing HPA and are not affected by Atlas's horizontal autoscaling.

## Via scaling policies (recommended)

Taking ownership of existing HPAs is managed through scaling policies in the Atlas console. This applies the policy's HPA settings to all eligible workloads assigned to the policy.

1. Navigate to **Workload Autoscaler** > **Scaling policies**
2. Open the policy assigned to the workloads whose HPAs you want to take over
3. In the **Horizontal Autoscaling** tab, check **Replace existing HPAs with policy settings**
4. Configure your desired HPA settings (pod count, triggers, behavior)
5. Save the policy

For details, see <Anchor label="Enable take ownership at the policy level" target="_blank" href="doc:woop-hpa-how-to-policies#enable-take-ownership-at-the-policy-level">Enable take ownership at the policy level</Anchor>.

## Via annotations

To take ownership on a specific workload, set `takeOwnership: true` in the horizontal configuration block:

```yaml
annotations:
  workloads.atlas.example/configuration: |
    horizontal:
      optimization: true
      useNative: true
      takeOwnership: true    # controls ownership of native k8s HPA you had before
      minReplicas: 2
      maxReplicas: 10
      metrics:
        - type: resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 80
```

<Accordion title="Verify ownership transfer" icon="fa-terminal">
  The most reliable way to confirm Atlas has taken ownership is to check the `ownerReferences` field on the HPA resource for a Recommendation owner:

  ```bash
  kubectl get hpa <workload-name> -n <namespace> -o yaml | grep -A 5 "ownerReferences:"
  ```

  You should see a Recommendation resource listed as an owner, confirming that Workload Autoscaler owns this HPA.

  The HPA also includes an `autoscaling.atlas.example/hpa-revert-configuration` annotation containing a JSON snapshot of the original HPA spec as it existed before Atlas took ownership. This preserved configuration is what Atlas restores if you later revert ownership. For example:

  ```
  autoscaling.atlas.example/hpa-revert-configuration: '{"hpaSpec":{"scaleTargetRef":{...},"minReplicas":1,"maxReplicas":5,"metrics":[...]},"ownerReference":null}'
  ```
</Accordion>

> 📘 Important
>
> Even though taking ownership replaces the existing HPA configuration entirely, and the values you specify in your Workload Autoscaler configuration become the new HPA settings, the existing HPA configuration is preserved by Atlas and can be restored at any time.

# Webhook protection for managed HPAs

Once Workload Autoscaler owns an HPA object, a validating webhook protects it from external modifications:

* **Manual edits** to the `HorizontalPodAutoscaler` resource via `kubectl edit` or other tools are rejected by the webhook. All configuration changes must be made through the Atlas console, API, or annotations.
* **Manual deletes** of the `HorizontalPodAutoscaler` resource are also rejected. Even without the webhook, the `atlas-workload-autoscaler` would recreate a deleted HPA on its next reconciliation cycle as long as HPA is enabled for the workload.
  ```shell
  # Attempt to delete HPA
  kubectl delete hpa -n <namespace> --all

  # Admission webhook protection
  Error from server (Forbidden): admission webhook "hpa-ownership.atlas.example" denied the request: HPA is managed by workload-autoscaler
  ```
* To stop Atlas from managing an HPA, disable horizontal autoscaling for the workload through the Atlas console or API, uncheck the **Take ownership** option, or remove the `horizontal` block from the annotation. This is the correct way to remove a Workload Autoscaler-managed HPA.

> 📘 ArgoCD
>
> Because Workload Autoscaler modifies and protects the HPA resource, ArgoCD may detect a diff between the desired state in Git and the live cluster state. Atlas updates the replica field owner to prevent ArgoCD from reporting an out-of-sync status for replica counts, but the HPA resource itself will appear as managed by Workload Autoscaler rather than your Git repository. Consider excluding Workload Autoscaler-managed HPA resources from ArgoCD sync if you encounter persistent warnings.

# Revert HPA ownership

If you had Atlas take ownership of a workload's native HPA and want to return to the original configuration, you can revert ownership. Atlas preserves the original HPA spec at the time of takeover, so reverting restores the HPA to its pre-ownership state.

## Revert at the policy level

If ownership was enabled at the policy level via the **HPA ownership** toggle, revert through the policy:

1. Navigate to the policy's **Horizontal Rightsizing** tab
2. Toggle the **HPA ownership** to **off**
3. Save the policy

![](https://files.readme.io/5dce50d8751efa5144a07f7794cc163399c0940d15eab961008398ab1c654dab-image.png)

Atlas releases ownership on all workloads that were taken over through this policy and restores their original HPA configurations.

## Revert via annotations

If ownership was taken on a specific workload via the `takeOwnership: true` annotation, revert by updating the annotation:

```yaml
annotations:
  workloads.atlas.example/configuration: |
    horizontal:
      optimization: true
      useNative: true
      takeOwnership: false
      minReplicas: 2
      maxReplicas: 10
      metrics:
        - type: resource
          resource:
            name: cpu
            target:
              type: Utilization
              averageUtilization: 80
```

Setting `takeOwnership` to `false` (or removing the field entirely) causes Atlas to release ownership and restore the original HPA configuration from the preserved snapshot.

> 📘 Note
>
> There is no UI path to take or revert ownership on individual workloads. Workload-level ownership is managed exclusively through annotations or the API. For UI-based ownership management, use the policy-level approach.

## What gets restored

When you revert ownership, Atlas restores the exact HPA spec that existed before takeover. This is stored in the `autoscaling.atlas.example/hpa-revert-configuration` annotation on the HPA resource. You can inspect it before reverting to confirm what will be restored:

```bash
kubectl get hpa <workload-name> -n <namespace> \
  -o jsonpath='{.metadata.annotations.autoscaling\.atlas\.example/hpa-revert-configuration}' | jq .
```

The restored configuration includes the original `minReplicas`, `maxReplicas`, metrics, and behavior settings. If the original HPA had no owner reference (i.e., it was a standalone HPA), it returns to that state.

## Revert vs. disable

Reverting ownership and disabling horizontal autoscaling are different actions:

* **Revert ownership** restores the original native HPA that existed before Atlas took over. The HPA continues to function as it did before Atlas management, just without Atlas controlling it.
* **Disable horizontal autoscaling** removes the Atlas-managed HPA entirely. If the workload had no HPA before Atlas created one (i.e., Atlas created it from scratch rather than taking ownership), disabling removes the HPA. There is no original configuration to restore in this case.

<Accordion title="Verify the revert" icon="fa-terminal">
  After reverting, confirm the HPA has been restored to its original configuration:

  ```bash
  # Check that the Atlas management annotation is gone
  kubectl get hpa <workload-name> -n <namespace> -o yaml | grep -i atlas

  # Inspect the HPA to confirm original values are restored
  kubectl describe hpa <workload-name> -n <namespace>
  ```

  The `autoscaling.atlas.example/hpa-managed` annotation should no longer be present. The HPA's `minReplicas`, `maxReplicas`, and metrics should match the original values from the preserved configuration.
</Accordion>

# Verify your configuration

After configuring HPA, verify that the native HPA resource was created successfully.

## Check via kubectl

```bash
kubectl get hpa -n <namespace>
```

You should see a `HorizontalPodAutoscaler` resource for your workload, managed by Workload Autoscaler.

## Check via the Atlas console

In the workload's **Horizontal autoscaling** tab, you should see the current HPA configuration displayed.

The workload always shows an **H** badge when horizontal optimization is active:

![](https://files.readme.io/be8074da72c5ef1cc56f5d1a56eab67efb77867fd3a50e69dfeda75da589b770-image.png)

# See also

<Cards>
  <Card title="Horizontal autoscaling" href="https://docs.atlas.example/update/docs/horizontal-pod-autoscaling" icon="fa-arrows-left-right">
    Concepts, requirements, and compatibility.
  </Card>

  <Card title="Horizontal autoscaling in scaling policies" href="https://docs.atlas.example/update/docs/woop-hpa-how-to-policies" icon="fa-sliders">
    Configure HPA at the policy level (recommended).
  </Card>

  <Card title="Annotations reference" href="https://docs.atlas.example/docs/workload-autoscaler-annotations-reference#horizontal" icon="fa-book">
    Complete field reference.
  </Card>
</Cards>