---
updatedAt: 2026-05-12T14:01:40.000Z
---

# TPU Instances (GKE)

Configure Atlas Autoscaler to scale your GKE cluster using TPU-optimized instances with Google Cloud TPU v5 support.

# Autoscaling using TPU instances

The Atlas Autoscaler can scale GKE clusters using TPU-optimized instances. This guide describes the steps needed to configure your cluster so that TPU nodes can join it and workloads that request TPU resources are scheduled correctly.

## Supported providers

| Provider | TPU versions supported |
| -------- | ---------------------- |
| GCP GKE  | v5e, v5p               |

> 📘 Note
>
> TPU instances are only supported on GCP GKE. AWS EKS and Azure AKS are not supported.

## How does it work?

Once a pending pod requests `google.com/tpu` resources, Atlas's Autoscaler detects the workload and provisions a TPU node that satisfies the request.

Atlas supports **single-host TPU slices only**. Multi-host (pod slice) topologies are not supported.

To enable provisioning of TPU nodes, you need:

* A pod that requests `google.com/tpu` resources
* A toleration for the `google.com/tpu=true:NoSchedule` taint
* Node selectors that specify the TPU version and topology

Atlas automatically selects the correct TPU instance type based on the node selectors in your pod specification — no device plugins or additional configuration are required. GKE handles TPU driver installation automatically on TPU nodes.

### TPU node labels

Atlas applies the following labels to TPU nodes, which you can use in node selectors or affinities:

| Label                                    | Example value          | Description                             |
| ---------------------------------------- | ---------------------- | --------------------------------------- |
| `cloud.google.com/gke-tpu-accelerator`   | `tpu-v5-lite-podslice` | TPU version and type                    |
| `cloud.google.com/gke-tpu-topology`      | `2x2`                  | Physical chip topology of the TPU slice |
| `cloud.google.com/gke-accelerator-count` | `4`                    | Number of TPU chips on the node         |

### Supported TPU versions and accelerator labels

| TPU version    | GKE machine type prefix | `gke-tpu-accelerator` label value | Topology format   | Chips per host |
| -------------- | ----------------------- | --------------------------------- | ----------------- | -------------- |
| TPU v5e (Lite) | `ct5lp-hightpu-{N}t`    | `tpu-v5-lite-podslice`            | 2D (e.g. `2x2`)   | 4              |
| TPU v5p        | `ct5p-hightpu-{N}t`     | `tpu-v5p-slice`                   | 3D (e.g. `2x2x1`) | 4              |

### Supported single-host topologies

**TPU v5e (2D topologies)**

| Chips | Topology |
| ----- | -------- |
| 1     | `1x1`    |
| 4     | `2x2`    |
| 8     | `2x4`    |
| 16    | `4x4`    |
| 32    | `4x8`    |
| 64    | `8x8`    |
| 128   | `8x16`   |
| 256   | `16x16`  |

**TPU v5p (3D topologies)**

| Chips | Topologies (most balanced first) |
| ----- | -------------------------------- |
| 1     | `1x1x1`                          |
| 4     | `2x2x1`                          |
| 8     | `2x2x2`, `2x4x1`                 |
| 16    | `2x2x4`, `4x2x2`, `4x4x1`        |
| 32    | `2x4x4`, `4x4x2`, `4x8x1`        |
| 64    | `4x4x4`, `8x8x1`                 |
| 128   | `4x4x8`, `8x8x2`, `4x8x4`        |
| 256   | `4x8x8`, `8x8x4`                 |

When multiple topologies are valid for a given chip count, Atlas prefers the most balanced (closest to cube/square) topology.

### Tainting of TPU nodes

TPU nodes will have the `google.com/tpu=true:NoSchedule` taint applied automatically. Your TPU workloads must include a matching toleration; without it, the pod will not be scheduled onto a TPU node, and Atlas will not provision one.

## Workload configuration examples

The following examples show common patterns for configuring TPU workloads.

### TPU v5e — 4 chips (2x2 topology)

```yaml
spec:
  nodeSelector:
    cloud.google.com/gke-tpu-accelerator: tpu-v5-lite-podslice
    cloud.google.com/gke-tpu-topology: 2x2
  tolerations:
    - key: "google.com/tpu"
      operator: Exists
      effect: NoSchedule
  containers:
    - image: my-tpu-image
      name: tpu-workload
      resources:
        requests:
          cpu: 4
          memory: 8Gi
          google.com/tpu: 4
        limits:
          cpu: 4
          memory: 8Gi
          google.com/tpu: 4
```

### TPU v5p — 4 chips (2x2x1 topology)

```yaml
spec:
  nodeSelector:
    cloud.google.com/gke-tpu-accelerator: tpu-v5p-slice
    cloud.google.com/gke-tpu-topology: 2x2x1
  tolerations:
    - key: "google.com/tpu"
      operator: Exists
      effect: NoSchedule
  containers:
    - image: my-tpu-image
      name: tpu-workload
      resources:
        requests:
          cpu: 4
          memory: 8Gi
          google.com/tpu: 4
        limits:
          cpu: 4
          memory: 8Gi
          google.com/tpu: 4
```

### TPU v5e — 8 chips (2x4 topology)

```yaml
spec:
  nodeSelector:
    cloud.google.com/gke-tpu-accelerator: tpu-v5-lite-podslice
    cloud.google.com/gke-tpu-topology: 2x4
  tolerations:
    - key: "google.com/tpu"
      operator: Exists
      effect: NoSchedule
  containers:
    - image: my-tpu-image
      name: tpu-workload
      resources:
        requests:
          cpu: 8
          memory: 16Gi
          google.com/tpu: 8
        limits:
          cpu: 8
          memory: 16Gi
          google.com/tpu: 8
```

## Required pod fields

The following fields in your pod specification are required for Atlas to correctly provision a TPU node:

| Field                                                       | Required | Description                                                                                      |
| ----------------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------ |
| `spec.nodeSelector["cloud.google.com/gke-tpu-accelerator"]` | Yes      | Selects the TPU version. Must match a supported accelerator label value.                         |
| `spec.nodeSelector["cloud.google.com/gke-tpu-topology"]`    | Yes      | Selects the chip topology. Must be a valid single-host topology for the chosen TPU version.      |
| `spec.tolerations` with key `google.com/tpu`                | Yes      | Required to allow scheduling onto TPU nodes, which are tainted `google.com/tpu=true:NoSchedule`. |
| `resources.limits["google.com/tpu"]`                        | Yes      | Number of TPU chips to request. Must match the chip count implied by the topology.               |

> 📘 Note
>
> The `google.com/tpu` resource must be set in both `requests` and `limits`, and both values must be equal.

## No device plugin required

Unlike GPU instances, TPU nodes on GKE do **not** require an additional device plugin. GKE manages TPU drivers and the `google.com/tpu` extended resource automatically on TPU nodes. Atlas does not validate for any TPU device plugin before autoscaling.

## Further reading

* [GKE TPU documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/tpus)
* [Plan your TPU configuration on GKE](https://cloud.google.com/kubernetes-engine/docs/concepts/plan-tpus)