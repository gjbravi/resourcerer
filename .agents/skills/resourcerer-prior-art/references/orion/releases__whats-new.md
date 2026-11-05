# What's new in Orion

> Monthly summaries of Orion updates

---

Entries say when a change needs a new Agent or CLI version; everything else applies automatically.

---

## September 2026

- **New Summary tab on the workload details page**\
The Summary tab replaces the Impact tab on the workload details page. It shows the workload's attributes and status messages, a table of current and recommended requests and cost, with the impact of each change, and CPU and memory usage charts for a single pod or the whole workload.

    ![Summary tab showing replica and HPA attributes, a status message, and a table of current and recommended per-pod and total requests and cost with the impact of each change][summary-tab]

- **GPU optimization (early access)**\
Orion now shows GPU usage for workloads that use NVIDIA GPUs: per-workload GPU memory and allocation on the Summary and Recommendation tabs, and device-level activity in a new GPU tab. Requires Orion Agent 3.2.0 and the new Node Agent, which is disabled by default. See [GPU optimization](recommendations__gpu.md) and [Node Agent configuration](installation__node-agent-config.md).

[summary-tab]: /img/docs/whats-new-2026-09-summary-tab.png

---

## August 2026

- **Redesigned Recommendation tab**\
The Recommendation tab on the workload details page now shows current and recommended requests and limits side by side for each container, with usage sparklines and the percentage change. Expand a container for CPU and memory charts with selectable time ranges, percentile readouts, and OOMKilled and CPU throttling markers.

    ![Recommendation tab showing current and recommended values for a container with CPU and memory usage charts][rec-tab]

[rec-tab]: /img/docs/whats-new-2026-08-recommendation-tab.png

---

## July 2026

- **Schedule time zone shown in workload settings**\
The Schedule section of a workload's Settings tab displays the time zone its schedule expressions are evaluated in.

- **Workload expiration extended from 14 to 28 days**\
Workloads that stop reporting now remain available for 28 days before they expire, doubled from the previous 14 days.

---

## June 2026

- **Orion Agent 3.0**\
A single `orion` Helm chart now installs the Agent and Applier together, replacing the separate `orion-agent` chart. Upgrading from v2 is a breaking change that requires migration. See [Orion Agent releases] and [Migrate v2 to v3].

[Orion Agent releases]: /docs/releases/orion/
[Migrate v2 to v3]: /docs/installation/migration/

---

## May 2026

- **CPU throttling visibility**\
The Workloads page includes a CPU throttling column, and the workload details page includes a CPU throttling chart, so you can see which containers are hitting their CPU limits and how often.

---

## March 2026

- **HPA settings shown on the workload details page**\
Workloads managed by a Horizontal Pod Autoscaler now show their HPA settings on the workload details page. See [HPA] for how Orion adjusts target utilization when it changes resource requests.

[HPA]: /docs/configure/settings/hpa/

---

## February 2026

- **Recommendations applied without restarting Pods**\
The webhook apply methods now use Kubernetes in-place Pod resizing by default, updating requests and limits on running Pods instead of triggering a rollout. Requires Applier v2.12.0 or later and Kubernetes v1.33 or later; on earlier Kubernetes versions these methods fall back to a rollout or defer the change until the next restart. See [Apply settings].

[Apply settings]: /docs/configure/settings/apply-settings/
