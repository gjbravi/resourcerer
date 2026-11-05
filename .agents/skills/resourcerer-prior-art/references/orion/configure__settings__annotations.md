# Annotations

> All available annotations for Orion configuration


## [Goal]

`live.orion.example/cpu.optimization-goal: "Balanced"`  
`live.orion.example/memory.optimization-goal: "Balanced"`

## [Schedule]

`live.orion.example/schedule: "@daily"`  
`live.orion.example/schedule.timezone: "UTC"`  
`live.orion.example/validity-period: ""`

## [Apply settings]

`live.orion.example/auto-deploy: "Disabled"`  
`live.orion.example/learning-period: "P7D"`  
`live.orion.example/auto-deploy.thresholds.cpu.percent: ""`  
`live.orion.example/auto-deploy.thresholds.memory.percent: ""`  
`live.orion.example/auto-deploy.thresholds.cpu.unit: ""`  
`live.orion.example/auto-deploy.thresholds.memory.unit: ""`  
`live.orion.example/apply.method: "PatchImmediateRollout"`  
`live.orion.example/apply.max-percent-increase: ""`  
`live.orion.example/apply.max-percent-decrease: ""`  
`live.orion.example/pod-template.metadata.patch-path: ""`  
`live.orion.example/pod-template.metadata.patch-format: ""`

## [Reliability]

`live.orion.example/reliability.oom.memory-bump-up.enabled: "False"`  
`live.orion.example/reliability.oom.memory-bump-up.period: "P4D"`  
`live.orion.example/reliability.oom.memory-bump-up.percent: "0"`  
`live.orion.example/reliability.oom.memory-bump-up.min: "0Mi"`  
`live.orion.example/reliability.oom.memory-bump-up.max: "2Gi"`  
`live.orion.example/reliability.oom.memory-bump-up.apply-immediately: "Never"`  

`live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled: "False"`  
`live.orion.example/reliability.cpu-pressure.cpu-bump-up.period: "P4D"`  
`live.orion.example/reliability.cpu-pressure.cpu-bump-up.percent: "0"`  
`live.orion.example/reliability.cpu-pressure.cpu-bump-up.min: "0m"`  
`live.orion.example/reliability.cpu-pressure.cpu-bump-up.max: "500m"`  
`live.orion.example/reliability.cpu-pressure.cpu-bump-up.apply-immediately: "Never"` 

## [CPU and memory]

`live.orion.example/containers.cpu.optimization-policy: "RequestsRaiseLimitsIfNeeded"`  
`live.orion.example/containers.memory.optimization-policy: "RequestsRaiseLimitsIfNeeded"`  
`live.orion.example/containers.cpu.limits.limit-request-ratio: "2.0"`  
`live.orion.example/containers.memory.limits.limit-request-ratio: "2.0"`  

`live.orion.example/containers.cpu.requests.min: "10m"`  
`live.orion.example/containers.cpu.requests.max: ""`  
`live.orion.example/containers.memory.requests.min: "32Mi"`  
`live.orion.example/containers.memory.requests.max: ""`  

`live.orion.example/containers.cpu.limits.min: ""`  
`live.orion.example/containers.cpu.limits.max: ""`  
`live.orion.example/containers.memory.limits.min: ""`  
`live.orion.example/containers.memory.limits.max: ""`  

`live.orion.example/containers.cpu.requests.patch-path: ""`  
`live.orion.example/containers.cpu.limits.patch-path: ""`  
`live.orion.example/containers.memory.requests.patch-path: ""`  
`live.orion.example/containers.memory.limits.patch-path: ""`

## [HPA]

`live.orion.example/hpa.cpu.target-utilization.min: "50"`  
`live.orion.example/hpa.cpu.target-utilization.max: "90"`  
`live.orion.example/hpa.memory.target-utilization.min: "50"`  
`live.orion.example/hpa.memory.target-utilization.max: "90"`  
`live.orion.example/hpa.preserve-horizontal-scale-point: "False"`  
`live.orion.example/hpa.ignore-horizontally-scaled-resources: "False"`  
`live.orion.example/hpa.ignore: "False"`

## [Runtime]

`live.orion.example/containers.java.max-heap.patch-path: -`  
`live.orion.example/containers.java.max-heap.patch-format: ""`  
`live.orion.example/containers.java.max-heap.min: "512Mi"`  
`live.orion.example/containers.java.max-heap.max: ""` 

## [Pod scheduling]

`live.orion.example/pod-scheduling.optimization-policy: DoNotOptimize`  
`live.orion.example/pod-scheduling.instance-categories.node-affinity.type: preferred`  
`live.orion.example/pod-scheduling.instance-categories.node-affinity.weight: "70"`  
`live.orion.example/pod-scheduling.instance-categories.node-affinity.patch-path: ""`  
`live.orion.example/pod-scheduling.instance-categories.node-affinity.patch-format: ""`

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/annotate-namespaces/">Configure namespaces</a>
                </h5>
                <p>Default optimization settings for every workload in a namespace, set with annotations</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/cluster-defaults/">Configure clusters</a>
                </h5>
                <p>Configure default optimization settings for the cluster</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span><a href="/docs/configure/annotate-workloads/">Configure workloads</a>
                </h5>
                <p>Optimization settings for a single workload, set with annotations</p>
            </div>
        </div>


[Goal]: /docs/configure/settings/goal/  
[Schedule]: /docs/configure/settings/schedule/  
[Apply settings]: /docs/configure/settings/apply-settings/  
[Reliability]: /docs/configure/settings/reliability/  
[CPU and memory]: /docs/configure/settings/cpu-and-memory/  
[HPA]: /docs/configure/settings/hpa/  
[Runtime]: /docs/configure/settings/runtime/  
[Pod scheduling]: /docs/configure/settings/pod-scheduling/
