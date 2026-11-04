---
title: Upgrade and uninstall
description: Upgrade both charts and their CRDs, and remove resourcerer cleanly.
---

## Upgrade

1. Read the [changelog](https://github.com/gjbravi/resourcerer/blob/main/CHANGELOG.md) for the versions you skip.
   While resourcerer is in beta, a minor version may rename a value or an annotation; the changelog says so under
   **Breaking changes**.
2. **Upgrade the CRDs first.** They are a release of their own, so this is what upgrades them:

    ```bash
    helm upgrade --install resourcerer-crds oci://ghcr.io/gjbravi/charts/resourcerer-crds --version <version> \
      --namespace resourcerer
    ```

3. Upgrade the release with the same values file:

    ```bash
    helm upgrade resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> \
      --namespace resourcerer -f resourcerer-values.yaml
    ```

Both charts are released together and take the same version. Recommendations survive an upgrade: they are custom
resources, and the new controller regenerates them on their next slot. Applied values are on the workloads themselves
and are not touched.

!!! warning "Upgrading from 0.9.0 or earlier: adopt the CRDs once"
    Those releases installed the CRDs from the chart's `crds/` directory, which Helm installs once, never upgrades and
    tracks no release for. The first `helm install resourcerer-crds` fails with *"exists and cannot be imported into
    the current release"*, and nothing changes. With Helm 4, pass `--take-ownership`. With Helm 3, label and annotate
    the three CRDs first:

    ```bash
    for crd in recommendations.resourcerer.io applypolicies.resourcerer.io clusterapplypolicies.resourcerer.io; do
      kubectl label crd "$crd" app.kubernetes.io/managed-by=Helm --overwrite
      kubectl annotate crd "$crd" meta.helm.sh/release-name=resourcerer-crds meta.helm.sh/release-namespace=<namespace> --overwrite
    done
    ```

    `<namespace>` is where the `resourcerer-crds` release lives, here `resourcerer`. Then upgrade the two releases as
    above. Upgrading `resourcerer` to the release without `crds/` deletes nothing: Helm never tracked those CRDs.

    That is the failure this release fixes. Up to 0.9.0 the CRDs were installed only on the first install, so an
    installation upgraded without a manual `kubectl apply` kept the old definitions: the apply policies ran ungated, or
    the API server pruned the decision written to `status`.

### With Argo CD or Flux

Install `resourcerer-crds` before `resourcerer`, and upgrade both in that order. A GitOps tool can otherwise try to
create the workloads that use the CRDs in the same sync, before the API server serves them.

- **Argo CD**: a second Application for `ghcr.io/gjbravi/charts/resourcerer-crds`, or a sync wave before the main
  Application. The static `resourcerer-crds-X.Y.Z.yaml` from the GitHub release works from a directory too.
- **Flux**: a second `HelmRelease` for `resourcerer-crds` that the main one waits for with
  `spec.dependsOn`, or the static file in a Kustomization applied before the main chart.

## Uninstall

```bash
helm uninstall resourcerer --namespace resourcerer
helm uninstall resourcerer-crds --namespace resourcerer
```

What stays:

- **Applied requests and limits stay on your workloads.** Uninstalling does not revert anything. To go back to the
  previous values, revert the workloads from the UI first, or let your next deploy from Git overwrite them.
- **HPA targets** changed by an apply stay as they are, for the same reason.
- **The CRDs and every `Recommendation`, `ApplyPolicy` and `ClusterApplyPolicy`.** Every CRD carries
  `helm.sh/resource-policy: keep`, so uninstalling `resourcerer-crds` does not delete them. Remove them when you are
  sure, which deletes every object of those kinds:

    ```bash
    kubectl delete crd recommendations.resourcerer.io applypolicies.resourcerer.io clusterapplypolicies.resourcerer.io
    ```

- **Annotations** `resourcerer.io/*` on namespaces and workloads, and the `resourcerer.io/enabled` labels. They do
  nothing without resourcerer; remove them with `kubectl annotate ... resourcerer.io/<key>-` if you want a clean slate.
