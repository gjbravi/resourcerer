# Security FAQ

> SOC 2 compliance, data collection and retention, SSO, and allowlist URLs


You asked:
- [How does one access Orion SOC 2 compliance reports?](#soc2-reports)
- [How is Orion deployed?](#how-is-orion-deployed)
- [What data do you collect?](#what-data-do-you-collect)
- [How do you collect data and where is it stored?](#how-do-you-collect-data-and-where-is-it-stored)
- [How long is data stored for?](#how-long-is-data-stored-for)
- [Who has access to customer data?](#who-has-access-to-customer-data)
- [Does Orion support federated single sign-on?](#federated-sso) 
- [Which Orion service URLs must be on the organization's allowlist?](#which-orion-service-urls-must-be-on-the-organizations-allowlist)
- [What metrics does Orion collect?](#metrics)

## How does one access Orion SOC 2 compliance reports? {#soc2-reports}

Orion is SOC 2 Type 2 compliant as determined by an audit completed by an accredited auditing firm. You can request access to the SOC 2 reports on the [SOC 2 Compliance Reports](https://www.orioncorp.io/report/soc-2-compliance/) page.

## How is Orion deployed?

- On your cluster, we deploy the following components:
    - **Orion Agent**, which reports on new workloads. The `oci://registry.orion.example/library/orion` Helm chart creates and uses a ServiceAccount called `orion-cluster-agent` and binds it to the Kubernetes `view` ClusterRole, granting **read-only** permissions to all resources in the cluster.
    - **Orion Forwarder**, which collects data and ships it to the Orion backend.
    - **Orion Applier**, which patches workloads with optimized resource utilization recommendations. The `oci://registry.orion.example/library/orion` Helm chart creates and uses a ServiceAccount called `orion-applier` and binds it to the Kubernetes `edit` ClusterRole, granting **update and patch** permissions to all optimizable workloads (and HPA, if enabled).
    - **Orion Node Agent**, which collects GPU metrics from GPU nodes. It is optional and disabled by default. When enabled, the Helm chart creates and uses a ServiceAccount called `orion-node-agent` and binds it to a ClusterRole granting **read-only** permissions (`get`, `list`, and `watch`) on pods and on `resourceslices.resource.k8s.io`. Because it reads GPU state per process, the DaemonSet also runs in the host PID namespace, holds the `SYS_ADMIN` capability while dropping all others, and mounts `/var/lib/kubelet/pod-resources` read-only. It runs with a read-only root filesystem and does not mount the container runtime socket or the host root filesystem.
- On our instances, we store the data and run machine learning to provide recommendations, which are presented in the Orion UI.  

## What data do you collect? {#what-data-do-you-collect}

From a targeted instance, we collect: 
- **Metadata**: node names, node UIDs, node instance types, kube-system namespace UID, namespace names, workload names, workload types, workload labels, pod names, pod requests and limits, container names, and container requests and limits. We also have the cluster name, which is provided by a user when installing Orion.

    - If you specify an allowNamespaces or a denyNamespaces list, we collect data accordingly. For example, we do not collect data about namespaces that you include in the denyNamespaces list.
    - You can choose to disable the collection of workload labels when you install Orion.
    - For workloads that request GPUs through Dynamic Resource Allocation, we collect the device identifiers from the resource claim: the driver and pool names, the device name, and the GPU UUID where it can be resolved.
    - If you enable the Node Agent, we also collect GPU device metadata: GPU UUIDs, GPU model names, and device indexes, plus Multi-Instance GPU (MIG) instance identifiers where MIG is in use.

- **Metrics**: We use the metadata above to build node, workload, and container metrics. For the complete list for metrics, see [What metrics does Orion collect?](#metrics) at the end of this topic. 

**We do not collect any personal data directly — only via social logins (such as Google or GitHub).**

 For details, see our [Privacy Policy][privacy-policy].

## How do you collect data and where is it stored?

The Orion Forwarder collects metrics data from a targeted instance via HTTPS requests, and then pushes the metrics to the Orion SaaS backend.
We store the parsed and ingested data in the Orion cloud. Data can be hosted in the US or EU region. Each customer has their own separate instance, and data is not shared.

## How long is data stored for?

How long we keep your data depends on the type of data:

- **Metric data**: We keep metrics collected from your clusters for about one year after collection.
- **Account data**: We keep your account, including user, cluster, and workload details, while your subscription is active and for about one year after it ends.

You can request that we delete your data at any time. For details on making a request, see our [Privacy Policy][privacy-policy].

## Who has access to customer data?

Access to production data is restricted to privileged Orion engineers on an as-needed temporary basis and only for the explicit intent of direct customer support.

When our Machine Learning team needs data for product improvement activities, we anonymize data. 
No external parties have access to customer data.

## Does Orion support federated single sign-on? {#federated-sso}

Yes. Orion supports OpenID Connect (OIDC), Security Assertion Markup Language (SAML), and other popular federated single sign-on (SSO) technologies through our authorization vendor. We can map groups from your authentication system into [Orion roles](reference__user-roles.md) (Viewer, Operator, Manager, Administrator).

For more information, contact your sales representative or [contact Orion sales](https://www.orioncorp.io/company/contact-us/).

## Which Orion service URLs must be on the organization's allowlist?

See the complete list in the [Orion installation prerequisites]( /docs/installation/install-adv/#prerequisites ).

## What metrics does Orion collect? {#metrics}

The following table lists the metrics that Orion collects from Kubernetes clusters.
- Workload-level metrics are generated by Orion [using metadata that we collect](#what-data-do-you-collect) and are prefixed with **orion**.
- Container-level metrics are built-in metrics provided by cAdvisor running on the Kubernetes node.

| Metric | Source | Why we collect it |
| ------------------------------------------------ | ----------------------------------- | -----------------------------------------------------------------------------|
| orion_node_allocated_requests     | Custom node metrics | The number of allocated requests on the node |
| orion_node_allocated_limits       | Custom node metrics | The number of allocated limits on the node   |
| orion_node_allocated_pods         | Custom node metrics | The number of non-terminated pods running on the node |
| orion_node_allocatable_resources  | Custom node metrics | The amount of resources available to be allocated on the node |
| orion_node_capacity_resources     | Custom node metrics | The total amount of resources that a node has. Used in calculating the average cluster CPU and memory utilization.   | 
| orion_node_labels                  | Custom node metrics | The labels on a node. Used to group nodes by things like instance type or node pool. |
| orion_workload_pod_owner          | Consolidated metric for ownership | With this metric, we have pod owner and workload, replacing KSM kube_pod_owner and kube_replicaset_owner |
| orion_workload_spec_replicas      | Consolidated metric for desired replicas number | With this metric, we have all desired replica metrics regardless type of pod owner. The pod owner must have the subresource scale. |
| orion_workload_status_replicas    | Consolidated metric for observed replicas number | With this metric, we have all observed replica metrics regardless type of pod owner. |
| orion_workload_pod_container_resource_requests | Consolidated pod metric with requests | With this metric, we have all requests metrics in a single metric. |
| orion_workload_pod_container_resource_limits | Consolidated pod metric with  limits | With this metric, we have all limits metrics in a single metric. |
| orion_workload_pod_resourceclaim_device_requests | Custom workload metrics | GPU devices a workload requests through Dynamic Resource Allocation |
| orion_pod_resourceclaim_device_requests | Custom workload metrics | GPU devices a pod requests through Dynamic Resource Allocation, and the device allocated to it |
| orion_workload_terminated_count   | Consolidated metric for workload termination events | With this metric, we track the count of times a workload was terminated (e.g. OOMKilled). |
| orion_workload_event_last_timestamp_seconds | Consolidated metric for reliability events | With this metric, we track when a reliability event, such as an OOMKill or node CPU pressure, was last observed for a workload. |
| orion_horizontalpodautoscaler_spec_min_replicas | KSM-like/horizontalpodautoscaler-metrics | Track minimum replicas for each HPA |
| orion_horizontalpodautoscaler_spec_max_replicas | KSM-like/horizontalpodautoscaler-metrics | Track maximum replicas for each HPA |
| orion_horizontalpodautoscaler_spec_target_metric | KSM-like/horizontalpodautoscaler-metrics | Track target metric for each HPA |
| container_cpu_usage_seconds_total  | cAdvisor  | Track CPU usage for each container |
| container_memory_working_set_bytes | cAdvisor  | Track memory usage for each container  |
| container_cpu_cfs_throttled_seconds_total | cAdvisor | Total time duration the container has been throttled |
| orion_container_cpu_pressure_waiting_seconds_total | cAdvisor | Total time that at least one task on the node was stalled waiting for CPU. Used to detect CPU pressure. |
| orion_container_cpu_pressure_stalled_seconds_total | cAdvisor | Total time that every task on the node was stalled waiting for CPU. Used to detect CPU pressure. |
| container_memory_max_usage_bytes  | cAdvisor | Maximum memory usage recorded |
| node_cpu_usage_seconds_total   | Resource metrics | Used in calculating the average cluster CPU utilization |
| node_memory_working_set_bytes  | Resource metrics | Used in calculating the average cluster memory utilization |

The following metrics are collected only when the [Node Agent](installation__node-agent-config.md) is enabled. The Node Agent reads them from the NVIDIA Management Library (NVML), from the kubelet PodResources API, and from NVIDIA profiling counters.

| Metric | Source | Why we collect it |
| ------------------------------------------------ | ----------------------------------- | -----------------------------------------------------------------------------|
| orion_container_gpu_memory_bytes | Node Agent (NVML) | GPU memory each container uses, so GPU usage can be attributed to a workload rather than to a shared device |
| orion_container_gpu_allocated | Node Agent (kubelet PodResources API) | The share of each GPU each container holds. Used to report GPU allocation and estimated GPU cost |
| orion_gpu_unallocated | Node Agent (kubelet PodResources API) | The fraction of each GPU that no container requested. Used to report unused GPU capacity |
| orion_gpu_memory_used_bytes, orion_gpu_memory_free_bytes, orion_gpu_memory_reserved_bytes, orion_gpu_memory_total_bytes | Node Agent (NVML) | Device-level GPU memory, used to show usage against device capacity |
| orion_gpu_utilization_ratio, orion_gpu_memory_copy_utilization_ratio | Node Agent (NVML) | Device-level GPU compute and memory-controller activity |
| orion_gpu_power_watts | Node Agent (NVML) | Device-level GPU power draw |
| orion_gpu_graphics_engine_active_ratio, orion_gpu_sm_active_ratio, orion_gpu_sm_occupancy_ratio, orion_gpu_tensor_core_active_ratio, orion_gpu_dram_active_ratio | Node Agent (NVIDIA profiling counters) | Device-level GPU engine activity, used to show how effectively a GPU is used |
| orion_gpu_scrape_errors, orion_gpu_allocation_scrape_errors, orion_gpu_allocated_unresolved_devices, orion_gpu_allocated_advertised_unknown_devices, orion_gpu_unallocated_clamped_devices | Node Agent | Collection health, so incomplete GPU data can be identified |

Source: Orion Helm chart README:
```sh
helm show readme oci://registry.orion.example/library/orion
```

[privacy-policy]: https://www.orioncorp.io/privacy-policy/
