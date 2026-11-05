---
updatedAt: 2026-09-22T12:07:51.000Z
---

# Legacy onboarding script reference

The Atlas console still offers this script as an option alongside [atlasctl](connect-with-atlasctl.md) (see [Connect using the Atlas console](connect-via-console.md)) for environments where atlasctl isn't yet supported. It creates the cloud IAM permissions and installs the Kubernetes components Atlas needs to manage nodes, not just monitor them.

If you connected with atlasctl instead, it creates largely the same IAM roles and components, automatically, in a single step. This page is still useful as a reference for what exists in your cloud account and cluster either way. The prerequisites below (`jq`, `yq`, `helm`, cloud CLIs) are specific to running the script manually: atlasctl bundles its own Helm logic and doesn't need any of them.

This script connects a cluster read-only. To turn on Node Autoscaling and/or Workload Autoscaling, run it again with those options selected once [the read-only agent](about-the-read-only-agent.md) is up: the script checks for it and exits with an error if it's missing.

<Callout icon="📘" theme="info">
  ### Turning on automation?

  Make sure the read-only agent is already running. The script exits with an error if it isn't.
</Callout>

Get the script by clicking the **Enable Atlas** button in the console, or via the [API](https://docs.atlas.example/atlas/reference/externalclusterapi_getcredentialsscripttemplate).

If you're using GCP or AWS, you can use our **Connect Cloud** feature to discover your clusters before agent installation. Click on **Connect Cloud** in the navigation menu and run the provided script in your cloud shell. This will discover all your GKE/EKS clusters with basic information before you decide which ones to fully connect with the agent. See [Getting started](getting-started.md) to learn more.

The following sections describe the prerequisites for the legacy script for each supported cloud provider and the actions it performs. Turning automation on or off for an already-connected cluster is a separate policy setting, not something this script controls: see the [Node Autoscaler checklist](autoscaler-checklist.md).

## EKS

### Prerequisites

<Callout icon="⚠️" theme="warn">
  ### EKS Auto Mode and Fargate

  EKS Auto Mode and EKS Fargate profiles are not supported for standard Atlas automation because AWS manages all compute resources automatically. For EKS clusters using Auto Mode or Fargate, use [Atlas Anywhere](atlas-anywhere-overview.md) instead, which provides workload optimization without requiring node management permissions.
</Callout>

* `AWS CLI`– a command-line tool for working with AWS services using commands in your command-line shell. For more details, see [Installing AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html).

* `jq` – a lightweight command-line JSON processor. For more details, see the [jq documentation](https://stedolan.github.io/jq/).

* `yq` - a lightweight command-line YAML processor. For more details, see [yq](https://github.com/mikefarah/yq).

* `kubectl` – a Kubernetes command-line tool that allows running commands against Kubernetes clusters. For more details, see [kubectl](https://kubernetes.io/docs/reference/kubectl/).

* `helm` – a command-line tool that simplifies deploying applications and services to Kubernetes clusters. For more details, see [helm](https://helm.sh/docs/intro/install/).
  <Helm_v4_error />

* *IAM permissions* – the IAM security principal you're using must have permission to work with AWS EKS, AWS IAM, and related resources. Additionally, you should have access to the EKS cluster that you wish to onboard into the Atlas console.

* The Atlas agent must be running on the cluster. Learn more about [Installing the Atlas agent](about-the-read-only-agent.md#how-the-agent-works-and-how-to-use-it)

Here's an example of a policy of least privilege for the administrator account, with permissions needed to run the onboarding script. This policy is used once per cluster during onboarding.

```json
{
    "Action": [
        "iam:CreateRole",
        "iam:CreatePolicy",
        "iam:GetPolicy",
      	"iam:GetPolicyVersion",
        "iam:ListPolicies",
        "iam:ListPolicyVersions",
        "iam:PutRolePolicy",
        "iam:AttachRolePolicy",
        "iam:CreateInstanceProfile",
        "iam:GetInstanceProfile",
        "iam:AddRoleToInstanceProfile",
        "iam:UpdateAssumeRolePolicy",
        "ec2:CreateSecurityGroup",
        "ec2:AuthorizeSecurityGroupEgress",
        "ec2:AuthorizeSecurityGroupIngress"
    ]
}
```

<Callout icon="📘" theme="info">
  ### Optional: permission preflight validation

  Before making any changes, the onboarding tooling (`atlasctl` and the onboarding script) runs a preflight check that calls `iam:SimulatePrincipalPolicy` to validate your permissions ahead of time. This permission is **not** part of the least-privilege policy above and is **not required** for onboarding to succeed — if it's missing, the check logs a warning and onboarding continues normally.
</Callout>

If you'd rather grant this up front than have the check warn about it, expand the list below for every permission the preflight check evaluates.

<Accordion title="Full permission set checked by the preflight validation (optional)">
  `iam:SimulatePrincipalPolicy` evaluates the following actions for a standard EKS autoscaler onboarding. Most already appear in the policy above; the rest are read-only lookups and a few permissions used only by optional onboarding steps.

  ```json
  {
      "Action": [
          "iam:SimulatePrincipalPolicy",
          "iam:GetRole",
          "iam:GetOpenIDConnectProvider",
          "iam:ListAttachedRolePolicies",
          "iam:CreatePolicyVersion",
          "iam:DeletePolicyVersion",
          "iam:CreateServiceLinkedRole",
          "iam:PassRole",
          "sts:GetCallerIdentity",
          "eks:DescribeCluster",
          "eks:ListNodegroups",
          "eks:DescribeNodegroup",
          "eks:ListAccessEntries",
          "eks:CreateAccessEntry",
          "eks:AssociateAccessPolicy",
          "eks:DescribeAccessEntry",
          "ec2:DescribeSecurityGroups",
          "ec2:CreateTags",
          "ec2:DescribeLaunchTemplateVersions",
          "autoscaling:DescribeAutoScalingGroups"
      ]
  }
  ```

  This list is optional and does not expand what's required for onboarding beyond the policy above.
</Accordion>

### Actions the onboarding script performs

The script performs the following actions:

1. Creates a `atlas-eks-*cluster-name*-cluster-role` IAM role and a `AtlasEKSPolicy` policy, attaching the AWS-managed `AmazonEC2ReadOnlyAccess` and `IAMReadOnlyAccess` policies plus custom permissions to manage instances, Auto Scaling groups, and EKS node groups within the cluster's VPC. See [Cloud permissions](cloud-permissions.md#aws) for the full policy.

2. Creates the following roles:
   * `atlas-*cluster-name*-eks-#######` used by instance profile with the following AWS-managed permission policies applied by default:
     * `AmazonEKSWorkerNodePolicy`
     * `AmazonEC2ContainerRegistryReadOnly`
     * `AmazonSSMManagedInstanceCore`
     * `AmazonEBSCSIDriverPolicy`
     * `AmazonEKS_CNI_Policy`
     * `AmazonEKS_CNI_IPv6_Policy` or create `AtlasEC2AssignIPv6Policy` for IPv6 support
     * Role ARN is printed and sent to the Atlas console. The platform then uses it to assume the role when making AWS programmatic calls.

3. Installs the Kubernetes components needed to manage the cluster. See [Hosted components](hosted-components.md) for the current, complete list.

<Callout icon="📘" theme="info">
  ### Optional permissions policies

  Please note that `AmazonEBSCSIDriverPolicy` and `AmazonSSMManagedInstanceCore` policies are optional. While they are included in the onboarding script by default, they can be disabled.
</Callout>

<Callout icon="📘" theme="info">
  ### The scope of permissions

  All `Write` permissions are scoped to the specified EKS cluster without access to the resources of any other clusters in the AWS account.

  You can grant extra permissions to Atlas roles if you need access to resources and actions not included in the onboarding script's policies.
</Callout>

### Manual credential onboarding

You can complete the steps listed above manually without using the script. However, keep in mind that when you create an Amazon EKS cluster, the IAM entity user or role (e.g., a federated user that creates the cluster) automatically gets `system:masters` permissions in the cluster's RBAC configuration in the control plane.

To allow additional AWS users or roles to interact with your cluster, you need to edit the `aws-auth` ConfigMap in Kubernetes. For more information, see [Managing users or IAM roles for your cluster](https://docs.aws.amazon.com/eks/latest/userguide/add-user-role.html).

### Using AWS services

Atlas relies on the agent running inside your cluster. The operation consumes the following services:

* A portion of the EC2 node resources from your cluster. The Atlas agent uses the [cluster proportional vertical autoscaler](https://github.com/kubernetes-sigs/cluster-proportional-vertical-autoscaler#calculation-of-resource-requests-and-limits) to consume the minimum required resources depending on the cluster size.
* Low amount of network traffic to communicate with the Atlas SaaS.
* EC2 instances, their storage, and intra-cluster network traffic to manage the Kubernetes cluster and perform autoscaling.
* IAM resources as detailed in the onboarding section

You can find a full overview of permissions used by the Atlas-created IAM role [here](cloud-permissions.md#aws).

## GKE

### Prerequisites

<Callout icon="⚠️" theme="warn">
  ### GKE Autopilot

  GKE Autopilot clusters are not supported for standard Atlas automation because Google manages node provisioning automatically. For GKE Autopilot clusters, use [Atlas Anywhere](atlas-anywhere-overview.md) instead, which provides workload optimization without requiring node management permissions. Workload Autoscaler automatically detects Autopilot clusters and adjusts recommendations to comply with [Autopilot resource constraints](workload-autoscaling-overview.md#gke-autopilot-support).
</Callout>

* `gcloud` – A command-line tool for working with GKE services using commands in your command-line shell. For more details, see [Installing gcloud](https://cloud.google.com/sdk/docs/install).

* `jq` – a lightweight command-line JSON processor. For more details, see the [jq documentation](https://stedolan.github.io/jq/).

* *IAM permissions* – The IAM security principal that you use to onboard the cluster must include:
  * Access to the project where the cluster is created.
  * Permission to work with IAM, GKE, and compute resources.

* `kubectl` – a Kubernetes command-line tool allowing commands to be run against Kubernetes clusters. For more details, see [kubectl](https://kubernetes.io/docs/reference/kubectl/).

* `helm` –  a command-line tool that simplifies the deployment of applications and services to Kubernetes clusters [helm](https://helm.sh/docs/intro/install/).

* The Atlas agent has to be running on the cluster. Learn more about [installing Atlas agent](about-the-read-only-agent.md#how-the-agent-works-and-how-to-use-it)

<Callout icon="📘" theme="info">
  ### Note

  GCP users can simplify discovery by using the **Cloud Connect** feature to see all available GKE clusters before deciding which ones to onboard with the agent.
</Callout>

Here's an example of the least privileged permissions for the administrator account needed to run the onboarding script, used once per cluster during its onboarding:

```shell
- serviceusage.services.enable
- servicemanagement.services.bind
- container.clusters.get
- container.clusters.list
- iam.serviceAccounts.get
- iam.serviceAccounts.create
- iam.roles.get
- iam.roles.update
- iam.roles.create
- iam.serviceAccountKeys.create
```

### Actions the onboarding script performs

The Phase 2 onboarding script performs several actions to get the permissions required to manage GKE and GCP resources on your behalf:

* Enables the necessary GCP services and APIs for the project.
* Creates the IAM service account and assigns the required roles to it.
* Generates an IAM service account key that Atlas components use to manage GKE and GCP resources on your behalf.
* Enables the following GCP services and APIs for the project in which the GKE cluster is running:

| GCP Service / API Group                                                                           | Description                                                          |
| ------------------------------------------------------------------------------------------------- | :------------------------------------------------------------------- |
| [`serviceusage.googleapis.com`](https://cloud.google.com/service-usage/docs/reference/rest)       | API to list, enable and disable GCP services.                        |
| [`iam.googleapis.com`](https://cloud.google.com/iam/docs/reference/rest)                          | API to manage identity and access control for GCP resources          |
| [`cloudresourcemanager.googleapis.com`](https://cloud.google.com/resource-manager/reference/rest) | API to create, read, and update metadata for GCP resource containers |
| [`container.googleapis.com`](https://cloud.google.com/kubernetes-engine/docs/reference/rest)      | API to manage GKE                                                    |
| [`compute.googleapis.com`](https://cloud.google.com/compute/docs/reference/rest/v1)               | API to manage GCP virtual machines                                   |

* Creates a dedicated GCP service account `atlas-gke-<cluster-name-hash>` for Atlas to request and manage GCP resources on your behalf.

* Creates a custom role `atlas.gkeAccess`. See [Cloud permissions](cloud-permissions.md#gcp) for the full permission list.

* Attaches required roles to the `atlas-gke-<cluster-name-hash>` service account.

* Installs the Kubernetes components needed to manage the cluster. See [Hosted components](hosted-components.md) for the current, complete list.

For a full overview of permissions a Atlas-created service account uses, see [GCP cloud permissions](cloud-permissions.md#gcp).

### GKE node pools created by Atlas

After the cluster is onboarded, Atlas will create two GKE node pools:

* `atlaspool` – gathers necessary data required for creating Atlas-managed GKE x86 nodes.
* `atlaspool-arm` – collects the data required for creating Atlas-managed GKE ARM64 nodes. It's created only if the cluster region supports ARM64 VMs.

## AKS

### Prerequisites

<Callout icon="⚠️" theme="warn">
  ### AKS Auto

  AKS Auto (automatic cluster management) is not supported for standard Atlas automation because Azure manages node provisioning and scaling automatically. For AKS Auto clusters, use [Atlas Anywhere](atlas-anywhere-overview.md) instead, which provides workload optimization without requiring node management permissions.
</Callout>

* `az CLI` – a command-line tool for working with Azure services using commands in your command-line shell. For more details, see [Installing az CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli).

* `jq` – a lightweight command-line JSON processor. For more details, see the [jq documentation](https://stedolan.github.io/jq/).

* *Azure AD permissions* – the Azure identity that you use to onboard the cluster must include:
  * Access to the subscription where the cluster was created.
  * Permission to work with AKS and compute resources.
  * Permission to create an App registration.

* `kubectl` – a Kubernetes command-line tool that allows running commands against Kubernetes clusters. For more details, see [kubectl](https://kubernetes.io/docs/reference/kubectl/).

* `helm` –  a command-line tool that simplifies the deployment of applications and services to Kubernetes clusters [helm](https://helm.sh/docs/intro/install/).

* The Atlas agent has to be running on the cluster. Learn more about [Installing Atlas agent](about-the-read-only-agent.md#how-the-agent-works-and-how-to-use-it)

### Actions the onboarding script performs

The script performs the following actions:

* Create a `AtlasAKSRole-${ATLAS_CLUSTER_ID:0:8}` role to manage onboarded AKS Cluster.

* Create an app registration `Atlas ${CLUSTER_NAME}-${ATLAS_CLUSTER_ID:0:8}"` that uses the role `AtlasAKSRole-${ATLAS_CLUSTER_ID:0:8}`.

* Installs the Kubernetes components needed to manage the cluster. See [Hosted components](hosted-components.md) for the current, complete list.

<Callout icon="📘" theme="info">
  ### The scope of permissions

  All `Write` permissions are scoped to the resource group (or groups) containing your cluster's resources, which can include the cluster's own resource group, its node resource group, and its VNet's resource group. See [Resource groups that AtlasAKSRole is assigned to](cloud-permissions.md#resource-groups-that-atlasaksrole-is-assigned-to) for the full picture.
</Callout>

For a full overview of permissions used by Atlas-created service accounts, see [Azure cloud permissions](cloud-permissions.md#azure).

### Azure agent pools created by Atlas

After the cluster is onboarded, Atlas will create two AKS agent pools:

* `atlaspool`– is used to schedule a node required for Atlas AKS image creation. Atlas AKS images are re-created after every AKS control plane upgrade or every 30 days. While the image creation process is in progress, the `atlaspool` node with the name `aks-atlaspool-xxxxxxxx-vmssxxxxxx` can be seen in the "Not Ready" state for some time.
* `atlasworkers`– is used as a container for Atlas-managed AKS nodes. Removing this agent pool results in removing all Atlas-created nodes.

### Azure CNI Overlay support

Atlas fully supports AKS clusters that use Azure CNI Overlay networking with Pod CIDR. Autoscaling decisions fully respect IP address limitations when adding nodes.