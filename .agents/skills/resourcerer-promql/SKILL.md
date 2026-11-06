---
name: resourcerer-promql
description: Write or change the Prometheus queries resourcerer runs (internal/prom, internal/cluster, internal/signals, server live checks). Use when adding a metric, a cost or placement query, a reliability check or a signal example, or when a query behaves differently on a real cluster than in kind.
---

# PromQL in resourcerer

Queries run per workload, on every regeneration, against Prometheus or Thanos in clusters with thousands of pods.
They must be scoped, deduplicated and bounded.

## Rules

1. **Scope to the workload.** Multiply per-pod metrics by `prom.Owner(ref)` (`owner(r)` inside the package):
   a vector of 1s keyed by `namespace, pod`, following ReplicaSets so pods of earlier rollouts count:

   ```promql
   sum by (container) (
     rate(container_cpu_usage_seconds_total{namespace="ns", container!="", container!="POD"}[40s])
     * on (namespace, pod) group_left () <owner>
   )
   ```

   Always filter by `namespace` in the inner selector too; it keeps the series set small before the join.
2. **Deduplicate what can overlap.** A pod recreated with the same name on another node leaves two
   `kube_pod_info` series; a node name reused by a new instance leaves two Karpenter series; Thanos downsampled data
   has no staleness markers. Keep one series per key before joining: `topk by (namespace, pod) (1, ...)`, or prefer
   the current series and fall back to the last value: `X or on (node_name) last_over_time(X[5m])`.
3. **Expect lag.** Karpenter metrics can reach Thanos a minute late; read them with `last_over_time(...[5m])`, and
   do not assume an instant query at "now" has data.
4. **Refuse partial results where it matters.** Thanos returns warnings instead of errors when a store is down. Use
   `Client.RangeComplete` for values that get stored or drive decisions (node allocation, placement); `Range` is fine
   for charts.
5. **Bound the cost.** Cap the points (node allocation reads its learning period in at most 336 points), derive the step
   from the window (`prom.Step`, `max(window/points, minStep)`), set a context timeout (10–30s), and cache live server
   queries (`s.cache.get(key, ttl, ...)`). Watch for wide metrics a join only needs a few series of:
   `karpenter_cloudprovider_instance_type_offering_price_estimate` has a series per offering of the region (tens of
   thousands), so range queries narrow it to the instance types the nodes had (`LabelValues` first). Run independent
   range queries in parallel.
6. **Missing is not zero.** An empty result usually means "does not apply" (a workload without Istio has no 5xx
   series). Return nil or skip, and make the UI say so; never store 0.
7. **Joins with optional sides.** `a + b` drops series missing on either side. To sum CPU and memory requests where
   one may be absent, union first and aggregate after: `sum by (namespace, pod) (cpu_expr or memory_expr)` (the
   `resource` label keeps them apart).
8. **No cluster label.** resourcerer's own queries assume the Prometheus it talks to sees one cluster. Do not add a
   `cluster` matcher in code.
9. **Say what it is for.** The client counts every query in `resourcerer_prometheus_requests_total` and
   `resourcerer_prometheus_samples_total` by `purpose`, read from the context: tag a new controller path with
   `prom.For(ctx, "<purpose>")` (`usage`, `signals`, `allocation`, `pricing` exist); server handlers get their route
   as the purpose already. Untagged queries count as `other`, which hides where load comes from.

## Test it

- Unit: assert on the generated string for the parts that matter (scoping, dedup, labels), as in
  `internal/prom/prom_test.go`; test the arithmetic on query results in pure functions with synthetic series
  (`internal/cluster/placement_test.go`).
- Live: port-forward the kind Prometheus (`make prom-forward`) and run the query from a scratch `main` under `cmd/`
  that you delete afterwards, or with `curl localhost:9090/api/v1/query --data-urlencode query=...`.
- Real data: on a real cluster, run read-only queries, start with an instant query on one cluster, check the run time
  before widening, and avoid repeating heavy range queries.

## Templates for declared signals

Signals use Go templates with `{{ .Namespace }}`, `{{ .Name }}`, `{{ .Kind }}`, `{{ .Window }}` and
`{{ .OwnedPods }}`; the contract is "the value over `{{ .Window }}`, one series or one per `container`". See
`docs/guides/reliability-signals.md` and keep its examples valid when changing the template variables.
