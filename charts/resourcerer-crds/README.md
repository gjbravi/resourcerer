# resourcerer-crds Helm chart

Installs the custom resource definitions of [resourcerer](https://github.com/gjbravi/resourcerer):

| CRD | Kind | Scope |
|---|---|---|
| `recommendations.resourcerer.io` | `Recommendation` | namespaced |
| `applypolicies.resourcerer.io` | `ApplyPolicy` | namespaced |
| `clusterapplypolicies.resourcerer.io` | `ClusterApplyPolicy` | cluster |

They are rendered as ordinary resources rather than files in `crds/`, so `helm upgrade` maintains them. Install this
chart as a release of its own, **before** `resourcerer`: the main chart can declare `ClusterApplyPolicy` objects from
`applyPolicies.cluster`, and Helm resolves every object of a release before creating any of them, so a CRD it needs
cannot come from that same release.

The chart has no values: `values.yaml` is empty and `values.schema.json` rejects every key.

## Install

```bash
helm install resourcerer-crds oci://ghcr.io/gjbravi/charts/resourcerer-crds --version <version> \
  -n resourcerer --create-namespace
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> \
  -n resourcerer --create-namespace -f resourcerer-values.yaml
```

The release namespace only records the release; the CRDs are cluster-scoped.

An upgrade runs the same two commands in the same order, with the same version: both charts are released together.
Recommendations and policies survive it.

## Without Helm

The same definitions are in [`files/crds.yaml`](files/crds.yaml), without Helm labels, for a team that owns its CRDs.
The release attaches that file as `resourcerer-crds-X.Y.Z.yaml`:

```bash
kubectl apply --server-side -f resourcerer-crds-<version>.yaml
```

A team that skips the chart and the file keeps the previous CRDs. That is fine until a release adds a field: the API
server prunes what the new version writes to `status`, silently.

## Adopting the CRDs on an existing installation

Releases up to 0.9.0 delivered the CRDs in the chart's `crds/` directory, which Helm installs once and never upgrades
and tracks no release for. On such a cluster the first `helm install resourcerer-crds` fails with *"exists and cannot
be imported into the current release"*, and nothing changes. Adopt them once:

```bash
# Helm 4
helm install resourcerer-crds oci://ghcr.io/gjbravi/charts/resourcerer-crds --version X.Y.Z --take-ownership

# Helm 3
for crd in recommendations.resourcerer.io applypolicies.resourcerer.io clusterapplypolicies.resourcerer.io; do
  kubectl label crd "$crd" app.kubernetes.io/managed-by=Helm --overwrite
  kubectl annotate crd "$crd" meta.helm.sh/release-name=resourcerer-crds meta.helm.sh/release-namespace=<namespace> --overwrite
done
```

`<namespace>` is where the `resourcerer-crds` release lives, here `resourcerer`. Then install the chart as above.
Upgrading `resourcerer` to the release without `crds/` deletes nothing: Helm never tracked those CRDs.

## Uninstall

```bash
helm uninstall resourcerer -n resourcerer
helm uninstall resourcerer-crds -n resourcerer
```

Uninstall `resourcerer` first: its applier reads the policies and its controller writes Recommendations.

Every CRD carries `helm.sh/resource-policy: keep`, so uninstalling `resourcerer-crds` leaves the CRDs and every
`Recommendation`, `ApplyPolicy` and `ClusterApplyPolicy` in place. Deleting them takes a deliberate command, which also
deletes every object of those kinds:

```bash
kubectl delete crd recommendations.resourcerer.io applypolicies.resourcerer.io clusterapplypolicies.resourcerer.io
```
