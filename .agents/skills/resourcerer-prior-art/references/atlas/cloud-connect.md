---
updatedAt: 2026-10-27T10:52:49.000Z
---

# Cloud Connect

Discover and inventory cloud resources across your entire cloud account without installing any agents.

Cloud Connect is an agentless, organization-level integration that gives Atlas read-only visibility into your cloud account. Instead of connecting clusters one by one, Cloud Connect discovers your Kubernetes clusters, virtual machines, and other cloud resources in a single setup — and keeps that inventory up to date with hourly syncs. The exact resources discovered depend on your cloud provider and the permission scope you choose.

If you need a one-time, read-only collection instead of a persistent integration, see [Take a snapshot of your AWS account](cloud-snapshot.md).

Cloud Connect creates IAM trust relationships in your cloud account (IAM roles, service accounts, or app registrations). Atlas uses these trust relationships to issue short-lived credentials on demand — no long-lived secrets are stored.

## How Cloud Connect works

Setting up Cloud Connect involves three steps:

1. **Select your provider and permission scope.** Choose which cloud provider to integrate and which [permission scope](#permission-scopes) matches your security requirements.
2. **Create IAM resources.** Run the generated setup script (or apply the equivalent configuration via the Atlas API) to create the necessary IAM roles, service accounts, or app registrations in your cloud account.
3. **Automatic sync begins.** Atlas starts discovering your cloud resources using short-lived credentials issued through the IAM trust relationship. Discovered clusters appear with a **Discovered** status in the console.

After the initial setup, Atlas syncs resources automatically every 60 minutes.

You can set up Cloud Connect in several ways:

* **Atlas console** — guided flow with a generated setup script. See [Connect using the Atlas console](connect-via-console.md#discover-clusters-with-cloud-connect).
* **Terraform** — declarative modules for <Anchor label="AWS" target="_blank" href="https://registry.terraform.io/modules/atlas/aws-cloud-connect/atlas/latest">AWS</Anchor> and <Anchor label="GCP" target="_blank" href="https://registry.terraform.io/modules/atlas/gcp-cloud-connect/atlas/latest">GCP</Anchor>.
* **Atlas API** — direct API calls for programmatic setup.

## Supported providers

Cloud Connect supports AWS, GCP, and Azure. Each provider differs in the scopes available and the resource types that get synced.

| Provider | Scopes                | Resources synced                                                                                                                                                                                              |
| -------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AWS      | Account, Organization | EKS clusters, EC2 instances, RDS databases, load balancers, Elastic IPs, Savings Plans, Reserved Instances, Capacity Reservations. Optionally: SageMaker, Bedrock, and other AI/ML services; CUR bucket data. |
| GCP      | Project, Organization | GKE clusters, Compute Engine VMs, persistent disks, Cloud SQL instances, CUDs, Flex CUDs, Capacity Reservations.                                                                                              |
| Azure    | Subscription          | AKS clusters, virtual machines, VM scale sets, Reserved Instances, Savings Plans.                                                                                                                             |

**Organization-level scopes** (AWS Organization and GCP Organization) let you discover resources across all accounts or projects in your organization from a single integration. Azure supports subscription-level scoping only.

## Permission scopes

When you set up Cloud Connect, you choose a permission scope that controls how much read access Atlas gets. All scopes are read-only — Cloud Connect never modifies your resources.

| Scope               | What it covers                                                                                | Available for            |
| ------------------- | --------------------------------------------------------------------------------------------- | ------------------------ |
| Full Access         | Broad read-only access across all services                                                    | AWS, GCP, Azure          |
| Minimal Permissions | Targeted access to core compute, Kubernetes, and database services                            | AWS, GCP, Azure          |
| Commitments         | Access to commitment and savings data (Reserved Instances, Savings Plans, CUDs)               | AWS, GCP, Azure          |
| AI Services         | List and describe permissions for AI/ML services (Bedrock, SageMaker, Comprehend, and others) | AWS                      |
| CUR Bucket          | Access to Cost and Usage Report S3 bucket                                                     | AWS (account scope only) |

You can combine multiple scopes during setup. For the full IAM policies, roles, and permissions for each scope, see [Permissions for Cloud Connect](cloud-permissions.md#permissions-for-cloud-connect).

## What happens to discovered clusters

Clusters found by Cloud Connect appear in the Atlas console with a **Discovered** status. For each discovered cluster, you can see basic metadata and optimization potential — without installing any components.

To start optimizing a discovered cluster, select it in the console and click **Connect**. This starts the standard connection flow described in [Connect using the Atlas console](connect-via-console.md).

## Cloud Connect and commitments

Cloud Connect also serves as the sync mechanism for cloud commitment data. When configured with the Commitments permission scope, Atlas automatically imports and refreshes your Reserved Instances, Savings Plans, CUDs, and Flex CUDs every 60 minutes.

This enables the Atlas autoscaler to factor your existing commitments into node provisioning decisions. For details, see [Commitments](commitments.md).

## Related resources

<Cards columns={2}>
  <Card title="Set up Cloud Connect" href="https://docs.atlas.example/docs/connect-via-console#discover-clusters-with-cloud-connect" icon="fa-play">
    Step-by-step setup instructions from the Atlas console.
  </Card>

  <Card title="Cloud Connect permissions" href="https://docs.atlas.example/docs/cloud-permissions#permissions-for-cloud-connect" icon="fa-shield">
    Full IAM policies, roles, and permission details for each provider and scope.
  </Card>
</Cards>