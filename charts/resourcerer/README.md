# resourcerer Helm chart

Installs [resourcerer](https://github.com/gjbravi/resourcerer), which creates and manages Vertical Pod Autoscalers
for your workloads and serves a UI to review their recommendations.

The [Vertical Pod Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) must be
installed separately.

```bash
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer -n resourcerer --create-namespace
```

## Values

| Key | Default | Description |
|---|---|---|
| `selection` | `all` | `all`: every workload outside `excludedNamespaces`, opt out with `resourcerer.io/enabled=false`. `label`: only workloads with `resourcerer.io/enabled=true` on the workload or namespace. |
| `defaultMode` | `Off` | VPA mode without a `resourcerer.io/mode` annotation: `Off` or `Auto` (`InPlaceOrRecreate`). |
| `excludedNamespaces` | `[kube-system, kube-public, kube-node-lease]` | Never managed. The release namespace is always added. |
| `ui.enabled` | `true` | Serve the embedded UI. |
| `ui.readOnly` | `false` | Disallow mode changes; also removes `patch` on workloads from the ClusterRole. |
| `replicaCount` | `1` | Leader election is enabled above 1. |
| `image.repository` | `ghcr.io/gjbravi/resourcerer` | |
| `image.tag` | chart `appVersion` | |
| `service.type` / `service.port` | `ClusterIP` / `80` | |
| `ingress.enabled` | `false` | The UI has no authentication; put it behind an authenticating proxy. |
| `metrics.enabled` | `true` | controller-runtime metrics on port 8081. |
| `metrics.serviceMonitor.enabled` | `false` | Create a Prometheus Operator ServiceMonitor. |
| `extraArgs` | `[]` | Extra flags, e.g. `--zap-log-level=debug`. |
| `resources` | 10m / 64Mi requests, 256Mi limit | |

Security context, node selector, tolerations, affinity and priority class are configurable; see
[values.yaml](values.yaml).
