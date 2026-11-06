---
name: resourcerer-dev-loop
description: Run resourcerer's kind environment, deploy a local change into it, run the e2e checks and look at the UI. Use when you need to test a controller, server, chart, query or UI change end to end, or when kind misbehaves (pods Pending, port-forward lost, metrics missing).
---

# resourcerer dev loop

## First time

```bash
make kind-up        # add EXTRA_CA_CERT=/path/to/ca.pem behind a TLS-intercepting proxy
make kind-deploy    # builds the image, loads it into kind, installs the chart with hack/resourcerer-values.yaml
make e2e
```

`kind-up` creates `kind-resourcerer` with Prometheus (10s scrape), Grafana (anonymous admin, dashboards from
labelled ConfigMaps), metrics-server, VPA, Argo Rollouts, the demo workloads in `hack/demo.yaml` and
`hack/fake-karpenter.sh`. The kind values enable the chart's Grafana dashboard: after a dashboard change
(`hack/grafana-dashboard.py`, then `make dashboard`), `make kind-deploy` and
`kubectl --context kind-resourcerer -n monitoring port-forward svc/grafana 3000:80`, then
http://localhost:3000/d/resourcerer.

## After each change

```bash
make lint test
make kind-deploy && make e2e
kubectl --context kind-resourcerer -n resourcerer port-forward svc/resourcerer 8080:80
```

- `kind-deploy` rolls the pods, which breaks a running port-forward: restart it.
- The kind values use short periods so things move quickly: `learning-period: PT15M`, `schedule: "@every 1m"`,
  `rateWindow: 40s`, `minStep: 5s`. A recommendation leaves `Pending` once it has 10 samples at the query step,
  so within about two minutes of a workload starting, counting the 1-minute schedule.
- cAdvisor stamps its samples every 10–17s whatever the scrape interval, so a shorter `scrape_interval` adds no CPU
  or memory data, and a `rateWindow` under ~35s leaves gaps in CPU usage.
- `@every 1m` is the shortest valid schedule; shorter values are rejected.

## Fake Karpenter

`hack/fake-karpenter.sh` labels the first worker `on-demand` (m7g.xlarge) and every other node `spot`
(t3.xlarge), and serves `karpenter_nodes_allocatable` and `karpenter_cloudprovider_instance_type_offering_price_estimate`
from a busybox httpd. The file is `metrics.txt` so busybox serves `text/plain`, which Prometheus 3 requires. Node
names are read from the cluster, so it works with one or two workers. `pricing/on-demand-app`, `pricing/spot-app` and
`pricing/mixed-app` are pinned to give predictable costs.

## Decision models

`make e2e` applies `hack/fake-decider.yaml`: a busybox httpd whose CGI script answers every `POST /v1/systemone` with
a fixed verdict (`review`) and keeps the last request in `/tmp/last.json`, with a `DecisionModel` pointing at it. The
checks read what was sent with `kubectl -n resourcerer exec deploy/fake-decider -- cat /tmp/last.json`, then delete it
and restore any `DecisionModel` that was there.

To look at a real model, run a decider on your machine and apply `hack/decision-model.yaml`, whose comment has the
commands; pods reach your machine at `host.docker.internal`. On Apple Silicon decider-4b needs `DECIDER_IMAGES=0`
(the image part wants `torchvision`) and answers in one to three seconds. Bind it to `127.0.0.1`: it has no
authentication.

## Adding an e2e check

Edit `test/e2e/run.sh`. Use `wait_for "<what>" <seconds> <command>` for things the controller reaches eventually
and `pass`/`fail` for immediate assertions; query the API with `curl -fs "$API/..." | jq -e` and resources with
`kubectl --context $CTX ... -o json | jq -e`. Add new namespaces to the `fail()` dump when they help debugging.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| New resourcerer pods `Pending`, `Insufficient memory` | A local cluster with one worker fills up; delete leftover test namespaces you created. CI has two workers. |
| `make kind-install` times out | Same as above, or the image did not load: check `kubectl -n resourcerer describe pod`. |
| UI shows old code | The port-forward points at a terminated pod: restart it; hard-reload the page. |
| Karpenter series duplicated for a few minutes | Expected after relabeling nodes; queries keep one series per node. |
| A query works in kind but not on a real cluster | Real data has duplicates, gaps and Thanos lag; see the `resourcerer-promql` skill. |

## Looking at the UI

Check light and dark themes. To see loading states, throttle the network in the browser's DevTools; kind answers too
fast for them to show otherwise.

## Against a real cluster, read-only

Run the server locally with `--read-only` against a copy of the cluster's kubeconfig and a port-forward to its
Prometheus, and delete the copied kubeconfig afterwards. Never apply recommendations to a shared cluster from a dev
build.
