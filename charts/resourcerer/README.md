# resourcerer Helm chart

Installs [resourcerer](https://github.com/gjbravi/resourcerer): CPU and memory recommendations from Prometheus usage
history, applied on request or automatically, with a UI.

Requires a Prometheus-compatible API with cAdvisor and kube-state-metrics data (`prometheus.url`).

```bash
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> -n resourcerer --create-namespace \
  --set prometheus.url=http://prometheus-server.monitoring.svc
```

| Value | Default | Description |
|---|---|---|
| `prometheus.url` | `http://prometheus-server.monitoring.svc` | Prometheus-compatible API. |
| `prometheus.rateWindow` | `5m` | `rate()` window for CPU; at least 4× the scrape interval. |
| `prometheus.minStep` | `30s` | Finest query resolution, about the scrape interval. |
| `clusterName` | `kubernetes` | Name shown in the UI. |
| `selection` | `all` | `all` (opt out with `resourcerer.io/enabled=false`) or `label` (opt in with `=true`). |
| `excludedNamespaces` | `kube-system`, `kube-public`, `kube-node-lease` | Never managed; the release namespace is added. |
| `defaults` | `{}` | Cluster defaults for settings, keyed by annotation name without `resourcerer.io/`. |
| `cost.cpuCoreHour`, `cost.memoryGiBHour` | `0.0316`, `0.0042` | USD rates for estimates. |
| `auth.strategy` | `anonymous` | Only `anonymous` for now. |
| `readOnly` | `false` | Refuse every change from the UI and API. |
| `applier.enabled` | `true` | Disable for a recommend-only install. |
| `controller.replicas`, `applier.replicas`, `server.replicas` | `1` | Leader election above 1 for controller and applier. |
| `server.ui` | `true` | Serve the embedded UI. |
| `server.service`, `server.ingress` | ClusterIP, disabled | The Ingress has no authentication. |
| `metrics.scrapeAnnotations` | `true` | `prometheus.io/scrape` on the controller metrics Service. |
| `metrics.serviceMonitor.enabled` | `false` | Prometheus Operator ServiceMonitor instead. |

The CRD is in `crds/` and installed by Helm on first install; upgrade it with `kubectl apply -f charts/resourcerer/crds`.
