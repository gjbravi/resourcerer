---
updatedAt: 2026-09-22T08:46:05.000Z
---

# GKE service account impersonation

> 🚧 Limited Availability Feature
>
> This feature is currently available through feature flags. Contact us to enable access for your organization.

Atlas supports GKE service account impersonation as an alternative to using service account keys. This method enhances security by eliminating the need for key management while allowing Atlas to access your GCP resources.

## How impersonation works

Atlas uses a two-tier service account architecture for GKE impersonation:

**Organization-level impersonation service account**: Atlas creates one impersonation service account per organization, shared across all clusters within that organization. This service account uses the naming pattern `atlas-gke-<hash>@prod-atlas-identity.iam.gserviceaccount.com`.

**Cluster-specific service accounts**: Each cluster maintains its own dedicated GCP service account in your project, following the `atlas-gke-<cluster-name-hash>` naming convention for direct resource management.

## Prerequisites

Before setting up impersonation, ensure you have:

* An existing GCP service account with the necessary permissions for your GKE cluster
* The Atlas API token and cluster ID
* The `gcloud` CLI configured with appropriate permissions
* The `jq` command-line JSON processor installed

## Setup process

### Using the Terraform module (optional)

Atlas provides a [Terraform module for GKE IAM impersonation](https://registry.terraform.io/modules/atlas/iam-impersonate/gke/latest) that automates the IAM policy configuration process. This module handles the complex IAM policy bindings and conditions required for impersonation.

When you use the Terraform module, you **do not** need to call the `/gcp-create-sa` API endpoint manually. The module creates a `atlas_gke_cluster_id` resource with `client_service_account` configured, which causes the Atlas Terraform provider to call `/gcp-create-sa` automatically during `terraform apply`.

### Step 1: Register the impersonation service account (manual setups only)

If you are **not** using the Terraform module, call the `/gcp-create-sa` API endpoint to register your service account for impersonation. This step is required when you configure impersonation manually or through infrastructure-as-code that does not include a `atlas_gke_cluster_id` resource with `client_service_account` set.

If you use the Terraform module described above, this API call is performed automatically and you can skip this step.

```shell
curl -X POST \
  -H "X-API-Key: $ATLAS_API_TOKEN" \
  "$ATLAS_API_URL/v1/kubernetes/external-clusters/$ATLAS_CLUSTER_ID/gcp-create-sa" \
  -d '{
    "gke": {
      "project_id": "'"$PROJECT_ID"'",
      "gke_sa_impersonate": "'"$SERVICE_ACCOUNT_EMAIL"'"
    }
  }'
```

> 📘 Note
>
> This API call returns the same impersonation service account for all clusters in your organization. While the service account remains consistent, each API call is required to:
>
> * Link the GCP project to the specific cluster in Atlas's system
> * Associate the impersonated service account with the cluster configuration
> * Enable proper permissions mapping for cluster operations

### Step 2: Configure service account permissions

Grant the following permissions to your service account:

| Permission                             | Purpose                                      |
| -------------------------------------- | -------------------------------------------- |
| `roles/iam.serviceAccountUser`         | Allows Atlas to act as the service account |
| `roles/iam.serviceAccountTokenCreator` | Enables token generation for impersonation   |
| `compute.subnetworks.useExternalIp`    | Required for network operations              |
| `compute.networks.useExternalIp`       | Required for network operations              |

### Step 3: Set configuration variables

Configure your environment with the following settings:

* Set `ATLAS_IMPERSONATE=true` in your environment variables
* Include both `gkeSaImpersonate` and `projectId` fields in your GKE configuration block

## Implementation example

The following script demonstrates the complete impersonation setup process:

```shell
if [[ -n $ATLAS_IMPERSONATE ]]; then
  echo "Registering service account for impersonation: $SERVICE_ACCOUNT_EMAIL"
  
  # Register service account with Atlas
  RESPONSE=$(curl -sSL --write-out "HTTP_STATUS:%{http_code}" \
    -X POST -H "X-API-Key: $ATLAS_API_TOKEN" \
    "$ATLAS_API_URL/v1/kubernetes/external-clusters/$ATLAS_CLUSTER_ID/gcp-create-sa" \
    -d '{"gke":{"project_id":"'$PROJECT_ID'","gke_sa_impersonate":"'$SERVICE_ACCOUNT_EMAIL'"}}')
  
  RESPONSE_STATUS=$(echo "$RESPONSE" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')
  RESPONSE_BODY=$(echo "$RESPONSE" | sed -e 's/HTTP_STATUS\:.*//g')
  
  if [[ $RESPONSE_STATUS != "200" ]]; then
    echo "Failed to register service account for impersonation. HTTP status: $RESPONSE_STATUS"
    echo $RESPONSE_BODY
    exit 1
  fi

  # Extract Atlas service account from response
  ATLAS_SERVICE_ACCOUNT=$(echo "$RESPONSE_BODY" | jq -r '.serviceAccountEmail')
  echo "Atlas impersonation service account: $ATLAS_SERVICE_ACCOUNT"
  
  if [[ "$ATLAS_SERVICE_ACCOUNT" == "" || "$ATLAS_SERVICE_ACCOUNT" == "null" ]]; then
    echo "Failed to retrieve Atlas service account from response"
    echo $RESPONSE_BODY
    exit 1
  fi

  # Clean up existing IAM policy bindings
  echo "Removing existing IAM policy bindings"
  gcloud projects remove-iam-policy-binding $SERVICE_ACCOUNT_EMAIL \
    --member="serviceAccount:$ATLAS_SERVICE_ACCOUNT" \
    --project $PROJECT_ID \
    --role='roles/iam.serviceAccountUser' \
    --all --no-user-output-enabled >/dev/null 2>&1
    
  gcloud projects remove-iam-policy-binding $SERVICE_ACCOUNT_EMAIL \
    --member="serviceAccount:$ATLAS_SERVICE_ACCOUNT" \
    --project $PROJECT_ID \
    --role='roles/iam.serviceAccountTokenCreator' \
    --all --no-user-output-enabled >/dev/null 2>&1
  
  # Grant token creator permissions
  echo "Configuring impersonation permissions"
  gcloud iam service-accounts add-iam-policy-binding $SERVICE_ACCOUNT_EMAIL \
    --member="serviceAccount:$ATLAS_SERVICE_ACCOUNT" \
    --role="roles/iam.serviceAccountTokenCreator" \
    --condition="title=AlwaysTrueCondition,description=This condition is always true,expression=true" \
    --project $PROJECT_ID

  # Grant impersonation permissions with conditional access
  gcloud iam service-accounts add-iam-policy-binding $SERVICE_ACCOUNT_EMAIL \
    --member="serviceAccount:$ATLAS_SERVICE_ACCOUNT" \
    --role="roles/iam.serviceAccountUser" \
    --condition="title=SpecificServiceAccountCondition,description=Allow impersonation only for Atlas service account,expression=request.auth.claims.email == \"$ATLAS_SERVICE_ACCOUNT\"" \
    --project $PROJECT_ID
    
  echo "Waiting for IAM permissions to propagate (180 seconds)"
  sleep 180

  # Update cluster configuration
  echo "Updating cluster configuration with impersonation settings"
  RESPONSE=$(curl -sSL --write-out "HTTP_STATUS:%{http_code}" \
    -X POST -H "X-API-Key: $ATLAS_API_TOKEN" \
    "$ATLAS_API_URL/v1/kubernetes/external-clusters/$ATLAS_CLUSTER_ID" \
    -d '{"credentials":"{}"}')
    
  RESPONSE_STATUS=$(echo "$RESPONSE" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')
  RESPONSE_BODY=$(echo "$RESPONSE" | sed -e 's/HTTP_STATUS\:.*//g')

  if [[ $RESPONSE_STATUS -eq 200 ]]; then
    echo "Impersonation setup completed successfully"
  else
    echo "Failed to update cluster configuration with impersonation settings"
    echo "Error details: HTTP $RESPONSE_STATUS - $RESPONSE_BODY"
    exit 1
  fi
fi
```

## Troubleshooting

### HTTP 500 Internal Server Error during cluster update

**Symptoms**: The cluster update fails with a 500 Internal Server Error after successfully configuring IAM permissions.

**Common causes**:

* The `/gcp-create-sa` API endpoint was not called before attempting the cluster update (skip this check if you are using the Terraform module, which calls it automatically)
* Mixing impersonation and non-impersonation authentication methods in the same script
* Insufficient IAM permissions on the service account
* IAM permission changes have not yet propagated

**Resolution steps**:

1. Verify that you called `/gcp-create-sa` before updating the cluster configuration
2. Ensure your script uses either impersonation or key-based authentication consistently
3. Confirm all required permissions are properly configured on your service account
4. Wait at least 3 minutes after configuring IAM permissions before updating the cluster