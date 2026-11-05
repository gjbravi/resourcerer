# Version 3.3.0

Site [llms.txt](/llms.txt)

---

### Added

* **Node CPU pressure detection**: The Agent detects Kubernetes nodes under sustained CPU pressure and identifies the workloads responsible. It reads each node's Pressure Stall Information (PSI) from cAdvisor. When a node's CPU pressure exceeds the detection threshold, the Agent finds the containers on that node using more CPU than they request, maps them to their workloads, and creates a recommendation with the `CPUPressureDetected` reason. Each recommendation lists the node, pod, and container, and how far CPU usage exceeded the request. Detection is always on. See [CPU pressure response](configure__settings__reliability.md#cpu-pressure).

  The Agent reports these metrics for CPU pressure detection:

  * `orion_container_cpu_pressure_waiting_seconds_total`
  * `orion_container_cpu_pressure_stalled_seconds_total`
  * `orion_container_cpu_pressure_waiting_rate`
  * `orion_container_cpu_pressure_stalled_rate`
  * `orion_workload_event_last_timestamp_seconds`

### Changed

* **Dependency updates**

### Fixed

* **Node metrics refresh on provider ID change**: Node metrics now pick up provider IDs that are assigned late or change. Previously, node updates that changed only the provider ID were ignored.
* **Provider ID sync to Orion API**: When cluster data is enabled, the Agent sends a node's provider ID to the Orion API whenever it changes. Previously, it was sent only when a node was added or its labels changed.
