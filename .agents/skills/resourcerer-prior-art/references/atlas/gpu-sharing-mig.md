---
updatedAt: 2026-05-12T14:01:40.000Z
---

# Multi-Instance GPU (MIG)

Atlas supports NVIDIA Multi-Instance GPU (MIG) technology, enabling you to partition powerful GPUs into smaller, isolated instances. This feature maximizes GPU utilization and cost efficiency by allowing multiple workloads to share a single physical GPU while maintaining hardware-level isolation.

# What is MIG?

Multi-Instance GPU (MIG) allows you to securely partition select NVIDIA GPUs into up to seven separate GPU instances. Each MIG instance provides:

* **Hardware-level isolation** - Dedicated memory, cache, and Streaming Multiprocessors (SM)
* **Quality of service** - Guaranteed resources for each workload
* **Security** - Complete isolation between different workloads
* **Fault tolerance** - One workload cannot affect another's performance

## Why MIG matters for your workloads

Many GPU-intensive tasks do not always require a high-end GPU's full performance and resources. This underutilization of GPU resources leads to inefficiencies, increased costs, and unpredictable performance. MIG directly addresses these challenges by enabling more efficient GPU sharing.

Consider these real-world scenarios where MIG provides significant value:

**Inference workloads at scale**: Multiple workloads share a single GPU with guaranteed performance isolation. For example, an AI startup comprising three specialized teams can customize GPU allocations to precisely match each team's requirements. The image recognition team might be assigned two 1g.5gb slices, the natural language processing team could use two 2g.10gb slices, while the video analytics team benefits from three 3g.20gb slices — all operating concurrently on a singular GPU.

**Development and testing environments**: In scenarios where developers and data scientists are prototyping, testing, or debugging models, they might not need continuous GPU access. MIG allows teams to provide isolated GPU resources without dedicating entire GPUs to intermittent workloads.

**Multi-tenant platforms**: Each GPU instance is isolated from the others, so one workload cannot affect the performance or stability of another. This isolation is crucial for running mixed workloads with different priorities and resource requirements. Service providers can offer GPU resources to multiple customers while ensuring complete isolation and predictable performance.

**Cost optimization through better density**: The combination of A100 MIG, Amazon EKS, and the P4d instance can speed up processing 2.5x compared to processing them without MIG enabled on the same instance. This improvement in throughput directly translates to cost savings and better resource utilization.

# Supported configurations

| Provider | MIG support | Compatible GPU types      | Notes                     |
| -------- | ----------- | ------------------------- | ------------------------- |
| GCP GKE  | ✓           | All MIG-capable GPU types | -                         |
| AWS EKS  | ✓           | All MIG-capable GPU types | Requires Bottlerocket AMI |

## MIG-compatible GPUs

MIG is supported on NVIDIA GPUs starting with the Ampere architecture (compute capability >= 8.0), including:

* **Ampere**: A100 (40GB/80GB), A30 (24GB)
* **Hopper**: H100 (multiple variants), H200 (141GB)
* **Blackwell**: B200 (180GB), GB200 (186GB)
* **Professional**: RTX PRO 6000 Blackwell series

For an exhaustive list of MIG-compatible GPUs, refer to relevant <Anchor label="NVIDIA documentation" target="_blank" href="https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html#supported-gpus">NVIDIA documentation</Anchor>.

**Cloud provider GPU documentation:**

* **GKE**: <Anchor label="GPU machine types and availability" target="_blank" href="https://cloud.google.com/compute/docs/gpus">GPU machine types and availability</Anchor>
* **EKS**: <Anchor label="GPU instance types" target="_blank" href="https://aws.amazon.com/ec2/instance-types/#accelerated-computing">GPU instance types</Anchor>

## Supported MIG partition sizes

Atlas supports NVIDIA MIG profiles using the single strategy, in which all GPU partitions must be the same size.

With the MIG single strategy, you can create MIG devices of the same size. For instance, on a P4d.24XL, the options include creating 56 slices of 1g.5gb, 24 slices of 2g.10gb, 16 slices of 3g.20gb, or a single slice of either 4g.40gb or 7g.40gb.

The exact partition sizes and memory allocations vary by GPU model:

| Partition type | Compute fraction | Memory/cache fraction | Max instances per GPU |
| -------------- | ---------------- | --------------------- | --------------------- |
| 1g partitions  | 1/7              | Varies by model       | 7                     |
| 2g partitions  | 2/7              | Varies by model       | 3                     |
| 3g partitions  | 3/7              | Varies by model       | 2                     |
| 4g partitions  | 4/7              | Varies by model       | 1                     |
| 7g partitions  | 7/7              | Full GPU              | 1                     |

**Supported configurations:**

* ✅ Single partition size per GPU (e.g., **all** `1g.5gb` partitions or **all** `2g.10gb` partitions)
* ❌ <Anchor label="Mixed partition sizes on same GPU" target="_blank" href="https://docs.nvidia.com/datacenter/cloud-native/kubernetes/latest/index.html#the-mixed-strategy">Mixed partition sizes on same GPU</Anchor> (e.g., mixing `1g.5gb` and `2g.10gb`)
* ❌ Media extension profiles (profiles ending with `+me`)
* ❌ Extended profiles with media engines (`+me`, `+me.all`)
* ❌ Graphics-enabled profiles (`+gfx`)
* ❌ Media-excluded profiles (`-me`)

For complete details on MIG profiles, see the <Anchor label="NVIDIA MIG User Guide" target="_blank" href="https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html#supported-mig-profiles">NVIDIA MIG User Guide</Anchor>.

> 📘 Note
>
> Some MIG profiles require specific NVIDIA driver versions.

# How MIG works with Atlas

Atlas simplifies MIG deployment by automating the entire provisioning and configuration process:

1. **Automatic detection** - Atlas detects MIG-configured workloads through node selectors:

   ```yaml
   # GKE nodeSelector
   nodeSelector:
     cloud.google.com/gke-gpu-partition-size: 1g.5gb

   # NVIDIA standard nodeSelector
   nodeSelector:
     nvidia.com/gpu.mig-partition-1g.5gb: "true"
   ```

2. **Automated provisioning** - When the autoscaler detects a pod requesting a MIG partition, it:
   * Provisions a MIG-capable GPU node
   * Automatically configures the GPU partitioning at provision time
   * Applies appropriate labels and taints

3. **Partition configuration** - The partitioning happens automatically during node provisioning:

4. **Resource calculation** - GPU capacity is calculated based on the requested partition size. If a GPU is partitioned into 7 isolated MIG partitions, those are treated as separate GPUs in autoscaling decisions. For example, a workload with the following configuration would be given 1/7 partitions in such a scenario:
   ```yaml
   resources:
     limits:
       nvidia.com/gpu: 1
   ```

5. **Seamless scaling** - Additional MIG-enabled nodes are created when partitions are exhausted. If you have a deployment with 7 replicas running on a single A100 node that is partitioned into 7 MIG partitions, and you scale that deployment to 8 replicas, the autoscaler will provision a new node to accommodate it.

## Partition size configuration

The partition size is specified differently depending on your cloud provider, which affects how you configure your workloads:

### GKE

Partition size goes in the node selector **value**:

```yaml
nodeSelector:
  cloud.google.com/gke-gpu-partition-size: 1g.5gb  # Size specified as the value
```

This means when switching between partition sizes on GKE, you only change the value:

* For 1/7 GPU: `cloud.google.com/gke-gpu-partition-size: 1g.5gb`
* For 2/7 GPU: `cloud.google.com/gke-gpu-partition-size: 2g.10gb`
* For 3/7 GPU: `cloud.google.com/gke-gpu-partition-size: 3g.20gb`

### EKS

Partition size is part of the node selector **key**:

```yaml
nodeSelector:
  nvidia.com/gpu.mig-partition-1g.5gb: "true"  # Size embedded in the key name
```

This means when switching between partition sizes on EKS, you must change the entire key:

* For 1/7 GPU: `nvidia.com/gpu.mig-partition-1g.5gb: "true"`
* For 2/7 GPU: `nvidia.com/gpu.mig-partition-2g.10gb: "true"`
* For 3/7 GPU: `nvidia.com/gpu.mig-partition-3g.20gb: "true"`

<Callout icon="💡" theme="default">
  ### Tip

  When designing workloads, consider this difference if you need portability between GKE and EKS. You may want to use the NVIDIA standard nodeSelector, which is cloud-agnostic, or abstract these provider-specific configurations entirely.
</Callout>

# Configuring MIG workloads

## Basic MIG configuration

To run workloads on MIG partitions, configure your pods with the appropriate node selector and toleration:

```yaml GKE
nodeSelector:
  cloud.google.com/gke-gpu-partition-size: 1g.5gb
tolerations:
  - key: "nvidia.com/gpu.mig"
    operator: "Exists"
    effect: "NoSchedule"
```
```yaml EKS
nodeSelector:
  nvidia.com/gpu.mig-partition-1g.5gb: "true"
tolerations:
  - key: "nvidia.com/gpu.mig"
    operator: "Exists"
    effect: "NoSchedule"
```

Refer to <Anchor label="NVIDIA documentation" target="_blank" href="https://docs.nvidia.com/datacenter/tesla/mig-user-guide/index.html#supported-mig-profiles">NVIDIA documentation</Anchor> for all supported partition sizes.

## GKE configuration

```yaml GKE configuration
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mig-workload
spec:
  replicas: 7
  selector:
    matchLabels:
      app: mig-workload
  template:
    metadata:
      labels:
        app: mig-workload
    spec:
      nodeSelector:
        cloud.google.com/gke-gpu-partition-size: 1g.5gb # Uses GKE-specific nodeSelector
      tolerations:
        - key: "nvidia.com/gpu.mig"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
      - name: gpu-workload
        image: your-gpu-image
        resources:
          limits:
            nvidia.com/gpu: 1
```
```yaml Example for testing
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cuda-simple-app
spec:
  replicas: 7 # Scale up to trigger new node provisioning
  selector:
    matchLabels:
      app: cuda-simple-app
  template:
    metadata:
      labels:
        app: cuda-simple-app
    spec:
      nodeSelector:
        cloud.google.com/gke-gpu-partition-size: 1g.5gb
      tolerations:
        - key: "nvidia.com/gpu.mig"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
      - name: cuda-simple
        image: nvidia/cuda:11.0.3-base-ubi7
        command:
        - bash
        - -c
        - |
          /usr/local/nvidia/bin/nvidia-smi -L
          /usr/local/nvidia/bin/nvidia-smi -q -d MEMORY | head -20
          sleep 300
        resources:
          limits:
            nvidia.com/gpu: 1
```

## Alternative GKE node selector

You can also use the NVIDIA-standard node selector:

```yaml
nodeSelector:
  nvidia.com/gpu.mig-partition-1g.5gb: "true"
tolerations:
  - key: "nvidia.com/gpu.mig"
    operator: "Exists"
    effect: "NoSchedule"
```

## EKS configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mig-workload-eks
spec:
  replicas: 7
  selector:
    matchLabels:
      app: mig-workload-eks
  template:
    metadata:
      labels:
        app: mig-workload-eks
    spec:
      nodeSelector:
        nvidia.com/gpu.mig-partition-1g.5gb: "true" # Uses NVIDIA standard nodeSelector
      tolerations:
        - key: "nvidia.com/gpu.mig"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
      - name: gpu-workload
        image: your-gpu-image
        resources:
          limits:
            nvidia.com/gpu: 1
```

## Optional GPU type selection

You can target specific GPU models for your MIG workloads:

### GKE-specific GPU selection

```yaml
nodeSelector:
  cloud.google.com/gke-gpu-partition-size: 1g.5gb
  cloud.google.com/gke-accelerator: "nvidia-tesla-a100"
```

### Provider-agnostic GPU selection

```yaml
nodeSelector:
  nvidia.com/gpu.mig-partition-1g.5gb: "true"
  nvidia.com/gpu.name: "nvidia-tesla-a100"
```

# MIG with GPU time-slicing

Atlas supports combining MIG with GPU time-slicing for maximum resource utilization. This powerful combination allows multiple workloads to share each MIG partition through time-based scheduling, dramatically increasing the number of workloads you can run per GPU.

## Why combine MIG with time-slicing?

While MIG provides hardware-level isolation by partitioning a GPU into up to 7 instances, time-slicing multiplies this capacity by allowing multiple workloads to share each partition.

**Development and testing environments**: Development teams often need GPU access for short bursts—compiling CUDA code, running quick experiments, or debugging ML models. With MIG + time-slicing, a single A100 GPU can support an entire development team. For example, with 7 MIG partitions and 4x time-slicing, 28 developers can have GPU access simultaneously.

**Batch inference workloads**: Many inference workloads don't fully utilize even a small MIG partition. By combining MIG with time-slicing, you can run many lightweight inference services on a single GPU. This is useful for serving multiple model versions or A/B testing scenarios.

**Educational and training environments**: Academic institutions can provide GPU access to large numbers of students without investing in massive GPU clusters. A single node with 8 A100 GPUs can support over 200 concurrent users with an appropriate MIG and time-slicing configuration.

## Configuring MIG with time-slicing

1. **Create a node template** with both GPU-enabled instances and <Anchor label="time-slicing" target="_blank" href="doc:gpu-sharing-time-slicing">time-slicing</Anchor> enabled
2. **Configure the sharing multiplier** (e.g., 4 shared clients per GPU)
3. **Deploy workloads** targeting the template with MIG selectors

Example configuration for 28 pods on a single A100 (7 MIG partitions × 4 time-slices):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mig-timeslicing-workload
spec:
  replicas: 28
  selector:
    matchLabels:
      app: mig-timeslicing-workload
  template:
    metadata:
      labels:
        app: mig-timeslicing-workload
    spec:
      nodeSelector:
        scheduling.atlas.example/node-template: "your-mig-timeslicing-template"
        cloud.google.com/gke-gpu-partition-size: 1g.5gb
      tolerations:
        - key: "nvidia.com/gpu.mig"
          operator: "Exists"
          effect: "NoSchedule"
        - key: "scheduling.atlas.example/node-template"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
      - name: gpu-workload
        image: your-gpu-image
        resources:
          limits:
            nvidia.com/gpu: 1
```

## Resource calculation with MIG and time-slicing

When combining MIG and time-slicing, Atlas calculates total GPU resources as:

**Total GPU resources = MIG partitions × Time-slicing multiplier**

Examples:

* A100 with 1g.5gb partitions and 4× time-slicing: `7 × 4 = 28` GPU resources
* A100 with 2g.10gb partitions and 2× time-slicing: `3 × 2 = 6` GPU resources

# Node labels and taints

Atlas automatically applies the following labels and taints to MIG-enabled nodes:

## Labels

| Label                                     | Example value | Description                          |
| ----------------------------------------- | ------------- | ------------------------------------ |
| `nvidia.com/gpu.mig`                      | `true`        | Node has MIG enabled                 |
| `nvidia.com/gpu.mig-partition-*`          | `true`        | Available MIG partition sizes        |
| `cloud.google.com/gke-gpu-partition-size` | `1g.5gb`      | GKE-specific partition size          |
| `scheduling.atlas.example/gpu-shared`           | `4`           | Time-slicing multiplier (if enabled) |

## Taints

| Taint                              | Effect       | Description                                |
| ---------------------------------- | ------------ | ------------------------------------------ |
| `nvidia.com/gpu.mig=true`          | `NoSchedule` | Prevents non-MIG workloads from scheduling |
| `scheduling.atlas.example/node-template` | `NoSchedule` | Applied to custom node templates           |

# Troubleshooting

## Common issues

### Pods stuck in Pending state

**Cause**: Missing or incorrect MIG toleration

**Solution**: Ensure your pods include the MIG toleration:

```yaml
tolerations:
  - key: "nvidia.com/gpu.mig"
    operator: "Exists"
    effect: "NoSchedule"
```

### No nodes provisioned

**Cause**: Node template not recognized or unavailable

**Solution**:

1. Verify the node template name matches exactly
2. Check that the template is enabled in the Atlas console
3. Ensure the template includes GPU-enabled instances

### MIG-capable GPUs not available

**Cause**: No MIG-capable GPU instances available in your region/zone

**Solution**: Verify that MIG-capable GPUs are available in your cluster's region:

* **GKE**: <Anchor label="GPU regions and zones" target="_blank" href="https://cloud.google.com/compute/docs/gpus/gpu-regions-zones">GPU regions and zones</Anchor>
* **EKS**: <Anchor label="GPU instance availability by region" target="_blank" href="https://aws.amazon.com/ec2/instance-types/">GPU instance availability by region</Anchor>

Consider moving your cluster to a region with MIG-capable GPU availability, or use a different GPU instance type that supports MIG.

## Verification commands

Check the MIG configuration on the nodes.

List MIG-enabled nodes:

```shell
# EKS
kubectl get nodes -l nvidia.com/gpu.mig=true

# GKE
kubectl get nodes -l cloud.google.com/gke-gpu-partition-size
```

Check available partitions:

```shell
kubectl describe node <node-name> | grep nvidia.com/gpu
```