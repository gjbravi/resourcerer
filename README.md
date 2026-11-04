<p align="center">
  <img src="docs/logo.svg" alt="resourcerer logo" width="112" height="112">
</p>

<h1 align="center">resourcerer</h1>

<p align="center">
  <a href="https://github.com/gjbravi/resourcerer/actions/workflows/ci.yaml"><img src="https://github.com/gjbravi/resourcerer/actions/workflows/ci.yaml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="License"></a>
  <img src="https://img.shields.io/badge/status-beta-7c5cff.svg" alt="Status: beta">
</p>

resourcerer recommends CPU and memory requests and limits for your Kubernetes workloads from their usage history
in Prometheus, shows the projected impact in a UI, and applies recommendations on request or automatically.

It runs entirely in your cluster. There is no machine learning: every recommendation is a documented percentile
of usage plus headroom, so you can see exactly why a value was chosen.

> Status: beta. APIs, annotations and the CRD may change.

## Features

- A recommendation per Deployment, StatefulSet, DaemonSet and [Argo Rollout](https://argoproj.github.io/rollouts/),
  regenerated on a schedule spread across workloads.
- Optimization goals per resource (Savings, Balanced, Reliability), limit policies, minimums, skipped sidecars.
- Reliability response: after an OOM kill, memory is raised above the limit that killed the container.
- HPA-aware: when the HPA scales on CPU utilization, its target is adjusted with the requests to keep the scale point.
- Settings cascade: Helm defaults → namespace annotations → workload annotations, all GitOps-friendly.
- UI: cluster overview, workload list with bulk apply, per-workload impact, usage percentile charts with current and
  recommended requests over time, Kustomize/Helm/kubectl patches, and effective settings with their source.
- Apply Now, Revert and auto-deploy with thresholds, through server-side apply with field manager `resourcerer`.
- Apply policies: apply windows in a time zone, freezes, and how many rollouts run at once, as `ClusterApplyPolicy`
  and `ApplyPolicy` objects a team writes itself.
- No database: state lives in a `Recommendation` CRD, history in Prometheus.

## Install

Requires Kubernetes 1.33+ and a Prometheus-compatible API (Prometheus, Thanos, Mimir, VictoriaMetrics) with cAdvisor
and kube-state-metrics data, retaining at least the learning period (7 days by default).

```bash
helm install resourcerer-crds oci://ghcr.io/gjbravi/charts/resourcerer-crds --version <version> \
  -n resourcerer --create-namespace
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> -n resourcerer \
  --set prometheus.url=http://prometheus-server.monitoring.svc
kubectl -n resourcerer port-forward svc/resourcerer 8080:80
```

Versions are listed on the [releases page](https://github.com/gjbravi/resourcerer/releases). Images and charts are
signed with cosign; see [SECURITY.md](SECURITY.md#verifying-releases).

Authentication is `anonymous` by default, like Kiali's anonymous strategy: anyone who reaches the UI can use it, so
keep the Service internal (port-forward) or put an authenticating proxy in front. `auth.strategy: oidc` signs callers
in at an OpenID Connect provider instead: groups from the token decide who may read and who may apply, every apply
records who asked, and the parameters live in their own ConfigMap while the client secret and the session key stay in
Secrets. The [chart README](charts/resourcerer/README.md) has the values.

## Documentation

The documentation is at **[gjbravi.github.io/resourcerer](https://gjbravi.github.io/resourcerer/)**, built from
[`docs/`](docs/):

- [Quickstart](https://gjbravi.github.io/resourcerer/getting-started/quickstart/) and
  [installation](https://gjbravi.github.io/resourcerer/getting-started/installation/)
- [How it works](https://gjbravi.github.io/resourcerer/concepts/how-it-works/) and
  [how recommendations are computed](https://gjbravi.github.io/resourcerer/concepts/recommendations/)
- [Configuration](https://gjbravi.github.io/resourcerer/configuration/overview/) and the
  [settings reference](https://gjbravi.github.io/resourcerer/reference/settings/)
- Guides: [auto-deploy](https://gjbravi.github.io/resourcerer/guides/auto-deploy/),
  [apply policies](https://gjbravi.github.io/resourcerer/guides/apply-policies/),
  [reliability signals](https://gjbravi.github.io/resourcerer/guides/reliability-signals/),
  [GitOps](https://gjbravi.github.io/resourcerer/guides/gitops/),
  [Argo Rollouts](https://gjbravi.github.io/resourcerer/guides/argo-rollouts/),
  [Karpenter costs](https://gjbravi.github.io/resourcerer/guides/karpenter-costs/)
- [Architecture](https://gjbravi.github.io/resourcerer/architecture/), with diagrams of each component and flow

## Development

```bash
make kind-up        # kind cluster with Prometheus, kube-state-metrics, metrics-server, Argo Rollouts and demo workloads
make kind-deploy    # build the image, load it, install the chart with hack/resourcerer-values.yaml
make e2e            # end-to-end checks
kubectl --context kind-resourcerer -n resourcerer port-forward svc/resourcerer 8080:80
```

The kind values use a 15-minute learning period and a 2-minute schedule, so recommendations show up within minutes.
Behind a TLS-intercepting proxy, pass its root CA: `EXTRA_CA_CERT=/path/ca.pem make kind-up`.

To run the components locally against the kind cluster: `make prom-forward` in one terminal, `make run` in another,
and `make web-dev` for the UI with hot reload.

The documentation site: `make docs-dev` serves it with live reload on http://localhost:8000/resourcerer/.

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Apache-2.0](LICENSE)
