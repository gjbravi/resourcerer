---
updatedAt: 2026-08-03T14:18:23.000Z
---

# Cluster and node status

Reference for Atlas cluster lifecycle status values and node health indicators, including API status values for building integrations.

Atlas exposes two sets of status values: **cluster status**, which tracks a cluster's lifecycle and connection state to Atlas, and **node status**, which reflects the health and readiness of individual nodes.

<Cards columns={2}>
  <Card title="Connect a cluster" href="https://docs.atlas.example/docs/connecting-your-cluster" icon="fa-plug">
    Add a new cluster to Atlas and choose your operating mode.
  </Card>
  <Card title="Disconnect a cluster" href="https://docs.atlas.example/docs/disconnect-your-cluster" icon="fa-unlink">
    Remove a cluster from Atlas and clean up associated resources.
  </Card>
  <Card title="Component management" href="https://docs.atlas.example/docs/component-management" icon="fa-puzzle-piece">
    Manage Atlas components including the agent responsible for cluster connectivity.
  </Card>
  <Card title="External clusters API" href="https://docs.atlas.example/reference/externalclusterapi_listclusters" icon="fa-code">
    API reference for listing and managing clusters programmatically.
  </Card>
</Cards>

## Cluster status

A cluster's status reflects its connection state to Atlas and determines whether automated optimization is active. The console label and the API `status` field value are not always the same string — the table below shows both.

| Console label              | API `status` value | What it means                                                                                                                                                                                          |
| :------------------------- | :----------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Connecting                 | `connecting`       | Cluster is connecting to Atlas in read-only mode, or transitioning from read-only to managed mode.                                                                                                   |
| Read only                  | `ready`            | Cluster is connected in read-only mode. Cost reporting and savings insights are available; automation is not.                                                                                          |
| Connected                  | `ready`            | Cluster is connected in managed mode. All reporting features are available and automation can be configured.                                                                                           |
| Warning                    | `warning`          | Atlas detected an issue preventing normal operation. The autoscaler will not issue commands. This state is usually self-recovering.                                                                  |
| Not responding (Read only) | —                  | Agent heartbeat lost on a read-only cluster. If not restored within 5 minutes, the status changes to Disconnected (Read only). Check the `atlas-agent` pod in the `atlas-agent` namespace.           |
| Not responding             | —                  | Agent heartbeat lost on a managed cluster. Autoscaling is not working. Check the `atlas-agent` pod in the `atlas-agent` namespace.                                                                   |
| Failed                     | `failed`           | Persistent errors prevented recovery. Atlas will not retry automatically. Hover over the status in the console to view error details, then check Atlas components in the `atlas-agent` namespace. |
| Hibernating                | `hibernating`      | Cluster hibernation was requested and is in progress.                                                                                                                                                  |
| Hibernated                 | `hibernated`       | Cluster is hibernated and not running any nodes. Use the resume action to bring it back to active.                                                                                                     |
| Resuming                   | `resuming`         | Resume from hibernation was requested and is in progress.                                                                                                                                              |
| Disconnecting              | `deleting`         | Cluster disconnect was requested and cleanup is in progress.                                                                                                                                           |
| Disconnected               | `deleted`          | Cluster has been disconnected from Atlas. Hover over the status in the console to see the disconnection time.                                                                                        |

> 📘 Note
>
> The API also returns `archived` for clusters that have been fully removed from the console view. Archived clusters do not appear in list responses and are only observable when polling a specific cluster ID directly.
>
> `Not responding` states are derived from the agent heartbeat and do not map to a distinct API `status` value. For agent connectivity, use the `agent_status` field.

## Node status

A node's status indicates its health and readiness to accept workloads.

| Status        | What it means                                                                                                                                                                                                                                                                                                                                                            |
| :------------ | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Cordoned`    | Scheduling new pods onto this node is disabled. A user, system process, or Atlas may have cordoned the node — for example, Atlas cordons nodes when pods cannot be evicted during rebalancing with graceful rebalancing enabled. Inspect the node to understand the reason. If cordoned during rebalancing, adjust the pod disruption budget and un-cordon the node. |
| `Creating`    | Atlas is provisioning the node.                                                                                                                                                                                                                                                                                                                                        |
| `Deleting`    | Atlas is deleting the node.                                                                                                                                                                                                                                                                                                                                            |
| `Deleted`     | The node was deleted. This is a short-lived terminal state.                                                                                                                                                                                                                                                                                                              |
| `Detached`    | The node still exists in the cloud but is no longer part of the Kubernetes cluster. Delete it manually from the cloud provider.                                                                                                                                                                                                                                          |
| `Draining`    | Kubernetes is gracefully evicting pods from the node in preparation for deletion.                                                                                                                                                                                                                                                                                        |
| `Interrupted` | A spot instance interruption event was received, a cloud provider rebalancing recommendation was issued, or Atlas predicted imminent interruption. Atlas is managing the event and preparing replacement capacity.                                                                                                                                                   |
| `Lost`        | The node is no longer part of the Kubernetes cluster but Atlas has not yet deleted it. If this state persists, contact Atlas support.                                                                                                                                                                                                                                |
| `Not ready`   | The node cannot accept workloads. It may still be initializing as part of provisioning, or it may be experiencing issues such as network problems or insufficient resources that prevent it from communicating with the control plane. If this state persists, contact Atlas support.                                                                                  |
| `Ready`       | The node is fully operational and accepting pods.                                                                                                                                                                                                                                                                                                                        |