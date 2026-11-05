---
updatedAt: 2026-08-26T10:21:57.000Z
---

# Multi-Process Service (MPS)

Atlas supports GPU sharing through <Anchor target="_blank" href="https://docs.nvidia.com/deploy/mps/latest/index.html">NVIDIA Multi-Process Service (MPS)</Anchor>, which enables multiple CUDA processes to concurrently utilize a single GPU with improved performance for compute-bound workloads.

Monitor your GPU sharing efficiency with <Anchor target="_blank" href="doc:gpu-utilization">GPU utilization metrics</Anchor> once configured.

# What is NVIDIA MPS?

NVIDIA Multi-Process Service (MPS) is an alternative, binary-compatible implementation of the CUDA API that enables multiple CUDA applications to share a single GPU simultaneously.
Unlike [GPU time-slicing,](gpu-sharing-time-slicing.md) which rapidly switches execution between processes, MPS allows truly concurrent execution of GPU kernels from different processes through a client-server architecture, where:

* Multiple processes run on the GPU at the same time
* The MPS server shares a single set of GPU scheduling resources across all clients

For more information on NVIDIA MPS, see <Anchor target="_blank" href="https://docs.nvidia.com/deploy/mps/latest/index.html">NVIDIA MPS documentation</Anchor>.

# Supported configurations

| Provider  | MPS support       | Notes |
| --------- | ----------------- | ----- |
| GCP GKE   | ✓                 | -     |
| AWS EKS   | Not yet supported | -     |
| Azure AKS | Not yet supported | -     |

> 📘 Note
>
> For MPS support on AWS or Azure, use [Dynamic Resource Allocation (DRA)](dra.md). DRA-based MPS is supported on all providers.

# How Atlas provisions MPS nodes

1. **Configuration**: Enable GPU sharing in your node template with MPS strategy and sharing parameters
2. **Resource calculation**: Atlas calculates extended GPU capacity as `GPU_COUNT * SHARED_CLIENTS_PER_GPU`
3. **Node provisioning**: The autoscaler provisions nodes with MPS configured
4. **Workload scheduling**: Pods continue to request `nvidia.com/gpu: 1`. On Volta and newer GPUs (compute capability ≥ 7.0), no changes to pod specifications are required. On pre-Volta GPUs,
   pods must set `hostIPC: true` to communicate with the MPS control daemon

# Configuring GPU MPS

GPU sharing with MPS can be configured through multiple methods.

## API

Use the <Anchor target="_blank" href="ref:nodetemplatesapi_createnodetemplate">Node Templates API</Anchor> to configure GPU sharing programmatically.
Include the `gpu` object with the `sharingStrategy` set to `mps`:

```json
{
  "gpu": {
    "sharingStrategy": "GPU_SHARING_STRATEGY_MPS",
    "defaultSharedClientsPerGpu": 4,
    "sharingConfiguration": {
      "nvidia-tesla-t4": {
        "sharedClientsPerGpu": 4
      },
      "nvidia-tesla-a100": {
        "sharedClientsPerGpu": 8
      }
    }
  }
}
```

## Terraform

Configure GPU sharing using the <Anchor target="_blank" href="https://registry.terraform.io/providers/atlas/atlas/latest/docs/resources/node_template">Atlas Terraform provider</Anchor>.
Add the `gpu` block with `sharing_strategy` set to `mps`:

```toml Terraform
resource "atlas_node_template" "example" {
  # ... other configuration

  gpu {
    sharing_strategy               = "mps"
    default_shared_clients_per_gpu = 4
    
    sharing_configuration = {
      "nvidia-tesla-t4" = {
        shared_clients_per_gpu = 4
      }
      "nvidia-tesla-a100" = {
        shared_clients_per_gpu = 8
      }
    }
  }
} 
```

# Workload configuration

When using GPU MPS, pods continue to request GPUs using the standard `nvidia.com/gpu` resource. The key difference is targeting nodes with MPS enabled through node selectors and tolerations.

## Basic MPS workload

```yaml
spec:
  nodeSelector:
    scheduling.atlas.example/node-template: "gpu-sharing-template"
  tolerations:
    - key: "gpu-sharing-template"
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

## Volta and newer GPUs (compute capability ≥ 7.0)

No additional pod configuration is required beyond standard GPU resource requests.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-mps-workload
spec:
  replicas: 4  # Can schedule 4 pods on a single GPU with 4x sharing
  selector:
    matchLabels:
      app: gpu-mps-workload
  template:
    metadata:
      labels:
        app: gpu-mps-workload
    spec:
      nodeSelector:
        scheduling.atlas.example/node-template: "gpu-mps-template"
      tolerations:
        - key: "scheduling.atlas.example/node-template"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
      - name: gpu-workload
        image: nvidia/samples:nbody
        resources:
          limits:
            nvidia.com/gpu: 1
```

## Pre-Volta GPUs (compute capability < 7.0)

Pre-Volta GPUs require `hostIPC: true` so that containers can communicate with the MPS control daemon through the host's Inter-Process Communication namespace.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-mps-workload
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gpu-mps-workload
  template:
    metadata:
      labels:
        app: gpu-mps-workload
    spec:
      hostIPC: true
      nodeSelector:
        scheduling.atlas.example/node-template: "gpu-mps-template"
      tolerations:
        - key: "scheduling.atlas.example/node-template"
          operator: "Exists"
          effect: "NoSchedule"
      containers:
      - name: gpu-workload
        image: nvidia/samples:nbody
        resources:
          limits:
            nvidia.com/gpu: 1
```

# Node labels and taints

Atlas automatically applies the following labels to MPS-enabled nodes:

| Label                                     | Example value | Description                                        |
| ----------------------------------------- | ------------- | -------------------------------------------------- |
| `scheduling.atlas.example/gpu-shared`           | `4`           | Number of max shared clients per GPU               |
| `scheduling.atlas.example/gpu-sharing-strategy` | `mps`         | Set to indicate MPS sharing strategy is configured |

# Monitoring GPU utilization

Once GPU MPS is configured, monitor your GPU sharing efficiency with <Anchor target="_blank" href="doc:gpu-utilization">GPU utilization metrics</Anchor>.
These metrics help you:

* Track GPU compute utilization across shared workloads
* Identify GPU memory waste
* Analyze cost efficiency of your sharing configuration
* Optimize sharing multipliers based on actual usage patterns

<br />