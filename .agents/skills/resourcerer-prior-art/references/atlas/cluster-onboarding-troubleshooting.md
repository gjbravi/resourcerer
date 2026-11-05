---
updatedAt: 2026-09-22T11:21:49.000Z
---

# Troubleshooting cluster onboarding

Solutions for resolving issues when connecting clusters to Atlas, including connectivity problems, authentication errors, and cloud provider-specific onboarding issues.

## Your cluster does not appear on the Connect Cluster screen

If the cluster does not appear on the Connect your cluster screen after you've connected with atlasctl, perform the following steps:

**1. Check agent container logs:**

````
```shell
kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-agent -c agent
```
````

**2. You might get output similar to this:**

````
```text
time="2021-05-06T14:24:03Z" level=fatal msg="agent failed: registering cluster: getting cluster name: describing instance_id=i-026b5fadab5b69d67: UnauthorizedOperation: You are not authorized to perform this operation.\n\tstatus code: 403, request id: 2165c357-b4a6-4f30-9266-a51f4aaa7ce7"
```
or

```text
time="2021-05-06T14:24:03Z" level=fatal msg=agent failed: getting provider: configuring aws client: NoCredentialProviders: no valid providers in chain"
```
or

```text
time="2023-08-18T18:44:49Z" level=error msg="agent failed: getting provider: configuring aws client: getting instance region: EC2MetadataRequestError: failed to get EC2 instance identity document\ncaused by: RequestError: send request failed\ncaused by: Get \"http://169.254.169.254/latest/dynamic/instance-identity/document\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
````

These errors have two possible causes:

**Missing cluster metadata env vars** — The agent identifies your EKS cluster using `EKS_ACCOUNT_ID`, `EKS_REGION`, and `EKS_CLUSTER_NAME` environment variables. atlasctl and the Terraform module set these automatically. If you installed the agent manually via Helm, you must set them yourself — without them the agent cannot register the cluster and will log `EC2MetadataRequestError` timeouts.

**Missing AWS credentials or IAM permissions** — The agent requires AWS credentials to call `ec2:DescribeInstances`. If no credentials are configured, you'll see `NoCredentialProviders`. If the IAM role lacks `ec2:DescribeInstances` permission, you'll see `UnauthorizedOperation`. The agent uses this API call to detect whether nodes are spot or on-demand.

The Atlas agent uses the official AWS SDK, which supports all variables for customizing authentication, as mentioned in its [documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

### Provide cluster metadata env vars

If the agent is missing the `EKS_*` env vars, add them to the Helm values (atlasctl and the Terraform module do this automatically):

```yaml
- name: EKS_ACCOUNT_ID
  value: "000000000000"    # your AWS account ID
- name: EKS_REGION
  value: "eu-central-1"    # your EKS cluster region
- name: EKS_CLUSTER_NAME
  value: "staging-example" # your EKS cluster name
```

If you're instead using GCP GKE, you can provide the following environment variables to overcome the lack of access to VM metadata:

```yaml
- name: GKE_PROJECT_ID
  value: your_project_id
- name: GKE_CLUSTER_NAME
  value: your_cluster_name
- name: GKE_REGION
  value: your_cluster_region
- name: GKE_LOCATION
  value: your_cluster_az
```

### Provide AWS credentials

The agent requires `ec2:DescribeInstances` to detect whether nodes are spot or on-demand. The default `AmazonEC2ReadOnlyAccess` policy includes this permission. Provide AWS API access by adding these variables to the Atlas Agent secret:

```yaml
AWS_ACCESS_KEY_ID = xxxxxxxxxxxxxxxxxxxx
AWS_SECRET_ACCESS_KEY = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Alternatively, if you use [IAM roles for service accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), you can annotate the atlas-agent service account instead of providing AWS credentials with your IAM role.

```shell
kubectl annotate serviceaccount -n atlas-agent atlas-agent eks.amazonaws.com/role-arn="arn:aws:iam::111122223333:role/iam-role-name"
```

***

## Disconnected or Not responding cluster

If the cluster has a `Not responding` status, most likely the Atlas agent deployment is missing. Press **Reconnect** and follow the instructions provided.

The `Not responding` state is temporary, and unless fixed, the cluster will enter the `Disconnected` state. If your cluster is disconnected, you can reconnect or delete it from the console, as shown below.

The delete action only removes the cluster from the Atlas console, leaving it running in the cloud service provider.

***

## TLS handshake timeout issue

In some edge cases, due to a specific cluster network setup, the agent might fail with the following message in the agent container logs:

```json
time="2021-11-13T05:19:54Z" level=fatal msg="agent failed: registering cluster: getting namespace \"kube-system\": Get \"https://100.10.1.0:443/api/v1/namespaces/kube-system\": net/http: TLS handshake timeout" provider=eks version=v0.22.1
```

You can resolve this issue by deleting the `atlas-agent` pod. The deployment will recreate the pod and resolve the issue.

***

## Refused connection to control plane

When enabling automated cluster optimization for the first time, the user runs a pre-generated script to grant Atlas the required permissions. The error message **No access to Kubernetes API server, please check your firewall settings** indicates that a firewall prevents communication between the control plane and Atlas.

To solve this issue, allow access to Atlas IP `35.221.40.21` and then enable automated optimization again.

***

## AKS versions not supported

During the cluster onboarding to Atlas managed mode, the onboarding process will create a new Node Pool. Microsoft Azure Cloud enforces certain restrictions for Node Pool creations:

* Node Pool can NOT be newer than your AKS cluster control plane version.
* Microsoft supports only a few minor/patch Kubernetes versions for Node Pool creation. [Azure documentation.](https://learn.microsoft.com/en-us/azure/aks/supported-kubernetes-versions?tabs=azure-cli#what-version-does-the-control-plane-support-if-the-node-pool-isnt-in-one-of-the-supported-aks-versions)

You can check the list of supported AKS versions in your region:

```shell
❯ az aks get-versions --location eastus --output table
KubernetesVersion    Upgrades                 
-------------------  -----------------------  
1.29.2               None available           
1.29.0               1.29.2                   
1.28.5               1.29.0, 1.29.2           
1.28.3               1.28.5, 1.29.0, 1.29.2   
1.27.9               1.28.3, 1.28.5           
1.27.7               1.27.9, 1.28.3, 1.28.5   
1.26.12              1.27.7, 1.27.9           
1.26.10              1.26.12, 1.27.7, 1.27.9  
```

If your AKS cluster control plane version is not in the supported list, no new Node Pools can be created (Atlas or not). To continue Atlas onboarding, upgrade the AKS control plane to a supported version and re-run atlasctl. There is no need to upgrade your existing nodes, just the Control Plane.

***

## AKS fails to pull images from Azure Container Registry to Azure Kubernetes Service cluster

If the cluster is already attached to the ACR after onboarding on Atlas, the Service Principal created to manage the cluster might not have the correct permissions to pull images from the private ACRs. This may result in `failed to pull and unpack image, failed to fetch oauth token: unexpected status: 401 Unauthorized` when creating new nodes.

Microsoft has detailed documentation on troubleshooting and fixing the issue: [Fail to pull images from Azure Container Registry to Azure Kubernetes Service cluster](https://learn.microsoft.com/en-us/troubleshoot/azure/azure-kubernetes/extensions/cannot-pull-image-from-acr-to-aks-cluster)

In most cases, [Solution 1: Ensure AcrPull role assignment is created for identity](https://learn.microsoft.com/en-us/troubleshoot/azure/azure-kubernetes/extensions/cannot-pull-image-from-acr-to-aks-cluster#solution-1-make-sure-acrpull-role-assignment-is-created-for-identity) is enough to resolve it.

***

## AKS onboarding or node provisioning fails when the cluster's VNet is in a different resource group

**Symptoms**

* The onboarding script or `atlasctl` fails while creating the role assignment, with an Azure authorization error similar to `AuthorizationFailed` for `Microsoft.Authorization/roleAssignments/write`.
* Onboarding completes but Atlas cannot provision nodes, with errors that mention networking (creating network interfaces or joining a subnet).

**Cause**

Atlas's `AtlasAKSRole` includes networking actions (for example, `Microsoft.Network/networkInterfaces/*` and `Microsoft.Network/virtualNetworks/subnets/join/action`) that must apply to the resource group containing the cluster's VNet or subnet. When that VNet is in a *different* resource group than the cluster, onboarding tries to assign the role to that networking resource group too — but the identity running onboarding often has permission to create role assignments only in the cluster's resource group. This typically requires action from the team that owns the networking resource group.

**Solution**

1. Identify the resource group that contains your cluster's VNet or subnet. You can read it from the subnet ID of an agent pool:

   ```shell
   az aks show --name <cluster-name> --resource-group <cluster-resource-group> \
     --query "agentPoolProfiles[0].vnetSubnetId" --output tsv
   ```

2. Grant the identity running onboarding the ability to create role assignments in that networking resource group (a role that allows `Microsoft.Authorization/roleAssignments/write`, such as `Owner` or `User Access Administrator`, scoped to the networking resource group).

3. Re-run the onboarding script or `atlasctl`. For Terraform, ensure the VNet's resource group is passed via the `additional_resource_groups` variable of the [Atlas Terraform module for AKS](https://registry.terraform.io/modules/atlas/aks/atlas/latest).

   Alternatively, assign `AtlasAKSRole-<cluster-id>` to the VNet's resource group directly:

   ```shell
   az role assignment create \
     --assignee <atlas-service-principal-object-id> \
     --role "AtlasAKSRole-<cluster-id>" \
     --scope "/subscriptions/<subscription-id>/resourceGroups/<vnet-resource-group>"
   ```

For more detail, see [Resource groups that `AtlasAKSRole` is assigned to](cloud-permissions.md#resource-groups-that-atlasaksrole-is-assigned-to).