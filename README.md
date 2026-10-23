# resourcerer

[![CI](https://github.com/gjbravi/resourcerer/actions/workflows/ci.yaml/badge.svg)](https://github.com/gjbravi/resourcerer/actions/workflows/ci.yaml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

resourcerer creates and manages [Vertical Pod Autoscalers](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler)
for your workloads and gives you a UI to review their recommendations and choose, per workload, whether VPA only
recommends or also applies.

It is in the spirit of [Goldilocks](https://github.com/FairwindsOps/goldilocks): one VPA per workload, no
configuration needed to get started.

## Features

- One VPA per Deployment, StatefulSet and DaemonSet, owned by the workload and deleted with it.
- Two selection modes: **all** workloads (opt out with a label) or only **labeled** workloads and namespaces.
- Per-workload mode: **Recommend** (VPA `Off`), **Auto** (VPA `InPlaceOrRecreate`) or **Disabled**.
- Web UI: list with filters (under/over-provisioned, learning), recommendation detail with VPA bounds,
  live pod resources and events, light and dark theme.
- Everything is driven by labels and annotations, so it works with GitOps; the UI writes the same keys.
- Never touches VPAs it did not create.

## Requirements

- Kubernetes 1.33+ (in-place pod resize)
- [Vertical Pod Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) 1.4+:
  the recommender for recommendations, plus the updater and admission controller for Auto mode.

## Install

```bash
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer \
  --namespace resourcerer --create-namespace
kubectl -n resourcerer port-forward svc/resourcerer 8080:80
```

Or from a checkout: `helm install resourcerer ./charts/resourcerer -n resourcerer --create-namespace`.

See [charts/resourcerer/values.yaml](charts/resourcerer/values.yaml) for all options.

## Configuration

### Which workloads get a VPA

| `selection` | Behavior |
|---|---|
| `all` (default) | Every workload outside `excludedNamespaces` gets a VPA. Opt out with `resourcerer.io/enabled=false`. |
| `label` | Only workloads with `resourcerer.io/enabled=true` on the workload or its namespace. |

The label is read from the workload first, then from its namespace, so a workload can override its namespace in
either direction:

```bash
# selection=label: enable a whole namespace, except one workload
kubectl label namespace shop resourcerer.io/enabled=true
kubectl -n shop label deployment legacy resourcerer.io/enabled=false

# selection=all: disable a namespace
kubectl label namespace batch resourcerer.io/enabled=false
```

### VPA mode

| Mode | VPA `updateMode` | Effect |
|---|---|---|
| `Off` (default) | `Off` | VPA only computes recommendations. |
| `Auto` | `InPlaceOrRecreate` | VPA updates pod requests, in place when possible, otherwise by recreating pods. |

Set it with the `resourcerer.io/mode` annotation on the workload or, as a default, on the namespace. The chart's
`defaultMode` applies when neither is set.

```bash
kubectl -n shop annotate deployment api resourcerer.io/mode=Auto
```

### Managed VPAs

resourcerer names each VPA `<kind>-<name>` (for example `deployment-api`), labels it `resourcerer.io/managed=true`
and sets the workload as its owner. It reconciles only `targetRef` and `updatePolicy.updateMode`, so other fields
you set on the VPA by hand (for example `resourcePolicy`) are kept.

If a VPA with the same name already exists without the managed label, resourcerer leaves it alone and emits a
`VPAConflict` event on the workload.

### Running alongside Goldilocks

Both tools create one VPA per workload. Two VPAs on the same pods conflict, so do not let them manage the same
namespaces: use `selection: label` here, or exclude namespaces in one of the two.

## Development

Requirements: Go 1.26, Node 24, Docker, kind, kubectl, Helm.

```bash
make help          # all targets
make test          # Go and UI unit tests
make lint          # go vet, golangci-lint, eslint, tsc, helm lint
make kind-up       # kind cluster with metrics-server, VPA and demo workloads
make kind-deploy   # build the image, load it into kind and install the chart
make e2e           # end-to-end checks against the kind cluster
make run           # run locally against the current kube context, UI on :8080
make web-dev       # Vite dev server with /api proxied to :8080
```

Behind a TLS-intercepting proxy, pass its root CA to the kind nodes: `EXTRA_CA_CERT=/path/ca.pem make kind-up`.

### Layout

```
cmd/resourcerer/     entrypoint and flags
internal/config/     selection and mode rules (labels, annotations, flags)
internal/controller/ reconcilers, one per workload kind
internal/vpa/        build, create, update and delete managed VPAs
internal/server/     HTTP API and embedded UI
internal/workload/   uniform view over Deployment, StatefulSet and DaemonSet
web/                 React + MUI UI, embedded with go:embed
charts/resourcerer/  Helm chart
hack/                kind setup and demo workloads
test/e2e/            end-to-end checks
```

## Roadmap

- Apply a recommendation to the workload spec on demand, with revert.
- Optimization goals (savings, balanced, reliability) via multiple VPA recommenders.
- Usage charts from Prometheus.
- Argo Rollouts and CronJob support.

## License

[Apache 2.0](LICENSE)
