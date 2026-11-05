---
updatedAt: 2026-09-29T11:16:11.000Z
---

# Node configuration

The Atlas provisioner allows you to set node configuration parameters that the platform will apply to provisioned nodes. Node configuration on its own does not influence workload placement. Its sole purpose is to apply user-provided configuration settings on the node during the provisioning process.

A cluster can have multiple Node Configurations linked to various [Node Templates](node-templates.md). However, you can select only one node configuration, which Atlas Autoscaler will use as the default.

<Callout icon="📘" theme="info">
  ### Note

  You can link node configuration to multiple node templates, but one node template can have just a single node configuration link.
</Callout>

You can manage node configurations via UI:**Autoscaler**>**Node configuration**, [API](https://docs.atlas.example/atlas/reference/nodeconfigurationapi_listconfigurations) or [Terraform](terraform-provider.md).

## Shared configuration options

The following table provides a list of supported cloud-agnostic configuration parameters:

| Configuration         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | Default value                                                                 |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Root volume ratio     | CPU to storage (GiB) ratio                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | 1 CPU: 0 GiB                                                                  |
| Initial disk size     | The base size of the disk attached to the node                                                                                                                                                                                                                                                                                                                                                                                                                                                                | 100 GiB                                                                       |
| Image                 | Image to be used when building a Atlas provisioned node. See [virtual machine image choice](#virtual-machine-image-choice)  below for cloud-specific behaviors.                                                                                                                                                                                                                                                                                                                                             | The latest available for Kubernetes release, based on an OS chosen by Atlas |
| SSH key               | Base64-encoded public key or AWS key ID                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | ""                                                                            |
| Subnets               | Subnet IDs for Atlas provisioned nodes                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | All subnets pointing to NAT/Internet Gateways inside the cluster's VPC        |
| Instance tags         | Tags/VM labels to be applied on Atlas provisioned nodes                                                                                                                                                                                                                                                                                                                                                                                                                                                     | ""                                                                            |
| Kubelet configuration | A [set of values](https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/#kubelet-config-k8s-io-v1beta1-KubeletConfiguration)  that will be added or overwritten in the Kubelet configuration                                                                                                                                                                                                                                                                                                 | JSON {}                                                                       |
| Init script           | An additional configuration for the node. It can be a bash script or settings in a format depending on the cloud and other settings                                                                                                                                                                                                                                                                                                                                                                           | ""                                                                            |
| Drain timeout         | Maximum time Atlas waits for a node to drain when removing it, such as during [rebalancing](rebalancing.md#pod-disruption-budget-pdb-handling). Once the timeout expires, the node is forcefully drained; if [graceful eviction](rebalancing.md#rebalancing-timeout-behavior) is enabled, the node is instead cordoned and kept. Accepts 0 to 180 minutes. Set via the console field **Drain timeout during rebalancing**, the `drainTimeoutSec` API field, or the `drain_timeout_sec` Terraform argument | 20 minutes (0 = platform default)                                             |

## EKS-specific configuration options

### Instance type compatibility

When configuring nodes in Atlas, certain instance types may have limitations or specific requirements based on their underlying hardware architecture and features. Understanding these limitations is crucial for proper cluster configuration.

#### Instance type identification

You can identify instance types and their underlying architecture through:

1. Cloud Service Provider (CSP) documentation for each instance family
2. Instance type naming patterns
3. CLI commands, which vary for each CSP

#### AWS Nitro System

The AWS Nitro System is Amazon's virtualization infrastructure that provides enhanced networking, high-performance storage, and improved security features. Instance types based on the Nitro system have different capabilities compared to non-Nitro instances.

#### AWS Bottlerocket images

Bottlerocket is a Linux distribution designed and optimized for container orchestration. It contains only the essential components required for this purpose, minimizing the attack surface and enforcing container best practices.

Bottlerocket has specific instance type requirements and limitations.

<Callout icon="📘" theme="info">
  ### Note

  When using Bottlerocket AMIs, ensure your node templates include only Nitro-based instance types. Non-Nitro instances are not supported and will fail to provision.
</Callout>

### Subnet requirements

<Callout icon="📘" theme="info">
  ### Note

  In EKS only subnets which match one of the rules below are allowed to be added to Node Configuration:

  - association with a route table that has a 0.0.0.0/0 route to Internet Gateway, it's known as a public subnet. Subnet also must have "MapPublicIpOnLaunch: true" set
  - association with a route table that has a 0.0.0.0/0 route to Transit Gateway, it's known as a private subnet
  - association with a route table that has a 0.0.0.0/0 route to NAT Gateway, it's known as a private subnet

  If Atlas cannot detect a routable subnet (a subnet that has access to the Internet), you can add a tag `atlas.example/routable=true` to the subnet. Atlas will then consider a subnet with this tag as having Internet access.
</Callout>

| Configuration                                                                                       | Description                                                                                                                                                                         | Default value                                                                                 |
| :-------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :-------------------------------------------------------------------------------------------- |
| Security groups                                                                                     | Security group IDs for Atlas-provisioned nodes.                                                                                                                                   | [Tagged](https://docs.aws.amazon.com/eks/latest/userguide/sec-group-reqs.html) and Atlas SG |
| Instance profile ARN                                                                                | Instance profile ARN for Atlas-provisioned nodes. Not recommended to share with EKS-managed node pools.                                                                           | `atlas-<cluster-name>-eks-<cluster-id>` (last 8 digits of cluster ID)                          |
| Dns-cluster-ip                                                                                      | Overrides the IP address used for DNS queries within the cluster. Supports IPv4 and IPv6.                                                                                           | `""`                                                                                          |
| Container runtime                                                                                   | Container runtime engine: Docker or `containerd`. `containerd` is required for [Container Live Migration](clm-overview.md).                                                        | Unspecified                                                                                   |
| Docker configuration                                                                                | Values overwritten in the [Docker daemon configuration](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file).                                   | `{}`                                                                                          |
| Volume type                                                                                         | EBS volume type for provisioned nodes.                                                                                                                                              | `gp3`                                                                                         |
| Volume IOPS                                                                                         | EBS volume IOPS for provisioned nodes.                                                                                                                                              | `3000`                                                                                        |
| KMS Key ARN                                                                                         | Customer-managed KMS key for EBS volume encryption. See [KMS key for EBS volume](#kms-key-for-ebs-volume).                                                                          | Unspecified                                                                                   |
| Volume throughput                                                                                   | EBS volume throughput in MiB/s.                                                                                                                                                     | `125`                                                                                         |
| Use [IMDSv1](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/environments-cfg-ec2-imds.html) | When `True`, IMDSv1 and v2 are both allowed. When `False`, only IMDSv2 is allowed.                                                                                                  | `True`                                                                                        |
| Target groups                                                                                       | List of ARN and port. Nodes register with these load balancer target groups on creation. See [Load balancer target group prerequisites](#load-balancer-target-group-prerequisites). | Unspecified                                                                                   |
| Image family                                                                                        | OS family for provisioned nodes. See [Image family](#image-family).                                                                                                                 | `FAMILY_AL2023` for k8s 1.33+; `FAMILY_AL2` for k8s 1.32 and below                            |
| Max pods formula                                                                                    | Dynamic formula for kubelet max pod count. See [Maximum pods formula](#maximum-pods-formula).                                                                                       | Unspecified (EKS default)                                                                     |
| Threads per CPU                                                                                     | Controls simultaneous multithreading (SMT/Hyper-Threading). See [Threads per CPU](#threads-per-cpu).                                                                                | Enabled                                                                                       |
| Node group ARN                                                                                      | IAM role ARN for temporary managed node groups. See [Node group ARN](#node-group-arn).                                                                                              | Unspecified                                                                                   |

<Callout icon="🚧" theme="warn">
  ### Kubelet configuration

  The `kubeReserved` and `evictionHard` settings are excluded from the kubelet configuration on **Amazon Linux 2 (AL2)** nodes. These settings are fully supported on **Amazon Linux 2023 (AL2023)** and **Bottlerocket** image families.
</Callout>

<Callout icon="🚧" theme="warn">
  ### Important: AWS is ending support for Amazon Linux 2 (AL2) with Kubernetes v1.32

  Starting with Kubernetes v1.33, only AL2023 and Bottlerocket AMIs will be supported by AWS for EKS. This change has the following implications:

  - **Atlas-managed nodes** already default to AL2023 on clusters running Kubernetes 1.33 or higher
  - **Existing AL2-based nodes** won't be automatically upgraded - manual migration is required
  - The upgrade deadline, based on AWS's initial guidance, is expected to be around November 2025
  - **Container Live Migration** requires AL2023 and is not compatible with AL2 or Bottlerocket

  To prepare for this change, we recommend:

  1. Update your Node Configuration to use `FAMILY_AL2023` for Container Live Migration compatibility, or `FAMILY_BOTTLEROCKET` for security-focused workloads
  2. Test the new OS images in non-production environments first
  3. Plan a gradual migration of your production workloads before the deadline

  Do not hesitate to contact Atlas Customer Success for any additional guidance and support.
</Callout>

<Callout icon="⚠️" theme="warn">
  ### Warm pool compatibility

  Atlas does not support managing nodes provisioned from AWS Auto Scaling warm pools. If your EKS cluster uses warm pools for faster node provisioning, those nodes should be managed directly through AWS tools rather than through Atlas.
</Callout>

### Image family

The OS family used when provisioning nodes. Possible values:

* `FAMILY_AL2` — Amazon Linux 2. **Being deprecated** with Kubernetes v1.33. New deployments should use `FAMILY_AL2023` or `FAMILY_BOTTLEROCKET`.
* `FAMILY_AL2023` — Amazon Linux 2023. Required for [Container Live Migration](clm-overview.md).
* `FAMILY_BOTTLEROCKET` — Bottlerocket.

**Default:** `FAMILY_AL2023` for Kubernetes v1.33+; `FAMILY_AL2` for v1.32 and below.

### Node group ARN

Specifies a dedicated IAM role ARN to be used when creating temporary managed node groups (such as during [cluster hibernation](autoscaling-cluster-hibernation.md) resume). Required for clusters where the node IAM role is granted cluster access via the `aws-auth` ConfigMap, to prevent IAM role removal from the ConfigMap when temporary node groups are deleted. Not required when node IAM roles use EKS access entries for authentication.

<Callout icon="📘" theme="info">
  ### Note

  Despite its name, this field expects an IAM role ARN, not a node group ARN. This field is only available via Terraform (`node_group_arn`) or API (`nodeGroupArn`).
</Callout>

### KMS key for EBS volume

The key that you provide for the encryption of EBS volume must have the following policy:

```json JSON
  {
            "Sid": "Allow access through EBS for all principals in the account that are authorized to use EBS",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Action": [
                "kms:ReEncrypt*",
                "kms:GenerateDataKey*",
                "kms:Encrypt",
                "kms:DescribeKey",
                "kms:Decrypt",
                "kms:CreateGrant"
            ],
            "Resource": "*",
            "Condition": {
                "StringEquals": {
                    "kms:CallerAccount": "<<account_ID>",
                    "kms:ViaService": "ec2.<<region>>.amazonaws.com"
                }
            }
}
    
```
```Text Terraform
module "kms" {
  source = "terraform-aws-modules/kms/aws"

  description = "EBS key"
  key_usage   = "ENCRYPT_DECRYPT"

  # Policy

  key_statements = [
    {
      sid =  "Allow access through EBS for all principals in the account that are authorized to use EBS",
      principals = [
        {
          type        = "AWS"
          identifiers = ["*"]
        }
      ]
      actions = [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:ReEncrypt*",
        "kms:GenerateDataKey*",
        "kms:CreateGrant",
        "kms:DescribeKey"
      ],
      resources =  ["*"],
      conditions = [
        {

          test     = "StringEquals"
          variable = "kms:ViaService"
          values   = [
            "ec2.${var.cluster_region}.amazonaws.com",
          ]
        },
        {
          test     = "StringEquals"
          variable = "kms:CallerAccount"
          values   = [
            data.aws_caller_identity.current.account_id
          ]
        }
      ]}
  ]

# Aliases
  aliases = ["mycompany/ebs"]

  tags = {
    Terraform   = "true"
    Environment = "dev"
  }
}
```

### Load balancer target group prerequisites

To use load balancer target groups with Atlas-managed nodes, you need to:

1. Configure the appropriate IAM permissions
2. Specify target groups in your node configuration

#### Required IAM permissions

The Atlas IAM role requires additional permissions to register and deregister nodes with target groups. Add the following IAM policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "atlas-target-group",
            "Effect": "Allow",
            "Action": [
                "elasticloadbalancing:RegisterTargets",
                "elasticloadbalancing:DeregisterTargets"
            ],
            "Resource": "arn:aws:elasticloadbalancing:<region>:<account>:targetgroup/*/*"
        }
    ]
}
```

<Callout icon="📘" theme="info">
  ### Note

  - This example policy allows access to all target groups. For enhanced security, replace the wildcards `*` with specific target group ARNs.
  - If your cluster uses IPv6, ensure your nodes have the necessary IPv6 address assignment permissions configured. See the AWS documentation for [IPv6 networking](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-ip-addressing.html#ipv6-addressing) and Atlas [cloud permissions](cloud-permissions.md#ipv6-support-for-eks-clusters) documentation.
</Callout>

#### How target groups work

* When specified in the node configuration, new instances are automatically registered with the load balancer target groups upon creation
* Without proper permissions, nodes will still be created but won't be registered with target groups
* If registration fails, you'll receive a notification detailing which target groups weren't updated

### Using IP prefix delegation with Atlas

[Prefix delegation mode](https://docs.aws.amazon.com/eks/latest/userguide/cni-increase-ip-addresses.html) in AWS CNI is a way to increase pod density on instances beyond the normal limits. Atlas detects when this setting is turned on in the CNI and adjusts the calculations about subnet and node capacity accordingly. See [AWS documentation](https://docs.aws.amazon.com/eks/latest/userguide/choosing-instance-type.html) for more information on how the setting affects IP address management.

If prefix delegation is used, a Maximum Pods formula should be set to properly configure the kubelet's maximum pod count. See *Prefix delegation formula* below for the default formula that matches AWS recommendations. This can be adjusted if needed (for example, to reserve an IP for extra ENIs on the nodes for other purposes). If the formula is not set, it is possible to see lower pod density than expected, even though there are sufficient IP slots available on the node.

<Callout icon="📘" theme="info">
  ### Container Live Migration compatibility

  Container Live Migration supports both standard and prefix delegation networking configurations. When using prefix delegation with CLM, ensure your node configuration maintains consistent networking setup across all CLM-enabled nodes.
</Callout>

### Maximum Pods formula

The Maximum Pods formula dynamically determines the maximum number of pods that can run on a node in your EKS cluster. This calculation serves two primary purposes:

* It informs the `kubelet` about the maximum number of pods that can be hosted on a node.
* It assists the Atlas Autoscaler in planning and optimizing cluster resources.

You can optimize your cluster's pod distribution and resource utilization by fine-tuning this formula.

When using AWS CNI for IP address management, setting a static value for max pods directly in the `kubelet` configuration is not recommended. The same node configuration is usually used for different instance types, and the max pods differ per instance type. Use the Maximum Pods formula to make the value dynamic while still configuring the `kubelet` accordingly or remove the setting from the `kubelet`'s configuration (the value will be calculated dynamically at node bootstrap time by default).

#### Formula variables

Maximum Pods formulas are constructed using the following variables:

| Variable                 | Description                                                                                                                                | Range/Value       | Default |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ | ----------------- | ------- |
| `NUM_IP_PER_PREFIX`      | Number of IPv4 addresses per prefix. Affects calculation when using prefix delegation. Specified as a parameter in the node configuration. | `0-256`           | `1`     |
| `NUM_MAX_NET_INTERFACES` | Maximum number of network interfaces.                                                                                                      | Instance-specific | N/A     |
| `NUM_IP_PER_INTERFACE`   | Number of IPv4 addresses per interface.                                                                                                    | Instance-specific | N/A     |
| `NUM_CPU`                | Number of CPUs.                                                                                                                            | Instance-specific | N/A     |
| `NUM_RAM_GB`             | Amount of RAM in GB.                                                                                                                       | Instance-specific | N/A     |

Atlas provides several preset formulas by default from which to choose.

#### Example formulas

The Maximum Pods formula can be tailored to different cluster configurations and requirements. Here are some common examples:

**Default EKS formula**

```
NUM_MAX_NET_INTERFACES * (NUM_IP_PER_INTERFACE - 1) + 2
```

This is often the default formula for EKS clusters. It calculates the maximum pods based on available network interfaces and IPs, reserving one IP per interface and adding 2 for system pods.

**Prefix delegation formula**

```
NUM_MAX_NET_INTERFACES * (NUM_IP_PER_INTERFACE - 1) * NUM_IP_PER_PREFIX + 2
```

This formula accounts for the additional IPs available through prefix delegation. Since the values of the formula can get unrealistically high for large instance types, Atlas provides a preset that also puts a reasonable upper limit:

```
math.least((NUM_MAX_NET_INTERFACES - 1) _ (NUM_IP_PER_INTERFACE - 1) _ NUM_IP_PER_PREFIX + 2, 300)
```

<Callout icon="📘" theme="info">
  ### Note

  The 300-pod cap is a safeguard against node instability that can occur with extremely high pod counts. While AWS instances can theoretically support higher IP counts, running hundreds of pods on a single node can lead to resource constraints and performance issues, especially for the kubelet process. AWS's own calculator caps at 250 pods for similar stability reasons.
  If your workload is specifically designed to handle higher pod densities and you have tested node stability at these counts, you can create a custom formula without this cap using the API or Terraform.
</Callout>

<Callout icon="📘" theme="info">
  ### Note

  AWS prefix delegation is not supported on non-Nitro instance types. If you're using prefix delegation in your cluster, ensure your node templates exclude non-Nitro instance families. This limitation applies regardless of which autoscaling solution you use, as it is an AWS-imposed limitation.

  If you're not using prefix delegation, you can safely use max pods formulas (without the `NUM_IP_PER_PREFIX` variable) with any instance type.
</Callout>

**Custom networking formula**

```
(NUM_MAX_NET_INTERFACES - 1) * (NUM_IP_PER_INTERFACE - 1) + 2
```

Use this formula when `AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG=true`. In custom networking mode, AWS reserves the primary ENI for the node's own networking, so only secondary ENIs carry pod IPs.

**Security Groups for Pods (SGPP) formula**

Atlas does not auto-detect `ENABLE_POD_ENI=true` (AWS Security Groups for Pods). When SGPP is enabled, AWS reserves one ENI as a trunk ENI for the VPC resource controller. That trunk ENI cannot be used for regular pod IP allocation, so subtract one ENI from the formula:

```
(NUM_MAX_NET_INTERFACES - 1) * (NUM_IP_PER_INTERFACE - 1) + 2
```

With prefix delegation:

```
math.least((NUM_MAX_NET_INTERFACES - 1) * (NUM_IP_PER_INTERFACE - 1) * NUM_IP_PER_PREFIX + 2, 300)
```

If you use SGPP together with custom networking, subtract one more ENI (one for the primary/node networking and one for the trunk):

```
(NUM_MAX_NET_INTERFACES - 2) * (NUM_IP_PER_INTERFACE - 1) + 2
```

With prefix delegation:

```
math.least((NUM_MAX_NET_INTERFACES - 2) * (NUM_IP_PER_INTERFACE - 1) * NUM_IP_PER_PREFIX + 2, 300)
```

<Callout icon="📘" theme="info">
  ### Note

  This is a conservative recommendation. Pods that actually use Security Groups for Pods consume IPs from branch interfaces on the trunk ENI, not from regular ENI slots, so in clusters where most pods use SGPP the `-1` adjustment can be slightly pessimistic. However, it prevents the more serious failure mode of non-SGPP pods scheduling and then failing to get an IP because the trunk ENI is unavailable for regular pod traffic. See the [AWS SGPP best-practices guide](https://docs.aws.amazon.com/eks/latest/best-practices/sgpp.html) and [AWS Security Groups for Pods documentation](https://docs.aws.amazon.com/eks/latest/userguide/security-groups-for-pods.html) for details.
</Callout>

**Fixed value**

```
512
```

Simple scalar values are allowed. This example sets a fixed maximum of `512` pods per node, regardless of other factors.

These examples demonstrate the flexibility of the Maximum Pods formula. You can create custom formulas that best suit your cluster's specific needs and constraints.

#### How the Maximum Pods formula works

The MaxPods formula is an expression that evaluates a number. We use [Common Expression Language (CEL)](https://cel.dev) to evaluate it. Here is how it operates:

* **Input**: The formula takes the node's characteristics as input variables (e.g., `NUM_MAX_NET_INTERFACES`, `NUM_CPU`, scalers, etc.).
* **Calculation**: It performs evaluation using the CEL library.
* **Output**: The calculation result is used as an integer. This number becomes the maximum number of pods allowed on the node.
* **Application**: The calculated value is used by the `kubelet` to limit pod scheduling and by the Atlas Autoscaler for capacity planning.

For example, let's break down a simple formula:

```
NUM_MAX_NET_INTERFACES * (NUM_IP_PER_INTERFACE - 1) + 2
```

If a node has 4 network interfaces (`NUM_MAX_NET_INTERFACES` = `4`) and `10` IPs per interface (`NUM_IP_PER_INTERFACE` = `10`), the calculation would be:

```
4 * (10 - 1) + 2 = 4 * 9 + 2 = 36 + 2 = 38
```

Thus, this node would be allowed to run a maximum of `38` pods.

The formula's flexibility allows for complex calculations that can account for various resource constraints and operational requirements. When you select or create a formula, you define the logic for this pod limit calculation.

#### Custom formulas

While the Atlas Console UI provides several preset formulas for configuring an EKS node, you can define a custom formula using Terraform for more advanced configurations.

### Threads per CPU

The Threads per CPU setting allows you to control whether Hyper-Threading (simultaneous multithreading) is enabled or disabled on your AWS EKS worker nodes.

![](https://files.readme.io/e71adc7257646e398410aec08c23e2ab69a96fdf1173f1c73963cb6a7b695477-image.png)

Disabling Hyper-Threading can be beneficial for high-performance computing (HPC) workloads that require predictable performance and minimal resource contention.

#### How to configure

To disable Hyper-Threading:

* **UI**: Check the **Disable Hyper-Threading** checkbox in the **Advanced Settings** section of your **Node Configuration**
* **API**: Set the `threadsPerCpu` parameter to `1` when <Anchor target="_blank" href="https://docs.atlas.example/reference/nodeconfigurationapi_updateconfiguration">updating a node configuration</Anchor>. To keep Hyper-Threading enabled, leave this parameter undefined
* **Terraform**: Use the `threads_per_cpu` parameter in the `eks` block of your `atlas_node_configuration` resource

#### How it works

* **Enabled (default)**: Hyper-Threading is active, allowing each physical CPU core to handle multiple threads simultaneously
* **Disabled**: Hyper-Threading is turned off, restricting each physical core to handle only one thread at a time

When Hyper-Threading is disabled, the number of available vCPUs will equal the number of physical CPU cores on the instance.

<Callout icon="🚧" theme="warn">
  ### Important

  Changing the Threads per CPU setting only affects newly provisioned nodes. To apply this configuration to existing nodes, you must either delete and replace them or perform a cluster <Anchor target="_blank" href="doc:rebalancing">rebalancing</Anchor> operation.
</Callout>

### Using Amazon Linux 2023 with Atlas

[Amazon Linux 2023 (AL2023)](https://docs.aws.amazon.com/linux/al2023/ug/what-is-amazon-linux.html) is the next generation of Amazon Linux, a Linux distribution provided by Amazon Web Services (AWS). AL2023 is the required image family for Container Live Migration functionality.

#### Passing Custom Configuration via Init Script

It is possible to pass a shell script to run during node initialization or pass a `nodeadm` NodeConfig to apply to the new nodes. Note that the values passed as NodeConfig will apply and possibly overwrite values computed by the Atlas provisioner.

##### Passing a shell script

Init script:

```shell
#!/bin/bash

echo "a script has run during node provisioning" > /var/log/node-configuration-init-script-out.txt
```

##### Passing a `nodeadm` Node Config

Init script specified as MIME multipart with `nodeadm` NodeConfig. Refer to [AWS Node Configuration API](https://awslabs.github.io/amazon-eks-ami/nodeadm/doc/api/).

```
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="BOUNDARY"

--BOUNDARY
Content-Type: application/node.eks.aws

---
apiVersion: node.eks.aws/v1alpha1
kind: NodeConfig
spec:
  kubelet:
    config:
      evictionMaxPodGracePeriod: 30
--BOUNDARY--

```

##### Passing a shell script and `nodeadm` Node Configuration

Using MIME multipart, it is possible to pass `nodeadm` NodeConfig and shell scripts together:

```
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="BOUNDARY"

--BOUNDARY
Content-Type: application/node.eks.aws

---
apiVersion: node.eks.aws/v1alpha1
kind: NodeConfig
spec:
  kubelet:
    config:
      evictionMaxPodGracePeriod: 30
--BOUNDARY
Content-Type: text/x-shellscript; charset="us-ascii"

#!/bin/bash

echo "a script has run during node provisioning" > /var/log/node-configuration-init-script-out.txt

--BOUNDARY--
```

### Using Bottlerocket with Atlas

Bottlerocket is a Linux distribution designed and optimized for container orchestration. It contains only the essential components required for this purpose, minimizing the attack surface and enforcing container best practices.

<Callout icon="📘" theme="info">
  ### Container Live Migration compatibility

  Container Live Migration is currently not supported with Bottlerocket AMIs. If you require Container Live Migration functionality, use Amazon Linux 2023 (`FAMILY_AL2023`) instead.
</Callout>

<Callout icon="⚠️" theme="warn">
  ### Slow node startup with Bottlerocket

  If Bottlerocket nodes take 5+ minutes to reach `Ready` state, the most common cause is a missing **Dns-cluster-ip** value in your Node Configuration. Without this value, Bottlerocket's bootstrap process attempts to discover it via IMDS, which can time out and delay the node join by \~5 minutes. Set the **Dns-cluster-ip** field to your cluster's `kube-dns` ClusterIP to resolve this. See [Troubleshooting: Bottlerocket nodes take 5+ minutes to reach Ready state](autoscaler-troubleshooting.md#eks-bottlerocket-nodes-take-5-minutes-to-reach-ready-state) for details.
</Callout>

This section provides guidance on using Bottlerocket with Atlas, including instructions on passing custom configurations and important considerations to remember. Useful resources are linked throughout to further understand Bottlerocket and its use cases. To start, here is a Bottlerocket overview:
[Container Host - Bottlerocket - Amazon Web Services](https://aws.amazon.com/bottlerocket/?amazon-bottlerocket-whats-new.sort-by=item.additionalFields.postDateTime\&amazon-bottlerocket-whats-new.sort-order=desc)

#### Using Bottlerocket

You can use Bottlerocket on Atlas by specifying the appropriate Amazon Machine Image (AMI) ID or name in Atlas node configurations. This works out of the box for standard setups, but applying custom configurations requires a slightly different approach due to Bottlerocket's unique architecture.

<Image src="https://files.readme.io/027b59a32d344006d70059ea62e23c379bcf22725073ec5810bd8a594b3daa0a-image.png" alt="Specifying the appropriate Amazon Machine Image (AMI) ID" align="center" caption="Specifying the appropriate Amazon Machine Image (AMI) ID" />

<br />

<Image src="https://files.readme.io/a1afe7a5970a46b757188c946814583cc0f59fdd7ae4f3be684d00d597b00b8c-image.png" alt="Specifying the appropriate Amazon Machine Image (AMI) name" align="center" caption="Specifying the appropriate Amazon Machine Image (AMI) name" />

#### Limitations with Advanced Node Configuration

When using Bottlerocket AMIs, many of the fields in the advanced section of Atlas's node configurations do not apply and cannot be used to pass custom configurations. This limitation is due to Bottlerocket's security-focused architecture, which does not provide a default shell in its container. Consequently, any changes to the AMI configuration must be passed through the settings exposed in the Bottlerocket API.

For a full list of settings available in Bottlerocket's API, refer to the [Bottlerocket Settings API Reference](https://bottlerocket.dev/en/os/1.20.x/api/settings-index/).

#### Passing Custom Configurations via Init Script

Outside of Atlas, custom configurations are typically passed to Bottlerocket AMIs in TOML format. Here's an example:

```toml
[settings.kubernetes]
api-server = "${endpoint}"
cluster-certificate = "${cluster_auth_base64}"
cluster-name = "${cluster_name}"
${additional_userdata}

[settings.kubernetes.node-labels]
"ingress" = "allowed"
"environment" = "prod"

[settings.kubernetes.system-reserved]
cpu = "100m"
memory = "256Mi"
ephemeral-storage = "2Gi"

[settings.kubernetes.kube-reserved]
cpu = "100m"
memory = "256Mi"
ephemeral-storage = "512Mi"

[settings.kubernetes.eviction-hard]
"memory.available" = "10%"

# Hardening based on <https://github.com/bottlerocket-os/bottlerocket/blob/develop/SECURITY_GUIDANCE.md>

[settings.kernel]
lockdown = "integrity"

[settings.host-containers.admin]
enabled = false
source = "328549459982.dkr.ecr.eu-central-1.amazonaws.com/bottlerocket-admin:v0.7.2"

[settings.host-containers.control]
enabled = false
```

#### Converting TOML Settings to Init Script Format

In Atlas, you can pass custom Bottlerocket configurations through the init script in the advanced section of node configurations. However, the settings must be converted to a single-line format compatible with the init script to do this.

Below is an example that converts the TOML-style settings listed above into a format compatible with the init script:

```toml
settings.kubernetes.node-labels."ingress"="allowed"
settings.kubernetes.node-labels."environment"="prod"

settings.kubernetes.system-reserved.cpu="100m"
settings.kubernetes.system-reserved.memory="256Mi"
settings.kubernetes.system-reserved.ephemeral-storage="2Gi"

settings.kubernetes.kube-reserved.cpu="100m"
settings.kubernetes.kube-reserved.memory="256Mi"
settings.kubernetes.kube-reserved.ephemeral-storage="512Mi"

settings.kubernetes.eviction-hard."memory.available"="10%"

settings.kernel.lockdown="integrity"

settings.host-containers.admin.enabled=false
settings.host-containers.admin.source="328549459982.dkr.ecr.eu-central-1.amazonaws.com/bottlerocket-admin:v0.7.2"

settings.host-containers.control.enabled=false
```

<Callout icon="🚧" theme="warn">
  ### Bottlerocket GPU nodes

  When using GPU nodes with Bottlerocket, you must keep the default parameter `settings.kernel.lockdown="none"`.
  There is a known Bottlerocket issue with self-signing the NVIDIA drivers: <Anchor target="_blank" href="https://github.com/bottlerocket-os/bottlerocket/issues/4218">https://github.com/bottlerocket-os/bottlerocket/issues/4218</Anchor>.

  **GPU sharing configuration**: When using GPU time-slicing with Bottlerocket, Atlas automatically configures the required NVIDIA device plugin settings. For configuration details, see <Anchor target="_blank" href="doc:gpu-sharing-time-slicing">GPU sharing with time-slicing</Anchor>.
</Callout>

#### Docker Authentication Configuration

When using private container registries with Bottlerocket nodes, you need to configure Docker authentication credentials through the container registry settings for pulling images from Docker Hub or other private registries.

##### Configuration Format

To configure Docker authentication for Bottlerocket nodes, use the following format in your init script:

```toml
settings.container-registry.credentials = [{registry = "docker.io", auth = "base64_encoded_credentials"}]
```

Or with username and password:

```toml
settings.container-registry.credentials = [{registry = "docker.io", username = "your_username", password = "your_password"}]
```

##### Example Configuration

Here's a complete example for Docker Hub authentication:

```toml
settings.container-registry.credentials = [{registry = "docker.io", auth = "dXNlcm5hbWU6cGFzc3dvcmQ="}]
```

##### Multiple Registry Support

You can configure multiple container registries by adding additional entries to the credentials array:

```toml
settings.container-registry.credentials = [
  {registry = "docker.io", auth = "base64_encoded_docker_credentials"},
  {registry = "your-private-registry.com", auth = "base64_encoded_private_credentials"}
]
```

For more information on Bottlerocket's container registry settings, refer to the [Bottlerocket Container Registry API documentation](https://bottlerocket.dev/en/os/1.20.x/api/settings/container-registry/#credentials).

#### Bottlerocket Settings Reference

The remaining fields in the advanced section of Atlas's node configuration cannot be used to pass these configurations. Instead, you must use the settings exposed in the Bottlerocket API and set them within the init script.

Here is a visual of what sections of advanced node configuration settings can be used with Bottlerocket:

<Image src="https://files.readme.io/0bb59e4732d9fbcd242bbe3e1003203968abc603c5cd78eba1efc278709762f6-Screenshot_2025-07-16_at_10.20.21.png" align="center" />

Customers migrating from other AMIs to Bottlerocket must convert their custom configurations into this Bottlerocket-compatible style and can only use the settings exposed in the API reference shared above.

#### Using Bootstrap Containers for Advanced Configurations

Any custom configurations outside of those settings, or the ability to change settings dynamically at boot, require the use of what Bottlerocket calls a bootstrap container.

Here is [an example](https://github.com/kisahm/bottlerocket-bootstrap-custom-max-pods/tree/main) that dynamically adjusts the maximum number of pods per node based on the instance type:

```toml
settings.bootstrap-containers.max-pods-calculator.source = "docker.io/kisahm/bottlerocket-bootstrap-max-pods:v0.2"
settings.bootstrap-containers.max-pods-calculator.essential = false
settings.bootstrap-containers.max-pods-calculator.mode = "always"
settings.bootstrap-containers.max-pods-calculator.user-data = "ZXhwb3J0IEFERElUSU9OQUxfT1BUSU9OUz0iLS1jdXN0b20tY25pIGNpbGl1bSAtLWNpbGl1bS1maXJzdC1pbnRlcmZhY2UtaW5kZXggMSIK"
```

See the GitHub repository for detailed instructions: [Bottlerocket Bootstrap Custom Max Pods](https://github.com/kisahm/bottlerocket-bootstrap-custom-max-pods/tree/main)

#### Connecting to Bottlerocket Nodes

Connecting to Bottlerocket nodes requires using an admin or control host container with special privileges. Guidance is provided in the Bottlerocket documentation:

* [Bottlerocket FAQ](https://bottlerocket.dev/en/faq/#1_2)
* [Bottlerocket GitHub Repository - Exploration](https://github.com/bottlerocket-os/bottlerocket#exploration)

Using Bottlerocket with Atlas provides a secure foundation for container orchestration. While Bottlerocket's minimalistic design limits the use of traditional configuration methods, you can effectively pass custom configurations through the init script by converting them into the appropriate format. For advanced or dynamic configurations, bootstrap containers offer a solution to extend functionality.

## GKE-specific configuration options

| Configuration                          | Description                                                                                                                                                               | Default value             |
| :------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | :------------------------ |
| Network tags                           | String added to the `tags` field of GCP VM resources.                                                                                                                     | Empty                     |
| Max pods per node                      | Maximum pods hosted on a node.                                                                                                                                            | `110`                     |
| Boot disk                              | Boot disk storage type. See [GCP documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/custom-boot-disks).                                                | `balanced`                |
| Use local SSD-backed ephemeral storage | Attach local ephemeral storage backed by Local SSD volumes. See [GCP documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/local-ssd). | `False`                   |
| Max pods per node formula              | Advanced formula to calculate maximum pods per node. Affects Pod CIDR space reservation. See [GKE max pods per node configuration](#gke-max-pods-per-node-configuration). | Empty (uses static value) |

### GKE max pods per node configuration

GKE offers two ways to configure the maximum number of pods per node:

1. **Max pods per node** (recommended): Set a static number that applies to all nodes in the configuration
2. **Max pods per node formula** (advanced): Use a dynamic formula that calculates the limit based on node characteristics

#### When to use the formula

The formula field is an advanced option that dynamically calculates pod limits based on node properties. Use it only when:

* You need different pod limits for different instance types within the same node configuration
* You have specific requirements that cannot be met with a static value

For most use cases, setting a static `Max pods per node` value provides better predictability and is easier to manage.

#### Formula variables

The following variables are available for use in the formula and will be bound to numeric values before evaluation:

| Variable     | Description                                      |
| ------------ | ------------------------------------------------ |
| `NUM_CPU`    | Number of CPUs available on the node             |
| `NUM_RAM_GB` | Amount of RAM in gigabytes available on the node |

#### Example formula

If you want the smallest value between 5 times the CPUs, 5 times the RAM, or a cap of 110, your formula would be:

```Text Example
math.least(110, 5*NUM_CPU, 5*NUM_RAM_GB)
```

For a node with 8 CPUs and 16 GB RAM, this calculates to:

* 40 (5×8 CPUs)
* 80 (5×16 GB RAM)
* 110 (cap)

The formula picks the smallest value: **40 pods**.

#### How GKE reserves IP addresses

GKE uses a specific method to reserve Pod CIDR space:

1. Takes your max pods per node value (static or calculated from formula)
2. Doubles this value
3. Rounds up to the next power of 2 to determine the CIDR size

This approach reduces IP address reuse as Pods are created and deleted, improving stability.

**Example:** If max pods per node is 11:

* GKE reserves 22 IPs (11 × 2)
* Next power of 2 is 32
* Pod CIDR will be /27

For more information, see Google's documentation on [Flexible Pod CIDR ranges for clusters](https://cloud.google.com/kubernetes-engine/docs/how-to/flexible-pod-cidr#cidr_ranges_for_clusters).

## AKS-specific configuration options

| Configuration          | Description                                                                                                                                                                                                                                                    | Default value             |
| :--------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :------------------------ |
| Max pods per node      | Maximum pods hosted on a node.                                                                                                                                                                                                                                 | `30`                      |
| OS disk                | Type of managed [OS disk](https://learn.microsoft.com/en-us/azure/virtual-machines/disks-types#disk-type-comparison). Not used when Ephemeral OS disk is enabled.                                                                                              | Standard SSD              |
| Ephemeral OS disk      | Places the OS disk on the VM's local cache, resource, or NVMe disk instead of a managed disk. See [Ephemeral OS disk for AKS](#ephemeral-os-disk-for-aks).                                                                                                     | Disabled                  |
| Image family           | OS family for provisioned nodes. See [Image family](#aks-image-family).                                                                                                                                                                                        | `FAMILY_UBUNTU`           |
| Pod subnet ID          | Azure subnet ID for pod IP allocation. Supports multiple pod subnets within a single AKS cluster. If not specified, the cluster default is used.                                                                                                               | Empty (cluster default)   |
| Load balancers         | Load balancers to attach nodes to. Setting this disables Atlas's default load balancer auto-discovery. See [Load balancer attachment](#load-balancer-attachment).                                                                                            | Empty                     |
| Accelerated networking | Controls SR-IOV <Anchor target="_blank" href="https://learn.microsoft.com/en-us/azure/virtual-network/accelerated-networking-overview">accelerated networking</Anchor> on the node NIC. See [Accelerated networking for AKS](#accelerated-networking-for-aks). | Enabled on supported SKUs |

### Ephemeral OS disk for AKS

By default, AKS nodes use a managed disk for the OS. An ephemeral OS disk instead places the OS disk on the VM's local storage. This removes the OS disk cost and speeds up node provisioning and reimaging. The OS disk is ephemeral: its contents do not persist when the node stops, gets replaced, or gets reimaged. Kubernetes OS disk data does not need to persist, so this is safe for nodes.

When you enable Ephemeral OS disk, you choose a **Placement**. The placement decides which local disk holds the OS disk:

| Placement     | Local disk used                                                                                                          |
| :------------ | :----------------------------------------------------------------------------------------------------------------------- |
| Resource disk | The VM's temporary (resource) disk.                                                                                      |
| Cache disk    | The VM's premium storage cache.                                                                                          |
| NVMe disk     | The VM's local NVMe disk.                                                                                                |
| AKS default   | Azure picks the placement per VM SKU. Atlas resolves it in this order: cache disk, then resource disk, then NVMe disk. |

The cache type is **Read Only**. Atlas deploys every ephemeral OS disk in Azure's partial caching mode with read-only caching. Full caching mode is not available through Atlas.

#### Instance type filtering

Enabling Ephemeral OS disk filters the instance types the Autoscaler can pick. Only instance types that support the chosen placement stay eligible. Atlas excludes instance types without a matching local disk. This filtering applies to every node the node configuration provisions, because the OS disk setting applies to all of them.

The breadth of filtering depends on the placement:

* **Resource disk** and **NVMe disk** need a local disk on the VM. Only instance types with local storage stay eligible.
* **Cache disk** works on VMs with premium storage cache but no local data disk. It does not limit nodes to local-storage instance types.

Ephemeral OS disk filtering does not add the storage-optimized label or taint. To get nodes with local disks but no `scheduling.atlas.example/storage-optimized=true` label or taint, use a resource disk or NVMe disk placement, not the [storage-optimized nodes](storage-optimized-nodes.md) constraint.

#### Disk size limit

The node's [Initial disk size](#shared-configuration-options) sets the requested size of the OS disk (the default is 100 GiB). Each VM SKU allows a maximum OS disk size per placement, and this maximum comes from Azure and varies by SKU. Atlas cannot extend it. Cache disk placements usually allow a smaller size than resource disk placements. Leave Initial disk size at the default unless you need a larger OS disk. If you set Initial disk size larger than a SKU's maximum, nodes of that SKU fail to provision.

> ⚠️ Warning
>
> Ephemeral OS disk has these Azure limits:
>
> * It has no fallback to a managed disk. If a VM SKU lacks the chosen placement, Atlas does not provision that SKU.
> * Keep Initial disk size within the placement's per-SKU maximum. A larger Initial disk size fails to provision.
> * You cannot stop and start (deallocate) a node that uses an ephemeral OS disk.

For more on Azure ephemeral OS disks, placement options, and per-SKU limits, see the <Anchor target="_blank" href="https://learn.microsoft.com/en-us/azure/virtual-machines/ephemeral-os-disks">Azure ephemeral OS disks documentation</Anchor>.

### Kubelet configuration for AKS

You can customize kubelet settings on AKS nodes by providing a kubelet configuration in your Node Configuration. The configuration applies to all nodes provisioned under that Node Configuration. If no kubelet config is provided, AKS defaults are used.

Only the kubelet settings that are supported by AKS are supported in Node Configuration. For the full list of supported parameters and their allowed values, see the <Anchor target="_blank" href="https://learn.microsoft.com/en-us/azure/aks/custom-node-configuration?tabs=linux-node-pools#supported-custom-configuration-parameters">Azure Kubernetes Service (AKS) custom node configuration</Anchor> documentation.

### AKS image family

The OS family used for provisioning nodes. Possible values:

* `FAMILY_UNSPECIFIED`
* `FAMILY_UBUNTU` — Atlas auto-picks the Ubuntu version: Ubuntu 22.04 for k8s ≤ 1.35, Ubuntu 24.04 for k8s ≥ 1.36.
* `FAMILY_UBUNTU_2204` — Ubuntu 22.04, supported on k8s 1.25–1.36.
* `FAMILY_UBUNTU_2404` — Ubuntu 24.04, supported on k8s 1.32+.
* `FAMILY_AZURE_LINUX` — Azure Linux 3.0, available from AKS v1.32 onward.
* `FAMILY_WINDOWS_2019`
* `FAMILY_WINDOWS_2022`
* `FAMILY_WINDOWS_2025` — supported on k8s 1.32+.

**Default:** `FAMILY_UBUNTU`

> 📘 Note
>
> When verifying the OS version of a provisioned node, do not rely on the Kubernetes label — log in to the node and check it directly.

### Pod subnet configuration for AKS

To configure a custom pod subnet for a node configuration:

1. **Via Terraform**: Use the `pod_subnet_id` parameter in your Atlas AKS Terraform configuration:

```toml Short
[...]
pod_subnet_id = azurerm_subnet.pod_subnet_1.id
[..]
```
```toml Full
module "atlas-aks-cluster" {
  source = "atlas/aks/atlas"
  
  aks_cluster_name    = var.aks_cluster_name
  aks_cluster_region  = var.aks_cluster_region
  node_resource_group = azurerm_kubernetes_cluster.example.node_resource_group
  resource_group      = azurerm_kubernetes_cluster.example.resource_group_name
  subscription_id     = data.azurerm_subscription.current.subscription_id
  tenant_id           = data.azurerm_subscription.current.tenant_id
  
  node_configurations = {
    default = {
      disk_cpu_ratio = 25
      subnets        = [azurerm_subnet.internal.id]
      pod_subnet_id  = azurerm_subnet.pod_subnet_1.id
      tags = {
        "node-config" = "default"
      }
    }
    custom_pod_subnet = {
      disk_cpu_ratio = 25
      subnets        = [azurerm_subnet.internal.id]
      pod_subnet_id  = azurerm_subnet.pod_subnet_2.id
      tags = {
        "node-config" = "custom-pod-subnet"
      }
    }
  }
  
  # other configuration...
}
```

2. **Via API**: Include the `pod_subnet_id` field in your node configuration API requests

### Load balancer attachment

A list of load balancers to attach nodes to. Setting this field disables Atlas's default load balancer auto-discovery.

Each entry accepts the following fields:

* `id` — The full ID of the Azure load balancer.
* `ipBasedBackendPools` — A list of IP-based backend pool names for attaching each node's IP.
* `nicBasedBackendPools` — A list of NIC-based backend pool names for attaching each node's NIC.

### Accelerated networking for AKS

By default, Atlas provisions AKS nodes with <Anchor target="_blank" href="https://learn.microsoft.com/en-us/azure/virtual-network/accelerated-networking-overview">accelerated networking</Anchor> enabled whenever the underlying VM SKU supports it. Accelerated networking offloads packet processing to the host hardware (SR-IOV), reducing latency and CPU overhead.

You can opt out per node configuration by setting **Accelerated networking** to `Disabled`. This forces accelerated networking off on every node provisioned from that configuration, even when the SKU supports it. Use this only when you have a specific reason to disable it — for example, compatibility constraints with custom networking stacks or third-party CNIs.

To disable accelerated networking:

* **API**: Set the `acceleratedNetworking` parameter to `ACCELERATED_NETWORKING_MODE_DISABLED` in the `aks` block when <Anchor target="_blank" href="https://docs.atlas.example/reference/nodeconfigurationapi_createconfiguration">creating a node configuration</Anchor>. To follow the default behavior, leave this parameter undefined.

```json
{
  "aks": {
    "acceleratedNetworking": "ACCELERATED_NETWORKING_MODE_DISABLED"
  }
}
```

* **Terraform**: Set `accelerated_networking = "disabled"` inside the `aks` block of your `atlas_node_configuration` resource:

```toml Terraform
resource "atlas_node_configuration" "default" {
  name       = "default"
  cluster_id = atlas_aks_cluster.example.id
  subnets    = [azurerm_subnet.internal.id]

  aks {
    max_pods_per_node      = 30
    accelerated_networking = "disabled"
  }
}
```

When the field is omitted (or set to an empty value), Atlas sends no preference to the API and the default behavior applies.

## Virtual machine image choice

When Atlas provisions a node, it must choose an appropriate VM image. This choice is crucial because the OS and image version determine the correct bootstrapping logic and instance type support, and are critical to ensuring the node joins the cluster successfully. For advanced use cases, Atlas offers several options in the node configuration.

### EKS

EKS supports a combination of the Image and Image Family fields to control OS choice.

* **Image family**: Determines the provisioning logic based on OS. If not provided, a default family is used for all operations (see table above for defaults). Some init scripts require selecting AMI Family explicitly. For example, if one wanted to provide a `nodeadm` config via MIME multipart settings, one would need to set AMI Family to Amazon Linux 2023. **For Container Live Migration,**`FAMILY_AL2023`**&#x20;is required**.
* **Image**: Used to determine the actual image choice more precisely. The field accepts one of three forms:

1. **AMI ID** (e.g., `ami-1234567890abcdef0`): Pins provisioning to that exact AMI. The value must be exactly `ami-` followed by 17 lowercase alphanumeric characters — any other format is treated as a search string. If the AMI architecture does not match the instance type, provisioning will fail. Use architecture restrictions in the [Node template](node-templates.md) to avoid this scenario. The AMI must match the image family (default or provided value), or provisioning will fail.
2. **Search string** (e.g., `amazon-eks-node-*`): Matched against the `name` filter in [aws describe-images](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-images.html). Supports `*` and `?` wildcards. When multiple images match, Atlas selects the newest by creation date, preferring images whose name contains the cluster's Kubernetes version. Architecture is filtered automatically based on the instance type — you do not need to include an architecture segment in the search string for standard (non-GPU, non-Neuron) nodes. If no images match the instance type architecture or the images are from a different family than the Image family field, provisioning will fail.
3. **Empty**: A default search is performed based on the Image family. This search targets public Amazon-owned images and selects the latest image matching the instance type architecture and Kubernetes version.

> ⚠️ Warning
>
> The Image field accepts one value at a time — either a single AMI ID or a single search string. Entering multiple values separated by commas is not supported and will silently return zero results. To cover multiple architectures with a pinned image version, create a separate node configuration for each architecture and use node template architecture constraints to direct traffic to the correct one.

#### How Atlas selects an image from a search string

When you provide a search string, Atlas passes it verbatim as the `name` filter to the AWS DescribeImages API, then applies the following selection logic to the results:

1. All returned images are sorted newest-first by creation date.
2. Images that do not match the provisioned node's architecture are discarded.
3. The first remaining image whose name contains the cluster's Kubernetes version string (e.g., `1.32`) is selected.
4. If no image matches the Kubernetes version, the newest architecture-compatible image is used and a warning is logged.

This means a search string that matches more variants than intended (for example, one that matches both standard and GPU images) will not cause an error — Atlas picks the newest architecture-compatible result. However, if your wildcard is too broad, you may get GPU or FIPS variants when you intended standard images. Make your search string specific enough to target only the intended image type.

#### Name string examples

The following patterns illustrate how to construct search strings for common EKS image types. These are name filter strings, not AMI IDs.

| OS           | Goal                                                                                                   | Image field value                                         |
| :----------- | :----------------------------------------------------------------------------------------------------- | :-------------------------------------------------------- |
| AL2023       | Latest for current k8s version, x86\_64 (equivalent to empty field + `FAMILY_AL2023` on x86\_64 nodes) | `amazon-eks-node-al2023-x86_64-standard-{k8s_version}-v*` |
| AL2023       | Latest, architecture resolved automatically by Atlas                                                 | `amazon-eks-node-al2023-*-standard-{k8s_version}-v*`      |
| Bottlerocket | Latest stable, x86\_64                                                                                 | `bottlerocket-aws-k8s-{k8s_version}-x86_64-*`             |
| Bottlerocket | Pin to a specific release (version-locked, any build hash)                                             | `bottlerocket-aws-k8s-{k8s_version}-x86_64-v1.57.0-*`     |
| Bottlerocket | Pin to a specific release, both architectures covered automatically                                    | `bottlerocket-aws-k8s-{k8s_version}-*-v1.57.0-*`          |
| AL2          | Latest (AL2 reaches end of support with k8s 1.33)                                                      | `amazon-eks-node-{k8s_version}-*`                         |

> 📘 Note
>
> The Bottlerocket image name does not embed the Kubernetes version in the same position as AL2/AL2023 names. Atlas applies k8s version preference logic only when the version string appears in the image name — for Bottlerocket, it appears in the pattern segment (e.g., `k8s-1.32`) and is matched correctly.

#### GPU and Neuron variants

When the Image field is **empty**, Atlas automatically selects the appropriate GPU or Neuron image variant based on the instance type. When you provide a search string or AMI ID, this automatic variant selection does not apply — the value you provide is used directly. If you set an explicit Image field on a node configuration used by GPU or Neuron instance types, ensure your search string or AMI ID targets the correct image variant (e.g., `nvidia` or `neuron` images for those workloads).

#### Custom and third-party AMIs

Atlas determines the correct node bootstrapping logic by detecting the OS family from the resolved image name. If the image name does not match a known pattern (for example, a fully custom or hardened AMI), Atlas falls back to the default family for the cluster's Kubernetes version (AL2023 for 1.33+, AL2 for 1.32 and below). If that fallback does not match the actual OS of your AMI, the generated userdata may be incorrect and the node may fail to join the cluster. Set the Image family field explicitly when using custom AMIs to avoid this.

Sample scenarios and suggested configuration:

<Accordion title="Hands-off approach — let Atlas choose">
  Leave both the Image and Image family fields empty. Atlas selects the latest compatible image for the cluster's Kubernetes version and instance architecture.
</Accordion>

<Accordion title="Use a specific OS family, let Atlas choose the image">
  Select an Image family. Leave the Image field empty. Atlas queries Amazon-owned images and picks the latest match for the instance architecture and Kubernetes version.
</Accordion>

<Accordion title="Enable Container Live Migration">
  Set Image family to `FAMILY_AL2023`. Leave the Image field empty. Set Container runtime to `containerd`.
</Accordion>

<Accordion title="Pin to a specific OS release version, let Atlas handle architecture">
  Add a version-specific search string to the Image field (for example, `bottlerocket-aws-k8s-{k8s_version}-*-v1.57.0-*`). Set Image family if different from the default.
</Accordion>

<Accordion title="Use private or third-party AMIs, architecture resolved automatically">
  Add a search string to the Image field that matches the required images. Set Image family if different from the default. For multi-architecture node configurations, the search string must match images for both arm64 and x86.
</Accordion>

<Accordion title="Use private or third-party AMIs without architecture-agnostic builds">
  Add a search string to the Image field. Set Image family if different from the default. Add architecture constraints to the node template to prevent mismatches.
</Accordion>

<Accordion title="Pin to a specific golden AMI">
  Enter the AMI ID in the Image field. Select the Image family that matches the OS. Add architecture constraints to the node template.
</Accordion>

<Accordion title="Support multiple architectures with a pinned image version">
  Create separate node configurations — one per architecture — each with an architecture-specific search string or AMI ID. Use node template architecture constraints to direct provisioning to the correct configuration. Comma-separated values are not supported.
</Accordion>

### GKE

For GKE, the image field can be used to control the node bootstrapping logic (for Linux).

* The reference must point to a specific image.
* If the image does not match the instance type architecture (for example, an ARM64 image for an x86 node), node provisioning will fail.
* Changing the value might require a successful reconciliation to recreate Atlas-owned node pools.
* If an image is not provided, the default behavior is to use the OS image captured when creating the `atlaspool` node pool.

<Callout icon="📘" theme="info">
  ### Note

  Atlas does not officially support selecting an OS family (for example, Ubuntu vs Container-Optimized OS) on GKE. Unlike EKS and AKS, GKE does not expose a family-based image selector — images live in a private GCP project with no wildcard or family-based resolution. If you need a specific OS family, use one of the options below.
</Callout>

#### Choosing an OS for GKE nodes

<Accordion title="Use Ubuntu for all Atlas-provisioned nodes">
  Pin a specific Ubuntu image path in the Image field (for example, `projects/ubuntu-os-gke-cloud/global/images/ubuntu-gke-2404-1-35-amd64-v20260416`). After a control plane upgrade or new image publish, update the reference manually.
</Accordion>

<Accordion title="Use Ubuntu without manual image maintenance">
  Delete the existing `atlaspool` node pool in GKE, recreate it selecting Ubuntu as the node image, then trigger a Atlas reconcile. Atlas reads the OS image from the pool and uses it for all nodes it provisions going forward.
</Accordion>

<Accordion title="Keep the default (Container-Optimized OS)">
  Leave the Image field empty. Atlas uses the OS image from the `atlaspool` node pool.
</Accordion>

<Callout icon="📘" theme="info">
  ### Note

  If you need Ubuntu to support btrfs (for example, for Atlas disk resizing), use the atlaspool recreation approach. Container-Optimized OS does not support btrfs natively.
</Callout>

### AKS

Custom VM images for AKS are not supported.

## How to create a node configuration

A default node configuration is created during cluster onboarding in the Atlas-managed mode.

You can choose to modify this configuration or create a new one. If you add a new node configuration that will be applied to all newly provisioned nodes, you must mark it as default.

Node configurations are versioned, and when the Atlas provisioner adds a new node, the latest version of the node configuration is applied.

A new configuration can't be applied to an existing node. If you want to upgrade node configuration on a node or a set of nodes, you need to delete an existing node and wait until Autoscaler replaces it with a new one or [rebalance](rebalancing.md) the cluster (fully or partially).

### Kubelet configuration examples

The Kubernetes documentation—[Kubelet Configuration](https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/#kubelet-config-k8s-io-v1beta1-KubeletConfiguration) —contains all available Kubelet settings. Please refer to the version of your cluster.

For example, if you want to add some specific custom taints during node startup, you could do it with the following snippet:

```json
{
    "registerWithTaints": [
        {
            "effect": "NoSchedule",
            "key": "nodes-service-critical",
            "value": "true"
        }
    ]
}
```

<Callout icon="⚠️" theme="warn">
  ### Use Node Template taints instead of init script taints

  If you set taints via `--register-with-taints` kubelet flags in an init script, be aware that kubelet treats duplicate `--register-with-taints` flags by keeping only the last value. **Earlier taints are silently dropped**. This means Atlas's own provisioner taints (such as `scheduling.atlas.example/spot`) can be overridden by a later `--register-with-taints=` in your init script.

  **Recommendation:** Apply custom taints through [Node Templates](node-templates.md) rather than init script flags. Node Template taints are applied reliably alongside Atlas's own provisioner taints without risk of override.
</Callout>

The second example involves configuring the`kubelet` image pulling and setting kube API limits like the following:

```json
{
    "eventBurst": 20,
    "eventRecordQPS": 10,
    "kubeAPIBurst": 20,
    "kubeAPIQPS": 10,
    "registryBurst": 20,
    "registryPullQPS": 10
}
```

Since `kubeReserved` is not supported on **Amazon Linux 2 (AL2)** nodes, `systemReserved` can be configured instead in the **Kubelet configuration** under **Advanced Settings** in **NodeConfig**, as shown in the following example:

```json
{
    "systemReserved": {
        "cpu": "500m",
        "memory": "1Gi"
    }
}
```

### Create node configuration with the Atlas Terraform provider

Use the resource `atlas_node_configuration` from [Atlas terraform provider](https://github.com/atlas/terraform-provider-atlas).

Reference example:

```hcl Terraform
resource "atlas_node_configuration" "test" {
  name           = local.name
  cluster_id     = atlas_eks_cluster.test.id
  disk_cpu_ratio = 5
  subnets        = aws_subnet.test[*].id
  tags           = {
    env = "development"
  }
  eks {
    instance_profile_arn = aws_iam_instance_profile.test.arn
    dns_cluster_ip       = "10.100.0.10"
    security_groups      = [aws_security_group.test.id]
  }
}

 
```

## Container Live Migration requirements

When configuring nodes for [Container Live Migration](clm-overview.md) on EKS, specific infrastructure requirements must be met to ensure successful pod migration between nodes.

### Image family requirements

**Amazon Linux 2023 required**: Nodes must use the Amazon Linux 2023 image family (`FAMILY_AL2023`) for Container Live Migration compatibility. Other image families are not supported.

```json
{
  "imageFamily": "FAMILY_AL2023"
}
```

### Container runtime requirements

**Containerd v2+ required**: Nodes must use `containerd` as the container runtime engine. Docker and other runtimes are not supported for live migration due to specific integration requirements for checkpoint and restore functionality.

```json
{
  "containerRuntime": "containerd"
}
```

<br />