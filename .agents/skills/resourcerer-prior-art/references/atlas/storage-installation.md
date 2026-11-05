---
updatedAt: 2026-10-26T08:39:44.000Z
---

# Installation

Install Storage Autoscaling on GKE or EKS clusters with dry-run safety mode, cloud credentials, and Sentinel storage metrics.

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

## Prerequisites

Before you install Storage Autoscaling, confirm the following:

* A GKE or EKS cluster running Kubernetes 1.29 or later
* Helm 3 installed on your local machine
* kubectl configured to target your cluster
* A StorageClass with `allowVolumeExpansion: true`
* Sentinel with storage metrics enabled
* Cloud credentials for capacity expansion and IOPS/throughput management

### Check volume expansion support

Run the following command to verify that your StorageClasses allow volume expansion:

```shell
kubectl get storageclass -o yaml | grep allowVolumeExpansion
```

### Sentinel with storage metrics

Sentinel populates the Storage autoscaler dashboard in the Atlas console with savings estimates, storage type breakdowns, resize event logs, and per-PV usage data. Without it, the dashboard will be empty. Storage Autoscaling still functions without Sentinel, but you lose console visibility into what it is doing.

See [Configuring Sentinel: Kubernetes storage metrics](sec-configuring-sentinel.md#kubernetes-storage-metrics) for installation instructions.

> 📘 Note
> Sentinel also requires cloud volume metadata configuration to collect disk type, size, and provisioned IOPS data per volume.

## GKE identity and permissions

IOPS and throughput rightsizing on GKE requires a GCP service account with specific IAM roles, bound to the Kubernetes service account via Workload Identity Federation. The Helm chart also includes `nodes/proxy` RBAC for reading disk usage metrics.

### Required IAM roles

Create a GCP service account with the following IAM roles:

| IAM role                     | Purpose                                                                 |
| ---------------------------- | ----------------------------------------------------------------------- |
| `roles/compute.storageAdmin` | Read and modify disk IOPS, throughput, and size via the GCP Compute API |
| `roles/monitoring.viewer`    | Fetch actual disk IOPS/throughput consumption from GCP Cloud Monitoring |

### Create the GCP service account

Create the service account and add the IAM bindings:

```shell
PROJECT_ID=YOUR_PROJECT_ID # update with your project ID

gcloud iam service-accounts create atlas-storage-autoscaler \
  --project="${PROJECT_ID}" \
  --display-name="atlas-storage-autoscaler controller"

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:atlas-storage-autoscaler@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/compute.storageAdmin"

gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:atlas-storage-autoscaler@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/monitoring.viewer"
```

### Configure Workload Identity Federation

Bind the Kubernetes service account to the GCP service account using Workload Identity Federation:

```shell
NAMESPACE=atlas-agent
KSA_NAME=atlas-storage-autoscaler

gcloud iam service-accounts add-iam-policy-binding \
  "atlas-storage-autoscaler@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/iam.workloadIdentityUser" \
  --member="serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"
```

## EKS identity and permissions

IOPS and throughput rightsizing on EKS requires an IAM policy with specific EC2 and CloudWatch permissions, attached to the Kubernetes service account via IAM Roles for Service Accounts (IRSA). The Helm chart also includes `nodes/proxy` RBAC for reading disk usage metrics.

### Required IAM permissions

Create an IAM policy with the following permissions:

| IAM permission                     | Purpose                                                           | Required for           |
| ---------------------------------- | ----------------------------------------------------------------- | ---------------------- |
| `ec2:DescribeVolumes`              | Read current EBS volume state                                     | capacity + performance |
| `ec2:DescribeVolumesModifications` | Poll in-flight ModifyVolume operations                            | capacity + performance |
| `ec2:ModifyVolume`                 | Issue size/IOPS/throughput changes                                | capacity + performance |
| `ec2:DescribeInstanceTypes`        | Discover per-VM EBS performance caps                              | performance only       |
| `cloudwatch:GetMetricData`         | Fetch actual IOPS/throughput consumption over the lookback window | performance only       |

### Create the IAM policy

Create the policy and role using IAM Roles for Service Accounts (IRSA):

```shell
export AWS_REGION="YOUR_AWS_REGION" # update with your region
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export POLICY_NAME="AtlasStorageAutoscalerPolicy"

aws iam create-policy \
  --policy-name ${POLICY_NAME} \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "EBS",
        "Effect": "Allow",
        "Action": [
          "ec2:DescribeVolumes",
          "ec2:DescribeVolumesModifications",
          "ec2:ModifyVolume",
          "ec2:DescribeInstanceTypes"
        ],
        "Resource": "*"
      },
      {
        "Sid": "CloudWatch",
        "Effect": "Allow",
        "Action": [
          "cloudwatch:GetMetricData"
        ],
        "Resource": "*"
      }
    ]
  }'
```

### Create the IRSA role

Create an IAM role with a trust policy that allows the `atlas-storage-autoscaler` service account to assume it via OIDC federation:

```shell
export CLUSTER_NAME="<YOUR_CLUSTER_NAME>" # update with your cluster name
export NAMESPACE="atlas-agent"
export SERVICE_ACCOUNT_NAME="atlas-storage-autoscaler"

export OIDC_PROVIDER=$(aws eks describe-cluster \
  --name "${CLUSTER_NAME}" \
  --query "cluster.identity.oidc.issuer" \
  --output text | sed 's|https://||')

aws iam create-role \
  --role-name AtlasStorageAutoscalerRole \
  --assume-role-policy-document "{
    \"Version\": \"2012-10-17\",
    \"Statement\": [{
      \"Effect\": \"Allow\",
      \"Principal\": {
        \"Federated\": \"arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}\"
      },
      \"Action\": \"sts:AssumeRoleWithWebIdentity\",
      \"Condition\": {
        \"StringEquals\": {
          \"${OIDC_PROVIDER}:sub\": \"system:serviceaccount:${NAMESPACE}:${SERVICE_ACCOUNT_NAME}\"
        }
      }
    }]
  }"

aws iam attach-role-policy \
  --role-name AtlasStorageAutoscalerRole \
  --policy-arn "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}"
```

## Install the atlas-storage-autoscaler Helm chart

Start with a dry-run deployment so that planned actions are logged without modifying your PVCs.

Add the Atlas Helm repository:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

<Tabs>
<Tab title="GKE">

Install in dry-run mode with the Workload Identity Federation annotation on the service account:

```shell
helm upgrade --install atlas-storage-autoscaler atlas-helm/atlas-storage-autoscaler \
  --namespace atlas-agent --create-namespace \
  --set atlas.clusterID="<YOUR_ATLAS_AI_CLUSTER_ID>" \
  --set atlas.organizationID="<YOUR_ORGANIZATION_ID>" \
  --set atlas.apiKey="<YOUR_ATLAS_API_KEY>" \
  --set serviceAccount.annotations."iam\.gke\.io/gcp-service-account"="atlas-storage-autoscaler@${PROJECT_ID}.iam.gserviceaccount.com" \
  --set config.dryRun=true
```

</Tab>
<Tab title="EKS">

Install in dry-run mode with the IRSA role annotation on the service account:

```shell
helm upgrade --install atlas-storage-autoscaler atlas-helm/atlas-storage-autoscaler \
  --namespace atlas-agent --create-namespace \
  --set atlas.clusterID="<YOUR_ATLAS_AI_CLUSTER_ID>" \
  --set atlas.organizationID="<YOUR_ORGANIZATION_ID>" \
  --set atlas.apiKey="<YOUR_ATLAS_API_KEY>" \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"="arn:aws:iam::${AWS_ACCOUNT_ID}:role/AtlasStorageAutoscalerRole" \
  --set config.dryRun=true
```

</Tab>
</Tabs>

If Terraform or another tool already creates the Kubernetes service account with the annotation applied, add the following flags to skip creating a second one:

```shell
--set serviceAccount.create=false --set serviceAccount.name=<CUSTOM_SA_NAME>
```

> 📘 Note
> Atlas API key can also be provided from a preexisting kubernetes secret, via `--set atlas.apiKeySecretRef.name=SECRET_NAME` and `--set atlas.apiKeySecretRef.key=FIELD_KEY_IN_SECRET`.

> 📘 Note
> The installation creates a default PVPolicy in dry-run mode that matches all PVCs. No changes are made until you explicitly enable live mode. See [How it works](storage-how-it-works.md) for details.

## Verify installation

### Check pods

Confirm that the deployment is running. You should see two pods: one active leader and one standby replica.

```shell
kubectl get pods -n atlas-agent
```

### Check the default PVPolicy

Verify that the Helm chart created the default PVPolicy. You should see a policy with priority 0 and dry-run enabled.

```shell
kubectl get pvpolicy -n atlas-agent
```

### Check the matched policy on a PVC

Once the pods are running, Storage Autoscaling begins evaluating PVCs. Verify that a PVC has been matched to a policy:

```shell
kubectl get pvc <pvc-name> -o jsonpath='{.metadata.annotations.storage\.atlas\.example/matched-pvpolicy}'
```

If the output is empty, no policy matched. See [No PVPolicy matching a PVC](storage-troubleshooting.md#no-pvpolicy-matching-a-pvc) for troubleshooting steps.

### Review dry-run logs

Follow the logs to see what actions would be taken. Confirm that PVCs are being evaluated and that planned actions appear:

```shell
kubectl logs -n atlas-agent \
  -l app.kubernetes.io/name=atlas-storage-autoscaler \
  --follow
```

### Scrape Prometheus metrics (optional)

If you use Prometheus, forward the metrics port and confirm that PVCs are being tracked:

```shell
kubectl port-forward -n atlas-agent svc/atlas-storage-autoscaler-metrics 8080:8080
curl http://localhost:8080/metrics | grep atlas_storage_autoscaler_pvcs_monitored
```

Key metrics to watch:

* `atlas_storage_autoscaler_pvcs_monitored`
* `atlas_storage_autoscaler_pvc_usage_percent`
* `atlas_storage_autoscaler_pvc_growth_total`

### Inspect annotations on a PVC

Storage Autoscaling writes status annotations to each managed PVC. Inspect them to confirm that the PVC is being tracked:

```shell
kubectl get pvc <pvc-name> -o jsonpath='{.metadata.annotations}' | jq
```

Look for the following annotations:

* `storage.atlas.example/disk-type` confirms the cloud disk type was detected.
* `storage.atlas.example/current-iops` shows the current provisioned IOPS.
* `storage.atlas.example/operation-status` shows whether a resize is in progress.

On AWS, also check `storage.atlas.example/modifications-history` to see how many of the 4 ModifyVolume slots have been consumed in the rolling 24-hour window.

## Enable live mode

Once you have verified the installation and reviewed the dry-run logs, enable live mode.

### Option A: Patch the default PVPolicy

The quickest way to go live is to patch the default policy. This enables all three dimensions for every matched PVC at once:

```shell
kubectl patch pvpolicy default -n atlas-agent --type=merge -p '{"spec":{"dryRun":false}}'
```

### Option B: Create a scoped PVPolicy (recommended for production)

For more control, create a new PVPolicy with `dryRun: false` and a higher `priority` targeting only the PVCs you want to manage live. This leaves the default policy in dry-run mode for everything else, so you can roll out incrementally. See [Configuration](storage-configuration.md) for examples.

> ⚠️ Warning
> The default PVPolicy enables all three dimensions: `storageEnabled`, `iopsEnabled`, and `throughputEnabled`. Every eligible PVC is managed once you disable dry-run.
>
> Before enabling live mode, review the dry-run logs. You can also create a higher-priority PVPolicy with dry-run enabled to keep specific PVCs in observation mode, or use per-PVC `storage.atlas.example/*-enabled: "false"` annotations.

## Trigger a test resize (optional)

To verify that the full resize pipeline works end to end, set a desired capacity annotation on a PVC. Storage Autoscaling will expand the PVC to at least the specified size through the cloud API.

```shell
kubectl annotate pvc <pvc-name> \
  storage.atlas.example/desired-capacity="200Gi" \
  --overwrite
```

Watch the operation status. It should transition from `pending` to `running` to `done`:

```shell
kubectl get pvc <pvc-name> -w -o \
  jsonpath='{.metadata.annotations.storage\.atlas\.example/operation-status}'
```

If the status shows `failed`, check the `storage.atlas.example/last-error` annotation for details. See [Troubleshooting](storage-troubleshooting.md) for common issues.

## Next steps

<Cards columns={2}>
  <Card title="Configuration" href="https://docs.atlas.example/docs/storage-configuration" icon="fa-sliders">
    Configure PVPolicy Custom Resources, PVC annotations, and Helm values.
  </Card>
  <Card title="Troubleshooting" href="https://docs.atlas.example/docs/storage-troubleshooting" icon="fa-wrench">
    Resolve common issues with capacity expansion and performance rightsizing.
  </Card>
</Cards>