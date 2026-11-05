---
updatedAt: 2026-10-09T11:52:46.000Z
---

# Troubleshooting node autoscaling

Solutions for resolving node provisioning, pod scheduling, and cloud provider-specific issues with the Atlas autoscaler.

## EKS: Max pod count on AWS CNI

There are situations when you can get VMs in AWS which will have a low upper limit of max pod count, i.e. `58` for `c6g.2xlarge`. Complete list of ENI limitations per instance type available at [eni-max-pods.txt](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/misc/eni-max-pods.txt).

This can be mitigated in two ways:

1. Setting Min CPU constraints in Node Templates to 16 CPUs, as the issue only exists on nodes that have fewer CPUs (i.e., 8 CPU nodes)
2. Increasing the pods per node limits, you can do it by executing the following within your cluster context:

* `kubectl set env daemonset aws-node -n kube-system ENABLE_PREFIX_DELEGATION=true`
* [Amazon VPC CNI plugin increases pods per node limits](https://aws.amazon.com/blogs/containers/amazon-vpc-cni-increases-pods-per-node-limits/)
* [Increase the amount of available IP addresses for your Amazon EC2 nodes](https://docs.aws.amazon.com/eks/latest/userguide/cni-increase-ip-addresses.html)

***

## EKS: Missing instance profile IAM role from `aws-auth` ConfigMap

For clusters that utilize `ConfigMap` access mode, nodes require an entry in `aws-auth` ConfigMap with the IAM instance profile role to properly access the cluster. If nodes remain unhealthy because `kubelet` cannot access the `api-server`, check if the role is present.

Sharing the instance profile and role for Atlas-managed nodes with EKS-managed node groups is not recommended. This is because deleting all managed node groups that use an instance profile removes the instance profile role from `aws-auth`, which can break Atlas-managed nodes that utilize the role. The observed symptom is nodes becoming `NotReady` shortly after, with `kubelet` receiving unauthorized errors when accessing the `api-server`.

To resolve the problem after it appears:

* For clusters utilizing `ConfigMap` access mode, add the role back in the `aws-auth`.
* For clusters utilizing `EKS API and ConfigMap` access mode, adding an access entry either in the EKS access entry list or in the `aws-auth` ConfigMap is sufficient.
* Clusters utilizing only `EKS API`, access mode should not be affected as EKS does not delete entries in EKS API.

Sample `aws-auth` entry:

```yaml
- "groups":
  - "system:bootstrappers"
  - "system:nodes"
  "rolearn": "arn:aws:iam::account-id:role/instance-profile-role"
  "username": "system:node:{{EC2PrivateDNSName}}"
```

References:

* [Delete a managed node group from your cluster](https://docs.aws.amazon.com/eks/latest/userguide/delete-managed-node-group.html)

***

## EKS: Warm pool nodes are not supported

Atlas does not support deleting nodes that were provisioned from AWS autoscaling warm pools. Attempting to delete warm pool nodes through the Atlas console will result in a deletion failure with an error message.

**What are warm pools?** AWS autoscaling warm pools are pre-initialized EC2 instances that remain in a stopped or hibernated state to reduce scale-out latency when scaling your cluster.

**Resolution:** Warm pool nodes must be managed directly through the AWS console or CLI rather than through Atlas. If you need to remove warm pool nodes, use AWS tools to manage the autoscaling group configuration.

**Related:** If you're using warm pools and experiencing node management issues, consider using standard EKS managed node groups or self-managed node groups without warm pools for nodes that Atlas will manage.

***

## EKS: Bottlerocket nodes take 5+ minutes to reach Ready state

Bottlerocket nodes on EKS may consistently take 5-6 minutes to reach `Ready` state instead of the typical 2-3 minutes. This delay occurs regardless of instance type or region.

### Symptoms

* New Bottlerocket nodes take 5-6 minutes to join the cluster, while Amazon Linux nodes join in 2-3 minutes.
* Running `systemd-analyze blame` on a Bottlerocket node shows `pluto.service` taking \~5 minutes to complete.
* The `journalctl -u pluto` output shows a long gap between the `Starting Generate additional settings for Kubernetes...` and `Finished` log lines.

### Root cause

Bottlerocket uses a system service called `pluto` to generate Kubernetes settings during node bootstrap. One of the values `pluto` resolves is the DNS cluster IP — the ClusterIP of the `kube-dns` service. To discover this value, `pluto` queries the [Instance Metadata Service (IMDS)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) and the EKS API.

When this lookup fails or times out (due to network configuration, security group rules, or IMDS access restrictions), `pluto` retries for approximately 5 minutes before giving up. Because kubelet depends on `pluto` completing, the entire node bootstrap stalls until the timeout expires.

### Solution

Set the **Dns-cluster-ip** field in your Atlas [Node Configuration](node-configuration.md) to the ClusterIP of the `kube-dns` service in your cluster. When this value is explicitly provided, Bottlerocket skips the IMDS lookup and boots without delay.

To find the correct value, run:

```shell
kubectl get svc kube-dns -n kube-system -o jsonpath='{.spec.clusterIP}'
```

For most EKS clusters, this value is `172.20.0.10` or `10.100.0.10`.

Then set it in your Node Configuration:

* **UI**: Go to **Autoscaler** > **Node configuration** > select your configuration > **Advanced Settings** > set the **Dns-cluster-ip** field.
* **Terraform**: Set the `dns_cluster_ip` parameter in the `eks` block:

```hcl
resource "atlas_node_configuration" "default" {
  name       = "default"
  cluster_id = atlas_eks_cluster.example.id

  eks {
    dns_cluster_ip = "172.20.0.10"
  }
}
```

* **API**: Include the `dnsClusterIp` field when creating or updating a node configuration.

### Verification

After applying the change, trigger a new node (for example, by scaling a workload or running a rebalance). The node should reach `Ready` state within 2-3 minutes. You can confirm the improvement in the Atlas console under **Autoscaler** > **Node list** by checking the node creation timeline.

***

## GKE: Pod startup failures with PD Standard on C3/C3D nodes

Pods fail to start when PersistentVolumeClaims using `pd-standard` storage are scheduled onto Nodes running C3 or C3D machine types. These instance families do not support `pd-standard` disks, which prevents volume attachment.

### Symptoms

Pods remain in `ContainerCreating` state and display this warning:

```shell
Warning  FailedAttachVolume  attachdetach-controller  
AttachVolume.Attach failed for volume "pvc-example" : 
rpc error: code = InvalidArgument desc = Failed to Attach: 
failed cloud service attach disk call: googleapi: 
Error 400: pd-standard disk type cannot be used by c3d-standard-90 machine type., badRequest
```

### Root cause

When your default StorageClass is configured with `volumeBindingMode: Immediate` and `type: pd-standard`, Kubernetes provisions the PersistentVolume (PV) before it schedules the Pod. This creates a timing issue:

1. Kubernetes creates the PV as `pd-standard`
2. The scheduler places the Pod on an available Node (potentially C3/C3D)
3. The Node attempts to attach the volume, but cannot support `pd-standard`
4. Volume attachment fails, and the Pod cannot start

C3 and C3D machine types support only `pd-balanced`, `pd-ssd`, and `pd-extreme` disk types.

### Solution

Prevent Pods requiring `pd-standard` storage from scheduling onto incompatible Nodes by implementing both configuration changes below.

#### 1. Update StorageClass volume binding mode

Change your StorageClass to use `WaitForFirstConsumer`. This ensures Kubernetes provisions the PV only after the Pod is scheduled, allowing the scheduler to consider Node constraints, including disk type compatibility.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-standard
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
reclaimPolicy: Delete
```

#### 2. Add Node Selector to Pods using PD Standard

Add this NodeSelector to Pod specifications that use `pd-standard` volumes:

```yaml
nodeSelector:
  volume.scheduling.atlas.example/pd-standard: "true"
```

This ensures these Pods only schedule onto NodePools that support `pd-standard` disks.

#### Why both changes are required

| Configuration          | Purpose                                                                                                            |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `WaitForFirstConsumer` | Delays PV provisioning until Kubernetes selects a Node, enabling the scheduler to evaluate disk type compatibility |
| Node selector          | Restricts pod placement to nodes that support `pd-standard`, preventing scheduling onto C3/C3D machine types       |

With this configuration:

* Pods using PD Standard Volumes schedule only on compatible Node types
* Kubernetes provisions volumes after selecting a valid Node
* Atlas's autoscaler respects these constraints when placing workloads
* Volume attachment failures are prevented

For details on disk type compatibility with GCP machine types, see <Anchor label="Google Cloud documentation on persistent disk types" target="_blank" href="https://cloud.google.com/compute/docs/disks#disk-types">Google Cloud documentation on persistent disk types</Anchor>.

***

## Autoscaler does not provision nodes for custom resource requests

### Symptom

Pods that request a custom or extended resource (for example, a resource exposed by a customer-owned device plugin) stay in `Pending`, and the Atlas autoscaler does not provision new nodes for them.

### Cause

The Atlas autoscaler cannot reason about the capacity of arbitrary custom resources on instance types, so it cannot provision new nodes specifically to satisfy those requests.

### Resolution

Atlas can still schedule or move pods onto existing nodes that already advertise the custom resource (for example, during rebalancing).

For GPU-related custom resources that require custom-resource-aware autoscaling, use one of the supported alternatives:

* [Multi-Process Service (MPS)](gpu-sharing-mps.md) on GKE
* [Dynamic Resource Allocation (DRA)](dra.md) with MPS on AWS, GKE, or AKS

For non-GPU custom resources, or if neither alternative fits your use case, contact [Atlas support](https://atlas.example/contact/) to discuss your requirements.

***

## FailedUpscaling: topology spread blocks node creation

### Symptom

A pod with a hard zone topology spread constraint (`whenUnsatisfiable: DoNotSchedule`) stays `Pending`, and the autoscaler does not create a node for it. The pod records a `FailedUpscaling` event. Inspect it with:

```shell
kubectl describe pod <pod-name>
```

The event is reported in one of these forms:

| Event                                                                                                                                                                   | Meaning                                                                           |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `Node template "<name>" would have been used to upscale, but its node configuration does not have subnets in the following zones required for topology spread: <zones>` | A zone the pod could run in is not covered by the Node Configuration's subnets.   |
| `Node template constrains provisioning to AZs [<zones>], but placing pod in those AZs would violate topology spread constraint (maxSkew=<value>)…`                      | The Node Template's AZ restrictions leave no placement that satisfies the spread. |
| `pod is bound to zone(s)…`                                                                                                                                              | The pod's PVCs bind it to zones that cannot be provisioned.                       |
| `…is constrained to AZs <v>, but pod's topology spread requires AZs <v> which have no overlap`                                                                          | The Node Template's AZ list and the pod's required zones do not intersect.        |
| `pod affinity and/or node selector configuration conflicts with topology spread…`                                                                                       | No zone matches the pod's selector and the spread requirements at the same time.  |

### Cause

A hard zone spread constraint limits the zones in which the autoscaler may create a node for the pod. The autoscaler derives the zones the pod can use as follows:

1. Start from the zones configured on the cluster, keeping those that match the pod's zone `nodeSelector` or required `nodeAffinity`.
2. If the selected [Node Template](node-templates.md) lists allowed AZs, keep only zones allowed by both the template's AZ list and the pod's requirements.
3. Check that spreading across the remaining zones can satisfy `maxSkew`. PVCs bind a pod to the zones of its volumes, and those zones must remain allowed.

On AWS clusters (EKS or self-managed Kubernetes on EC2), the zones the autoscaler can create a node in come from the subnets in the [Node Configuration](node-configuration.md) linked to the selected Node Template. Only zones that hold at least one node the pod could run on require subnet coverage. A zone whose existing nodes the pod's `nodeSelector`, `nodeAffinity`, or taints exclude cannot host the pod. Such a zone demands no subnet, but the autoscaler still removes it from the zones the pod may use. On other providers, subnets do not restrict spreading. A subnet-covered zone is a valid target for a new node even when no node the pod could run on exists there today. The autoscaler creates the node with the Node Template's labels and taints, the pod schedules onto it, and the zone becomes a counted spread domain.

When the remaining zones cannot satisfy the spread, the autoscaler does not create a node, and the pod stays `Pending` with a `FailedUpscaling` event. For a step-by-step walkthrough of the scheduler and autoscaler counting, with a decision table per zone, see [Topology spread constraints and node provisioning](autoscaler.md#topology-spread-constraints-and-node-provisioning).

### Solution

1. Confirm which zones the autoscaler can use for the pod. On AWS, give the [Node Configuration's subnets](node-configuration.md#subnet-requirements) coverage in each zone the pod requires.
2. If the selected [Node Template](node-templates.md) lists allowed AZs, widen the list, or narrow the pod's zone `nodeSelector` or required `nodeAffinity`, so the allowed zones can satisfy `maxSkew`.
3. After changing the configuration, start a reconciliation, then run `kubectl describe pod <pod-name>` to confirm the `FailedUpscaling` event is gone.