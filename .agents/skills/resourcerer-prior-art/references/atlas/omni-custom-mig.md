---
updatedAt: 2026-11-02T08:06:54.000Z
---

# Custom MIG configurations

OMNI can provision GPU instances with MIG partitioning and optional DRA (Dynamic Resource Allocation) support. When you create an edge node, you can request GPUs with a MIG configuration and DRA enabled — see [Create an edge](https://docs.atlas.example/reference/edgesv2betaapi_createedge). With DRA enabled, GPU resources are advertised through Kubernetes ResourceSlices and allocated via ResourceClaims, so you can create a ResourceClaim, assign it to a workload definition, and consume a single MIG slice.

> 📘 Note
>
> The OMNI API applies one MIG configuration to the whole instance: the same partition sizes are applied to every GPU attached to the VM. You cannot define a non-homogeneous MIG layout — different profiles per GPU, or a mix of profiles on a single GPU — through the API alone.

This guide shows how to achieve custom, mixed MIG layouts on OMNI edge nodes by customizing the `nvidia-gpu-operator` add-on installed on the edge cluster.

# Before you begin

* OMNI enabled on your cluster. See [Getting started with OMNI](omni-getting-started.md).
* At least one [edge location](omni-getting-started.md#step-2-create-and-onboard-an-edge-location) with the `nvidia-gpu-operator` add-on enabled.
* An edge instance with MIG-capable GPUs — for example, a GCP `a2-highgpu-2g` machine type with 2× A100-40GB GPUs. MIG is supported on NVIDIA GPUs from the Ampere architecture onward; see [Multi-Instance GPU (MIG)](gpu-sharing-mig.md) and the [NVIDIA MIG documentation](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html) for supported GPUs and partition sizes.

# How MIG is managed on edge nodes

MIG management on an edge node is performed by the MIG Manager, a component of the NVIDIA GPU Operator installed through the `nvidia-gpu-operator` add-on. The MIG Manager:

1. Watches the `nvidia.com/mig.config` node label.
2. Looks up the label value in a mig-parted configuration — a ConfigMap on the edge cluster — to determine the MIG layout to apply.
3. Applies the layout using [mig-parted](https://github.com/NVIDIA/mig-parted).

By default, the add-on uses the `default-mig-parted-config` ConfigMap that ships with the GPU Operator. You can point the add-on at your own ConfigMap using the `migManager.config.name` Helm value. When a configuration mixes different MIG profiles (for example, `1g.5gb` and `2g.10gb` on the same GPU), the GPU Operator must run with the `mixed` MIG strategy (`mig.strategy=mixed`); the `single` strategy applies one uniform profile to all GPUs.

For details, see the [NVIDIA GPU Operator documentation](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/gpu-operator-mig.html).

# Step 1: Configure the gpu-operator add-on

<Tabs>
  <Tab title="Atlas console">
    1. In the Atlas console, navigate to **Autoscaler → Edge locations**.
    2. Create a new edge location, or edit an existing one, and enable the **NVIDIA GPU Operator** add-on.
    3. In the add-on values, set the MIG strategy to `mixed` and the mig-parted ConfigMap name to `custom-mig-parted-config`.
  </Tab>
  <Tab title="API">
    Patch the edge location so the add-on uses a custom ConfigMap and the `mixed` MIG strategy:

    ```shell
    curl --request PATCH \
      'https://api.atlas.example/omni-provisioner/v1beta/organizations/<ORG_ID>/clusters/<CLUSTER_ID>/edge-locations/<EDGE_LOCATION_ID>' \
      --header 'Content-Type: application/json' \
      --header 'Accept: application/json' \
      --header 'X-API-Key: <API_KEY>' \
      --data '{
        "edgeClusterSpec": {
          "addons": [
            {
              "name": "nvidia-gpu-operator",
              "values": {
                "mig": {
                  "strategy": "mixed"
                },
                "migManager": {
                  "config": {
                    "name": "custom-mig-parted-config"
                  }
                }
              }
            }
          ]
        }
      }'
    ```
  </Tab>
</Tabs>

The console and the API write the same [edge location add-on values](https://docs.atlas.example/reference/edgelocationsapi_createedgelocation) (`edgeClusterSpec.addons`) — the values are deep-merged over the add-on's default Helm values.

> 📘 Note
>
> The `addons` list defines the complete set of optional add-ons installed on the edge cluster — when the list is set, only the listed add-ons are installed. If the edge location already has optional add-ons enabled (for example, `nvidia-dra`), include them in the list alongside `nvidia-gpu-operator`, otherwise the update removes them.

# Step 2: Apply the custom MIG ConfigMap

Create the ConfigMap the add-on now references. The ConfigMap must contain a `config.yaml` key. In the examples below, `0x20B010DE` is the device ID of the A100-SXM4-40GB GPU type.

## Single GPU with mixed MIG profiles

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-mig-parted-config
  namespace: gpu-operator
data:
  config.yaml: |-
    version: v1
    mig-configs:
      custom-mixed:
        - device-filter: ["0x20B010DE"]
          devices: all
          mig-enabled: true
          mig-devices:
            "1g.5gb": 4
            "2g.10gb": 1
```

## Two GPUs with different MIG profiles

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-mig-parted-config
  namespace: gpu-operator
data:
  config.yaml: |-
    version: v1
    mig-configs:
      custom-mixed:
        - device-filter: ["0x20B010DE"]
          devices: [0]
          mig-enabled: true
          mig-devices:
            "1g.5gb": 7
        - device-filter: ["0x20B010DE"]
          devices: [1]
          mig-enabled: true
          mig-devices:
            "2g.10gb": 3
```

Apply the manifest:

```shell
kubectl apply -f custom-mig-config.yaml
```

# Step 3: Trigger the MIG reconfiguration

Label the edge node with the custom profile:

```shell
kubectl label nodes <NODE_NAME> nvidia.com/mig.config=custom-mixed --overwrite
```

The MIG Manager picks up the change and applies the layout. To watch it work:

```shell
kubectl logs -n gpu-operator -l app=nvidia-mig-manager -c nvidia-mig-manager -f
```

When the reconfiguration completes, confirm the node reports a successful state:

```shell
kubectl get node <NODE_NAME> -o jsonpath="{.metadata.labels['nvidia.com/mig.config.state']}"
```

The value should be `success`. To inspect the resulting MIG devices:

```shell
kubectl exec -it -n gpu-operator ds/nvidia-driver-daemonset -- nvidia-smi -L
```

> ⚠️ Warning
>
> MIG Manager terminates workloads that use the GPUs before applying a new layout, and on some cloud providers the node must reboot to enable MIG mode. Reconfiguration is per node — repeat the labeling step on every edge node that should use the custom layout.

# Related documentation

* [Multi-Instance GPU (MIG)](gpu-sharing-mig.md)
* [Manual edge provisioning](omni-manual-edge-provisioning.md)
* [Custom edge locations](omni-custom-edge-location.md)
* [Create an edge location](https://docs.atlas.example/reference/edgelocationsapi_createedgelocation)
* [Create an edge](https://docs.atlas.example/reference/edgesv2betaapi_createedge)
* [Update edge location](https://docs.atlas.example/reference/edgelocationsapi_updateedgelocation)