---
updatedAt: 2026-08-26T10:21:57.000Z
---

# Dynamic Resource Allocation (DRA)

Configure Atlas Autoscaler to use Kubernetes Dynamic Resource Allocation (DRA) for flexible resource management

# Dynamic Resource Allocation (DRA)

Atlas supports <Anchor target="_blank" href="https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/">Dynamic Resource Allocation (DRA)</Anchor>,
a powerful feature that enables flexible resource allocation for hardware accelerators like GPUs.
DRA provides more sophisticated resource management compared to traditional device plugin approaches, with support for resource sharing, fine-grained device selection, and simplified workload configuration.

## What is Dynamic Resource Allocation?

Dynamic Resource Allocation is a Kubernetes feature that enables pods to request and share specialized hardware resources in a flexible manner. DRA works through a driver-based architecture where:

* **DRA device drivers** advertise available hardware resources to the cluster
* **DeviceClasses** categorize and define resource types (like GPUs) that can be allocated
* **ResourceClaims/ResourceClaimTemplates** represent specific resource requests from workloads
* **ResourceSlices** advertise available resources from nodes to the API server

Unlike traditional device plugins that require explicit resource quantities in each container, DRA allows:

* **Flexible device selection** using attribute-based filtering
* **Easy resource sharing** between containers and pods
* **Simplified pod specifications** with centralized resource definitions
* **Better resource visibility** through structured resource advertisement

For more information on DRA architecture, see the <Anchor target="_blank" href="https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/">Kubernetes DRA documentation</Anchor>.

## Supported providers and drivers

| Provider  | DRA Support | Supported Drivers | Notes                                                      |
| --------- | ----------- | ----------------- | ---------------------------------------------------------- |
| AWS EKS   | ✓           | NVIDIA GPU        | Kubernetes 1.34+ required. Bottlerocket not yet supported. |
| GCP GKE   | ✓           | NVIDIA GPU        | Kubernetes 1.34+ required.                                 |
| Azure AKS | ✓           | NVIDIA GPU        | Kubernetes 1.34+ required.                                 |

## Prerequisites for NVIDIA GPU DRA

Before using DRA with Atlas, ensure you have:

**Required versions:**

* Kubernetes version: 1.34 or higher
* Atlas agent: v0.109.0 or higher (helm chart 0.132.0+)
* Cluster: GKE or EKS cluster connected to Atlas

**Tools required:**

* `kubectl` CLI
* `helm` CLI (for driver installation)

You can verify your Kubernetes version:

```shell
kubectl version --short
```

Expected output should show server version `v1.34.x` or higher.

## Installing NVIDIA DRA driver

DRA requires a driver to be installed in your cluster to manage resource allocation. NVIDIA provides <Anchor target="_blank" href="https://github.com/NVIDIA/k8s-dra-driver-gpu">DRA driver for GPUs</Anchor>.

> ⚠️ Important: The NVIDIA device plugin and NVIDIA DRA kubelet plugin cannot coexist on the same GPU nodes
>
> **NVIDIA device plugin and NVIDIA DRA kubelet plugin cannot run on the same GPU nodes.** If you have a NVIDIA device plugin installed, you must update its configuration to exclude nodes with the `nvidia.com/gpu.dra=true` label before installing the DRA driver.
>
> **Note for GKE users:** This is not required for GKE clusters using the default NVIDIA device plugin. Atlas automatically creates DRA GPU nodes with the `gke-no-default-nvidia-gpu-device-plugin` label, which prevents the default NVIDIA device plugin from running on those nodes.
>
> See the <Anchor target="_blank" href="#updating-existing-nvidia-device-plugin">updating existing NVIDIA device plugin</Anchor> section below for detailed configuration instructions.

### Updating existing NVIDIA device plugin

If you have an existing NVIDIA device plugin installed in your cluster (common in EKS clusters), you must update its affinity configuration to prevent it from scheduling on DRA-enabled GPU nodes.

The NVIDIA device plugin and NVIDIA DRA kubelet plugin cannot coexist on the same GPU nodes. Add the following affinity configuration to your existing NVIDIA device plugin installation to exclude nodes with the `nvidia.com/gpu.dra=true` label:

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            # Discrete GPU label
            - key: feature.node.kubernetes.io/pci-10de.present
              operator: In
              values:
                - "true"
            # Exclude DRA GPU nodes
            - key: nvidia.com/gpu.dra
              operator: NotIn
              values:
                - "true"
        - matchExpressions:
            # Tegra / CPU vendor NVIDIA
            - key: feature.node.kubernetes.io/cpu-model.vendor_id
              operator: In
              values:
                - "NVIDIA"
            # Exclude DRA GPU nodes
            - key: nvidia.com/gpu.dra
              operator: NotIn
              values:
                - "true"
        - matchExpressions:
            # Forced GPU label
            - key: "nvidia.com/gpu.present"
              operator: In
              values:
                - "true"
            # Exclude DRA GPU nodes
            - key: nvidia.com/gpu.dra
              operator: NotIn
              values:
                - "true"
```

### Installing NVIDIA DRA driver on GKE

The following installation process:

* Installs GPU drivers on nodes via DaemonSet
* Creates the `nvidia` namespace for DRA driver components
* Applies ResourceQuota
* Installs NVIDIA DRA driver via Helm (controller + kubelet plugin)

```shell GKE installation script
#!/bin/bash

# Install GPU drivers (GKE documentation: https://cloud.google.com/kubernetes-engine/docs/how-to/gpus#installing_drivers)
kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/container-engine-accelerators/master/nvidia-driver-installer/cos/daemonset-preloaded-latest.yaml

# Install NVIDIA DRA driver
kubectl create namespace nvidia

kubectl apply -n nvidia -f - << EOF
apiVersion: v1
kind: ResourceQuota
metadata:
  name: nvidia
spec:
  hard:
    pods: 100
  scopeSelector:
    matchExpressions:
    - operator: In
      scopeName: PriorityClass
      values:
      - system-node-critical
      - system-cluster-critical
EOF

cat <<EOF > values.yaml
resources.gpus.enabled: "true"
gpuResourcesEnabledOverride: "true"
nvidiaDriverRoot: /home/kubernetes/bin/nvidia/

controller:
  affinity: null

kubeletPlugin:
  tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
  nodeSelector:
    nvidia.com/gpu.present: "true"
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "nvidia.com/gpu.present"
            operator: "Exists"
          - key: nvidia.com/gpu.dra
            operator: In
            values:
            - "true"
EOF

helm upgrade -i nvidia-dra-driver-gpu \
  oci://registry.k8s.io/dra-driver-nvidia/charts/dra-driver-nvidia-gpu \
  --namespace nvidia \
  -f values.yaml
```

### Installing NVIDIA DRA driver on EKS/AKS

The following installation process:

* Creates the `nvidia` namespace for DRA driver components
* Applies ResourceQuota
* Installs NVIDIA DRA driver

> ⚠️ AKS clusters with GRID drivers
>
> If your AKS cluster uses or will use GPU instances with GRID drivers (currently all instances with NVIDIA A10), the NVIDIA DRA driver requires special configuration. See <Anchor target="_blank" href="#nvidia-dra-driver-on-aks-with-grid-drivers">NVIDIA DRA driver on AKS with GRID drivers</Anchor> for details.

```shell EKS/AKS installation script
#!/bin/bash

kubectl create ns nvidia

kubectl apply -n nvidia -f - << EOF
apiVersion: v1
kind: ResourceQuota
metadata:
  name: nvidia
spec:
  hard:
    pods: 100
  scopeSelector:
    matchExpressions:
    - operator: In
      scopeName: PriorityClass
      values:
      - system-node-critical
      - system-cluster-critical
EOF

cat <<EOF > values.yaml
resources.gpus.enabled: "true"
gpuResourcesEnabledOverride: "true"

controller:
  affinity: null

kubeletPlugin:
  tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
  nodeSelector:
    nvidia.com/gpu.present: "true"
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "nvidia.com/gpu.present"
            operator: "Exists"
          - key: nvidia.com/gpu.dra
            operator: In
            values:
            - "true"
EOF

helm upgrade -i nvidia-dra-driver-gpu \
  oci://registry.k8s.io/dra-driver-nvidia/charts/dra-driver-nvidia-gpu \
  --namespace nvidia \
  -f values.yaml
```

#### NVIDIA DRA driver on AKS with GRID drivers

GRID drivers on AKS are incompatible with the `ComputeDomainCliques` and `IMEXDaemonsWithDNSNames` features of the NVIDIA DRA driver.
If your cluster uses or will use GPU instances with GRID drivers on AKS (currently all instances with NVIDIA A10), disable both features by adding the following to your `values.yaml`:

```yaml
featureGates:
  ComputeDomainCliques: false
  IMEXDaemonsWithDNSNames: false
```

### Verifying DRA driver installation

After installation, verify the DRA driver is functioning:

**Check DeviceClasses:**

```shell
kubectl get DeviceClasses
```

Expected output:

```
NAME             AGE
gpu.nvidia.com   Xs
```

**Check DRA driver pods:**

```shell
kubectl get pods -n nvidia
```

You should see running pods including:

* `nvidia-dra-driver-gpu-controller-*` (controller pod)
* `nvidia-dra-driver-gpu-kubelet-plugin-*` (kubelet plugin DaemonSet)

**View controller logs (optional):**

```shell
kubectl logs -n nvidia deployment/nvidia-dra-driver-gpu-controller
```

## Configuring workloads with DRA

DRA workloads use ResourceClaims and ResourceClaimTemplates to request access to hardware resources. ResourceClaims and ResourceClaimTemplates reference a DeviceClass and specify requirements.

> 📘 Supported GPU attributes for DRA
>
> Atlas supports the following GPU attributes for use with DRA ResourceClaims, allowing you to specify precise GPU requirements:
>
> * **productName** - Specific GPU model
> * **device count** - Number of GPU devices required
> * **architecture** - GPU architecture (e.g., "Ampere", "Hopper")
> * **brand** - GPU brand classification (e.g., "Tesla", "Nvidia")
> * **type** - GPU type category
> * **memory** - GPU memory capacity
>
> These attributes can be used in ResourceClaim selectors to target specific GPU characteristics for your workloads.
>
> **Note:** Multi-Instance GPU (MIG) is not currently supported with DRA.

### Basic workload configuration

Here's a simple example with one deployment requesting dedicated GPU access:

```yaml Single GPU workload
---
apiVersion: v1
kind: Namespace
metadata:
  name: gpu-workload

---
apiVersion: resource.k8s.io/v1
kind: ResourceClaim
metadata:
  namespace: gpu-workload
  name: single-gpu
spec:
  devices:
    requests:
    - name: gpu
      firstAvailable:
      - name: nvidia-gpu
        deviceClassName: gpu.nvidia.com

---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: gpu-workload
  name: gpu-deployment
  labels:
    app: gpu-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gpu-deployment
  template:
    metadata:
      labels:
        app: gpu-deployment
    spec:
      containers:
      - name: gpu-container
        image: ubuntu:22.04
        command: ["bash", "-c"]
        args: ["nvidia-smi -L; trap 'exit 0' TERM; sleep 9999 & wait"]
        resources:
          claims:
          - name: gpu
      resourceClaims:
      - name: gpu
        resourceClaimName: single-gpu
      tolerations:
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"
```

### GPU sharing across multiple pods

DRA also allows multiple pods to share the same GPU by referencing a common ResourceClaim:

```yaml Cross-pod GPU sharing
---
apiVersion: v1
kind: Namespace
metadata:
  name: multi-pod-shared

---
apiVersion: resource.k8s.io/v1
kind: ResourceClaim
metadata:
  namespace: multi-pod-shared
  name: global-shared-gpu
spec:
  devices:
    requests:
    - name: gpu
      firstAvailable:
      - name: nvidia-gpu
        deviceClassName: gpu.nvidia.com

---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: multi-pod-shared
  name: gpu-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gpu-deployment
  template:
    metadata:
      labels:
        app: gpu-deployment
    spec:
      containers:
      - name: container
        image: ubuntu:22.04
        command: ["bash", "-c"]
        args: ["nvidia-smi -L; trap 'exit 0' TERM; sleep 9999 & wait"]
        resources:
          claims:
          - name: shared-gpu
      resourceClaims:
      - name: shared-gpu
        resourceClaimName: global-shared-gpu
      tolerations:
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: multi-pod-shared
  name: deployment-2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: deployment-2
  template:
    metadata:
      labels:
        app: deployment-2
    spec:
      containers:
      - name: container
        image: ubuntu:22.04
        command: ["bash", "-c"]
        args: ["nvidia-smi -L; trap 'exit 0' TERM; sleep 9999 & wait"]
        resources:
          claims:
          - name: shared-gpu
      resourceClaims:
      - name: shared-gpu
        resourceClaimName: global-shared-gpu
      tolerations:
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"
```

### Advanced GPU selection using CEL expressions

DRA supports using CEL (Common Expression Language) expressions in ResourceClaims and ResourceClaimTemplates to filter GPUs based on multiple attributes simultaneously.
This example demonstrates selecting GPUs by brand, architecture, and memory capacity:

```yaml GPU selection with CEL expression
---
apiVersion: v1
kind: Namespace
metadata:
  name: gpu-cel-example

---
apiVersion: resource.k8s.io/v1
kind: ResourceClaim
metadata:
  namespace: gpu-cel-example
  name: gpu-cel-filter
spec:
  devices:
    requests:
    - name: gpu
      firstAvailable:
      - name: nvidia-gpu
        deviceClassName: gpu.nvidia.com
        selectors:
        # Filter by brand, architecture, and memory using combined CEL expression
        - cel:
            expression: |
              device.attributes['gpu.nvidia.com'].brand == 'Tesla' &&
              device.attributes['gpu.nvidia.com'].architecture == 'Volta' &&
              device.capacity['gpu.nvidia.com'].memory.isGreaterThan(quantity('10Gi'))

---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: gpu-cel-example
  name: gpu-workload
  labels:
    app: gpu-cel-example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gpu-cel-example
  template:
    metadata:
      labels:
        app: gpu-cel-example
    spec:
      containers:
      - name: gpu-workload
        image: ubuntu:22.04
        command: ["bash", "-c"]
        args:
          - |
            echo "GPU allocated using CEL expression filter:"
            nvidia-smi -L
            nvidia-smi --query-gpu=name,architecture,memory.total,uuid --format=csv
            sleep 3600
        resources:
          claims:
          - name: gpu-claim
      resourceClaims:
      - name: gpu-claim
        resourceClaimName: gpu-cel-filter
      tolerations:
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"
```

### GPU sharing with MPS

> ⚠️ Important
>
> When using MPS through DRA, all replicas of the same Deployment or StatefulSet that share the same GPU must run on the same node.

DRA supports GPU sharing using [Multi-Process Service (MPS)](gpu-sharing-mps.md), which allows multiple pods to concurrently execute GPU kernels on a shared GPU.

MPS support requires enabling the `MPSSupport` feature gate in the NVIDIA DRA driver Helm chart. Add the following to your `values.yaml`:

```yaml
featureGates:
  MPSSupport: true
```

The ResourceClaim configures MPS sharing parameters including active thread percentage and pinned device memory limit per client:

```yaml GPU sharing with MPS
---
apiVersion: v1
kind: Namespace
metadata:
  name: gpu-mps-example

---
apiVersion: resource.k8s.io/v1
kind: ResourceClaim
metadata:
  namespace: gpu-mps-example
  name: gpu-mps-shared
spec:
  devices:
    requests:
    - name: mps-gpu
      exactly:
        deviceClassName: gpu.nvidia.com
    config:
    - requests: ["mps-gpu"]
      opaque:
        driver: gpu.nvidia.com
        parameters:
          apiVersion: resource.nvidia.com/v1beta1
          kind: GpuConfig
          sharing:
            strategy: MPS
            mpsConfig:
              defaultActiveThreadPercentage: 33
              defaultPinnedDeviceMemoryLimit: 1Gi

---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: gpu-mps-example
  name: gpu-mps-workload
  labels:
    app: gpu-mps-workload
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
      containers:
      - name: gpu-workload
        image: ubuntu:22.04
        command: ["bash", "-c"]
        args: ["nvidia-smi -L; trap 'exit 0' TERM; sleep 9999 & wait"]
        resources:
          claims:
          - name: mps-gpu
            request: mps-gpu
      resourceClaims:
      - name: mps-gpu
        resourceClaimName: gpu-mps-shared
      tolerations:
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"
```

## Monitoring and verification

### Watching Atlas provision GPU nodes

After deploying DRA workloads, monitor as Atlas detects requirements and provisions nodes:

```shell
# Watch pod status
kubectl get pods -n <namespace> -w

# Watch for new GPU nodes
kubectl get nodes -w
```

### Verifying ResourceSlices

After GPU nodes join the cluster, verify ResourceSlices are created:

```shell
kubectl get ResourceSlices
```

You should see ResourceSlice objects advertising available GPUs from provisioned nodes.

## Troubleshooting

### Pods stuck in Pending state

Check pod events:

```shell
kubectl describe pod <pod-name> -n <namespace>
```

Common issues:

* Atlas is still provisioning GPU nodes (wait a few minutes)
* DRA driver not installed or not functioning
* Insufficient quota for GPU instances in your cloud provider
* DeviceClass not found or misconfigured

### Atlas not provisioning GPU nodes

1. Verify Atlas autoscaler is enabled:

```shell
kubectl get deployment -n atlas-agent
```

2. Check Atlas agent logs:

```shell
kubectl logs -n atlas-agent deployment/atlas-agent
```

3. Ensure your Atlas configuration includes GPU instance types in allowed instance families

### ResourceSlices not appearing

1. Verify GPU nodes for NVIDIA DRA are running:

```shell
kubectl get nodes -l nvidia.com/gpu.dra=true
```

2. Check DRA kubelet plugin is running on GPU nodes:

```shell
kubectl get pods -n nvidia -o wide
```

3. Check kubelet plugin logs:

```shell
kubectl logs -n nvidia <nvidia-dra-driver-gpu-kubelet-plugin-pod>
```

### DeviceClass not found

Verify the DRA driver created the DeviceClass:

```shell
kubectl get DeviceClasses
```

If no DeviceClasses exist, reinstall the DRA driver and check controller logs for errors.

## References

* <Anchor target="_blank" href="https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/">Kubernetes Dynamic Resource Allocation Documentation</Anchor>
* <Anchor target="_blank" href="https://github.com/NVIDIA/k8s-dra-driver">NVIDIA DRA Driver GitHub</Anchor>

<br />