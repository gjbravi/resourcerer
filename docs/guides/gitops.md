---
title: GitOps
description: Use resourcerer with Argo CD, Flux or Helm-managed workloads without the tools undoing each other's changes.
---

If your workloads are deployed from Git, a change made in the cluster is drift, and your GitOps tool may put the Git
value back on its next sync. There are two ways to work, and they can be mixed per namespace.

## Option 1: commit the recommendation

Keep Git the only source of resources. Turn auto-deploy off (the default), and take each recommendation from the
**Patches** tab of the workload page:

- **Kustomize patch**: a strategic merge patch with the containers' `resources`, for an overlay;
- **Helm values**: the `resources:` block per container, in the layout most charts use;
- **`kubectl`**: a command, for a quick test.

Commit it, let the tool sync, and the next regeneration sees the new requests: the workload becomes "Already
right-sized". Optionally set `readOnly: true` so nobody applies from the UI by mistake.

## Option 2: let resourcerer own the resources

Apply from the UI or with auto-deploy, and tell the GitOps tool to leave the fields resourcerer owns alone. The applier
uses **server-side apply** with the field manager **`resourcerer`**, so its fields can be told apart.

### Argo CD

Ignore the fields managed by `resourcerer` on the kinds it changes:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  ignoreDifferences:
    - group: apps
      kind: Deployment
      managedFieldsManagers: [resourcerer]
    - group: apps
      kind: StatefulSet
      managedFieldsManagers: [resourcerer]
    - group: apps
      kind: DaemonSet
      managedFieldsManagers: [resourcerer]
    - group: argoproj.io
      kind: Rollout
      managedFieldsManagers: [resourcerer]
    # The HPA's CPU target, adjusted with the requests (a merge patch, not server-side apply).
    - group: autoscaling
      kind: HorizontalPodAutoscaler
      jqPathExpressions:
        - .spec.metrics[] | select(.resource.name == "cpu") | .resource.target.averageUtilization
  syncPolicy:
    syncOptions:
      - RespectIgnoreDifferences=true
```

`RespectIgnoreDifferences=true` makes a sync keep the ignored fields as they are in the cluster instead of only hiding
the diff. Or set the same `ignoreDifferences` once for all applications in `argocd-cm`
(`resource.customizations.ignoreDifferences.all`).

### Flux

Flux's kustomize-controller applies manifests with server-side apply and corrects drift on the fields set in Git. To
let resourcerer own the resources:

- **leave `resources` out of the manifests** of the workloads resourcerer manages (or out of the overlay for that
  environment), so Flux has no value to put back; or
- mark the workload with `kustomize.toolkit.fluxcd.io/ssa: IfNotPresent`, so Flux creates it but does not override it
  afterwards, at the cost of not syncing any later change from Git to that object.

For a `HelmRelease`, set the chart's `resources` values empty for those workloads, or use option 1.

### Helm without a GitOps tool

`helm upgrade` renders the chart's `resources` again. With values that set them, the next upgrade replaces what
resourcerer applied; the next regeneration then finds the requests changed and recommends again. Take the Helm values
from the **Patches** tab into your values file to stop the back and forth.

## Settings in Git

Settings are annotations (`resourcerer.io/<key>`) on namespaces and workloads, so they belong in Git with the
manifests. A change from the UI's **Settings** tab writes the same annotation in the cluster, which your GitOps tool
sees as drift. Either make settings changes in Git only, or ignore the annotations you change from the UI, such as the
auto-deploy switch (`~1` stands for `/` in a JSON pointer):

```yaml
ignoreDifferences:
  - group: apps
    kind: Deployment
    jsonPointers:
      - /metadata/annotations/resourcerer.io~1auto-deploy
```

Apply policies are custom resources: keep them in Git like any other object, or in the chart values under
`applyPolicies.cluster`.
