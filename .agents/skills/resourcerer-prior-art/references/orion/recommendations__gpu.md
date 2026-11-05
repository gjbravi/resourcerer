# GPU optimization

> GPU visibility for Kubernetes workloads


Orion can collect GPU metrics for workloads that use NVIDIA GPUs and show each workload's own GPU usage alongside its CPU and memory usage. The Node Agent collects these metrics. See [Node Agent configuration][node-agent-config].

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Early access feature</div>


To share feedback on GPU optimization, [contact Orion support](https://www.orioncorp.io/company/contact-us/).
</div>


## What Orion shows for GPU workloads

When a workload uses GPUs, the workload details page adds the following.

| Tab | What it shows |
| --- | ------------- |
| **Summary** | A GPU attribute indicator, a GPU row in the resource table with per-pod and total GPU requests, and a Per-Pod GPU Memory chart plotting Usage against device memory Capacity. |
| **Recommendation** | A GPU column in Container Values showing the current per-container GPU count, and a GPU Memory usage chart alongside CPU and Memory showing per-container usage. |
| **GPU** | Aggregated GPU Device Activity for the specific GPU devices the workload ran on. Aggregation is broken down by GPU Model. |

The Summary and Recommendation tabs show *per-pod* data: they show GPU information specifically attributable to the workload's own containers. The GPU tab is *device-level*: it shows total activity across whole GPU devices the workload was allocated, which other workloads may also be contributing to when devices are shared. The GPU tab's Allocated GPUs line counts the distinct physical GPUs the workload's pods hold: [MIG][mig] devices and shared GPUs count as their physical GPU, and each physical GPU counts once.

For more information about the workload details page, see [View recommendations][view].

## How GPU metrics are collected

The Node Agent is a DaemonSet that runs on GPU nodes. On each node it reads GPU state directly from the NVIDIA driver, reads GPU memory per process, and maps each process back to the container that owns it. This is what makes per-workload attribution possible on a GPU shared between pods. Because it reads the driver directly, you do not need to run `dcgm-exporter` or any other GPU exporter alongside it.

Device-level engine activity metrics, such as streaming multiprocessor and tensor core activity, depend on NVIDIA profiling counters. Where a GPU or driver does not support profiling, those metrics are absent and the rest of the GPU metrics are unaffected.

To turn collection on, see [Node Agent configuration][node-agent-config].

## How GPU workloads are identified

Orion marks a workload as using GPUs when any of its containers requests a GPU extended resource, such as `nvidia.com/gpu`, or when workloads request GPUs through a [Dynamic Resource Allocation][dra] `ResourceClaim` or `ResourceClaimTemplate`.

Detection needs no configuration and happens whether or not the Node Agent is enabled. Collecting GPU metrics requires the Node Agent.

To list GPU workloads from the command line with Orion CLI version 6.2.8 or later:

```sh
orion view workloads --gpu
```

## Enabling GPU visibility

Prerequisites:

* Orion Agent version 3.2.0 or later
* Helm version 3.14.0 or later (required for the `--reset-then-reuse-values` parameter)
* Nodes with NVIDIA GPUs, running the NVIDIA container runtime
* GPU nodes on the `linux/amd64` architecture

The Node Agent needs node access that some managed and hardened Kubernetes platforms do not permit. It runs in the host PID namespace, holds the `SYS_ADMIN` capability, and mounts a host path. If your cluster restricts these, confirm that its policies admit the DaemonSet before you enable it. See [Permissions and host access][permissions].

### Steps

1. Enable the Node Agent. In an existing cluster, run:

    ```sh
    helm upgrade orion oci://registry.orion.example/library/orion \
      -n orion-system \
      --reset-then-reuse-values \
      --set nodeAgent.enabled=true
    ```

    By default the Node Agent is scheduled onto nodes labeled `nvidia.com/gpu.present: "true"` and tolerates all taints. If your GPU nodes are labeled differently, override the selector as described in [Node selection][node-selection]. See [Node Agent configuration][node-agent-config] for the full set of parameters.

2. Verify that the DaemonSet is running on your GPU nodes:

    ```sh
    kubectl get pods -n orion-system -l app.kubernetes.io/component=node-agent -o wide
    ```

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Installation<span class="related-topics-divider">/</span><a href="/docs/installation/node-agent-config/">Node Agent configuration</a>
                </h5>
                <p>Configure the Orion Node Agent component</p>
            </div>
        </div>


[node-agent-config]: /docs/installation/node-agent-config/
[permissions]: /docs/installation/node-agent-config/#permissions
[node-selection]: /docs/installation/node-agent-config/#node-selection
[view]: /docs/recommendations/view/
[dra]: https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/
[mig]: https://docs.nvidia.com/datacenter/tesla/mig-user-guide/
