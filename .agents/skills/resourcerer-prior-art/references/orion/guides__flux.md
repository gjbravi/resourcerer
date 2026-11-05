# How to use Orion and Flux together

> Use a webhook-based apply method, and configure drift detection ignore rules for HPAs


Orion and [Flux][flux] are fully compatible:

- **Flux-managed workloads:** Recommendations are injected into Pods at admission time when using a [webhook-based apply method][webhook-applier]. This leaves Flux's reconciliation of the parent workload resource undisturbed.
- **Workloads with an HPA:** Orion writes directly to `spec.metrics`. `driftDetection.ignore` is required, otherwise Flux overwrites those changes on each reconciliation.

## Configure a webhook-based apply method

Set the `live.orion.example/apply.method` annotation to a webhook-based value:

```yaml
live.orion.example/apply.method: WebhookImmediateRollout
```

See [Apply method][apply-method] for the full list of webhook rollout options and how to enable the webhook on the Applier.

## Configure `driftDetection.ignore` for HPAs

[Configure a patch in your root Flux Kustomization][kustomization-patches] to propagate the setting to every `HelmRelease`:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: flux-system
  namespace: flux-system
spec:
  patches:
  - target:
      kind: HelmRelease
    patch: |
      apiVersion: helm.toolkit.fluxcd.io/v2
      kind: HelmRelease
      metadata:
        name: not-used
      spec:
        driftDetection:
          ignore:
          - paths: ["/spec/metrics"]
            target:
              kind: HorizontalPodAutoscaler
```

The patch reaches every `HelmRelease` managed by the root Kustomization, including ones added later.

If you prefer to configure this per `HelmRelease`, or if your cluster does not use a root Kustomization, add the `driftDetection.ignore` block directly to each `HelmRelease` that manages HPAs:

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: <your-helm-release>
  namespace: <your-namespace>
spec:
  driftDetection:
    ignore:
    - paths: ["/spec/metrics"]
      target:
        kind: HorizontalPodAutoscaler
```

For more on the underlying mechanism, see:

- [Server-Side Apply][ssa] in the Kubernetes docs
- [Drift detection][drift-detection] in the Flux docs

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
                    
                    Guides<span class="related-topics-divider">/</span><a href="/docs/guides/argo-cd/">Orion and Argo CD</a>
                </h5>
                <p>Apply method and ignoreDifferences settings for using Orion with Argo CD</p>
            </div>
        </div>


[apply-method]:     /docs/configure/settings/apply-settings/#apply-method
[kustomization-patches]: https://fluxcd.io/flux/components/kustomize/kustomizations/#patches
[drift-detection]:  https://fluxcd.io/flux/components/helm/helmreleases/#drift-detection
[flux]:             https://fluxcd.io/
[ssa]:              https://kubernetes.io/docs/reference/using-api/server-side-apply/
[webhook-applier]:  /docs/installation/webhook-config/#webhook-enabled
