---
updatedAt: 2026-10-09T12:56:52.000Z
---

# Subnets

# Subnet selection

## Feature availability

|                                                  | EKS       | GKE | AKS |
| ------------------------------------------------ | --------- | --- | --- |
| **Random subnet selection (from available set)** | **+**     | -   | -   |
| **Subnet selection by usage\***                  | **+\*\*** | -   | -   |

> 📘 Note
>
> \*\*\* Select the subnet with the greatest number of free IP addresses
>
> \*\* Available if AWS cloud CNI is used with specific settings

## Available subnets detection

Cluster subnets with subnet IP CIDR and availability zone are synced periodically with Atlas. The autoscaler, based on various rules, decides from which subnets to choose when constructing in-memory nodes for autoscaling. The selection is influenced by:

* Pod node selector for `topology.atlas.example/subnet-id` label.
* Pod node affinity for `topology.atlas.example/subnet-id` label.
* Availability zone on in-memory node(would choose only subnets in the same zone).
* Choose a zone based on constraints from other parts—e.g., the Persistent Volume zone affects the in-memory node zone.
* Choose the least allocated zone (by CPU cores). If IP allocation calculation is enabled, also consider subnet IP availability.
* Choose a random zone from equally allocated zones.

**If subnet calculation is supported** and we detect that all the available subnets are full, the pod will get a pod event with a message `there is no subnet with enough available IP addresses.`

## Zone selection and node distribution

### How zone allocation works

When the autoscaler selects availability zones for new nodes, it follows this priority order:

1. **Primary factor**: Total CPU core allocation - zones with fewer allocated CPU cores are preferred, provided the zone's subnets have enough free IP addresses for the expected pod density
2. **IP availability check**: For EKS clusters with subnet usage calculation enabled, Atlas verifies that the target zone's subnets have sufficient available IP addresses for the expected pod density
3. **Fallback mechanism**: If the least allocated zone lacks sufficient IP capacity, that zone is removed from the available pool, and the next least allocated zone is evaluated
4. **Final fallback**: If no zones have adequate IP capacity, node provisioning is aborted, preventing nodes that cannot support the required number of pods due to IP constraints

### Understanding zone imbalances

Zone imbalances can occur due to several factors:

1. **Large node bias**: Zones with fewer large nodes may appear less balanced by node count but are actually more balanced by resource allocation
2. **StatefulSet volume constraints**: Persistent volumes are zone-specific, which can force related pods (and their bin-packed companions) into the same availability zone
3. **IP capacity constraints**: Zones may be skipped if their subnets lack sufficient IP addresses, concentrating nodes in zones with higher IP availability

### Achieving better zone balance

If you need more even node distribution across zones, consider:

1. **Topology Spread Constraints**: Add topology spread constraints to your pod specifications to explicitly request zone distribution:
   ```yaml
   topologySpreadConstraints:
   - maxSkew: 1
     topologyKey: topology.kubernetes.io/zone
     whenUnsatisfiable: DoNotSchedule
     labelSelector:
       matchLabels:
         app: your-app
   ```

On AWS, a hard zone spread constraint also interacts with subnets. The autoscaler can create a node only in a zone where your Node Configuration has a subnet. A zone without subnets is removed from the zones available to a spread pod. See [Topology spread constraints and node provisioning](autoscaler.md#topology-spread-constraints-and-node-provisioning).

## Subnet usage calculation

Subnet usage calculation is available only for EKS and when AWS Cloud CNI is used for networking.

Subnet usage is calculated based on CNI settings and instance-type networking capabilities([max ENI count on instance type and ipv4 count per ENI](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#AvailableIpPerENI)).

CNI settings used to calculate used IP addresses:

| Name                           | Description                                                                             | Default |
| ------------------------------ | --------------------------------------------------------------------------------------- | ------- |
| **WARM\_ENI\_TARGET**          | How many free ENIs should be attached as reserve.                                       | 1       |
| **WARM\_IP\_TARGET**           | How many free secondary IPs should be kept as reserve.                                  |         |
| **MINIMUM\_IP\_TARGET**        | Minimum IP count to be requested when adding node.                                      |         |
| **MAX\_ENI**                   | Additional capping on instance type max ENI count.                                      |         |
| **ENABLE\_PREFIX\_DELEGATION** | Enables prefix (/28, 16 IPs per prefix) allocation instead of individual secondary IPs. | `false` |
| **WARM\_PREFIX\_TARGET**       | How many free prefixes should be kept as reserve when prefix delegation is enabled.     |         |

CNI settings that disable subnet usage calculation:

| Name                                         | Supported values  |
| -------------------------------------------- | ----------------- |
| **AWS\_VPC\_K8S\_CNI\_CUSTOM\_NETWORK\_CFG** | `None` or `false` |
| **ENABLE\_IPv6**                             | `None` or `false` |

> 📘 Zone Selection Impact
>
> When subnet usage calculation is enabled, IP availability directly affects zone selection. Zones with insufficient IP capacity are automatically removed from consideration during the autoscaling process, which may contribute to zone imbalances if some zones consistently have lower IP availability than others.

## How does the calculation work

The source of documentation is [here](https://github.com/aws/amazon-vpc-cni-k8s#eni-allocation).

Each instance type in AWS has limits on [how many ENIs can be attached and how many IPs each ENI can have](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#AvailableIpPerENI). Atlas periodically synchronizes those numbers, which this algorithm uses.

Some key points:

* Each ENI uses 1 IP for itself, and all other IPs are secondary and can be used for pods, so always (max IPs for pods in ENI = max IPs per ENI - 1).
* If we just attach ENI, 1 IP will always be used regardless of CNI settings.
* If `WARM_IP_TARGET` is specified, `WARM_ENI_TARGET` is not used.
* If `MAX_ENI` < instances max ENI count, it works as an override for instance setting. Otherwise, instance setting is used.
* If `ENABLE_PREFIX_DELEGATION=true`, IPs are allocated and counted in /28 prefixes (16 IPs per prefix), plus `WARM_PREFIX_TARGET` reserved prefixes. Each new node therefore requires more free subnet IPs than with per-IP allocation, and subnets with less free space than a node needs are skipped. `WARM_ENI_TARGET` does not apply; `WARM_IP_TARGET` / `MINIMUM_IP_TARGET` override `WARM_PREFIX_TARGET` when set.
* All the pods that have `hostNetwork:true` don't get secondary IP and host IP is used for communication (as an example AWS CNI and kube-proxy) PS. They are still counted as PODs and are capped by the podCount constraint on the node.

Here is a detailed description of how [WARM\_ENI\_TARGET, WARM\_IP\_TARGET, and MINIMUM\_IP\_TARGET work](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/eni-and-ip-target.md).

# Troubleshooting

## Useful commands for investigations

Command to get subnet IP allocation—We consider that the subnet is used only for this Kubernetes cluster (some worker groups or security groups might use some IPs if they were created with this subnet, and this could result in a few IPs difference between calculation and actual allocation, bringing failed node creation instead of pod event in some edge cases). Using the same subnets for anything else than this cluster will make this feature work incorrectly.

```shell
aws ec2 describe-network-interfaces --filters Name=subnet-id,Values=subnet_id > subnet_id.yaml
```

Command to print all pods with IP information, sorted by node.

```shell
kubectl get pod -o=custom-columns=NAME:.metadata.name,STATUS:.status.phase,NODE:.spec.nodeName,POD-IP:.status.podIP,HOST-IP:.status.hostIP --sort-by=.spec.nodeName  --all-namespaces
```

## Investigating zone imbalances

If you notice uneven node distribution across zones:

1. **Check CPU allocation per zone**:
   ```shell
   kubectl get nodes -o wide --show-labels | grep topology.kubernetes.io/zone
   kubectl describe nodes | grep -E "(Name:|cpu:|zone)"
   ```
2. **Identify StatefulSets with persistent volumes**:
   ```shell
   kubectl get pvc --all-namespaces -o wide
   kubectl get statefulsets --all-namespaces
   ```
3. **Review large nodes**: Look for nodes with significantly higher CPU allocations that might be skewing zone selection.

### Common causes of zone imbalance

* **Large instance types**: A single large node can make a zone appear "more allocated" than zones with multiple smaller nodes
* **Volume constraints**: StatefulSets with zone-specific volumes can concentrate workloads in particular zones
* **Bin-packing effects**: When pods are packed together with volume-constrained pods, they inherit zone restrictions
* **IP exhaustion**: Zones with insufficient subnet IP addresses are automatically excluded from consideration