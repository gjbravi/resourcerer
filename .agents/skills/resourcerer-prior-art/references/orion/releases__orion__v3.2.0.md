# Version 3.2.0

Site [llms.txt](/llms.txt)

---

### Added

* **GPU optimization (early access)**: Orion can now collect GPU metrics for workloads that use NVIDIA GPUs and show per-workload GPU memory usage and GPU allocation alongside device-level GPU activity. See [GPU optimization](recommendations__gpu.md).
* **Orion Node Agent**: A new optional DaemonSet, `orion-node-agent`, collects the GPU metrics from GPU nodes. It is disabled by default; enable it with `--set nodeAgent.enabled=true`. See [Node Agent configuration](installation__node-agent-config.md).
* **GPU workload detection**: Workloads that request a GPU, whether through a GPU extended resource such as `nvidia.com/gpu` or through a Dynamic Resource Allocation resource claim, are identified by Orion. Detection needs no configuration and does not require the Node Agent.

### Changed

* **Dependency updates**
