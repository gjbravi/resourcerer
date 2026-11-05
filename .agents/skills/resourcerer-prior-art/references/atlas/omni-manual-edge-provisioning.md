---
updatedAt: 2026-10-30T13:47:53.000Z
---

# Manual edge provisioning

# Manual edge provisioning

By default, edge nodes are created automatically: the [Autoscaler provisions them](omni-getting-started.md#edge-node-provisioning) whenever a node template with edge locations enabled needs capacity. Manual edge provisioning lets you create a single edge node directly, with an exact instance type, GPU sharing configuration, and edge configuration of your choosing — outside of the Autoscaler's normal decision-making.

## When to use manual provisioning

Manual provisioning is useful when you want a specific outcome rather than the most cost-effective one, for example:

* **Validating a new edge location** before adding it to a node template, without needing a pending pod to trigger provisioning.
* **Reserving a specific instance type or GPU** (e.g. a specific MIG partition layout) for a known workload.
* **Testing GPU sharing configurations** (time-slicing or MIG) on a given instance type before rolling them out via node templates.
* **One-off or ad hoc capacity** that shouldn't be subject to Autoscaler bin-packing or downscaling decisions.

> 📘 Note
> Manually provisioned edge nodes still appear as standard edge nodes in your cluster (see [How edge nodes appear in your cluster](omni-getting-started.md#how-edge-nodes-appear-in-your-cluster)) and are subject to the same [Evictor behavior](omni-getting-started.md#evictor-behavior-with-edge-nodes) as Autoscaler-provisioned ones.

## Prerequisites

* A [Phase 2 cluster](https://docs.atlas.example/docs/cluster-onboarding) with [OMNI enabled](omni-getting-started.md#step-1-onboard-your-cluster-to-omni).
* At least one [edge location](omni-getting-started.md#step-2-create-and-onboard-an-edge-location) that has completed onboarding (status **Incomplete setup** or **In use** — not **Pending** or **Failed**).

## Provision an edge node using the console

1. In the Atlas console, navigate to **Autoscaler → Edge locations**.

2. Select the edge location you want to provision the node in, and open **Provision an edge node**.

3. Configure the node:

   * **Edge location** — the edge location to provision into. Pre-filled if you started from a specific edge location.
   * **Instance type** — the cloud instance type for the node.
   * **Availability zone** — filtered to zones available for the selected instance type within this edge location. Select the instance type first.
   * **Edge configuration** — the [edge configuration](omni-edge-configuration.md) to use for this node (boot disk, image, instance tags, container runtime socket). If none is selected, the edge location's default configuration is used.
   * **Resource offering** — **On-Demand** or **Spot**.

4. *(Optional, GPU instance types only)* Configure GPU sharing via the standard NVIDIA device plugin:

   * **GPU time-slicing** — set **Replicas** to time-slice a GPU into that many virtual replicas for shared access.
   * **Multi-instance GPU (MIG)** — select **Partition sizes** to split the GPU into isolated slices with dedicated memory. Available for A100/H100 instances only.

   These two settings can be combined: on a MIG-capable instance type, you can partition the GPU into MIG slices and additionally time-slice each MIG partition, exposing more schedulable GPU replicas from the same physical device. This is standard NVIDIA device plugin behavior and is separate from [Dynamic Resource Allocation (DRA)](dra.md).

5. *(Optional)* Configure **Kubernetes orchestration**:

   * **Taint nodes** — applies a taint so pods without a matching toleration won't be scheduled on this node.
   * **Use custom labels** — apply additional Kubernetes labels to the node.

6. *(Optional)* Under **Edge configuration overrides → Storage**, set a **Boot disk size** in GiB. Leave blank to inherit the size from the linked edge configuration.

7. Click **Provision** to submit the request.

The node is created asynchronously; it appears in **kubectl get nodes** once provisioning completes (see [Edge node provisioning time](omni-getting-started.md#edge-node-provisioning-time)).

## Provision an edge node using the API

You can provision an edge node programmatically with the [Create an edge](https://docs.atlas.example/reference/edgesv2betaapi_createedge) endpoint.

```
POST https://api.atlas.example/omni-provisioner/v2beta/organizations/{organizationId}/clusters/{clusterId}/edge-locations/{edgeLocationId}/edges
```

### Path parameters

| Parameter        | Type   | Description                                    |
| ---------------- | ------ | ---------------------------------------------- |
| `organizationId` | string | The ID of the organization.                    |
| `clusterId`      | string | The cluster ID.                                |
| `edgeLocationId` | string | The edge location ID to provision the node in. |

### Body parameters

| Parameter          | Type             | Required | Description                                                                                                                                                                                                                      |
| ------------------ | ---------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`             | string           | Yes      | Name of the edge node.                                                                                                                                                                                                           |
| `instanceType`     | string           | Yes      | Instance type of the edge node.                                                                                                                                                                                                  |
| `schedulingType`   | enum             | Yes      | `ON_DEMAND` or `SPOT`. Defaults to on-demand.                                                                                                                                                                                    |
| `zone`             | string           | Yes      | The zone for the edge node. Must be part of the edge location.                                                                                                                                                                   |
| `nodeArchitecture` | enum             | Yes      | `X86_64` or `ARM64`.                                                                                                                                                                                                             |
| `configurationId`  | string           | No       | Reference to an edge configuration to use. If omitted, the default configuration is used.                                                                                                                                        |
| `gpuConfig`        | object           | No       | GPU configuration for the instance — corresponds to the console's **GPU time-slicing** and **Multi-instance GPU (MIG)** settings (replica count and/or partition sizes). Both can be set together on MIG-capable instance types. |
| `kubernetesLabels` | object           | No       | Kubernetes labels to add to the node. Overrides labels from the edge configuration on key conflicts.                                                                                                                             |
| `kubernetesTaints` | array of objects | No       | Kubernetes taints to add to the node. Overrides taints from the edge configuration on key conflicts.                                                                                                                             |
| `bootDiskGib`      | uint32           | No       | Boot disk size in GiB. Leave unset to inherit from the linked edge configuration.                                                                                                                                                |
| `imageId`          | string           | No       | Image ID for the edge node.                                                                                                                                                                                                      |
| `instanceLabels`   | object           | No       | Cloud provider instance labels/tags added to the edge node.                                                                                                                                                                      |

### Example request

```bash
curl -X POST \
  "https://api.atlas.example/omni-provisioner/v2beta/organizations/${ORGANIZATION_ID}/clusters/${CLUSTER_ID}/edge-locations/${EDGE_LOCATION_ID}/edges" \
  -H "X-API-Key: ${ATLAS_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpu-edge-node-01",
    "instanceType": "a2-highgpu-1g",
    "schedulingType": "ON_DEMAND",
    "zone": "us-central1-a",
    "nodeArchitecture": "X86_64",
    "bootDiskGib": 200,
    "kubernetesLabels": {
      "workload-type": "gpu-inference"
    }
  }'
```

A successful request returns `200 OK` with the created edge object. See the [API reference](https://docs.atlas.example/reference/edgesv2betaapi_createedge) for the full response schema and error format.

## Verifying the node

```bash
kubectl get nodes
```

A manually provisioned node shows up the same way as an Autoscaler-provisioned edge node — as a virtual node with role `agent`, tainted `virtual-node.omni.atlas.example/not-allowed=true:NoExecute` by default, and labeled with `omni.atlas.example/edge-location-name`, `omni.atlas.example/edge-id`, and `omni.atlas.example/csp`. See [How edge nodes appear in your cluster](omni-getting-started.md#how-edge-nodes-appear-in-your-cluster) for the full list of labels and taints.

To schedule workloads on it, follow the same steps as for any edge node — see [Scheduling workloads on edge nodes](omni-getting-started.md#scheduling-workloads-on-edge-nodes).

## Troubleshooting

| Issue                                     | Possible cause                                                                     | Solution                                                                                  |
| ----------------------------------------- | ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| Instance type dropdown is empty           | Edge location not fully onboarded                                                  | Verify the edge location status is **Incomplete setup** or **In use**                     |
| Availability zone dropdown stays disabled | No instance type selected yet                                                      | Select an instance type first                                                             |
| GPU time-slicing / MIG fields disabled    | Selected instance type has no GPU, or doesn't support MIG (MIG requires A100/H100) | Choose a GPU-capable instance type, or a MIG-supported instance type for MIG specifically |
| Node never appears in `kubectl get nodes` | Provisioning request failed, or edge location is in a **Failed** state             | Check the edge location status; re-run the edge location onboarding script if needed      |
| Duplicate nodes from retried requests     | No `clientRequestId` supplied on retry                                             | Always pass the same `clientRequestId` when retrying a request                            |

## FAQ

**Does a manually provisioned edge node count toward node template constraints?**
Yes. It's a regular node in the cluster and is counted like any other node when Autoscaler and Evictor evaluate constraints and bin-packing.

**Can I combine GPU time-slicing and MIG on the same node?**
Yes, on MIG-capable instance types. MIG partitions the physical GPU into isolated slices, and time-slicing can then be applied on top of each MIG partition to expose additional virtual replicas.

**Will the Autoscaler remove a manually provisioned node?**
The [Evictor](omni-getting-started.md#evictor-behavior-with-edge-nodes) can evict workloads off it like any edge node, but it doesn't delete manually provisioned nodes outright — the same downscaling behavior that applies to other nodes in your cluster applies here.