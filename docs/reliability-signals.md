# Reliability signals

Reliability signals tell when a workload's requests or limits may be too tight, or when it is unhealthy for other
reasons. They show in three places:

- the **workload list**, as a Reliability column (critical, warning, fixed by apply, healthy) you can sort and filter by;
- the **Reliability tab** of a workload, checked live over 15m to the learning period, with a chart and the query;
- the **verdict** on the Impact tab: a signal the recommendation does not fix turns "Safe to apply" into "Review
  before applying".

There are two kinds: built-in signals, which resourcerer computes itself, and signals you declare as PromQL templates.

## Built-in signals

| id | What it measures | Default |
|---|---|---|
| `throttling` | Share of CFS periods throttled by the CPU limit. Critical from 25%, where the recommendation raises the limit. | warn at 0.10 |
| `memoryHeadroom` | Peak working set as a share of the memory limit. | warn at 0.90 |
| `restarts` | Container restarts other than OOM kills. | on |
| `crashLoop` | Containers in CrashLoopBackOff. | on |
| `hpaAtMax` | The HPA runs at its maximum replicas. | on |
| `bestEffort` | Pods without requests or limits. | on |

OOM kills are always on: the memory recommendation depends on them.

```yaml
reliability:
  builtin:
    throttling: { enabled: true, warn: 0.20 }
    bestEffort: { enabled: false }
```

## Declared signals

A declared signal is a PromQL template that resourcerer fills in for each workload. Up to 20 signals.

```yaml
reliability:
  signals:
    - id: http-5xx                 # lowercase letters, digits and dashes
      name: HTTP 5xx rate
      description: Share of requests answered with a 5xx.
      unit: ratio                  # ratio | count | seconds | bytes
      above: { warn: 0.01, critical: 0.05 }
      runbook: https://example.com/runbooks/5xx
      query: |
        sum(rate(istio_requests_total{response_code=~"5..",
          destination_workload_namespace="{{ .Namespace }}", destination_workload="{{ .Name }}"}[{{ .Window }}]))
        /
        sum(rate(istio_requests_total{
          destination_workload_namespace="{{ .Namespace }}", destination_workload="{{ .Name }}"}[{{ .Window }}]))
```

### The query contract

The query returns **the signal's value over `{{ .Window }}`**:

- one series, or one series per `container` label to break the signal down by container (the worst one is shown);
- no data means the signal does not apply to the workload, which then does not show it. Declare Istio, JVM or queue
  signals freely: they only show where their metrics exist.

resourcerer evaluates the same query in two ways:

- **at each regeneration**, as an instant query with `Window` set to the learning period. The result is stored in the
  Recommendation's `status.reliability.signals` and drives the list and the verdict;
- **in the Reliability tab**, as an instant query over the window the user picks, and as a range query with `Window`
  set to the chart step, for the chart.

Use `rate()` or `increase()` for counters and `max_over_time()` or `avg_over_time()` for gauges, so the value means the
same over 15 minutes and over 7 days.

### Template variables

| Variable | Value |
|---|---|
| `{{ .Namespace }}` | The workload's namespace. |
| `{{ .Name }}` | The workload's name. |
| `{{ .Kind }}` | `Deployment`, `StatefulSet`, `DaemonSet` or `Rollout`. |
| `{{ .Window }}` | A PromQL duration: the learning period, the chosen window or the chart step. |
| `{{ .OwnedPods }}` | A vector of 1s per pod of the workload, keyed by `namespace` and `pod`, to keep a per-pod metric to the workload's pods. |

`{{ .OwnedPods }}` follows ReplicaSets, so pods from earlier rollouts count too:

```promql
sum(rate(container_cpu_usage_seconds_total{namespace="{{ .Namespace }}", container!=""}[{{ .Window }}])
  * on (namespace, pod) group_left () {{ .OwnedPods }})
```

### Levels

Set `above` for values that should stay low, such as error rates or latency, or `below` for values that should stay
high, such as availability. Each takes `warn`, `critical` or both.

### Overrides per namespace or workload

Annotations on a namespace or a workload change the signals for it; the workload's win over the namespace's:

```bash
# Turn signals off, built-in or declared.
kubectl annotate namespace batch resourcerer.io/reliability.disabled=http-5xx,hpaAtMax
# Tune a declared signal's levels.
kubectl -n search annotate deploy autocomplete resourcerer.io/reliability.http-5xx.critical=0.10
```

### Validation

The chart schema checks the structure. On start, every template is parsed, and the Settings page test-runs each query
against Prometheus every few minutes: a signal whose template fails is left out, one whose query fails shows the error.
A failing signal never blocks a recommendation.

### Cost

One instant query per signal and workload at each regeneration, spread over the schedule like the recommendations,
with a 10s timeout each.

## Examples

### p99 latency from Istio

```yaml
- id: latency-p99
  name: p99 latency
  unit: seconds
  above: { warn: 0.5, critical: 1 }
  query: |
    histogram_quantile(0.99, sum by (le) (rate(istio_request_duration_milliseconds_bucket{
      destination_workload_namespace="{{ .Namespace }}", destination_workload="{{ .Name }}"}[{{ .Window }}]))) / 1000
```

### Memory above request, per container

```yaml
- id: memory-over-request
  name: Memory above request
  unit: ratio
  above: { warn: 1.0, critical: 1.5 }
  query: |
    max by (container) (max_over_time(container_memory_working_set_bytes{namespace="{{ .Namespace }}", container!="", container!="POD"}[{{ .Window }}])
      * on (namespace, pod) group_left () {{ .OwnedPods }})
    /
    max by (container) (kube_pod_container_resource_requests{namespace="{{ .Namespace }}", resource="memory"}
      * on (namespace, pod) group_left () {{ .OwnedPods }})
```

### JVM GC pause share

```yaml
- id: gc-pause
  name: Time in GC
  unit: ratio
  above: { warn: 0.05, critical: 0.15 }
  query: |
    sum(rate(jvm_gc_pause_seconds_sum{namespace="{{ .Namespace }}"}[{{ .Window }}]) * on (namespace, pod) group_left () {{ .OwnedPods }})
    / count({{ .OwnedPods }})
```

### Consumer lag

```yaml
- id: consumer-lag
  name: Consumer lag
  unit: count
  above: { warn: 1000, critical: 10000 }
  query: |
    max(max_over_time(kafka_consumergroup_lag{consumergroup="{{ .Name }}"}[{{ .Window }}]))
```
