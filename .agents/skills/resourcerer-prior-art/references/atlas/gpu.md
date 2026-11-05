---
updatedAt: 2026-08-26T10:21:57.000Z
---

# GPU Instances

Configure Atlas Autoscaler to scale your cluster using GPU-optimized instances across AWS EKS, GCP GKE, and Azure AKS with NVIDIA GPU support.

# Autoscaling using GPU instances

The Atlas Autoscaler can scale the cluster using GPU-optimized instances. This guide describes the steps needed to configure the cluster so that GPU nodes can join it.

## Supported providers

| Provider  | GPUs supported |
| --------- | -------------- |
| AWS EKS   | NVIDIA         |
| GCP GKE   | NVIDIA         |
| Azure AKS | NVIDIA         |

## How does it work?

Once activated, Atlas's Autoscaler detects workloads requiring GPU resources and starts provisioning them.

To enable the provisioning of GPU nodes, you need a few things:

* Choose a GPU instance type or attach a GPU to the instance type
* Install NVIDIA device plugin (needed for EKS and AKS)
* Expose the GPU to Kubernetes as a consumable resource.

Atlas ensures that the correct GPU instance type is selected - all you have to do is define GPU resources and add a GPU or a node template toleration. You can also target specific GPU characteristics using node selectors or affinities to the GPU labels.

| Label                   | Value Example     | Description                        |
| :---------------------- | :---------------- | :--------------------------------- |
| `nvidia.com/gpu`        | `true`            | Node has an NVIDIA GPU attached    |
| `nvidia.com/gpu.name`   | `nvidia-tesla-t4` | Attached GPU type                  |
| `nvidia.com/gpu.count`  | `1`               | Attached GPU count                 |
| `nvidia.com/gpu.memory` | `15258`           | Avaialble single GPU memory in Mib |

> 📘 Tainting of GPU nodes
>
> GPU nodes will have the nvidia.com/gpu=true:NoSchedule taint applied automatically, except when the Node Template has GPU constraints configured or limits instance types to GPU instances only. In such cases, the taint is not applied since the template is specifically designed for only GPU workloads.

## GPU sharing

Atlas supports GPU sharing to enable multiple workloads to utilize GPU resources more efficiently.
Multiple sharing methods are available, each optimized for different use cases:

* **<Anchor label="Time-slicing" target="_blank" href="doc:gpu-sharing-time-slicing">Time-slicing</Anchor>** - Share GPUs through rapid context switching
* **<Anchor label="Multi-Instance GPU (MIG)" target="_blank" href="doc:gpu-sharing-mig">Multi-Instance GPU (MIG)</Anchor>** - Partition GPUs with hardware-level isolation
* **<Anchor label="Multi-Process Service (MPS)" target="_blank" href="doc:gpu-sharing-mps">Multi-Process Service (MPS)</Anchor>** - Concurrent GPU kernel execution for multiple CUDA processes
* **<Anchor label="Fractional GPUs" target="_blank" href="doc:gpu-sharing-fractional">Fractional GPUs</Anchor>** - Pre-partitioned GPU portions from AWS (1/8 to full GPU)

For a detailed comparison and guidance on choosing the right method, see the <Anchor label="GPU sharing overview" target="_blank" href="doc:gpu-sharing">GPU sharing overview</Anchor>.

Atlas also supports GPU sharing through [Dynamic Resource Allocation (DRA)](dra.md).

## Custom and extended resources

Atlas supports scheduling workloads onto existing nodes that advertise custom or extended resources (for example, resources exposed by a customer-owned device plugin). However, the Atlas autoscaler cannot provision new nodes based on arbitrary custom resource requests, because it does not know how much of a custom resource each instance type provides.

If nodes already exist in the cluster that advertise the required custom resource, Atlas can schedule or move pods onto those nodes during rebalancing.

For workloads that require custom-resource-aware autoscaling, use one of the supported alternatives:

* [Multi-Process Service (MPS)](gpu-sharing-mps.md) on GKE
* [Dynamic Resource Allocation (DRA)](dra.md) with MPS on AWS, GKE, or AKS

If neither alternative fits your use case, contact [Atlas support](https://atlas.example/contact/) to discuss your requirements.

## NVIDIA device plugin

After creating a node of an instance type with a GPU, the node becomes part of the cluster, but GPU resources are not immediately usable. To make GPUs accessible to Kubernetes, you must install the NVIDIA device plugin on the node.

Atlas validates to ensure that the NIDIA device plugin exists on a cluster before performing any kind of autoscaling. If it doesn't detect the NVIDIA device plugin, it creates a pod event with details on how to resolve the problem.

### NVIDIA device plugin detection

Atlas assumes that the NVIDIA device plugin is installed if it finds a DaemonSet that matches the plugin's characteristics, and a pod created from that DaemonSet can run on a node.

Atlas supports NVIDIA device plugins whose DaemonSet names match specific patterns. Any DaemonSet matching one of these patterns is considered a NVIDIA device plugin.

* `*nvidia-device-plugin*`
* `*nvidia-gpu-device-plugin*`

Atlas also recognizes custom device plugin DaemonSets by the pod-template label `nvidia-device-plugin: "true"`. This label should only be applied to pods that actually run the NVIDIA device plugin — any DaemonSet carrying it will be treated as a device plugin provider.

> 📘 **NVIDIA device plugin requires successful pod scheduling**
>
> Atlas considers a NVIDIA device plugin present only if a pod from the matching DaemonSet can actually **run on the GPU node**.
> A DaemonSet that exists in the cluster but whose pods cannot be scheduled onto a node — for example, because the node has a taint that the DaemonSet does not tolerate — will not satisfy NVIDIA device plugin detection.
> This is a common source of the `NVIDIA Device Plugin is required` error message.
>
> If your Node Template includes custom taints, the NVIDIA device plugin DaemonSet **must** have a matching toleration for each of those taints.
> Without it, the device plugin pod will not be scheduled on newly provisioned GPU nodes, and Atlas will report a detection failure even if the plugin is otherwise correctly installed.
>
> If you see pod events with error message `NVIDIA Device Plugin is required`, check that the NVIDIA device plugin DaemonSet has tolerations for:
>
> * `nvidia.com/gpu:NoSchedule` (applied automatically by Atlas on most GPU nodes)
> * All Atlas system taints: `scheduling.atlas.example/spot`, `scheduling.atlas.example/scoped-autoscaler`, `scheduling.atlas.example/node-template`
> * Any custom taints defined in your Node Templates
>
> **Example:** If your Node Template has the taint `team: ml-workloads:NoSchedule`, add a specific toleration:
>
> ```yaml
> tolerations:
>   - key: "team"
>     operator: "Equal"
>     value: "ml-workloads"
>     effect: "NoSchedule"
> ```
>
> Alternatively, you can use a wildcard toleration that matches all taints:
>
> ```yaml
> tolerations:
>   - operator: Exists  # Tolerates ALL taints
> ```
>
> Note that the wildcard is permissive — it allows the pod to be scheduled on any node regardless of its taints. Use specific tolerations in production where possible.

### NVIDIA device plugin

By default, EKS and AKS clusters do not come with a NVIDIA device plugin installed. There are several ways to add a NVIDIA device plugin to a cluster:

> 📘 GKE clusters come with the NVIDIA device plugin pre-installed by default. No manual installation is required unless you want to use a custom NVIDIA device plugin.

#### During Atlas onboarding (only for EKS)

Atlas allows you to enable the GPU device plugin at the cluster onboarding stage. By ticking the checkbox in the UI, you can install plugins automatically during onboarding.

<Image align="center" border={true} src="https://files.readme.io/9118da7-image.png" className="border" />

#### Manually installing NVIDIA device plugin

Alternatively, you can manually install the plugin from the NVIDIA Helm repository.

```shell
helm repo add nvdp https://nvidia.github.io/k8s-device-plugin
helm repo update
```

Install the plugin using a values file. Copy the command below, which pipes the values directly via stdin:

```shell
helm upgrade -i nvdp nvdp/nvidia-device-plugin \
  --namespace atlas-agent \
  --create-namespace \
  -f - <<'EOF'
nodeSelector:
  nvidia.com/gpu: "true"

tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
  - key: scheduling.atlas.example/spot
    operator: Exists
  - key: scheduling.atlas.example/scoped-autoscaler
    operator: Exists
  - key: scheduling.atlas.example/node-template
    operator: Exists

affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        # Discrete GPU nodes (PCI vendor ID 10de = NVIDIA)
        - matchExpressions:
            - key: feature.node.kubernetes.io/pci-10de.present
              operator: In
              values:
                - "true"
            - key: nvidia.com/gpu.dra
              operator: NotIn
              values:
                - "true"
        # Tegra-based systems (CPU vendor NVIDIA)
        - matchExpressions:
            - key: feature.node.kubernetes.io/cpu-model.vendor_id
              operator: In
              values:
                - "NVIDIA"
            - key: nvidia.com/gpu.dra
              operator: NotIn
              values:
                - "true"
        # Manually labeled GPU nodes
        - matchExpressions:
            - key: nvidia.com/gpu.present
              operator: In
              values:
                - "true"
            - key: nvidia.com/gpu.dra
              operator: NotIn
              values:
                - "true"
EOF
```

#### Custom GPU device plugin

Atlas assumes a custom plugin controls the driver installation process and node management. If you wish to autoscale GPU nodes using a custom plugin, it must be <Anchor label="detectable to Atlas" target="_blank" href="#driver-detection">detectable to Atlas</Anchor>.

#### NVIDIA drivers and Amazon Machine Images (AMIs)

The NVIDIA device plugin requires that NVIDIA drivers and the `nvidia-container-toolkit` already exist on the machine, or it will fail to start, or expose GPU resources properly. By default, Atlas detects GPU-enabled nodes and uses the <Anchor label="EKS-optimized and GPU-enabled AMI by Amazon" target="_blank" href="https://docs.aws.amazon.com/eks/latest/userguide/eks-optimized-ami.html#gpu-ami">EKS-optimized and GPU-enabled AMI by Amazon</Anchor>, which already bundles these. However, if custom AMIs are used, then the installation of these prerequisites must also be included in the AMI building process or the node's user data scripts.

##### Amazon Linux 2023 AMI

**Amazon Linux 2023 (AL2023) AMIs support GPU time-slicing.** When using AL2023 for GPU nodes with time-slicing enabled, Atlas automatically configures the required NVIDIA device plugin settings.

See the <Anchor label="GPU sharing with time-slicing" target="_blank" href="doc:gpu-sharing-time-slicing#aws-eks">GPU sharing with time-slicing</Anchor> documentation for detailed AL2023 configuration instructions.

##### Bottlerocket AMI considerations

**Bottlerocket AMIs come with pre-installed NVIDIA device plugin.** Do not install additional NVIDIA device plugins on clusters where you plan to use Bottlerocket GPU nodes, as they may conflict with the pre-installed NVIDIA device plugin and fail to start.
If you do, you need to ensure these NVIDIA device plugin daemonsets do not run on those Bottlerocket nodes through tolerations.

When using Bottlerocket for GPU nodes, the OS handles driver management automatically. For more information, see <Anchor label="Bottlerocket support for NVIDIA GPUs" target="_blank" href="https://aws.amazon.com/blogs/containers/bottlerocket-support-for-nvidia-gpus/">Bottlerocket support for NVIDIA GPUs</Anchor>.

See <Anchor label="Node Configuration documentation" target="_blank" href="doc:node-configuration#virtual-machine-image-choice">Node Configuration documentation</Anchor> for more details on AMI choice.

> 🚧 Known issues
>
> NVIDIA dropped support for Kepler architecture GPUs after driver version 470. Since Atlas uses AMIs that bundle newer driver versions, these AMIs cannot be used with GPU instances that utilize such GPUs (P2 instance types). In order to use those instance types, an older GPU-enabled AMI or custom AMI must be set in the node configuration.

## Workload configuration examples

The following examples show common patterns for configuring GPU workloads. For GPU sharing-specific configurations, see the <Anchor label="time-slicing" target="_blank" href="doc:gpu-sharing-time-slicing#workload-configuration">time-slicing</Anchor> and <Anchor label="MIG" target="_blank" href="doc:gpu-sharing-mig#configuring-mig-workloads">MIG</Anchor> documentation.

```yaml GPU toleration
spec:
  tolerations:
    - key: "nvidia.com/gpu"
      operator: Exists
  containers:
    - image: my-image
      name: gpu-test
      resources:
        requests:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
        limits:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
```
```yaml Node template
spec:
  nodeSelector:
    scheduling.atlas.example/node-template: "gpu-node-template"
  tolerations:
    - key: "gpu-node-template"
      value: "template-affinity"
      operator: "Equal"
      effect: "NoSchedule"
  containers:
    - image: my-image
      name: gpu-test
      resources:
        requests:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
        limits:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
```
```yaml Specific GPU selection
spec:
  nodeSelector:
    nvidia.com/gpu.name: "nvidia-tesla-t4"
  tolerations:
    - key: "nvidia.com/gpu"
      operator: Exists
  containers:
    - image: my-image
      name: gpu-test
      resources:
        requests:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
        limits:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
```
```yaml Node template & specific GPU
spec:
  nodeSelector:
    scheduling.atlas.example/node-template: "gpu-node-template"
    nvidia.com/gpu.name: "nvidia-tesla-p4"
  tolerations:
    - key: "scheduling.atlas.example/node-template"
      value: "gpu-node-template"
      operator: "Equal"
      effect: "NoSchedule"
  containers:
    - image: my-image
      name: gpu-test
      resources:
        requests:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
        limits:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
```
```yaml GPU selection based on memory
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: nvidia.com/gpu.memory
            operator: Gt
            values:
            - "10000"
  tolerations:
    - key: "nvidia.com/gpu"
      operator: Exists
  containers:
    - image: my-image
      name: gpu-test
      resources:
        requests:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
        limits:
          cpu: 1
          memory: 1Gi
          nvidia.com/gpu: 1
```