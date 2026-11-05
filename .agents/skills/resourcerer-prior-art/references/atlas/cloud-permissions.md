---
updatedAt: 2026-09-22T11:35:10.000Z
---

# Cloud permissions

Permissions used by Cloud Providers (AWS/GCP/Azure/Oracle)

# Permissions setup used by cloud providers

Atlas requires different permission sets depending on the integration type: monitoring, cluster onboarding and automation, *Cloud Connect* resource discovery, *Discover cluster* one-time discovery, or Commitments data import.

## Monitoring permissions

Comprehensive monitoring of a cluster requires a holistic approach to metrics, including access to various information available only through a cloud API. Atlas Sentinel, among many features, provides both ephemeral and persistent storage monitoring.
To properly track Persistent Volumes usage and cost, Sentinel controller requires cloud API access to inspect metadata (e.g., disk size, type, provisioned IOPS and throughput, location) of cloud disks backing these volumes.

### AWS

Cloud permissions for monitoring AWS EBS volumes are granted through the IRSA (IAM Roles for Service Accounts). Inline policy used in the IRSA role is as follows:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeVolumes"
      ],
      "Resource": "*"
    }
  ]
}
```

## Permissions for cluster onboarding and automation

When a cluster enters automated cost optimization mode, Atlas performs operations at the cloud provider level, such as adding or removing nodes. These operations require specific credentials and permissions.

### AWS

> 📘 Note
>
> Atlas strongly advises creating a dedicated IAM role and instance profile for Atlas-managed nodes.
>
> Reusing IAM roles with EKS Managed Node Groups is not recommended and can lead to unexpected service outages if AWS deletes the profile and role when the last managed node group is removed.

#### AWS permissions with access granted using a cross-account IAM role

When enabling cost optimization for a connected cluster, you grant permissions using a cross-account IAM role.

This feature creates a dedicated cluster user in the Atlas AWS account with a trust policy that can assume the role defined in your AWS account.

Keeping role definitions and users in separate AWS accounts enables storing the user's credentials on the Atlas side without handing them over when running the onboarding script, which improves security levels.

You can verify the set of permissions and trust relationship on your IAM role using the following commands:

<Accordion title="Verify IAM role permissions">
  ```shell
  aws iam get-role --role-name <role name>
  aws iam list-attached-role-policies --role-name <role name>
  aws iam list-role-policies --role-name <role name>
  ```
</Accordion>

The following AWS-managed permissions are required for Atlas to work:

* [AmazonEC2ReadOnlyAccess](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonEC2ReadOnlyAccess.html)
* [IAMReadOnlyAccess](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/IAMReadOnlyAccess.html)

In addition, the following custom policy is necessary:

<Accordion title="Custom IAM policy for Atlas">
  ```json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Sid": "PassRoleEC2",
              "Action": "iam:PassRole",
              "Effect": "Allow",
              "Resource": "arn:aws:iam::*:role/*",
              "Condition": {
                  "StringEquals": {
                      "iam:PassedToService": "ec2.amazonaws.com"
                  }
              }
          },
          {
              "Sid": "PassRoleEKS",
              "Action": "iam:PassRole",
              "Effect": "Allow",
              "Resource": "arn:aws:iam::*:role/*",
              "Condition": {
                  "StringEquals": {
                      "iam:PassedToService": "eks.amazonaws.com"
                  }
              }
          },
          {
              "Sid": "NonResourcePermissions",
              "Effect": "Allow",
              "Action": [
                  "iam:CreateServiceLinkedRole",
                  "ec2:CreateKeyPair",
                  "ec2:DeleteKeyPair",
                  "ec2:CreateTags",
                  "ec2:ImportKeyPair"
              ],
              "Resource": "*"
          },
          {
              "Sid": "RunInstancesPermissions",
              "Effect": "Allow",
              "Action": "ec2:RunInstances",
              "Resource": [
                  "arn:aws:ec2:*:<account-id>:network-interface/*",
                  "arn:aws:ec2:*:<account-id>:security-group/*",
                  "arn:aws:ec2:*:<account-id>:volume/*",
                  "arn:aws:ec2:*:<account-id>:key-pair/*",
                  "arn:aws:ec2:*::image/*"
              ]
          },
          {
              "Sid": "RunInstancesTagRestriction",
              "Effect": "Allow",
              "Action": "ec2:RunInstances",
              "Resource": "arn:aws:ec2:<region>:<account-id>:instance/*",
              "Condition": {
                  "StringEquals": {
                      "aws:RequestTag/kubernetes.io/cluster/<cluster-name>": "owned"
                  }
              }
          },
          {
              "Sid": "RunInstancesVpcRestriction",
              "Effect": "Allow",
              "Action": "ec2:RunInstances",
              "Resource": "arn:aws:ec2:<region>:<account-id>:subnet/*",
              "Condition": {
                  "StringEquals": {
                      "ec2:Vpc": "arn:aws:ec2:<region>:<account-id>:vpc/<vpc-id>"
                  }
              }
          },
          {
              "Sid": "InstanceActionsTagRestriction",
              "Effect": "Allow",
              "Action": [
                  "ec2:TerminateInstances",
                  "ec2:StartInstances",
                  "ec2:StopInstances",
                  "ec2:CreateTags"
              ],
              "Resource": "arn:aws:ec2:<region>:<account-id>:instance/*",
              "Condition": {
                  "StringEquals": {
                      "ec2:ResourceTag/kubernetes.io/cluster/<cluster-name>": [
                          "owned",
                          "shared"
                      ]
                  }
              }
          },
          {
              "Sid": "AutoscalingActionsTagRestriction",
              "Effect": "Allow",
              "Action": [
                  "autoscaling:UpdateAutoScalingGroup",
                  "autoscaling:CreateAutoScalingGroup",
                  "autoscaling:DeleteAutoScalingGroup",
                  "autoscaling:SuspendProcesses",
                  "autoscaling:ResumeProcesses",
                  "autoscaling:TerminateInstanceInAutoScalingGroup"
              ],
              "Resource": "arn:aws:autoscaling:<region>:<account-id>:autoScalingGroup:*:autoScalingGroupName/*",
              "Condition": {
                  "StringEquals": {
                      "autoscaling:ResourceTag/kubernetes.io/cluster/<cluster-name>": [
                          "owned",
                          "shared"
                      ]
                  }
              }
          },
          {
              "Sid": "EKS",
              "Effect": "Allow",
              "Action": [
                  "eks:Describe*",
                  "eks:List*",
                  "eks:TagResource",
                  "eks:UntagResource",
                  "eks:CreateNodegroup",
                  "eks:DeleteNodegroup"
              ],
              "Resource": [
                  "arn:aws:eks:<region>:<account-id>:cluster/<cluster-name>",
                  "arn:aws:eks:<region>:<account-id>:nodegroup/<cluster-name>/*/*"
              ]
          },
          {
            "Sid": "CreateLaunchTemplateWithTag",
            "Effect": "Allow",
            "Action": "ec2:CreateLaunchTemplate",
            "Resource": "*",
            "Condition": {
              "StringEquals": {
                "aws:RequestTag/kubernetes.io/cluster/<cluster-name>": "owned"
              }
            }
          },
          {
            "Sid": "ManageLaunchTemplatesAndCreateNodeGroupWithLaunchTemplate",
            "Effect": "Allow",
            "Action": [
              "ec2:DescribeLaunchTemplates",
              "ec2:DeleteLaunchTemplate",
              "ec2:CreateLaunchTemplateVersion",
              "ec2:DescribeLaunchTemplateVersions"
            ],
            "Resource": "*",
            "Condition": {
              "StringEquals": {
                "aws:ResourceTag/kubernetes.io/cluster/<cluster-name>": "owned"
              }
            }
          },
          {
            "Sid": "RunInstancesForEKSNodegroups",
            "Effect": "Allow",
            "Action": "ec2:RunInstances",
            "Resource": "arn:aws:ec2:*:*:launch-template/*",
            "Condition": {
              "StringEquals": {
                "aws:ResourceTag/kubernetes.io/cluster/<cluster-name>": "owned"
              }
            }
          }
      ]
  }
  ```
</Accordion>

The IAM role needs the following trust relationship to allow Atlas to assume it:

<Accordion title="Trust relationship policy">
  ```json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Effect": "Allow",
              "Principal": {
                  "AWS": "arn:aws:iam::809060229965:user/atlas-crossrole-<cluster-id>"
              },
              "Action": "sts:AssumeRole",
              "Condition": {
                  "StringEquals": {
                      "sts:ExternalId": "<cluster-id>"
                  }
              }
          }
      ]
  }
  ```
</Accordion>

#### EBS Volume Permissions

If your cluster uses Amazon EBS volumes, ensure that the IAM roles associated with your nodes have the necessary permissions to manage them, including the ability to attach and detach them.

To manage EBS volumes, add the following permissions to your IAM role:

<Accordion title="EBS volume management policy">
  ```json
  {
      "Sid": "ManageEBSVolumes",
      "Effect": "Allow",
      "Action": [
          "ec2:AttachVolume",
          "ec2:DetachVolume"
      ],
      "Resource": "*"
  }
  ```
</Accordion>

You can include these permissions in your existing IAM role policies or attach the `AmazonEBSCSIDriverPolicy`, which includes these permissions.

<Accordion title="Example standalone EBS policy">
  ```json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Sid": "ManageEBSVolumes",
              "Effect": "Allow",
              "Action": [
                  "ec2:AttachVolume",
                  "ec2:DetachVolume"
              ],
              "Resource": "*"
          }
      ]
  }
  ```
</Accordion>

Ensure that these permissions are attached to the IAM role used by the instances running in your cluster to avoid issues with persistent volumes.

#### IPv6 support for EKS clusters

If your cluster uses IPv6 addressing, you need an additional IAM policy to allow pods to assign IPv6 addresses.

Add the following policy to your instance profile:

<Accordion title="IPv6 policy">
  ```json
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Effect": "Allow",
              "Action": "ec2:AssignIpv6Addresses",
              "Resource": "*"
          }
      ]
  }
  ```
</Accordion>

> 📘 Note
>
> The Atlas onboarding script and Terraform module automatically create this policy by default. You only need to add it manually if you're setting up permissions outside of these standard deployment methods.

### GCP

#### The GCP service account used by Atlas

> 📘 Note
>
> To learn more about using service account impersonation on GKE, see [GKE service account impersonation](sa-impersonation-gke.md).

[The onboarding script](legacy-onboarding-script.md) creates a dedicated GCP service account that Atlas uses to request and manage GCP resources on your behalf.

The service account follows a `atlas-gke-<cluster-name-hash>` convention. You can verify the service account with the following command:

<Accordion title="Verify GCP service account">
  ```shell
  gcloud iam service-accounts describe atlas-gke-<cluster-name-hash>@<your-gcp-project>.iam.gserviceaccount.com
  ```
</Accordion>

The service account created by Atlas includes the following roles:

| Role name                | Description                                                                                                              |
| ------------------------ | :----------------------------------------------------------------------------------------------------------------------- |
| `atlas.gkeAccess`       | A Atlas-managed role used to handle Atlas add/delete node operations. You can find a full list of permissions below. |
| `iam.serviceAccountUser` | A GCP-managed role to allow running operations as a service account.                                                     |

> 📘 Important update for existing customers
>
> As of 2024-08-19, Atlas no longer requires the `roles/container.developer` role for new cluster onboarding. This change enhances security by reducing permissions.
> If you onboarded your cluster before this date, you may still see the `container.developer` role assigned to your Atlas service account. You can safely remove this role to align with our updated, more secure permissions model.
> To remove the role, use the following command:
>
> <Accordion title="Remove container.developer role">
>   ```shell
>   gcloud projects remove-iam-policy-binding YOUR_PROJECT_ID \
>       --member=serviceAccount:atlas-gke-<cluster-name-hash>@<your-gcp-project>.iam.gserviceaccount.com \
>       --role=roles/container.developer
>   ```
> </Accordion>
>
> This action will not affect Atlas's ability to manage your cluster.

#### IAM Conditions

When creating that service account, you can enforce conditional, attribute-based access on the `iam.serviceAccountUser` role.

It can access and act as all other service accounts or be scoped to those used by node pools in the GKE cluster, which is more secure and recommended. By default, [the onboarding script](legacy-onboarding-script.md) follows the more secure option.

When onboarding the cluster with **Terraform**, you can use the [atlas\_gke\_iam](https://registry.terraform.io/modules/atlas/gke-iam/atlas/latest) module to specify which method you want to use. You can find a [GKE IAM Terraform example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/gke/gke_atlas_iam) in the Atlas Terraform provider repository.

#### Required APIs to be enabled for a GCP project

| API                                   | Description                                                           |
| :------------------------------------ | :-------------------------------------------------------------------- |
| `cloudresourcemanager.googleapis.com` | API to create, read, and update metadata for GCP resource containers. |
| `serviceusage.googleapis.com`         | API to list, enable, and disable GCP services.                        |

#### List of `atlas.gkeAccess` role permissions:

You can verify the permissions on your Atlas service account using the following commands:

<Accordion title="Verify atlas.gkeAccess permissions">
  ```shell
  gcloud iam roles describe atlas.gkeAccess --project=<your-project-id>
  gcloud projects get-iam-policy <your-project-id> --filter="bindings.members:atlas-*" --flatten="bindings[].members"
  ```
</Accordion>

<Accordion title="Full atlas.gkeAccess role definition">
  ```shell
  » gcloud iam roles describe --project=<your-project-name> atlas.gkeAccess

  description: Role to manage GKE cluster via Atlas
  etag: example-tag
  includedPermissions:
  - compute.addresses.use
  - compute.disks.create
  - compute.disks.setLabels
  - compute.disks.use
  - compute.images.useReadOnly
  - compute.images.get
  - compute.instanceGroupManagers.get
  - compute.instanceGroupManagers.update
  - compute.instanceGroups.get
  - compute.instanceTemplates.create
  - compute.instanceTemplates.delete
  - compute.instanceTemplates.get
  - compute.instanceTemplates.list
  - compute.instances.create
  - compute.instances.delete
  - compute.instances.get
  - compute.instances.list
  - compute.instances.setLabels
  - compute.instances.setMetadata
  - compute.instances.setServiceAccount
  - compute.instances.setTags
  - compute.instances.start
  - compute.instances.stop
  - compute.instances.detachDisk (optional, used to properly detach PVC disks)
  - compute.networks.use
  - compute.networks.useExternalIp
  - compute.subnetworks.get
  - compute.subnetworks.use
  - compute.subnetworks.useExternalIp
  - compute.zones.get
  - compute.zones.list
  - compute.zoneOperations.get
  - compute.regionOperations.get
  - container.certificateSigningRequests.approve
  - container.clusters.get
  - container.clusters.update
  - container.operations.get
  - serviceusage.services.list
  - resourcemanager.projects.getIamPolicy
  name: projects/<your-project-name>/roles/atlas.gkeAccess
  stage: ALPHA
  title: Role to manage GKE cluster via Atlas
  ```
</Accordion>

### Azure

#### An overview of Azure permissions used by Atlas

> 📘 Note
>
> Atlas also supports [AKS workload identity impersonation](sa-impersonation-aks.md) as an alternative to App Registration. Workload Identity uses federated credentials instead of client secrets and is the recommended approach for Terraform-based onboarding with impersonation.

[The onboarding script](legacy-onboarding-script.md) creates a dedicated Azure app registration for Atlas to request and manage Azure resources on your behalf.

App registration naming follows this convention: `Atlas ${CLUSTER_NAME}-${ATLAS_CLUSTER_ID:0:8}"`.

Created Atlas app registration has a custom role bound to it. Custom role name follows this naming convention: `AtlasAKSRole-${ATLAS_CLUSTER_ID:0:8}`.

The role only has a predefined list of permissions to manage cluster resource groups.

#### Resource groups that `AtlasAKSRole` is assigned to

`AtlasAKSRole` is assigned to up to three resource groups:

* **Cluster resource group** — the resource group that contains the AKS cluster.
* **Node resource group** — the AKS-managed (`MC_*`) resource group that holds the cluster nodes.
* **Virtual network (VNet) resource group** — only when the cluster's VNet or subnet is in a *different* resource group than the cluster.

Atlas needs the networking actions (such as `Microsoft.Network/networkInterfaces/*` and `Microsoft.Network/virtualNetworks/subnets/join/action`) on the resource group that contains the cluster's VNet or subnet, so it can attach network interfaces and place provisioned nodes in the cluster's subnets.

All three onboarding methods detect this automatically and add the VNet resource group scope when it differs from the cluster resource group:

* The [console onboarding script](legacy-onboarding-script.md) and the `atlasctl` CLI read the VNet subnet ID from the AKS agent pool and derive its resource group.
* The [Atlas Terraform module for AKS](https://registry.terraform.io/modules/atlas/aks/atlas/latest) covers additional resource groups through the `additional_resource_groups` variable — add your VNet's resource group there when it differs from the cluster resource group.

> ⚠️ Warning
>
> If your cluster's VNet is in a different resource group than the cluster, the identity running onboarding must be able to **create role assignments in that networking resource group** (`Microsoft.Authorization/roleAssignments/write`). The networking resource group is often owned by a different team than the one managing the cluster, so arrange that access before onboarding to avoid provisioning failures.

You can verify the permissions on your Atlas role using the following commands:

<Accordion title="Verify Azure role permissions">
  ```shell
  az role definition list --name "AtlasAKSRole-<cluster-id>" --output table
  az role assignment list --assignee <app-registration-id> --output table
  ```
</Accordion>

#### List of `AtlasAKSRole` role permissions:

<Accordion title="AtlasAKSRole definition">
  ```shell
  ROLE_NAME="AtlasAKSRole-${ATLAS_CLUSTER_ID:0:8}"
  ROLE_DEF='{
     "Name": "'"$ROLE_NAME"'",
     "Description": "Atlas role used to manage '"$CLUSTER_NAME"' AKS cluster",
     "IsCustom": true,
     "Actions": [
         "Microsoft.Compute/*/read",
         "Microsoft.Compute/virtualMachines/*",
         "Microsoft.Compute/virtualMachineScaleSets/*",
         "Microsoft.Compute/disks/write",
         "Microsoft.Compute/disks/delete",
         "Microsoft.Compute/disks/beginGetAccess/action",
         "Microsoft.Compute/galleries/write",
         "Microsoft.Compute/galleries/delete",
         "Microsoft.Compute/galleries/images/write",
         "Microsoft.Compute/galleries/images/delete",
         "Microsoft.Compute/galleries/images/versions/write",
         "Microsoft.Compute/galleries/images/versions/delete",
         "Microsoft.Compute/snapshots/write",
         "Microsoft.Compute/snapshots/delete",
         "Microsoft.Network/*/read",
         "Microsoft.Network/networkInterfaces/write",
         "Microsoft.Network/networkInterfaces/delete",
         "Microsoft.Network/networkInterfaces/join/action",
         "Microsoft.Network/networkSecurityGroups/join/action",
         "Microsoft.Network/virtualNetworks/subnets/join/action",
         "Microsoft.Network/applicationGateways/backendhealth/action",
         "Microsoft.Network/applicationGateways/backendAddressPools/join/action",
         "Microsoft.Network/applicationSecurityGroups/joinIpConfiguration/action",
         "Microsoft.Network/loadBalancers/backendAddressPools/write",
         "Microsoft.Network/loadBalancers/backendAddressPools/join/action",
         "Microsoft.ContainerService/*/read",
         "Microsoft.ContainerService/managedClusters/start/action",
         "Microsoft.ContainerService/managedClusters/stop/action",
         "Microsoft.ContainerService/managedClusters/runCommand/action",
         "Microsoft.ContainerService/managedClusters/agentPools/*",
         "Microsoft.Resources/*/read",
         "Microsoft.Resources/tags/write",
         "Microsoft.Authorization/locks/read",
         "Microsoft.Authorization/roleAssignments/read",
         "Microsoft.Authorization/roleDefinitions/read",
         "Microsoft.ManagedIdentity/userAssignedIdentities/assign/action"
       ],
       "AssignableScopes": [
         "/subscriptions/'"$SUBSCRIPTION_ID"'/resourceGroups/'"$VNET_GROUP"'",
         "/subscriptions/'"$SUBSCRIPTION_ID"'/resourceGroups/'"$CLUSTER_GROUP"'",
         "/subscriptions/'"$SUBSCRIPTION_ID"'/resourceGroups/'"$NODE_GROUP"'"
       ]
  }'
  ```

  The `VNET_GROUP` scope is included only when the cluster's VNet is in a different resource group than the cluster; it is derived automatically from the cluster's subnet ID. See [Resource groups that `AtlasAKSRole` is assigned to](#resource-groups-that-atlasaksrole-is-assigned-to) above.
</Accordion>

## Permissions for Cloud Connect

Cloud Connect creates read-only IAM roles or service accounts to discover and inventory cloud resources. Permission scopes can be configured during setup to control the level of access granted.

> 📘 Note
>
> For AWS, Azure, and GCP, discovered resources are synchronized every hour.

### AWS

#### Default

Grants the AWS `ReadOnlyAccess` managed policy, providing comprehensive read-only access to all AWS services.

<Accordion title="Default AWS policy">
  ```json
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "*"
        ],
        "Resource": "*",
        "Condition": {
          "StringEquals": {
            "iam:PassedToService": [
              "ec2.amazonaws.com"
            ]
          }
        }
      }
    ]
  }
  ```
</Accordion>

#### Minimal permissions

Provides targeted permissions for core AWS services, including EC2, EKS, RDS, SageMaker, and commitments data.

<Accordion title="Minimal AWS permissions">
  ```json
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "ec2:DescribeInstances",
          "ec2:DescribeAddresses",
          "eks:ListClusters",
          "eks:DescribeCluster",
          "rds:DescribeDBInstances",
          "rds:DescribeDBClusters",
          "sagemaker:ListEndpoints",
          "sagemaker:DescribeEndpoint",
          "sagemaker:ListTransformJobs",
          "sagemaker:DescribeTransformJob",
          "sagemaker:ListTrainingJobs",
          "sagemaker:DescribeTrainingJob",
          "sagemaker:ListNotebookInstances",
          "sagemaker:DescribeNotebookInstance",
          "sagemaker:ListProcessingJobs",
          "sagemaker:DescribeProcessingJob",
          "cloudwatch:GetMetricData",
          "savingsplans:Describe*",
          "savingsplans:List*",
          "ec2:DescribeReservedInstances",
          "ec2:DescribeReservedInstancesListings",
          "ec2:DescribeReservedInstancesModifications",
          "ec2:DescribeReservedInstancesOfferings",
          "ec2:DescribeCapacityReservations",
          "organizations:ListAccounts",
          "organizations:DescribeOrganization",
          "account:ListRegions"
        ],
        "Resource": "*"
      }
    ]
  }
  ```
</Accordion>

#### AWS AI services

Provides `List*`, `Describe*`, and `Get*` permissions for AWS AI and machine learning services.

<Accordion title="AWS AI services permissions">
  ```json
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "bedrock:List*",
          "bedrock:Describe*",
          "bedrock:Get*",
          "cloudwatch:List*",
          "cloudwatch:Describe*",
          "cloudwatch:Get*",
          "codeguru:List*",
          "codeguru:Describe*",
          "codeguru:Get*",
          "codeguru-reviewer:List*",
          "codeguru-reviewer:Describe*",
          "codeguru-reviewer:Get*",
          "comprehend:List*",
          "comprehend:Describe*",
          "comprehend:Get*",
          "comprehendmedical:List*",
          "comprehendmedical:Describe*",
          "comprehendmedical:Get*",
          "devops-guru:List*",
          "devops-guru:Describe*",
          "devops-guru:Get*",
          "forecast:List*",
          "forecast:Describe*",
          "forecast:Get*",
          "frauddetector:List*",
          "frauddetector:Describe*",
          "frauddetector:Get*",
          "healthlake:List*",
          "healthlake:Describe*",
          "healthlake:Get*",
          "kendra:List*",
          "kendra:Describe*",
          "kendra:Get*",
          "lex:List*",
          "lex:Describe*",
          "lex:Get*",
          "lookoutequipment:List*",
          "lookoutequipment:Describe*",
          "lookoutequipment:Get*",
          "lookoutmetrics:List*",
          "lookoutmetrics:Describe*",
          "lookoutmetrics:Get*",
          "lookoutvision:List*",
          "lookoutvision:Describe*",
          "lookoutvision:Get*",
          "monitron:List*",
          "monitron:Describe*",
          "monitron:Get*",
          "personalize:List*",
          "personalize:Describe*",
          "personalize:Get*",
          "polly:List*",
          "polly:Describe*",
          "polly:Get*",
          "rekognition:List*",
          "rekognition:Describe*",
          "rekognition:Get*",
          "sagemaker:List*",
          "sagemaker:Describe*",
          "sagemaker:Get*",
          "q:List*",
          "q:Describe*",
          "q:Get*",
          "textract:List*",
          "textract:Describe*",
          "textract:Get*",
          "transcribe:List*",
          "transcribe:Describe*",
          "transcribe:Get*",
          "translate:List*",
          "translate:Describe*",
          "translate:Get*",
          "deepcomposer:List*",
          "deepcomposer:Describe*",
          "deepcomposer:Get*",
          "deepracer:List*",
          "deepracer:Describe*",
          "deepracer:Get*",
          "panorama:List*",
          "panorama:Describe*",
          "panorama:Get*"
        ],
        "Resource": "*"
      }
    ]
  }
  ```
</Accordion>

### GCP

#### Default

Uses the standard GCP roles for comprehensive resource access:

<Accordion title="Default GCP roles">
  ```yaml
  - roles/reader
  - roles/viewer
  ```
</Accordion>

#### Minimal permissions

Grants specific viewer roles for essential Atlas functionality:

<Accordion title="Minimal GCP permissions">
  ```yaml
  - roles/compute.viewer
  - roles/container.viewer
  - roles/cloudsql.viewer
  - roles/billing.viewer
  - roles/resourcemanager.projectViewer
  - roles/aiplatform.viewer
  ```
</Accordion>

### Azure

#### Default

Uses the built-in Azure `Reader` role for comprehensive read-only access to all resources in the subscription.

<Accordion title="Default Azure role">
  ```
  Reader
  ```
</Accordion>

#### Minimal permissions

Provides targeted read-only permissions for core Azure services, including virtual machines, scale sets, and AKS clusters.

<Accordion title="Minimal Azure permissions">
  ```json
  {
    "Actions": [
      "Microsoft.Compute/virtualMachines/read",
      "Microsoft.Compute/virtualMachineScaleSets/read",
      "Microsoft.Compute/virtualMachineScaleSets/virtualMachines/read",
      "Microsoft.ContainerService/managedClusters/read"
    ],
    "NotActions": [],
    "DataActions": [],
    "NotDataActions": []
  }
  ```
</Accordion>

### Oracle

#### Default

Oracle Cloud Infrastructure uses policy statements for resource access:

<Accordion title="Oracle Cloud policy statements">
  ```
  Allow group Atlas-Group to read compartments in tenancy
  Allow group Atlas-Group to read cluster-family in tenancy
  Allow group Atlas-Group to read instance-family in tenancy
  Allow group Atlas-Group to read volume-family in tenancy
  Allow group Atlas-Group to read database-family in tenancy
  Allow group Atlas-Group to read virtual-network-family in tenancy
  ```
</Accordion>

### Permission Configuration

Permission scopes are configurable during the Cloud Connect integration setup. Based on their security requirements, users can select from available permission levels.

**Example permission scope selection for Cloud Connect:**

<Image align="center" width="50% " src="https://files.readme.io/52a81487d5ff702115a4a22b73f7e62e5f1b2008a52698fc6dabdfac4c6bd6af-image.png" />

## Permissions for *Discover cluster*

*Discover cluster* lets you discover virtual machines, disks, and Kubernetes clusters across AWS, Azure, and GCP — without creating any cloud integrations or storing credentials.

Run the script once, and it uses your already authenticated local CLIs to collect a read-only snapshot of your infrastructure. The data is imported into Atlas so you can view your discovered resources and identify potential savings.

### What makes *Discover cluster* different from Cloud Connect

| Aspect        | Discover cluster                            | Cloud Connect                                                      |
| ------------- | ------------------------------------------- | ------------------------------------------------------------------ |
| Setup         | Run a script using authenticated local CLIs | Creates IAM roles or service accounts                              |
| Persistence   | One-time execution                          | Ongoing integration with periodic sync                             |
| Credentials   | Uses your already authenticated local CLIs  | IAM trust relationships — short-lived credentials issued on demand |
| Cloud changes | None. Read-only, no resources created       | Creates IAM roles/service accounts                                 |
| Scope         | VMs, disks, Kubernetes clusters only        | Broader: includes RDS, SageMaker, AI services, and more            |

### What the script does

1. Uses your already authenticated local CLIs to access cloud resources
2. Enumerates accounts, subscriptions, or projects
3. Lists compute resources and Kubernetes clusters
4. Sends collected data to Atlas
5. Populates the Atlas console with discovered clusters

### What the script does NOT do

* **Does not create** any IAM roles, service accounts, or cloud integrations
* **Does not modify** any cloud resources
* **Does not delete** anything in your environment
* **Does not store** credentials; uses your already authenticated local CLIs
* **Does not make persistent changes** – your environment remains unchanged after execution

### Cloud permissions

The script requires read-only permissions to list resources. These are the same permissions you'd use when running list or describe commands manually with your cloud CLI.

#### AWS

| Permission                   | Purpose                               |
| ---------------------------- | ------------------------------------- |
| `ec2:DescribeInstances`      | List EC2 instances                    |
| `ec2:DescribeVolumes`        | List EBS volumes                      |
| `ec2:DescribeRegions`        | Enumerate available regions           |
| `eks:ListClusters`           | List EKS clusters                     |
| `eks:DescribeCluster`        | Get EKS cluster details               |
| `eks:AccessKubernetesApi`    | Access cluster to list pods and nodes |
| `organizations:ListAccounts` | List accounts in organization         |

#### Azure

| Permission                                                                    | Purpose             |
| ----------------------------------------------------------------------------- | ------------------- |
| `Microsoft.Compute/virtualMachines/read`                                      | List VMs            |
| `Microsoft.Compute/virtualMachineScaleSets/read`                              | List VMSS           |
| `Microsoft.Compute/virtualMachineScaleSets/virtualMachines/read`              | List VMSS instances |
| `Microsoft.Compute/disks/read`                                                | List disks          |
| `Microsoft.ContainerService/managedClusters/read`                             | List AKS clusters   |
| `Microsoft.ContainerService/managedClusters/listClusterUserCredential/action` | Get AKS credentials |

#### GCP

| Permission                             | Purpose                          |
| -------------------------------------- | -------------------------------- |
| `compute.instances.list`               | List Compute Engine instances    |
| `compute.disks.list`                   | List persistent disks            |
| `container.clusters.list`              | List GKE clusters                |
| `container.clusters.get`               | Get GKE cluster details          |
| `container.clusters.getCredentials`    | Get GKE credentials              |
| `serviceusage.services.list`           | Check enabled APIs               |
| `resourcemanager.projects.get`         | List projects                    |
| `cloudasset.assets.searchAllResources` | Filter projects with compute API |

#### Kubernetes RBAC

For Kubernetes cluster access, the script also needs RBAC permissions to list nodes, pods, and pod metrics.

### Resources surfaced

| Cloud | Resources                                                           |
| ----- | ------------------------------------------------------------------- |
| AWS   | EC2 instances, EBS volumes, EKS clusters (with pods/nodes)          |
| Azure | VMs, VMSS instances, Managed disks, AKS clusters (with pods/nodes)  |
| GCP   | Compute instances, Persistent disks, GKE clusters (with pods/nodes) |

## Permissions for Commitments import

Commitments import provides access to reserved instances, savings plans, and cost commitment data.

### AWS

> 📘 Note
>
> For AWS, commitment data is synchronized every hour.

<Accordion title="AWS commitments permissions">
  ```json
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "savingsplans:Describe*",
          "savingsplans:List*",
          "ec2:DescribeReservedInstances",
          "ec2:DescribeReservedInstancesListings",
          "ec2:DescribeReservedInstancesModifications",
          "ec2:DescribeReservedInstancesOfferings",
          "ec2:DescribeCapacityReservations",
          "organizations:ListAccounts",
          "organizations:DescribeOrganization",
          "account:ListRegions"
        ],
        "Resource": "*"
      }
    ]
  }
  ```
</Accordion>

### Azure

Azure commitments require read access to subscription and billing data:

<Accordion title="Azure commitments permissions">
  ```json
  {
    "Name": "Atlas-Commitments-Reader",
    "Actions": [
      "Microsoft.Consumption/*/read",
      "Microsoft.Billing/*/read",
      "Microsoft.Commerce/*/read",
      "Microsoft.CostManagement/*/read",
      "Microsoft.Resources/subscriptions/read",
      "Microsoft.Resources/subscriptions/resourceGroups/read"
    ],
    "AssignableScopes": [
      "/subscriptions/{subscription-id}"
    ]
  }
  ```
</Accordion>