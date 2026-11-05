---
updatedAt: 2026-07-14T11:56:21.000Z
---

# Vertical & horizontal workload autoscaling

In Kubernetes, running workloads with both Vertical Pod Autoscaling (VPA) and Horizontal Pod Autoscaling (HPA) turned on at the same time can be challenging. VPA adjusts the resources allocated to individual pods while HPA changes the number of pod replicas – when operating independently, these mechanisms can work against each other. For example, VPA might increase resources per pod while HPA tries to scale down the number of pods, or vice versa, leading to sub-optimal resource allocation and potential issues.

Atlas has developed a unique approach to ensure these scaling mechanisms work together harmoniously for CPU scaling. Rather than letting them potentially conflict, our Workload Autoscaler automatically adjusts its behavior in the background to optimize resource allocation of workloads that are being scaled horizontally and vertically at the same time.

# How it works

Workload Autoscaler analyzes the historical metrics of the workload and searches for the CPU recommendation that works best at the moment, delivering a stable and cost-effective workload. Memory requests are coordinated separately — see [Memory-based horizontal scaling](#memory-based-horizontal-scaling).

For every CPU recommendation candidate, Atlas simulates how the workload would behave under the previously observed load, including when HPA would create and stop replicas. The simulation takes into account several factors:

* **Resource cost** — how much the workload would cost on average
* **CPU delays** — insufficient CPU at any given point means increased latency and error rate
* **Saturation** — a workload approaching its maximum replica count faces higher risk of running out of resources at new peaks
* **Stable scalability** — many short-term replica spikes can add significant cost and instability, sometimes making a slightly higher base CPU recommendation the better choice
* **Snowballing replicas** — for heavy-startup workloads without precise probes, HPA may detect artificial startup load and keep creating replicas up to the maximum, despite no real load to handle
* **Node overhead** — chaotic scaling behavior, especially with anti-affinity rules, can cause nodes to pack less efficiently, adding DaemonSet costs and unused capacity

Starting from the CPU percentile configured in your workload policy, the system calibrates the recommendation by evaluating nearby CPU candidates to find a more cost-effective or more stable result.

## Reading the results

When both vertical and horizontal scaling are active, you can inspect the simulation output directly from the event log. Open a **Recommendation generated** event and switch to the [JSON view](event-log.md#json-view). Look for the step with `"message": "Calibrating for HPA"`.

That step contains several `stats*` objects, each representing a different CPU candidate the simulation evaluated:

```json
{
  "message": "Calibrating for HPA",
  "reasoning": "raised above break-even/lowest-cost: 8% better score over original",
  "statsCalibrated": {
    "cpu": 3135,
    "delay": { "avg": 0, "max": 0 },
    "replicas": { "avg": 10.83, "max": 25, "min": 10 },
    "resources": { "cpu": 33949, "memory": 54598910197 },
    "scaling": { "cycles": 121, "cyclesSignificant": 90, "down": 162, "up": 156 },
    "score": 44.106
  },
  "statsOriginal": { ... },
  "statsLowestCost": { ... }
}
```

`statsCalibrated` is the chosen recommendation. Comparing its `score` and `replicas` against `statsOriginal` (the raw percentile) and `statsLowestCost` shows why the system moved away from the configured percentile value.

## Memory-based horizontal scaling

When vertical scaling changes the memory requests for a workload, it can affect how a memory-based Horizontal Pod Autoscaler (HPA) interprets utilization. In some cases, this can make HPA trigger earlier than intended, causing unnecessary scaling events.

The Workload Autoscaler prevents this by automatically adjusting memory requests to keep typical usage below the HPA's scaling threshold, even after vertical optimizations. This coordination works with your existing memory overhead settings, applying additional adjustments on top of any configured overhead to prevent frequent scale-up events. These corrections happen automatically in the background, without requiring any changes to your HPA or vertical scaling configurations.

# Limitations

When using vertical and horizontal autoscaling together in Atlas's Workload Autoscaler, the aforementioned algorithmic optimizations happen automatically in the background with no manual configuration needed or, in fact, available. There is no way to toggle this behavior on or off. Therefore, it is important to note that there are certain limitations that need to be understood.

## Custom metrics

When HPAs or KEDA `ScaledObjects` combine CPU metrics with custom metrics, Workload Autoscaler still runs the simulation but limits how far it will move the recommendation away from the originally detected percentile. This is because for workloads primarily driven by custom metrics, CPU-based corrections may not reflect actual throughput.

For more information, see <Anchor label="KEDA compatibility" target="_blank" href="doc:woop-hpa-third-party-support">KEDA compatibility</Anchor>.

## CPU or memory metrics requirement

Workload Autoscaler's vertical and horizontal autoscaling coordination works best when horizontal scaling includes **CPU or memory metrics**. When HPAs use only custom metrics (without CPU or memory), the automatic optimization still occurs to a limited degree — the system estimates the best CPU recommendation to keep the workload stable and models replica scaling behavior, but cannot fully account for throughput-driven scaling decisions. Note that CPU overhead is still managed automatically in this case and will override any user-configured CPU overhead setting — see [Resource overhead](#vpa-setting-availability).

> 📘 Custom metrics and third-party autoscalers
>
> For detailed information about how Atlas works with custom metrics and third-party autoscaling solutions like KEDA, see [KEDA compatibility](woop-hpa-third-party-support.md).

## VPA setting availability

Several vertical workload scaling settings available for the Workload Autoscaler become irrelevant as the system handles CPU allocation dynamically and balances between the two optimization strategies. Even if their values are set in the vertical scaling policy, they will be ignored.

**Recommendation percentile**:
When both vertical and horizontal scaling are enabled for a workload, instead of respecting the CPU values manually set by the user, the system will rely on its aforementioned algorithms to issue CPU recommendations. For Memory, the figure below can still be configured.

<Image border={false} src="https://files.readme.io/398abef72d334f831b9c76fa75fabdb546c3a9a284c62c8a33b6a278fda73eb9-image.png" />

**Resource overhead**:
When both vertical and horizontal scaling are enabled for a workload, CPU overhead is handled automatically through native overhead calculations that completely replace any user-configured CPU overhead settings. This automatic overhead ensures workloads maintain appropriate headroom below HPA scaling thresholds without requiring user intervention. Memory overhead can still be configured manually as that resource coordination works with existing user settings rather than replacing them.

<Image border={false} src="https://files.readme.io/9f36557fb34dfade60a70176079748ad7f4faac57c1e9a51b35b2b2180b04c82-image.png" />