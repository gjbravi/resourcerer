---
updatedAt: 2026-10-30T14:16:09.000Z
---

# Connect clusters with separate Kubernetes and cloud owners

Connect a cluster when the Kubernetes team and the cloud team have separate access: the Kubernetes team runs `atlasctl cluster connect --k8s-only` first, then the cloud team registers cloud credentials via the Atlas API.

In many organizations, access is split: one team manages the Kubernetes cluster, and another team manages the cloud account. Atlas supports this workflow:

1. **Kubernetes team** connects the cluster with `atlasctl cluster connect --k8s-only`. This installs only the in-cluster Atlas components; no cloud-side resources are created and the cluster runs in read-only mode (Cost Monitoring).
2. **Cloud team** provisions the required cloud resources and registers the credentials with the Atlas API, without needing access to the Kubernetes cluster.
3. The cluster starts to reconcile and ends up connected — Node Autoscaling becomes available.

The cloud team needs only a Atlas API key and the cluster ID (an org admin or the Kubernetes team can provide both).

> 📘 Note
>
> The steps below cover the manual API flow. If both parts can run from one identity, simply re-running `atlasctl cluster connect` without `--k8s-only` performs the whole cloud setup automatically — pass `--gcp-credential-method impersonation` (GKE) or `--azure-credential-method impersonation` (AKS) to use impersonation.

## Step 1 (Kubernetes team): connect with `--k8s-only`

Run:

```shell
atlasctl cluster connect --k8s-only
```

Then hand the cluster ID and a Atlas API key to the cloud team.

## Step 2 (Cloud team): register cloud credentials

Follow the section for your provider.

### EKS: cross-account IAM role

1. Get the Atlas jumper role ARN:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" \
     "<API root URL>/v1/kubernetes/external-clusters/<cluster ID>/assume-role-user"
   ```

2. Create an IAM role in your AWS account. Allow the Atlas jumper role to assume it and grant it the required permissions (see [Cloud permissions](cloud-permissions.md#aws)).

3. Upload the IAM role to Atlas:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/external-clusters/<cluster ID>" \
     -d '{"eks":{"assumeRoleArn":"<your IAM role ARN>"}}'
   ```

4. Create a node configuration. Since `--k8s-only` does not create the Atlas instance profile, point the configuration at your own existing instance profile:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/clusters/<cluster ID>/node-configurations" \
     -d '{
       "name": "default",
       "subnets": ["<subnet ID>"],
       "eks": {
         "securityGroups": ["<security group ID>"],
         "instanceProfileArn": "<instance profile ARN>"
       }
     }'
   ```

   > ⚠️ Warning
   >
   > The role inside your instance profile must be authorized in the EKS cluster (EKS access entry or the `aws-auth` ConfigMap), or nodes will fail to join. Atlas manages this entry only for its own default role. This step requires Kubernetes access — coordinate with the Kubernetes team.

5. Set the node configuration as default:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/clusters/<cluster ID>/node-configurations/<node config ID>/default"
   ```

### GKE: service account impersonation

Atlas's organization-level impersonation service account impersonates a service account in your project using short-lived tokens. See [GKE service account impersonation](sa-impersonation-gke.md) for background.

1. Get the Atlas service account that will impersonate your service account:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/external-clusters/impersonation-service-account" \
     -d '{"provider":"gke"}'
   ```

   Save the `email` from the response.

2. Create a GCP service account that Atlas will impersonate and grant it the necessary permissions (see [Cloud permissions](cloud-permissions.md)).

3. Allow the Atlas service account to impersonate your service account:

   ```shell
   gcloud iam service-accounts add-iam-policy-binding "<your SA email>" \
     --member="serviceAccount:<Atlas SA email>" \
     --role="roles/iam.serviceAccountUser" \
     --condition="title=SpecificServiceAccountCondition,description=Allow impersonation only for ATLAS_SERVICE_ACCOUNT,expression=request.auth.claims.email == \"<Atlas SA email>\"" \
     --project "<GCP project ID>"

   gcloud iam service-accounts add-iam-policy-binding "<your SA email>" \
     --member="serviceAccount:<Atlas SA email>" \
     --role="roles/iam.serviceAccountTokenCreator" \
     --condition="title=AlwaysTrueCondition,description=This condition is always true,expression=true" \
     --project "<GCP project ID>"
   ```

   GCP IAM bindings can take a few minutes to propagate.

4. Upload your service account details to Atlas (both fields must be set together):

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/external-clusters/<cluster ID>" \
     -d '{"gke":{"targetServiceAccountEmail":"<your SA email>","atlasServiceAccountEmail":"<Atlas SA email>"}}'
   ```

### AKS: workload identity impersonation

A user-assigned managed identity in your subscription is trusted by a federated identity credential for Atlas's impersonation service account. See [AKS workload identity impersonation](sa-impersonation-aks.md) for background and the Terraform module alternative.

1. Get the Atlas impersonation principal:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/external-clusters/impersonation-service-account" \
     -d '{"provider":"aks"}'
   ```

   Save the numeric `id` from the response — it is used as the federated credential's subject (not the `email`).

2. Create the identity to impersonate:

   ```shell
   az identity create \
     --name <your identity name> \
     --resource-group <resource group> \
     --location <location>
   ```

3. Create a federated credential on the identity, granting the Atlas principal permission to impersonate it:

   ```shell
   az identity federated-credential create \
     --name <federated credential name> \
     --identity-name <your identity name> \
     --resource-group <resource group> \
     --issuer "https://accounts.google.com" \
     --subject "<Atlas impersonation principal ID>" \
     --audiences "api://AzureADTokenExchange"
   ```

   Save the credential's resource ID (the `id` field in the output).

4. Create the Atlas role and assign it to the identity, scoped to the cluster and node resource groups (see [Cloud permissions](cloud-permissions.md)).

   > ⚠️ Warning
   >
   > Assign the role at each resource group scope individually — assignments inherited from the subscription level are not accepted by credential validation. Use the custom role with the documented action list; built-in roles such as Contributor do not satisfy it.

5. Upload the federated credentials to Atlas:

   ```shell
   curl -s -H "Authorization: Bearer <API key>" -X POST \
     "<API root URL>/v1/kubernetes/external-clusters/<cluster ID>" \
     -d '{"aks":{},"credentials":"{\"subscriptionId\":\"<subscription ID>\",\"tenantId\":\"<tenant ID>\",\"clientId\":\"<identity client ID>\",\"federationId\":\"<federated credential resource ID>\"}"}'
   ```

## Step 3: automatic reconciliation

Once the credentials validate, the cluster leaves read-only mode and becomes connected. Continue with node configurations and a node policy to enable [Node Autoscaling](autoscaler.md).