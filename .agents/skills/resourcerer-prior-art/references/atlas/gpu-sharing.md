---
updatedAt: 2026-05-26T10:42:39.000Z
---

# GPU sharing

GPU sharing allows multiple workloads to utilize GPU resources more efficiently by enabling GPUs to be shared among different processes or workloads.
Atlas supports three methods for GPU sharing, each optimized for different use cases and requirements.

# GPU sharing methods

## Supported methods

| Method                                             | Cloud provider   | Limitations                                    |
| -------------------------------------------------- | ---------------- | ---------------------------------------------- |
| [Time-slicing](gpu-sharing-time-slicing.md)       | AWS EKS, GCP GKE | EKS requires Amazon Linux 2023 or Bottlerocket |
| [Multi-Instance GPU (MIG)](gpu-sharing-mig.md)    | AWS EKS, GCP GKE | EKS requires Bottlerocket                      |
| [Fractional GPUs](gpu-sharing-fractional.md)      | AWS EKS          | AWS G6 instances only                          |
| [Multi-Process Service (MPS)](gpu-sharing-mps.md) | GCP GKE          | -                                              |

Atlas also supports GPU sharing through [Dynamic Resource Allocation (DRA)](dra.md).

## Time-slicing

<Anchor target="_blank" href="doc:gpu-sharing-time-slicing">Time-slicing</Anchor> allows multiple workloads to share a single physical GPU through rapid context switching. This approach enables better GPU utilization for workloads that don't continuously require GPU resources.

**Best for:**

* Development and testing environments
* Workloads with intermittent GPU usage
* Cost optimization when workloads don't need dedicated GPU access
* Scenarios where hardware isolation is not required

**Key characteristics:**

* Software-based sharing through context switching
* Memory shared between all processes
* Equal time allocation across workloads
* Simple configuration through node templates

<Anchor target="_blank" href="doc:gpu-sharing-time-slicing">Learn more about time-slicing →</Anchor>

## Multi-Instance GPU (MIG)

<Anchor target="_blank" href="doc:gpu-sharing-mig">MIG</Anchor> partitions powerful GPUs into smaller, hardware-isolated instances. Each MIG instance provides dedicated memory, cache, and compute resources with guaranteed performance.

**Best for:**

* Production workloads requiring guaranteed resources
* Multi-tenant environments needing isolation
* Workloads with consistent GPU requirements
* Scenarios requiring fault tolerance between workloads

**Key characteristics:**

* Hardware-level isolation
* Dedicated resources per instance
* Quality of service guarantees
* Available on select NVIDIA GPUs (Ampere architecture and newer)

[Learn more about MIG →](gpu-sharing-mig.md)

## Multi-Process Service (MPS)

<Anchor target="_blank" href="doc:gpu-sharing-mps">MPS</Anchor> enables multiple CUDA processes to concurrently utilize a single GPU with improved performance for compute-bound workloads.<br />Unlike time-slicing, which uses context switching, MPS allows concurrent GPU kernel execution.

**Best for:**

* Workloads that individually underutilize the GPU (small-batch inference, lightweight models)
* Small LLM inference serving (models under 3B parameters)
* MPI-based high-performance computing (HPC) and scientific simulations with many parallel processes
* Multi-tenant environments where multiple applications share a GPU

**Key characteristics:**

* Concurrent GPU kernel execution (spatial sharing, not time-slicing)
* Client-server architecture with shared GPU scheduling resources
* Binary-compatible — existing CUDA applications work unchanged

[Learn more about MPS →](gpu-sharing-mps.md)

# Combining sharing methods

Time-slicing and MIG can be <Anchor target="_blank" href="doc:gpu-sharing-mig#mig-with-gpu-time-slicing">combined</Anchor> for maximum resource utilization. This powerful combination allows multiple workloads to time-slice each MIG partition, dramatically increasing the number of workloads that can run per physical GPU.

**Example:** A single A100 GPU with 7 MIG partitions and 4× time-slicing can support 28 concurrent workloads (7 × 4 = 28).

# Choosing the right sharing method

| Consideration           | Time-slicing                               | MIG                                                                | MPS                                                                           |
| ----------------------- | ------------------------------------------ | ------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| **Isolation**           | None — memory shared between all processes | Hardware-based — dedicated memory, cache, and compute per instance | Full memory isolation on Volta+ (compute capability ≥ 7.0); none on pre-Volta |
| **Resource guarantees** | Shared, no guarantees                      | Dedicated, guaranteed                                              | Configurable on Volta+ (thread percentage, memory limits); none on pre-Volta  |
| **Max clients per GPU** | Depends on configuration                   | Up to 7 instances (GPU-dependent)                                  | Up to 60 on Volta+; up to 16 on pre-Volta                                     |
| **Setup complexity**    | Simple                                     | Moderate                                                           | Simple                                                                        |
| **GPU requirements**    | Any NVIDIA GPU                             | Ampere architecture or newer                                       | Any NVIDIA GPU (Volta+ recommended for full feature support)                  |
| **Best use case**       | Development, testing, variable workloads   | Production, multi-tenant, consistent workloads                     | Workloads that individually underutilize the GPU                              |

# Getting started

1. Review your workload requirements and choose the appropriate sharing method
2. Configure GPU sharing in your <Anchor target="_blank" href="doc:node-templates">node templates</Anchor>
3. Deploy your workloads with the appropriate node selectors and tolerations
4. Monitor GPU utilization with <Anchor target="_blank" href="doc:gpu-utilization">GPU metrics</Anchor>

For general GPU setup and driver installation, see the <Anchor target="_blank" href="doc:gpu">GPU instances</Anchor> documentation.