# resourcerer-crds Helm chart

Installs the custom resource definitions of [resourcerer](https://github.com/gjbravi/resourcerer):
`Recommendation`, `ApplyPolicy`, `ClusterApplyPolicy` and `DecisionModel`. They are ordinary templates rather than
files in `crds/`, so `helm upgrade` maintains them, and each carries `helm.sh/resource-policy: keep`, so uninstalling
leaves them and their objects in place. The chart has no values.

## Install

Install it as a release of its own, before `resourcerer` and with the same version:

```bash
helm install resourcerer-crds oci://ghcr.io/gjbravi/charts/resourcerer-crds --version <version> \
  -n resourcerer --create-namespace
```

The same definitions, without Helm labels, are in [`files/crds.yaml`](files/crds.yaml) for `kubectl apply`.
Upgrades, adopting the CRDs of 0.9.0 or earlier, GitOps and uninstall are in
[Upgrade and uninstall](https://gjbravi.github.io/resourcerer/getting-started/upgrade-and-uninstall/).
