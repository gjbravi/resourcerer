# Apply settings

> Enable automatic deployment, configure the learning period, set optional thresholds to skip low-impact recommendations, and choose an apply method


Define how Orion deploys generated recommendations.

**Apply settings:**

- [Auto-deploy](#auto-deploy)
- [Learning period](#learning-period)
- [Auto-deploy thresholds](#thresholds)
- [Apply method](#apply-method)
- [Pod template metadata](#pod-template-metadata)
- [Incremental rollout](#incremental-rollout)

---

## Deploy recommendations automatically {#deploy-recommendations-automatically}

These settings define whether Orion deploys new recommendations automatically.

### Auto-deploy {#auto-deploy}

When `"Enabled"`, Orion automatically applies new recommendations when they are generated on the configured [schedule].

| Annotation                       | Default value |
| -------------------------------- | ------------- |
| `live.orion.example/auto-deploy` | `"Disabled"`  |

#### Description {#auto-deploy-description}

When auto-deploy is enabled (see [Examples](#auto-deploy-examples) below), Orion automatically applies new recommended settings when they are generated on the configured [schedule].

[The Applier must be installed][applier] in order for recommendations to be applied automatically. 

#### Examples {#auto-deploy-examples}

- To enable auto-deploy: `"Enabled"`
- To disable auto-deploy: `"Disabled"`

#### Settings that require auto-deploy {#auto-deploy-required}

The following settings are in effect only when auto-deploy is enabled:

- [Auto-deploy thresholds](#thresholds)
- [Continuous reconciliation](installation__applier-config.md#continuous-rec)
- [OOM events response](configure__settings__reliability.md#oom)
- [CPU pressure response](configure__settings__reliability.md#cpu-pressure)

---

### Learning period {#learning-period}

Acts as an automation gate, defining how long Orion observes a new workload before providing recommendations that can be auto-deployed.

| Annotation                           | Default value |
| ------------------------------------ | ------------- |
| `live.orion.example/learning-period` | `"P7D"`       |

#### Description {#learning-period-description}

During this observation or _learning period_, Orion generates _preliminary recommendations_ based on the metrics it has collected. Preliminary recommendations are not auto-deployed, but they can be applied on demand.

If you want to automatically deploy recommendations sooner-for example, for short-running workloads, branch-based deploys, or other ephemeral environments-you can shorten the learning period. Otherwise, it typically isn't necessary to change this value.

The shortest valid duration is one day (`P1D`).

Recommendations generated after the learning period is complete are deployed automatically if [auto-deploy](#auto-deploy) is enabled.

#### Duration expressions {#learning-period-duration-examples}

Only valid ISO 8601 [Duration][duration] strings can be specified.

#### Examples {#learning-period-examples}

- "`P1D`" – One day (minimum value)
- "`P3D`" – Three days
- "`P7D`" – One week (default)

---

### Auto-deploy thresholds {#thresholds}

Optional. Prevent Orion from automatically applying low-impact recommendations by defining the minimum amount of change required between current and recommended values.

| Annotation                                                 | Default value |
| ---------------------------------------------------------- | ------------- |
| `live.orion.example/auto-deploy.thresholds.cpu.percent`    | None          |
| `live.orion.example/auto-deploy.thresholds.memory.percent` | None          |
| `live.orion.example/auto-deploy.thresholds.cpu.unit`       | None          |
| `live.orion.example/auto-deploy.thresholds.memory.unit`    | None          |

#### Description {#thresholds-description}

Auto-deploy thresholds are a constraint that can prevent automatic deployment of recommendations generated on a [schedule]. If a threshold is configured, the amount of change between the new recommended values and the values currently applied to the workload must meet the defined minimum for the recommendation to be automatically deployed.

#### Percent thresholds

To use percent thresholds, specify a percent value for **both** CPU **and** for memory. The threshold is met if the amount of change for either resource exceeds the configured minimum percent.

##### Valid values {#thresholds-percent-valid-values}

Any positive integer string. Examples:

- `"5"`
- `"10"`

#### Unit thresholds

To use unit thresholds, specify a unit value for **both** CPU **and** for memory. The threshold is met if the amount of change for either resource exceeds the configured minimum unit value.

##### Valid values {#thresholds-unit-valid-values}

Any positive Kubernetes quantity, appropriate to the resource. Examples:

- `"10m"` (for CPU)
- `"64Mi"` (for memory)

#### Key points {#thresholds-key-points}

- Thresholds have no effect unless the [`live.orion.example/auto-deploy`](#auto-deploy) annotation is set to `"Enabled"`.
- If, for **at least one container**, the recommendation satisfies the thresholds for **at least one resource** — as in, CPU (unit AND percent threshold) OR memory (unit AND percent threshold) — the recommendation will be applied automatically.
- It is highly recommended to set percent **and** unit thresholds. If you set a _percent_ threshold only, small unit changes for small containers can be a large enough percentage to meet the percent threshold, thus deploying the recommendation. Setting percent **and** unit thresholds ensures small changes don't trigger a deployment.
- To disable thresholds, set them to `"0"`.

---

## General apply options  {#general-apply-options}

These settings define how Orion applies recommendations, regardless of whether auto-deploy is enabled.

### Apply method {#apply-method}

Defines how the Applier applies recommendations to a workload, and what happens to running Pods.

| Annotation                        | Default value             |
| --------------------------------- | ------------------------  |
| `live.orion.example/apply.method` | `"PatchImmediateRollout"` |

#### Description {#apply-method-description}

When a recommendation is applied, the apply method controls:
- How the Applier communicates resource changes to Kubernetes
- What happens to running Pods

By default, Orion uses Server-Side Apply patches (with `PatchImmediateRollout`) to directly modify Deployments, StatefulSets, and other Pod-owning workload resources.

Alternatively, you can configure Orion to use a mutating admission webhook (with `WebhookImmediateRollout`, `WebhookHybridRollout`, or `WebhookDeferredRollout`) to modify Pods at admission.

`WebhookImmediateRollout` and `WebhookHybridRollout` support [**in-place Pod resizing**](https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/), a Kubernetes feature (enabled by default from Kubernetes v1.33) that allows resource requests and limits to be updated on running Pods without a restart.

If Pods can't be resized in-place (for example, on clusters running earlier Kubernetes versions), the supporting methods either trigger a rollout/restart or defer changes until the workload restarts by itself. See [Choosing an apply method](#choosing-an-apply-method) for details.

`PatchImmediateRollout` always triggers a rollout, and `WebhookDeferredRollout` always defers changes until the next restart.

Depending on the configured apply method and the cluster's Kubernetes version, applying a recommendation may result in:
- Pods being resized in-place
- A rollout/restart being triggered
- Changes being deferred until the next rollout or restart

You can configure the apply method at the [cluster], [namespace], or [workload] level.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Webhook-based methods</div>


- Webhook-based methods require `webhook.enabled=true` on the Applier. If the webhook feature isn't enabled, recommendations are still generated, but can't be applied. See [Enable or disable the webhook](installation__webhook-config.md#webhook-enabled).
- Ensure the namespace that uses the webhook isn't on the `webhook.denyNamespaces` list.
- If using Argo CD, configure an `ignoreDifferences` element for HPA objects in the `argocd-cm` ConfigMap. See the [How to use Orion and Argo CD together guide][argo-cd] for details.
</div>


#### Valid values {#apply-method-valid-values}

##### `PatchImmediateRollout` (default) {#patch-immediate-rollout}

Use Server-Side Apply patches to modify Deployments, StatefulSets, and other Pod-owning workload resources directly when a recommendation is applied.

This is the default apply method.

##### `WebhookImmediateRollout` {#webhook-immediate-rollout}

Use a mutating webhook to modify Pods at admission. Pods are resized in-place when a recommendation is applied. If Pods can't be resized in-place, a rollout/restart is triggered.

##### `WebhookHybridRollout` {#webhook-hybrid-rollout}

Use a mutating webhook to modify Pods at admission. Pods are resized in-place when a recommendation is applied. If Pods can't be resized in place, changes are deferred until the next rollout.

##### `WebhookDeferredRollout` {#webhook-deferred-rollout}

Use a mutating webhook to modify Pods at admission. Pods aren't modified and the workload doesn't restart when a recommendation is applied. Changes are deferred until the next restart.

#### Choosing an apply method {#choosing-an-apply-method}

| Apply method              | Mechanism                               | In-place resize | If Pods can't be resized in-place       |
| ------------------------- | --------------------------------------- | --------------- | --------------------------------------- |
| `PatchImmediateRollout`   | Patch workload resources                | No              | N/A                                     |
| `WebhookImmediateRollout` | In-place Pod resizing, mutating webhook | Yes             | Rollout/restart triggered               |
| `WebhookHybridRollout`    | In-place Pod resizing, mutating webhook | Yes             | Changes deferred until the next restart |
| `WebhookDeferredRollout`  | Mutating webhook                        | No              | N/A                                     |

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Deprecated apply methods</div>


With the introduction of support for in-place Pod resizing, the apply methods `PatchWorkloadResources` and `DynamicAdmissionWebhook` are deprecated. Use `PatchImmediateRollout`, `WebhookImmediateRollout`, `WebhookHybridRollout`, or `WebhookDeferredRollout` instead.
</div>


### Pod template metadata {#pod-template-metadata}

Applies only to `WebhookImmediateRollout`. Defines where and how Orion annotates a workload's Pod template to trigger a rollout.

| Annotation                                              | Default value |
| ------------------------------------------------------- | ------------- |
| `live.orion.example/pod-template.metadata.patch-path`   | None          |
| `live.orion.example/pod-template.metadata.patch-format` | None          |

#### Description {#pod-template-metadata-description}

When Pods can't be resized in place, `WebhookImmediateRollout` triggers a rollout by adding the `orion.example/recommendation-url` annotation to the workload's Pod template. The webhook then sets the recommended values on the new Pods.

For most built-in workload kinds, Orion writes the annotation to the Pod template's metadata by default. Change these settings when:

- A custom resource stores Pod annotations in a nonstandard location. Set the patch path to that location and the patch format to match. See [`patchTargetDefaults`][patch-target-defaults].
- A workload's Pod template can't change after creation, such as a standalone Job. Set the patch path to `"-"`. See [Missing patch path for a Job][missing-patch-path].

#### Valid values {#pod-template-metadata-valid-values}

- `patch-path`: A path in the workload, or `"-"` to skip the annotation. With `"-"`, if Pods can't be resized in place, changes are deferred until the next restart.
- `patch-format`: A Go template for the value written at the patch path, matched to what the path points to:
  - Pod metadata (default for most workloads): `{"annotations":{"orion.example/recommendation-url": "{{ .RecommendationURL }}"}}`
  - An annotations map, such as `"/spec/podAnnotations"`: `{"orion.example/recommendation-url": "{{ .RecommendationURL }}"}`

### Incremental rollout {#incremental-rollout}

Incremental rollout lets you configure Orion to apply changes to resource settings gradually over time, rather than all at once.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Cost projection impact</div>


Using this feature may affect the projected cost impact shown in the Orion UI.

Cost projections are based on the next change Orion will apply, not on the final optimization state. Using this feature therefore restricts the magnitude of the total impact projection shown.

Incremental rollout is eventually consistent. The optimization results with and without use of this feature will eventually converge, even if the cost projection doesn't show it.
</div>


| Annotation                                      | Default value |
| ----------------------------------------------- | ------------- |
| `live.orion.example/apply.max-percent-increase` | None          |
| `live.orion.example/apply.max-percent-decrease` | None          |

#### Description {#incremental-rollout-description}

Orion normally applies its rightsizing recommendations in full, regardless of how much larger than or smaller than current values they are.

Incremental rollout, controlled with max percent increase and max percent decrease options, allows users to restrict the magnitude of individual changes Orion applies. This results in Orion applying a series of incremental changes to workloads that eventually converge them with their optimal values, rather than applying a single large change.

The allowed percent increase or percent decrease is based on the workload's current request values. For example, if a workload currently has a CPU request of `100m`, Orion is recommending a CPU request of `50m`, but the `max-percent-decrease` is set to `"20"`, Orion will apply the following changes over time:

* **Change 1:** From `100m` to `80m` (reducing `100m` by 20%)
* **Change 2:** From `80m` to `64m` (reducing `80m` by 20%)
* **Change 3:** From `64m` to `52m` (reducing `64m` by 20%)
* **Change 4:** From `52m` to `50m` (reaching the final recommended CPU request)

When each change is applied depends on the workload's [schedule][schedule]. If the schedule is `@daily`, the recommendation will be applied incrementally over 4 days.

#### Examples {#incremental-rollout-examples}

- To let Orion always apply recommendations in full:

      # Don't configure either of these settings.

- To let Orion always increase requests by its full recommended amount, but limit any reduction of requests to 20% at a time:

      live.orion.example/apply.max-percent-decrease: "20"

- To prevent Orion from raising request values at all:

      live.orion.example/apply.max-percent-increase: "0"

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/apply/">Apply recommendations</a>
                </h5>
                <p>Apply recommendations using the method that suits your environment</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Guides<span class="related-topics-divider">/</span><a href="/docs/guides/argo-cd/">Orion and Argo CD</a>
                </h5>
                <p>Apply method and ignoreDifferences settings for using Orion with Argo CD</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Guides<span class="related-topics-divider">/</span><a href="/docs/guides/flux/">Orion and Flux</a>
                </h5>
                <p>Use a webhook-based apply method, and configure drift detection ignore rules for HPAs</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span>Settings<span class="related-topics-divider">/</span><a href="/docs/configure/settings/reliability/">Reliability</a>
                </h5>
                <p>Respond to OOM and CPU pressure events with a temporary resource bump-up</p>
            </div>
        </div>


[applier]: /docs/installation/install/#orion-applier
[schedule]: /docs/configure/settings/schedule/#schedule
[argo-cd]: /docs/guides/argo-cd/
[cluster]:  /docs/configure/cluster-defaults/
[workload]: /docs/configure/annotate-workloads/
[namespace]: /docs/configure/annotate-namespaces/
[duration]: <https://en.wikipedia.org/wiki/ISO_8601#Durations>
[patch-target-defaults]: /docs/recommendations/custom-workload-kinds/#patchtargetdefaults
[missing-patch-path]: /docs/reference/troubleshooting/#err-missing-patch-path
