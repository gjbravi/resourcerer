---
title: Try it on kind
description: Run resourcerer locally on a kind cluster with Prometheus, demo workloads and simulated Karpenter prices.
---

The repository has a complete local environment: a [kind](https://kind.sigs.k8s.io/) cluster with Prometheus,
kube-state-metrics, metrics-server, Argo Rollouts, demo workloads with uneven load, and a fake Karpenter exporter with
fixed prices. Recommendations show up within minutes, because the environment uses a 15-minute learning period and a
one-minute schedule.

## What you need

Docker, `kind`, `kubectl`, `helm`, Go and Node.js, and `make`.

## Run it

```bash
git clone https://github.com/gjbravi/resourcerer.git
cd resourcerer
make kind-up        # the cluster, Prometheus and demo workloads
make kind-deploy    # build the image, load it into kind, install the two charts
kubectl --context kind-resourcerer -n resourcerer port-forward svc/resourcerer 8080:80
```

Open [http://localhost:8080](http://localhost:8080). After a few minutes the demo workloads are `Ready`, with
savings, reliability signals and Karpenter-priced placement to look at.

Behind a TLS-intercepting proxy, pass its root CA: `EXTRA_CA_CERT=/path/to/ca.pem make kind-up`.

The values used are in `hack/resourcerer-values.yaml`: a good example of a configuration with Karpenter pricing,
node allocation and declared reliability signals.

## A decision model

To try [decision models](decision-models.md), start a decider server on your machine and point the cluster at it; the
commands are at the top of `hack/decision-model.yaml`:

```bash
kubectl --context kind-resourcerer apply -f hack/decision-model.yaml
```

kind's pods reach your machine at `host.docker.internal`, which Docker Desktop provides.

## Clean up

```bash
make kind-down
```

To work on resourcerer itself, see [CONTRIBUTING.md](https://github.com/gjbravi/resourcerer/blob/main/CONTRIBUTING.md).
