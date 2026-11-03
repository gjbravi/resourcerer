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
| `cost.source` | `auto` | `static` (`cost.rates`), `karpenter` (rates calibrated to Karpenter node prices, an error when missing) or `auto` (Karpenter when available). |
| `reliability.builtin.<id>` | all enabled | Turn a built-in signal off (`enabled: false`) or set its `warn` level. See `docs/guides/reliability-signals.md`. |
| `reliability.signals` | `[]` | Signals declared as PromQL templates, evaluated per workload (at most 20). |
| `applyPolicies.cluster` | `[]` | `ClusterApplyPolicy` objects created with the release: when an apply may start, and how many run at once. See `docs/guides/apply-policies.md`. |
| `cost.allocation` | `cluster` | `cluster` (every workload at the cluster's average rates) or `node` (each workload at the rates of the nodes its pods ran on; needs Karpenter prices). |
| `cost.rates.cpuCoreHour`, `cost.rates.memoryGiBHour` | `0.0316`, `0.0042` | USD rates; with Karpenter, the CPU to memory ratio. |
| `cost.karpenter.refreshInterval` | `10m` | How often rates are recalibrated from Karpenter prices; at least `30s`. |
| `auth.strategy` | `anonymous` | `anonymous` (anyone who reaches the server may use it, with the server's own permissions) or `oidc` (sign in at an OpenID Connect provider, roles from group claims). |
| `auth.cookieSecure` | `auto` | `Secure` on the session cookie: `auto` (when the request is HTTPS), `true` (behind a TLS-terminating proxy), `false` (plain HTTP only). |
| `auth.oidc.issuer`, `auth.oidc.clientID` | unset | The provider (an http(s) URL with a host) and the confidential client resourcerer presents there — a client secret is required. Required with `oidc`. |
| `auth.oidc.clientSecretSecretRef`, `auth.session.keySecretRef` | unset | Secrets with the client secret and the session key (standard base64 of 32 bytes: `openssl rand -base64 32`). Required with `oidc`; the chart renders references, never the material. |
| `auth.session.keyPreviousSecretRef` | unset | The key being retired, still accepted while a rotation is in flight. |
| `auth.oidc.allowedGroups` | `[]` | Which groups may use resourcerer. Empty gives read access to every user the provider authenticates. |
| `auth.oidc.roles.editorGroups` | `[]` | Which groups may apply, revert and change settings. Empty means nobody can. |
| `auth.oidc.roles.anyAuthenticatedIsViewer` | `true` | Whether an admitted user who is not an editor may read. |
| `auth.oidc.scopes`, `auth.oidc.usernameClaim`, `auth.oidc.groupsClaim` | `openid profile email`, `preferred_username`, `groups` | What to ask the provider for, and which claims name the user and the groups. |
| `auth.oidc.redirectURL` | derived | Absolute callback URL, for a proxy that rewrites the Host the server sees. Derived from the request otherwise, so every way of reaching it — port-forward, Ingress — needs its own callback registered at the provider. |
| `auth.oidc.sessionTTL` | `12h` | How long a session is valid without visiting the provider again: whole minutes (`5m` to `10080m`) or hours (`1h` to `168h`); nothing revokes it sooner. |
| `readOnly` | `false` | Refuse every change from the UI and API. |
| `applier.enabled` | `true` | Disable for a recommend-only install. |
| `controller.replicas`, `applier.replicas`, `server.replicas` | `1` | Leader election above 1 for controller and applier. |
| `server.ui` | `true` | Serve the embedded UI. |
| `server.service`, `server.ingress` | ClusterIP, disabled | The chart authenticates nothing at the Ingress; `auth.strategy` decides who may use the UI. |
| `metrics.scrapeAnnotations` | `true` | `prometheus.io/scrape` on the controller and applier metrics Services. |
| `metrics.serviceMonitor.enabled` | `false` | Prometheus Operator ServiceMonitor instead, for both. |

With `auth.strategy: oidc` the parameters are rendered into a ConfigMap named `<fullname>-auth` — release
`rz` gives `rz-resourcerer-auth` — under the key `oidc.yaml`, which the server reads at start and every 30s, and the
two Secrets are read as environment variables. Anyone who can write that ConfigMap can point `issuer` at a provider
they run and sign in as an editor, so treat writes to it as writes to resourcerer's own credentials. Change it through
the chart's values: Helm owns that ConfigMap, so a `kubectl edit` is reverted on the next sync.
`server.extraArgs` must not carry an `--auth-*` flag: a later duplicate on the command line would win over the ones
the chart derives from `auth.*`, and the render fails instead.

```yaml
auth:
  strategy: oidc
  oidc:
    issuer: https://sso.example.com/realms/example
    clientID: resourcerer
    allowedGroups: [engineering, sre]
    roles:
      editorGroups: [engineering]
    clientSecretSecretRef: { name: resourcerer-oidc, key: clientSecret }
  session:
    keySecretRef: { name: resourcerer-session, key: key }
```

Both Secrets are created outside the chart; the session key is standard base64 of exactly 32 bytes:

```bash
kubectl -n resourcerer create secret generic resourcerer-session --from-literal=key="$(openssl rand -base64 32)"
kubectl -n resourcerer create secret generic resourcerer-oidc --from-literal=clientSecret="$CLIENT_SECRET"
```

The CRDs are in `crds/` and installed by Helm on first install; upgrade them with
`kubectl apply --server-side -f charts/resourcerer/crds`. An installation upgraded without them runs ungated:
the applier says so in its log instead of refusing every apply.
