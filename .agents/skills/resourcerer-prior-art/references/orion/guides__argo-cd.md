# How to use Orion and Argo CD together

> Apply method and ignoreDifferences settings for using Orion with Argo CD


How you configure [Argo CD][argo-cd] depends on which Orion apply method you use:

- **[Webhook apply methods](#webhook) (recommended):** Orion sets recommendations on Pods at admission time and doesn't change the workload spec. Argo CD only needs to ignore Orion's changes to HPAs and ScaledObjects.
- **[Patch apply methods](#patch):** Orion writes recommendations into the workload spec. Argo CD needs to ignore those changes and keep them when it syncs, which has some [known issues](#patch-known-issues).

## Webhook apply methods {#webhook}

The webhook apply methods are `WebhookImmediateRollout`, `WebhookHybridRollout`, and `WebhookDeferredRollout`.

### Set a webhook apply method as the default

Set Orion's default apply method to a webhook apply method. We recommend setting it in the `defaults` rule of your cluster-wide ClusterOptimizationConfiguration:

```yaml
apiVersion: optimize.orion.example/v1
kind: ClusterOptimizationConfiguration
metadata:
  name: default-policy
spec:
  order: 10
  rules:
  - name: defaults
    settings:
    # ...your other default settings
    - name: apply.method
      value: WebhookImmediateRollout
```

See [Configure optimization using CRDs][optimization-configurations] for more on ClusterOptimizationConfigurations, and [Apply method][apply-method] for the webhook apply methods and how to enable the webhook on the Applier.

### Configure `ignoreDifferences` for HPAs and ScaledObjects

Orion writes HPA target utilization directly to HPA `spec.metrics` and ScaledObject `spec.triggers` so that workloads keep scaling at the same load after their Pods are resized. Add [`ignoreDifferences`][argo-cd-diffing] rules for these to the `argocd-cm` ConfigMap so that Argo CD doesn't treat Orion's changes as drift and revert them:

```yaml
# argocd-cm ConfigMap
data:
  resource.customizations.ignoreDifferences.autoscaling_HorizontalPodAutoscaler: |
    jqPathExpressions:
    - .spec.metrics[].resource.target.averageUtilization
    - .spec.metrics[].containerResource.target.averageUtilization
  resource.customizations.ignoreDifferences.keda.sh_ScaledObject: |
    jqPathExpressions:
    - .spec.triggers[] | select((.type == "cpu" or .type == "memory") and .metricType == "Utilization") | .metadata.value
```

The rules apply to every Application, including ones added later.

To configure the rules per Application instead, add them to `spec.ignoreDifferences` in each Application that includes HPAs or ScaledObjects:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: <your-application>
  namespace: argocd
spec:
  ignoreDifferences:
  - group: autoscaling
    kind: HorizontalPodAutoscaler
    jqPathExpressions:
    - .spec.metrics[].resource.target.averageUtilization
    - .spec.metrics[].containerResource.target.averageUtilization
  - group: keda.sh
    kind: ScaledObject
    jqPathExpressions:
    - .spec.triggers[] | select((.type == "cpu" or .type == "memory") and .metricType == "Utilization") | .metadata.value
```

### Keep `RespectIgnoreDifferences` disabled

Keep [`RespectIgnoreDifferences`][argo-cd-sync-options] at its default of `false` for Applications that manage HPAs or ScaledObjects:

```yaml
spec:
  syncPolicy:
    syncOptions:
    - RespectIgnoreDifferences=false
```

With this option enabled, Argo CD keeps the HPA's whole live `metrics` list, or the ScaledObject's whole live `triggers` list, when it syncs, so changes to those lists in Git are never applied. See [Argo CD issue #23283][argo-cd-23283].

## Patch apply methods {#patch}

The patch apply method is `PatchImmediateRollout`.

### Configure `ignoreDifferences` for workloads, HPAs, and ScaledObjects

Orion writes recommendations to the workload spec and to HPA and ScaledObject targets. Add `ignoreDifferences` rules for both to the `argocd-cm` ConfigMap:

```yaml
# argocd-cm ConfigMap
data:
  # Workload resource kinds
  resource.customizations.ignoreDifferences.apps_Deployment: |
    managedFieldsManagers:
    - orion
  resource.customizations.ignoreDifferences.apps_StatefulSet: |
    managedFieldsManagers:
    - orion
  resource.customizations.ignoreDifferences.apps_DaemonSet: |
    managedFieldsManagers:
    - orion

  # Autoscaler resource kinds
  resource.customizations.ignoreDifferences.autoscaling_HorizontalPodAutoscaler: |
    jqPathExpressions:
    - .spec.metrics[].resource.target.averageUtilization
    - .spec.metrics[].containerResource.target.averageUtilization
  resource.customizations.ignoreDifferences.keda.sh_ScaledObject: |
    jqPathExpressions:
    - .spec.triggers[] | select((.type == "cpu" or .type == "memory") and .metricType == "Utilization") | .metadata.value
```

The HPA and ScaledObject rules are the same as for [webhook apply methods](#webhook). Don't use `resource.customizations.ignoreDifferences.all`. The `managedFieldsManagers` rule is for workload kinds only.

### Enable `RespectIgnoreDifferences`

Add `RespectIgnoreDifferences=true` to each Application's sync options:

```yaml
spec:
  syncPolicy:
    syncOptions:
    - RespectIgnoreDifferences=true
```

Without it, Argo CD overwrites Orion's values with the values in Git each time it syncs. With it, changes in Git to a workload's container resources aren't applied while Orion manages that workload.

### Known issues {#patch-known-issues}

With `RespectIgnoreDifferences=true`, some changes in Git aren't applied ([Argo CD issue #23283][argo-cd-23283]). With patch apply methods, this means:

- **HPA and ScaledObject metric changes in Git aren't applied.** When you change an HPA's `metrics` or a ScaledObject's `triggers` in Git, the Application stays OutOfSync.
- **Argo Rollouts container changes in Git aren't applied.** Argo CD keeps the Rollout's live `containers` list, so image and other container changes in Git don't roll out. Without an ignore rule for Rollouts, Argo CD reverts Orion's changes instead.

To avoid both, use a webhook apply method, either as the default or only for affected workloads: Argo Rollouts and workloads with an HPA or ScaledObject.

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span>Settings<span class="related-topics-divider">/</span><a href="/docs/configure/settings/apply-settings/">Apply settings</a>
                </h5>
                <p>Enable automatic deployment, configure the learning period, set optional thresholds to skip low-impact recommendations, and choose an apply method</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Guides<span class="related-topics-divider">/</span><a href="/docs/guides/flux/">Orion and Flux</a>
                </h5>
                <p>Use a webhook-based apply method, and configure drift detection ignore rules for HPAs</p>
            </div>
        </div>


[apply-method]:          /docs/configure/settings/apply-settings/#apply-method
[argo-cd]:               https://argo-cd.readthedocs.io/
[argo-cd-23283]:         https://github.com/argoproj/argo-cd/issues/23283
[argo-cd-diffing]:       https://argo-cd.readthedocs.io/en/stable/user-guide/diffing/
[argo-cd-sync-options]:  https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/
[optimization-configurations]: /docs/configure/optimization-configurations/
