---
updatedAt: 2026-06-02T13:49:21.000Z
---

# AWS PrivateLink

# Connecting to Atlas through AWS PrivateLink

Atlas PrivateLink enables EKS clusters without internet access to securely communicate with Atlas using <Anchor label="AWS PrivateLink" target="_blank" href="https://aws.amazon.com/privatelink/">AWS PrivateLink</Anchor> technology. This solution is designed for clusters where nodes cannot reach the public internet through NAT gateways or similar methods.

> 📘 Reference Implementation
>
> This guide is based on the <Anchor label="Atlas PrivateLink AWS repository" target="_blank" href="https://github.com/atlas/privatelink-aws">Atlas PrivateLink AWS repository</Anchor>, which contains the latest Terraform examples, implementation details, and updates. Refer to the repository for the most current information and code samples.

## Prerequisites

* An EKS cluster in a VPC with private subnets
* Subnets tagged with `atlas.example/routable = true` to indicate they can communicate with Atlas
* Kubectl access to your cluster (either direct or through a bastion host)
* A private AWS ECR for hosting Atlas container images
* Terraform installed (if using the Infrastructure as Code (IaC) approach)

## Architecture

Atlas PrivateLink uses VPC endpoints to establish private connectivity between your EKS cluster and the Atlas platform. This connection allows Atlas agents to communicate with the central system without traversing the public internet.

<Image align="center" alt="AWS PrivateLink Architecture" border={false} src="https://files.readme.io/e27ec92e2390e3843c8a1ee01cbfbd304e418026d2363f33eb9694d9cb73a5c7-image.png" />

## Setup Process

### Step 1: Tag your subnets

Add the `atlas.example/routable` tag to all subnets where your EKS cluster nodes will run:

```shell
aws ec2 create-tags \
  --resources subnet-xxxxx \
  --tags Key=atlas.example/routable,Value=true
```

This tag informs Atlas that these subnets have connectivity to the Atlas platform (through the VPC endpoints you'll create in the next step).

For more information on tagging AWS resources, see <Anchor label="Tag your Amazon EC2 resources" target="_blank" href="https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html">Tag your Amazon EC2 resources</Anchor> in the AWS documentation.

### Step 2: Create VPC endpoints

Create VPC endpoints in your VPC to enable private access to Atlas services. The required endpoints vary by region.

Choose your region and create the corresponding VPC endpoints.

> 📘 India region endpoint service names
>
> For connecting to the Atlas India region via PrivateLink, contact Atlas Customer Success to obtain the VPC endpoint service names for your AWS region.

#### US East (N. Virginia) - us-east-1

| Endpoint Type         | Service Name                                              |
| --------------------- | --------------------------------------------------------- |
| REST API              | `com.amazonaws.vpce.us-east-1.vpce-svc-0f648001d494b9a46` |
| gRPC for Pod Pinner   | `com.amazonaws.vpce.us-east-1.vpce-svc-05e9e206a737781a7` |
| API via gRPC          | `com.amazonaws.vpce.us-east-1.vpce-svc-00d55140e30124b2f` |
| Downloading artifacts | `com.amazonaws.vpce.us-east-1.vpce-svc-0b611f59ee5494cc3` |

#### US East (Ohio) - us-east-2

| Endpoint Type         | Service Name                                              |
| --------------------- | --------------------------------------------------------- |
| REST API              | `com.amazonaws.vpce.us-east-2.vpce-svc-0504dcd21f12037fe` |
| gRPC for Pod Pinner   | `com.amazonaws.vpce.us-east-2.vpce-svc-0681d838ee3b5e496` |
| API via gRPC          | `com.amazonaws.vpce.us-east-2.vpce-svc-0678ed50f02956c7f` |
| Downloading artifacts | `com.amazonaws.vpce.us-east-2.vpce-svc-0d1fad36f0230d84c` |

#### Asia Pacific (Mumbai) - ap-south-1

| Endpoint Type         | Service Name                                               |
| --------------------- | ---------------------------------------------------------- |
| REST API              | `com.amazonaws.vpce.ap-south-1.vpce-svc-07c813d8840493bcf` |
| gRPC for Pod Pinner   | `com.amazonaws.vpce.ap-south-1.vpce-svc-028ad7f3e7879c47c` |
| API via gRPC          | `com.amazonaws.vpce.ap-south-1.vpce-svc-0dd30c956893d640e` |
| Downloading artifacts | `com.amazonaws.vpce.ap-south-1.vpce-svc-0c4777e1cc92bf8a8` |

#### Asia Pacific (Hyderabad) - ap-south-2

| Endpoint Type         | Service Name                                               |
| --------------------- | ---------------------------------------------------------- |
| REST API              | `com.amazonaws.vpce.ap-south-2.vpce-svc-0a2a5307bb44fab88` |
| gRPC for Pod Pinner   | `com.amazonaws.vpce.ap-south-2.vpce-svc-00d1d87c025531c96` |
| API via gRPC          | `com.amazonaws.vpce.ap-south-2.vpce-svc-009c9d91819071294` |
| Downloading artifacts | `com.amazonaws.vpce.ap-south-2.vpce-svc-05b723b65a7f5a0f9` |

> 📘 Note
>
> If your region is not listed, contact Atlas support to confirm availability and obtain the VPC endpoint service names.

#### Using Terraform

Refer to the [Atlas PrivateLink AWS repository](https://github.com/atlas/privatelink-aws) for Terraform examples that automate VPC endpoint creation.

#### Using AWS CLI

Create each VPC endpoint using the AWS CLI:

```shell
aws ec2 create-vpc-endpoint \
  --vpc-id vpc-xxxxx \
  --vpc-endpoint-type Interface \
  --service-name com.amazonaws.vpce.us-east-1.vpce-svc-0f648001d494b9a46 \
  --subnet-ids subnet-xxxxx subnet-yyyyy \
  --security-group-ids sg-xxxxx
```

Repeat this command for each VPC endpoint service listed for your region.

For detailed information on creating VPC endpoints, see <Anchor label="Create an interface endpoint" target="_blank" href="https://docs.aws.amazon.com/vpc/latest/privatelink/create-interface-endpoint.html">Create an interface endpoint</Anchor> in the AWS documentation.

### Step 3: Set up a private container registry

Since your cluster cannot access public container registries, copy all required Atlas images to your private AWS ECR:

1. Create a private ECR repository accessible from your cluster
2. Copy Atlas images from public repositories to your private ECR
3. Ensure your EKS nodes have the necessary IAM permissions to pull images from your private ECR

> 📘 Note
>
> Atlas is developing a script to automate the image copying process. Contact support for the latest tooling.

For guidance on setting up ECR and configuring permissions, see <Anchor label="Getting started with Amazon ECR" target="_blank" href="https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html">Getting started with Amazon ECR</Anchor> in the AWS documentation.

### Step 4: Connect your cluster

Generate the onboarding script from the Atlas Console and run it to deploy the Atlas agent:

1. Navigate to the Atlas Console and <Anchor label="start the cluster connection process" target="_blank" href="doc:connect-via-console">start the cluster connection process</Anchor>
2. Generate and run the onboarding script as usual
3. After the agent is deployed, update the deployment to use images from your private ECR:

```shell
kubectl set image deployment/atlas-agent \
  atlas-agent=<your-ecr-url>/atlas-agent:<version> \
  -n atlas-agent
```

Repeat this step for all Atlas components deployed in your cluster.

## Maintenance

When upgrading Atlas agents:

1. Pull the new agent images from Atlas's public registry
2. Push them to your private ECR
3. Update the deployments in your cluster to use the new image versions

Since Atlas cannot directly access your private ECR, you must manage agent upgrades manually.

## Additional Resources

* [Atlas PrivateLink AWS repository](https://github.com/atlas/privatelink-aws) – Terraform examples and additional guidance
* [AWS PrivateLink documentation](https://aws.amazon.com/privatelink/)