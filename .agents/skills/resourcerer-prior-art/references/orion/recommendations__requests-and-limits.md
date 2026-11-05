# Requests and limits

> Recommendations for CPU and memory resource settings


Every Orion recommendation includes CPU and memory requests that the machine learning has determined to be optimal for each container in a workload.

Orion generates CPU and memory recommendations using our patent pending machine learning. Our machine learning examines each container's resource usage metrics, and monitors usage patterns and scaling behavior to come up with the optimal CPU and memory resource settings.

## Requests

Request recommendations for CPU and memory are based on observed resource usage, constrained by optimization settings. After the machine learning generates recommendation candidates, an effective recommendation is created by:

1. Selecting a recommendation candidate based on each workload's configured [Optimization Goal][goal].
2. Applying all of the configured constraints, such as [min and max request bounds][req-bounds].

## Limits

Limit recommendations for CPU and memory are optional, and Orion offers a high degree of control over whether and how to set limits. If you would like to include limits in a recommendation, the recommended limits will usually be created by:

1. Multiplying the recommended request value by a configurable [limit-to-request-ratio][lrr], or LRR.
2. Applying all of the configured constraints, such as [min and max limit bounds][lim-bounds].

Whether and how to recommend limits is controlled by [Optimization Policy][opt-policy].

## HPA and target utilization

Orion is compatible with Horizontal Pod Autoscalers (HPA). For HPA-managed workloads, recommendations may include updates to target utilization values.

Target utilization recommendations are designed to preserve observed scaling behavior while vertically resizing pods.

The range within which Orion is allowed to adjust target utilization values can be constrained by [min and max bounds][hpa-bounds].

Orion can also be configured to:
- [Preserve the horizontal scale point][preserve-scale-point]
- [Ignore horizontally-scaled resources][ignore-hpa-resources]

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span>Settings<span class="related-topics-divider">/</span><a href="/docs/configure/settings/cpu-and-memory/">CPU and memory</a>
                </h5>
                <p>Optimization policy, limit-to-request ratio, and bounds for CPU and memory</p>
            </div>
        </div>


[goal]: /docs/configure/settings/goal/
[req-bounds]: /docs/configure/settings/cpu-and-memory/#requests-min-max
[lrr]: /docs/configure/settings/cpu-and-memory/#lrr
[lim-bounds]: /docs/configure/settings/cpu-and-memory/#limits-min-max
[opt-policy]: /docs/configure/settings/cpu-and-memory/#opt-policy
[hpa-bounds]: /docs/configure/settings/hpa/#target-min-max
[preserve-scale-point]: /docs/configure/settings/hpa/#preserve-scale-point
[ignore-hpa-resources]: /docs/configure/settings/hpa/#ignore-hpa-resources
