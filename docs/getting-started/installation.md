---
title: Installation
description: Install the Helm chart with a values file, verify the release signatures, and choose between a full and a recommend-only install.
---

resourcerer ships as one OCI Helm chart, `oci://ghcr.io/gjbravi/charts/resourcerer`, and one image,
`ghcr.io/gjbravi/resourcerer`. Versions are listed on the
[releases page](https://github.com/gjbravi/resourcerer/releases).

## Install with a values file

Keep your configuration in a values file, so it can live in Git next to the rest of the cluster:

```yaml title="resourcerer-values.yaml"
clusterName: production

prometheus:
  url: http://thanos-query.monitoring.svc:9090

# Cluster defaults for the optimization settings; namespaces and workloads override them with annotations.
defaults:
  cpu.optimization-goal: Balanced
  memory.optimization-goal: Balanced
  learning-period: P7D
  schedule: "@daily"
  auto-deploy: Disabled

# Namespaces that are never managed, on top of the release namespace.
excludedNamespaces:
  - kube-system
  - kube-public
  - kube-node-lease
  - monitoring
```

```bash
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> \
  --namespace resourcerer --create-namespace -f resourcerer-values.yaml
```

Every value is listed in [Helm values](../reference/helm-values.md). The chart's `values.schema.json` rejects unknown
keys, so a typo fails the install instead of being ignored.

To install from a checkout instead: `helm install resourcerer ./charts/resourcerer -n resourcerer --create-namespace`.

## Recommend-only and read-only installs

Two values let you start without anything that can change a workload:

| Install | Values | Effect |
|---|---|---|
| Full (default) | | The UI can apply, revert and edit settings; auto-deploy works where enabled. |
| Recommend-only | `applier.enabled: false` | No applier Deployment and no RBAC to patch workloads. Recommendations, the UI and patches still work, and nothing carries out an apply. |
| Read-only UI | `readOnly: true` | The server refuses every change from the UI and API (apply, revert, regenerate, settings), and its RBAC drops the `patch` verbs. The applier still auto-deploys where enabled. |

Use both together for a reporting install: recommendations and patches to copy into Git, nothing else. Without
`readOnly`, the UI of a recommend-only install still offers **Apply**, and the request it records waits for an
applier that is not there.

## Verify the release

Images and charts are signed with [cosign](https://github.com/sigstore/cosign) keyless signing from the release
workflow, and images carry an SBOM and SLSA provenance:

```bash
cosign verify ghcr.io/gjbravi/resourcerer:<version> \
  --certificate-identity-regexp '^https://github.com/gjbravi/resourcerer/\.github/workflows/release\.yaml@refs/(heads/main|tags/v.+)$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

cosign verify ghcr.io/gjbravi/charts/resourcerer:<version> \
  --certificate-identity-regexp '^https://github.com/gjbravi/resourcerer/\.github/workflows/release\.yaml@refs/(heads/main|tags/v.+)$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

## What gets installed

| Object | Purpose |
|---|---|
| CRDs `recommendations`, `applypolicies`, `clusterapplypolicies` (`resourcerer.io`) | One `Recommendation` per workload; apply policies |
| Deployments `<release>-controller`, `<release>-applier`, `<release>-server` | The three components, one image |
| ConfigMap `<release>-defaults` | Settings defaults, cost and reliability configuration, read by the controller |
| Service `<release>` | The UI and API, port 80 |
| Services `<release>-metrics`, `<release>-applier-metrics` | Prometheus metrics of the controller and the applier |
| ServiceAccounts, ClusterRoles, ClusterRoleBindings | One set per component, see [Permissions](../reference/permissions.md) |
| `ClusterApplyPolicy` objects | Only those listed under `applyPolicies.cluster` |
| Ingress, ServiceMonitor | Only when enabled |

The pods run as non-root with a read-only root filesystem, no privilege escalation and all capabilities dropped.

## Running more than one replica

`controller.replicas` and `applier.replicas` above 1 use leader election: one replica works, the others wait to take
over. `server.replicas` above 1 all serve requests.

## Next steps

- [Open the UI](../guides/exposing-the-ui.md) for your team, behind authentication.
- [Configure the settings](../configuration/overview.md) that matter most.
