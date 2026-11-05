---
updatedAt: 2026-07-14T11:43:38.000Z
---

# KEDA compatibility

Atlas's Workload Autoscaler is fully compatible with [KEDA](https://keda.sh/) (Kubernetes Event-driven Autoscaling) when KEDA is configured to scale standard Kubernetes workloads using KEDA `ScaledObjects`. Our system recognizes and works seamlessly with it, the same as with native Kubernetes HPAs.

# How KEDA and Workload Autoscaler work together

KEDA creates HPA objects based on your `ScaledObject` configurations. Atlas's Workload Autoscaler integrates with these KEDA-generated HPAs and applies different optimization strategies depending on which metrics are configured in your `ScaledObject`.

## Compatibility

Atlas's Workload Autoscaler works with KEDA `ScaledObjects`. KEDA `ScaledJobs` are not supported.

# KEDA scaling optimization strategies

The level of optimization applied by Workload Autoscaler depends on which metric types are configured in your KEDA `ScaledObject`.

## KEDA with only CPU or memory metrics

When your KEDA `ScaledObject` uses only CPU or memory metrics:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: cpu-only-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: my-deployment
  triggers:
  - type: cpu
    metadata:
      type: Utilization
      value: "70"
```

**Workload Autoscaler will:**

* Run the full <Anchor label="simulation-based optimization" target="_blank" href="doc:workload-autoscaler-hpa-vpa">simulation-based optimization</Anchor>, evaluating CPU candidates across cost, stability, and replica behavior factors
* Calibrate the CPU recommendation away from the raw percentile when a nearby value produces better workload behavior

## KEDA with CPU/memory combined with custom metrics

When your KEDA `ScaledObject` combines CPU or memory metrics with custom metrics:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: combined-metrics-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: my-deployment
  triggers:
  - type: cpu
    metadata:
      type: Utilization
      value: "70"
  - type: prometheus
    metadata:
      serverAddress: http://<prometheus-host>:9090
      threshold: '100'
      query: sum(rate(http_requests_total{deployment="my-deployment"}[2m]))
```

**Workload Autoscaler will:**

* Run the simulation and apply vertical scaling optimization (CPU/memory resource requests)
* Limit how far the recommendation moves away from the originally detected percentile
* Respect KEDA's custom metric scaling decisions

Because custom metrics primarily drive scaling decisions for these workloads, CPU-based corrections may not reflect actual throughput. The simulation still runs, but calibration precision is reduced compared to CPU- or memory-only configurations.

## KEDA with only custom metrics

When your KEDA `ScaledObject` uses **only** custom metrics (no CPU or memory):

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: prometheus-scaledobject
  namespace: default
spec:
  scaleTargetRef:
    name: my-deployment
  triggers:
  - type: prometheus
    metadata:
      serverAddress: http://<prometheus-host>:9090
      threshold: '100'
      query: sum(rate(http_requests_total{deployment="my-deployment"}[2m]))
```

**Workload Autoscaler will:**

* Continue vertical scaling optimization (CPU/memory resource requests)
* **Not apply** the <Anchor label="intelligent vertical and horizontal scaling optimization algorithms" target="_blank" href="doc:workload-autoscaler-hpa-vpa">intelligent vertical and horizontal scaling optimization algorithms</Anchor>
* Respect KEDA's custom metric scaling decisions

<br />