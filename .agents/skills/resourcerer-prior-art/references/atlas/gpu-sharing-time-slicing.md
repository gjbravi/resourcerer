---
updatedAt: 2026-05-25T09:06:08.000Z
---

# Time-slicing

Atlas supports GPU sharing through <Anchor label="time-slicing" target="_blank" href="https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/gpu-sharing.html">time-slicing</Anchor>, which allows multiple workloads to share a single physical GPU by rapidly switching between processes. This feature enables better GPU utilization and cost optimization for workloads that don't require dedicated GPU access.

Monitor your GPU sharing efficiency with <Anchor label="GPU utilization metrics" target="_blank" href="doc:gpu-utilization">GPU utilization metrics</Anchor> once configured.

# What is GPU time-slicing?

GPU time-slicing is achieved through rapid context switching, where:

* Each process gets an equal share of GPU time
* Compute resources are assigned to one process at a time
* GPU memory is shared between all processes
* The NVIDIA device plugin manages the sharing configuration

For more information on GPU time-slicing, see:

* <Anchor label="Nvidia documentation" target="_blank" href="https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/gpu-sharing.html">Nvidia documentation</Anchor>
* <Anchor label="Amazon AWS documentation" target="_blank" href="https://aws.amazon.com/blogs/containers/gpu-sharing-on-amazon-eks-with-nvidia-time-slicing-and-accelerated-ec2-instances/">Amazon AWS documentation</Anchor>

# Supported configurations

| Provider  | GPU time-slicing support | Notes                                                                                                                                                                                                                                                                                            |
| --------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| AWS EKS   | ✓                        | Available with <Anchor label="time-slicing configuration" target="_blank" href="https://aws.amazon.com/blogs/containers/gpu-sharing-on-amazon-eks-with-nvidia-time-slicing-and-accelerated-ec2-instances/">time-slicing configuration</Anchor>. Supported on Amazon Linux 2023 and Bottlerocket. |
| GCP GKE   | ✓                        | -                                                                                                                                                                                                                                                                                                |
| Azure AKS | Not yet supported        | -                                                                                                                                                                                                                                                                                                |

# How it works

1. **Configuration**: Enable GPU sharing in your node template with time-slicing parameters
2. **Resource calculation**: Atlas calculates extended GPU capacity as `GPU_COUNT * SHARED_CLIENTS_PER_GPU`
3. **Node provisioning**: The autoscaler provisions nodes in adherence to the specified sharing configuration
4. **Workload scheduling**: Pods continue to request `nvidia.com/gpu: 1` - no changes to pod specifications required

# Configuring GPU time-slicing

GPU sharing can be configured through multiple methods.

## Console UI

GPU sharing is configured through node templates in the Atlas console:

1. Create or edit a node template
2. Enable **Use GPU time-slicing**
3. Configure sharing parameters:
   * **Default shared clients per GPU**: The default number of workloads that can share each GPU (1-48)
   * **Sharing configuration per GPU type**: Override defaults for specific GPU models

<Image border={false} src="https://files.readme.io/849e1bf6c7e9bd107ea92324e792df2ff797866844ce24ba0e3a928a01676534-image.png" />

The maximum number of shared clients per physical GPU is 48.

## API

Use the <Anchor label="Node Templates API" target="_blank" href="ref:nodetemplatesapi_createnodetemplate">Node Templates API</Anchor> to configure GPU sharing programmatically. Include the `gpu` object in your node template configuration:

```json
{
  "gpu": {
    "enableTimeSharing": true,
    "defaultSharedClientsPerGpu": 10,
    "sharingConfiguration": {
      "nvidia-tesla-t4": {
        "sharedClientsPerGpu": 8
      },
      "nvidia-tesla-a100": {
        "sharedClientsPerGpu": 16
      }
    }
  }
}
```

## Terraform

Configure GPU sharing using the <Anchor label="Atlas Terraform provider" target="_blank" href="https://registry.terraform.io/providers/atlas/atlas/latest/docs/resources/node_template">Atlas Terraform provider</Anchor>. Add the `gpu` block to your node template resource:

```toml Terraform
resource "atlas_node_template" "example" {
  # ... other configuration

  gpu {
    enable_time_sharing             = true
    default_shared_clients_per_gpu  = 10
    
    sharing_configuration = {
      "nvidia-tesla-t4" = {
        shared_clients_per_gpu = 8
      }
      "nvidia-tesla-a100" = {
        shared_clients_per_gpu = 16
      }
    }
  }
} 
```

# Workload configuration

When using GPU time-slicing, pods continue to request GPUs using the standard `nvidia.com/gpu` resource. The key difference is targeting nodes with time-slicing enabled through node selectors and tolerations.

## Basic time-slicing workload

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

## Complete deployment example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-timeslicing-workload
spec:
  replicas: 10  # Can schedule 10 pods on a single GPU with 10x sharing
  selector:
    matchLabels:
      app: gpu-timeslicing-workload
  template:
    metadata:
      labels:
        app: gpu-timeslicing-workload
    spec:
      nodeSelector:
        scheduling.atlas.example/node-template: "gpu-sharing-template"
      tolerations:
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

# Provider-specific considerations

## AWS EKS

### Amazon Linux 2023 (AL2023)

**Amazon Linux 2023 (AL2023) AMIs support GPU time-slicing.** When using AL2023 for GPU nodes with time-slicing enabled, Atlas automatically configures the required NVIDIA device plugin settings.

#### Updating existing NVIDIA device plugin installations

If you have an existing NVIDIA device plugin Helm chart installation and want to use GPU time-slicing with AL2023, you must update your installation to prevent its pods from scheduling on nodes with GPU time-slicing enabled.

Add the following affinity configuration to your Helm values:

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        # On discrete-GPU based systems NFD adds the following label where 10de is the NVIDIA PCI vendor ID
        - key: feature.node.kubernetes.io/pci-10de.present
          operator: In
          values:
          - "true"
        # Atlas condition to exclude nodes with preinstalled NVIDIA device plugin
        - key: "scheduling.atlas.example/preinstalled-nvidia-driver"
          operator: DoesNotExist
      - matchExpressions:
        # On some Tegra-based systems NFD detects the CPU vendor ID as NVIDIA
        - key: feature.node.kubernetes.io/cpu-model.vendor_id
          operator: In
          values:
          - "NVIDIA"
        # Atlas condition to exclude nodes with preinstalled NVIDIA device plugin
        - key: "scheduling.atlas.example/preinstalled-nvidia-driver"
          operator: DoesNotExist
      - matchExpressions:
        # Allow GPU deployment to be forced by setting this label to "true"
        - key: "nvidia.com/gpu.present"
          operator: In
          values:
          - "true"
        # Atlas condition to exclude nodes with preinstalled NVIDIA device plugin
        - key: "scheduling.atlas.example/preinstalled-nvidia-driver"
          operator: DoesNotExist
```

This ensures that the NVIDIA device plugin pods run only on nodes without Atlas's preinstalled device plugin, preventing conflicts with GPU time-slicing configurations.

### Bottlerocket

**Bottlerocket AMIs come with pre-installed NVIDIA drivers.** Do not install additional NVIDIA device plugins on clusters where you plan to use Bottlerocket GPU nodes, as they may conflict with the pre-installed drivers and fail to start. If you do, you need to ensure these NVIDIA device plugin daemonsets do not run on those Bottlerocket nodes through tolerations.

When using Bottlerocket for GPU nodes, the OS handles driver management automatically. For more information, see <Anchor label="Bottlerocket support for NVIDIA GPUs" target="_blank" href="https://aws.amazon.com/blogs/containers/bottlerocket-support-for-nvidia-gpus/">Bottlerocket support for NVIDIA GPUs</Anchor>.

**GPU sharing configuration**: When using GPU time-slicing with Bottlerocket, Atlas automatically configures the required NVIDIA device plugin settings. Do not manually add GPU time-slicing settings to your init script, as these will conflict with Atlas's automatic configuration and prevent nodes from being created successfully.

See <Anchor label="Node Configuration documentation" target="_blank" href="doc:node-configuration#virtual-machine-image-choice">Node Configuration documentation</Anchor> for more details on AMI choice.

## GCP GKE

GPU time-slicing is supported on GKE with the standard Atlas GPU driver configuration. No additional provider-specific configuration is required beyond enabling time-slicing in your node template.

# Combining time-slicing with MIG

Atlas supports combining <Anchor label="Multi-Instance GPU (MIG)" target="_blank" href="doc:gpu-sharing-mig">Multi-Instance GPU (MIG)</Anchor> with time-slicing for maximum resource utilization. This powerful combination allows multiple workloads to share each MIG partition through time-based scheduling.

For detailed information on configuring MIG with time-slicing, see the <Anchor label="MIG with GPU time-slicing" target="_blank" href="doc:gpu-sharing-mig#mig-with-gpu-time-slicing">MIG with GPU time-slicing</Anchor> section in the MIG documentation.

# Monitoring GPU utilization

Once GPU time-slicing is configured, monitor your GPU sharing efficiency with <Anchor label="GPU utilization metrics" target="_blank" href="doc:gpu-utilization">GPU utilization metrics</Anchor>. These metrics help you:

* Track GPU compute utilization across shared workloads
* Identify GPU memory waste
* Analyze cost efficiency of your sharing configuration
* Optimize sharing multipliers based on actual usage patterns