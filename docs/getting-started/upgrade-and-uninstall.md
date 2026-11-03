---
title: Upgrade and uninstall
description: Upgrade the chart and its CRDs, and remove resourcerer cleanly.
---

## Upgrade

1. Read the [changelog](https://github.com/gjbravi/resourcerer/blob/main/CHANGELOG.md) for the versions you skip.
   While resourcerer is in beta, a minor version may rename a value or an annotation; the changelog says so under
   **Breaking changes**.
2. **Apply the CRDs first.** Helm installs the CRDs in the chart's `crds/` directory on the first install only and never
   upgrades them:

    ```bash
    helm pull oci://ghcr.io/gjbravi/charts/resourcerer --version <version> --untar
    kubectl apply --server-side -f resourcerer/crds
    ```

3. Upgrade the release with the same values file:

    ```bash
    helm upgrade resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> \
      --namespace resourcerer -f resourcerer-values.yaml
    ```

Recommendations survive an upgrade: they are custom resources, and the new controller regenerates them on their next
slot. Applied values are on the workloads themselves and are not touched.

!!! warning "CRDs left behind"
    A release upgraded without its CRDs keeps working, but the features that need the new fields do not. For apply
    policies, the applier logs that their CRDs are missing and **runs ungated**, applying as if no policy existed, rather
    than refusing every apply.


### With Argo CD or Flux

Both can install the CRDs for you: Argo CD applies the chart's `crds/` on every sync, and Flux's `HelmRelease` does with
`install.crds: CreateReplace` and `upgrade.crds: CreateReplace`.

## Uninstall

```bash
helm uninstall resourcerer --namespace resourcerer
```

What stays:

- **Applied requests and limits stay on your workloads.** Uninstalling does not revert anything. To go back to the
  previous values, revert the workloads from the UI first, or let your next deploy from Git overwrite them.
- **HPA targets** changed by an apply stay as they are, for the same reason.
- **The CRDs and every `Recommendation`, `ApplyPolicy` and `ClusterApplyPolicy`.** Helm never deletes CRDs. Remove them
  when you are sure, which deletes every object of those kinds:

    ```bash
    kubectl delete crd recommendations.resourcerer.io applypolicies.resourcerer.io clusterapplypolicies.resourcerer.io
    ```

- **Annotations** `resourcerer.io/*` on namespaces and workloads, and the `resourcerer.io/enabled` labels. They do
  nothing without resourcerer; remove them with `kubectl annotate ... resourcerer.io/<key>-` if you want a clean slate.
