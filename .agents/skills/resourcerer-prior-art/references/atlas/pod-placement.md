---
updatedAt: 2026-10-09T11:52:46.000Z
---

# Pod placement

Learn how to place pods using labels and other Kubernetes scheduling features

# Configure pod placement by topology

This guide describes placing pods on a particular node, zone, region, cloud, etc., using labels and advanced Kubernetes scheduling features.

Kubernetes supports this step with the following methods:

* [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector)
* [node affinity and anti-affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity)
* [topology spread constraints](https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/)
* [pod affinity and anti-affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#types-of-inter-pod-affinity-and-anti-affinity)

All these methods require using specific labels on each Kubernetes node.

## Supported labels

Atlas supports the following labels:

| Label                                                                          | Type          | Description                                                                                                                 | Example(s)                                                                    |
| ------------------------------------------------------------------------------ | ------------- | --------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `kubernetes.io/arch` and `beta.kubernetes.io/arch`                             | well-known    | Node CPU architecture                                                                                                       | `amd64`, `arm64`                                                              |
| `kubernetes.io/os` and `beta.kubernetes.io/os`                                 | well-known    | Node Operating System                                                                                                       | `linux`, `windows`                                                            |
| `kubernetes.io/hostname`                                                       | well-known    | Node Hostname                                                                                                               | `ip-192-168-32-94.eu-central-1.compute.internal`, `testcluster-31qd-gcp-3ead` |
| `topology.kubernetes.io/region` and `failure-domain.beta.kubernetes.io/region` | well-known    | Node region in the CSP                                                                                                      | `eu-central-1`, `europe-central1`                                             |
| `topology.kubernetes.io/zone` and `failure-domain.beta.kubernetes.io/zone`     | well-known    | Node zone of the region in the CSP                                                                                          | `eu-central-1a`,`europe-central1-a`                                           |
| `provisioner.atlas.example/managed-by`                                               | Atlas-specific | Atlas managed node                                                                                                        | `atlas.example`                                                                     |
| `provisioner.atlas.example/node-id`                                                  | Atlas-specific | Atlas node ID                                                                                                             | `816d634e-9fd5-4eed-b13d-9319933c9ef0`                                        |
| `scheduling.atlas.example/spot`                                                      | Atlas-specific | Node resource offering type - spot                                                                                          | `true`                                                                        |
| `scheduling.atlas.example/spot-fallback`                                             | Atlas-specific | A fallback for Spot Instance                                                                                                | `true`                                                                        |
| `topology.atlas.example/subnet-id`                                                   | Atlas-specific | Node subnet ID                                                                                                              | `subnet-006a6d1f18fc5d390`                                                    |
| `scheduling.atlas.example/storage-optimized`                                         | Atlas-specific | Local SSD attached node                                                                                                     | `true`                                                                        |
| `scheduling.atlas.example/compute-optimized`                                         | Atlas-specific | A compute-optimized instance                                                                                                | `true`                                                                        |
| `scheduling.atlas.example/cpu-manufacturer`                                          | Atlas-specific | CPU vendor (`amd`, `intel`, `aws`, `ampere`, `apple`, `google`). Not set when the instance has multiple or unknown vendors. | `amd`, `intel`                                                                |
| `scheduling.atlas.example/instance-cpu`                                              | Atlas-specific | vCPU count of the node (integer string)                                                                                     | `8`, `16`                                                                     |
| `scheduling.atlas.example/instance-family`                                           | Atlas-specific | Instance family identifier                                                                                                  | `c6a`, `m5` (AWS); `n2` (GCP)                                                 |
| `scheduling.atlas.example/instance-generation`                                       | Atlas-specific | Instance generation number. **AWS only.**                                                                                   | `5`, `6`                                                                      |
| `scheduling.atlas.example/instance-hypervisor`                                       | Atlas-specific | Hypervisor type. **AWS only.** Values: `nitro`, `xen`.                                                                      | `nitro`                                                                       |
| `scheduling.atlas.example/instance-memory`                                           | Atlas-specific | Total node memory in MiB (integer string)                                                                                   | `16384`, `65536`                                                              |

> 📘 Automating spot placement with pod mutations
>
> Instead of adding spot-related tolerations and node selectors to each deployment manifest manually, you can use [Pod mutations](pod-mutations-quickstart.md) to inject them automatically at pod creation time. Pod mutations support `only-spot`, `preferred-spot`, and `optional-spot` scheduling modes, and a `distributionPercentage` setting to split a workload's replicas between Spot and On-Demand Instances. See [Spot configuration](pod-mutations-reference.md#spot-configuration) in the Pod mutations reference.

## Scheduling using pod affinities

For pod affinity, Atlas supports the following labels:

| Label                                                                      | Type           | Description                            |
| :------------------------------------------------------------------------- | :------------- | :------------------------------------- |
| `topology.kubernetes.io/zone` and `failure-domain.beta.kubernetes.io/zone` | well-known     | Node zone of the region in the CSP     |
| `topology.gke.io/zone`                                                     | GCP specific   | Availability zone of a persistent disk |
| `topology.disk.csi.azure.com/zone`                                         | Azure specific | Availability zone of a persistent disk |
| `topology.ebs.csi.aws.com/zone`                                            | AWS specific   | Availability zone of a persistent disk |

> 🚧 Important!
>
> Currently, Atlas does not support pod affinity using the `kubernetes.io/hostname` topology key.

### Example

Let's consider an example of three pods (`app=frontend`, `app=backend`, `app=cache`) that should run in the same zone (any zone) but not on the same node.

If you want the pods to run in any zone, you can not just specify the `nodeSelector`, but also create a dependency chain between them.

First, you select one pod that will be the leading one (no affinities) so it can schedule freely. Let's say it's `app=backend`.

Then, you will need to add pod affinities for the pod `app=cache`:

```yaml
podAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchExpressions:
      - key: app
        operator: In
        values:
        - backend
    topologyKey: topology.kubernetes.io/zone
```

and for the `app=frontend`:

```yaml YAML
podAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchExpressions:
      - key: app
        operator: In
        values:
        - backend #to have the ability to run without cache
        - cache
    topologyKey: topology.kubernetes.io/zone
```

With this setup, all three pods will always run in the same zone.

Pod `app=backend` will choose a node and zone, and the other two will follow.

However, all these pods can still be scheduled on the same node. To fix this, you must specify their pod anti-affinity for each other.

Example for `app=frontend`:

```yaml YAML
podAntiAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchExpressions:
      - key: app
        operator: In
        values:
        - backend
        - cache
    topologyKey: kubernetes.io/hostname
```

## Highly-available pod scheduling

You can schedule pods in a highly available way using the topology spread constraints feature. Atlas supports the following topology keys:

* `topology.kubernetes.io/zone` - enables your pods to be spread between availability zones, taking advantage of cloud redundancy.
* `kubernetes.io/hostname` - enables your pods to be spread evenly between available hosts.

> 📘 Note
>
> Atlas will only create nodes in different fault domains when the `whenUnstatisfiable` property equals `DoNotSchedule`. The value `ScheduleAnyway` means that the spread is just a preference, so the autoscaler will keep binpacking those pods, which might result in scheduling them all on the same fault domain.

> 🚧 Important limitations
>
> * When multiple topology spread constraints are specified:
>   * Atlas prioritizes zonal topology keys (like `topology.kubernetes.io/zone`) for node provisioning decisions. If multiple zonal constraints exist, the last one in the list will be used.
>   * Non-zonal topology constraints (like `kubernetes.io/hostname`) influence node provisioning indirectly. Atlas uses the Kubernetes scheduler to validate if pods can fit on prospective nodes, for example, so these constraints may trigger additional node creation if the scheduler cannot place pods.
> * Atlas does not currently support the `minDomains` option for topology spread constraints.

For how the scheduler counts spread and how the autoscaler chooses zones for new nodes, see [Topology spread constraints and node provisioning](autoscaler.md#topology-spread-constraints-and-node-provisioning).

The deployment below will be spread and scheduled across all availability zones supported by your cluster:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: az-topology-spread
  name: az-topology-spread
spec:
  replicas: 30
  selector:
    matchLabels:
      app: az-topology-spread
  template:
    metadata:
      labels:
        app: az-topology-spread
    spec:
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchExpressions:
              - key: app
                operator: In
                values:
                  - az-topology-spread
      containers:
        - image: nginx
          name: nginx
```

## Scheduling on nodes with locally attached SSD

Storage-optimized nodes have local NVMe/SSD disks that provide higher throughput and lower latency than network-attached storage. Pods can request a storage-optimized node by adding a `nodeSelector` and toleration for the `scheduling.atlas.example/storage-optimized` label:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demopod
spec:
  nodeSelector:
    scheduling.atlas.example/storage-optimized: "true"
  tolerations:
    - key: scheduling.atlas.example/storage-optimized
      operator: Exists
      effect: NoSchedule
  containers:
    - name: app
      image: nginx
      resources:
        requests:
          ephemeral-storage: "2Gi"
      volumeMounts:
        - name: ephemeral
          mountPath: "/tmp"
  volumes:
    - name: ephemeral
      emptyDir: {}
```

For a detailed guide covering how storage optimization works, cloud provider differences, node template constraints, and important considerations, see [Storage-optimized nodes](storage-optimized-nodes.md).

## Scheduling on compute-optimized nodes

Compute-optimized instances are ideal for compute-bound applications that benefit from high-performance processors. They offer the highest consistent performance per core to support real-time application performance.

This pod will be scheduled on a compute-optimized instance:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demopod
spec:
  nodeSelector:
    scheduling.atlas.example/compute-optimized: "true"
  containers:
  - name: app
    image: nginx
```

## Scheduling on ARM nodes

ARM processors are designed to deliver the best price-performance for your workloads.

This pod will be scheduled on an ARM instance:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo-arm-pod
spec:
  nodeSelector:
    kubernetes.io/arch: "arm64"
  containers:
  - name: app
    image: nginx
```

When utilizing a multi-architecture Node template, also use the `nodeSelector` `kubernetes.io/arch: "arm64"` to ensure that the pod lands on an ARM node.

> 📘 Note
>
> Atlas added ARM nodes do not have any taint applied by default.

## Scheduling on Windows nodes

> 📘 Note
>
> Autoscaling using Windows nodes is supported only for AKS clusters and has to be first enabled for the organization. Contact the support team for more details.

When a workload has the `nodeSelector` `kubernetes.io/os: windows`, Atlas selects a default Windows Server version based on the cluster's Kubernetes version:

| Kubernetes version | Default Windows Server version |
| ------------------ | ------------------------------ |
| < 1.33             | Windows Server 2019            |
| 1.33–1.36          | Windows Server 2022            |
| 1.37+              | Windows Server 2025            |

To target a specific Windows version regardless of the cluster's Kubernetes version, provide the additional `nodeSelector` `node.kubernetes.io/windows-build`.

To schedule a pod on a Windows node with a specific version, use the following `nodeSelectors`:

```yaml Windows Server 2025
nodeSelector:
  kubernetes.io/os: windows
  node.kubernetes.io/windows-build: '10.0.26100'
```
```yaml Windows Server 2022
nodeSelector:
  kubernetes.io/os: windows
  node.kubernetes.io/windows-build: '10.0.20348'
```
```yaml Windows Server 2019
nodeSelector:
  kubernetes.io/os: windows
  node.kubernetes.io/windows-build: '10.0.17763'
```

> 📘 Note
>
> Windows Server 2025 is supported on Kubernetes 1.32 and later.

## Scheduling by hardware characteristics

Atlas applies a set of hardware labels to every node it provisions. You can use these labels in a workload's `nodeSelector` or `nodeAffinity` to steer that workload onto nodes with the right hardware — without creating a custom node template just for it. The default node template handles provisioning; Atlas picks an instance that satisfies the workload's label constraints.

| Label                                    | What it selects                      | Values                                             | Supported operators                  |
| ---------------------------------------- | ------------------------------------ | -------------------------------------------------- | ------------------------------------ |
| `scheduling.atlas.example/cpu-manufacturer`    | CPU vendor                           | `amd`, `intel`, `aws`, `ampere`, `apple`, `google` | `=` (nodeSelector), `In`, `NotIn`    |
| `scheduling.atlas.example/instance-cpu`        | vCPU count (integer string)          | e.g. `8`, `16`                                     | `=` (nodeSelector), `In`, `Gt`, `Lt` |
| `scheduling.atlas.example/instance-family`     | Instance family identifier           | e.g. `c6a`, `m5` (AWS); `n2` (GCP)                 | `=` (nodeSelector), `In`, `NotIn`    |
| `scheduling.atlas.example/instance-generation` | Instance generation — **AWS only**   | e.g. `5`, `6`                                      | `=` (nodeSelector), `In`, `NotIn`    |
| `scheduling.atlas.example/instance-hypervisor` | Hypervisor type — **AWS only**       | `nitro`, `xen`                                     | `=` (nodeSelector), `In`, `NotIn`    |
| `scheduling.atlas.example/instance-memory`     | Total memory in MiB (integer string) | e.g. `16384`, `65536`                              | `=` (nodeSelector), `In`, `Gt`, `Lt` |

### Exact-match: nodeSelector

Use `nodeSelector` when you need an exact value — a specific CPU manufacturer, instance family, or hypervisor. The `cpu-manufacturer` label is only present on nodes where the instance type has a single, known vendor; it is not set on multi-vendor or unknown-vendor types.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: amd-workload
spec:
  template:
    spec:
      nodeSelector:
        scheduling.atlas.example/cpu-manufacturer: amd
        scheduling.atlas.example/instance-family: c6a
      containers:
        - name: app
          image: my-app:latest
```

### Exclusion: nodeAffinity with NotIn

Use `NotIn` to keep a workload off a set of values, for example, specific instance families. Atlas never provisions an excluded instance for that workload: it picks a matching node from the remaining families, and the autoscaling event log shows which one it selected.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-avoiding-c5-m5-r5
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: scheduling.atlas.example/instance-family
                    operator: NotIn
                    values:
                      - c5
                      - m5
                      - r5
      containers:
        - name: app
          image: my-app:latest
```

This workload never runs on a `c5`, `m5`, or `r5` node. Atlas provisions a node from a family outside that list, such as `m6a`, and schedules the workload on it.

### Threshold match: nodeAffinity with Gt/Lt

For CPU count and memory, use `nodeAffinity` with `Gt` or `Lt` when you need a range rather than an exact value. Both operators are **strict** (not inclusive): `Gt: "15"` matches nodes with 16 or more vCPUs, not 15. To include the boundary value, subtract one for `Gt` or add one for `Lt`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-intensive-workload
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: scheduling.atlas.example/instance-memory
                    operator: Gt
                    values:
                      - "65536"   # nodes with more than 64 GiB RAM (strictly greater than)
                  - key: scheduling.atlas.example/instance-cpu
                    operator: Gt
                    values:
                      - "15"      # nodes with 16+ vCPUs (strictly greater than 15)
      containers:
        - name: app
          image: my-app:latest
```

> 📘 Use resource requests for most sizing needs
>
> For general CPU and memory sizing, set accurate `requests` and `limits` on your containers and let Atlas pick the right instance size automatically. Reserve hardware label constraints for cases where a specific CPU vendor, instance family, or hypervisor is a genuine requirement.

> 📘 When to use a custom node template instead
>
> Hardware label constraints on a workload scope the constraint to that workload only. If you need the same hardware restriction applied to an entire pool of nodes — or want to combine hardware constraints with spot configuration, capacity reservations, or custom taints — create a [custom node template](node-templates.md) instead.

> 🚧 Native cloud and Karpenter selectors are not supported
>
> Atlas's scheduling simulation only recognises its own `scheduling.atlas.example/*` labels for hardware constraints. Provider-native labels such as GKE's `cloud.google.com/machine-family` or Karpenter's `karpenter.k8s.aws/instance-family` are not evaluated by the autoscaler and will not influence which node type Atlas provisions. If your workloads use these labels, replace them with the equivalent `scheduling.atlas.example/*` labels to get the intended placement behaviour.

## How to isolate specific workloads

It's best practice to set workload requests and limits identically and distribute various workloads among all the nodes in the cluster so that the Law of Averages can provide the best performance and availability.

However, in some edge cases, it may be better to isolate volatile workloads to their nodes and avoid mixing them with other workloads in the same clusters. That's where you can use `affinity.podAntiAffinity`:

```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - topologyKey: kubernetes.io/hostname
      labelSelector:
        matchExpressions:
        - key: <any POD label>
          operator: DoesNotExist
```

Pod example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: worklow-jobs
  labels:
    app: workflows
spec:
  replicas: 2
  selector:
    matchLabels:
      app: workflows
  template:
    metadata:
      labels:
        app: workflows
        no-requests-workflows: "true"
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - topologyKey: kubernetes.io/hostname
            labelSelector:
              matchExpressions:
              - key: no-requests-workflows
                operator: DoesNotExist
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 300m  
```

# Troubleshooting

## Topology spread constraints not working as expected

Pods might not get scheduled as intended due to conflicting or incorrectly configured topology spread constraints.
Here's what you can do to troubleshoot this issue:

1. **Review constraints**: Ensure your topology spread constraints are configured correctly in the Pod specs. Verify the `maxSkew`, `topologyKey`, and `whenUnsatisfiable` parameters.
2. **Verify node labels**: ensure your nodes have accurate and required labels. If `topologyKey` references node labels, these must be consistent across all nodes.
3. **Check the scheduler logs**: inspect the Kubernetes scheduler logs for any issues or warnings related to Pod scheduling. Use `kubectl describe pod <pod-name>` to see events and errors.
4. **Inspect resource utilization**: ensure your nodes have enough resources to accommodate the Pods following spread constraints. Insufficient resources or tight constraints might lead to scheduling issues.

## Event: Cannot satisfy Zone Topology Spread

`topology.kubernetes.io/zone` is a supported topology spread constraint key in the Node Placer, allowing you to deploy Pods in a highly available way to take advantage of the cloud zone redundancy. The autoscaler adds a `Cannot Satisfy Zone Topology Spread` event to pods when it cannot scale up because new nodes would have to go in zones the cluster isn't configured for. This is the same mechanism as the `FailedUpscaling` case; for the zone-counting rules behind it, how the autoscaler chooses zones, and the full event table, see [Topology spread constraints and node provisioning](autoscaler.md#topology-spread-constraints-and-node-provisioning) and [FailedUpscaling: topology spread blocks node creation](autoscaler-troubleshooting.md#failedupscaling-topology-spread-blocks-node-creation).