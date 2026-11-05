# September 2026

Site [llms.txt](/llms.txt)

---

- **New Summary tab on the workload details page**\
The Summary tab replaces the Impact tab on the workload details page. It shows the workload's attributes and status messages, a table of current and recommended requests and cost, with the impact of each change, and CPU and memory usage charts for a single pod or the whole workload.

    ![Summary tab showing replica and HPA attributes, a status message, and a table of current and recommended per-pod and total requests and cost with the impact of each change][summary-tab]

- **GPU optimization (early access)**\
Orion now shows GPU usage for workloads that use NVIDIA GPUs: per-workload GPU memory and allocation on the Summary and Recommendation tabs, and device-level activity in a new GPU tab. Requires Orion Agent 3.2.0 and the new Node Agent, which is disabled by default. See [GPU optimization](recommendations__gpu.md) and [Node Agent configuration](installation__node-agent-config.md).

[summary-tab]: /img/docs/whats-new-2026-09-summary-tab.png
