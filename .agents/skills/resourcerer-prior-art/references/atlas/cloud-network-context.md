---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Cloud network context

Enrich network flow data with cloud provider region, zone, and subnet information.

Cloud network context enriches your network flow data with cloud provider context (VPC/subnet CIDRs, regions, availability zones, peering connections, and Transit Gateway topology). This lets you attribute network flows to specific cloud regions and availability zones for cost analysis and security visibility.

## Overview

When enabled, Sentinel's cloud state controller:

1. Fetches VPC network state from your cloud provider API on startup
2. Caches network state and refreshes it periodically (default: every hour)
3. Fetches cloud service IP ranges (e.g. `ip-ranges.amazonaws.com`) separately — no auth needed, refreshed daily
4. Enriches network flow events with VPC context (region, zone, subnet info)

IP lookup uses a three-tier priority: **static mappings > cloud-discovered > public cloud service ranges**.

> ⚠️ Warning
>
> Zone information is not available for GCP. GCP subnets are regional, not zonal, so enriched flows will include region but not availability zone.

### Before you start

* Sentinel installed and running ([Installation guide](sec-sentinel-installation-and-upgrading.md))
* Network traffic monitoring enabled (`--set agent.extraArgs.netflow-enabled=true`) — see [Configuring Sentinel features](sec-configuring-sentinel.md)
* Your cloud provider VPC ID (AWS) or VPC network name (GCP)
* IAM permissions configured (see below)

## Configure cloud provider authentication

<Tabs>
  <Tab title="AWS (EKS)">
    ### IAM policy

    Create an IAM policy with the following permissions. This covers VPC, peering, and Transit Gateway discovery. If you don't use Transit Gateway, the `TransitGateway*` and `SearchTransitGatewayRoutes` actions are unused but harmless.

    ```json
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "ec2:DescribeVpcs",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcPeeringConnections",
            "ec2:DescribeTransitGatewayAttachments",
            "ec2:DescribeTransitGatewayPeeringAttachments",
            "ec2:DescribeTransitGatewayRouteTables",
            "ec2:SearchTransitGatewayRoutes"
          ],
          "Resource": "*"
        }
      ]
    }
    ```

    ### IRSA (recommended)

    **Step 1: Create IAM policy**

    ```bash
    export AWS_REGION="us-east-1"
    export POLICY_NAME="SentinelVPCReaderPolicy"

    cat > sentinel-vpc-policy.json <<'EOF'
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "ec2:DescribeVpcs",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcPeeringConnections",
            "ec2:DescribeTransitGatewayAttachments",
            "ec2:DescribeTransitGatewayPeeringAttachments",
            "ec2:DescribeTransitGatewayRouteTables",
            "ec2:SearchTransitGatewayRoutes"
          ],
          "Resource": "*"
        }
      ]
    }
    EOF

    aws iam create-policy \
      --policy-name ${POLICY_NAME} \
      --policy-document file://sentinel-vpc-policy.json \
      --region ${AWS_REGION}
    ```

    **Step 2: Create IAM role for service account**

    <Accordion title="Option A: Use existing Sentinel service account (recommended)" icon="fa-star">

    ```bash
    export CLUSTER_NAME="your-eks-cluster"
    export NAMESPACE="atlas-agent"
    export SERVICE_ACCOUNT_NAME="atlas-sentinel-controller"
    export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

    eksctl utils associate-iam-oidc-provider \
      --cluster ${CLUSTER_NAME} --region ${AWS_REGION} --approve

    export OIDC_PROVIDER=$(aws eks describe-cluster --name ${CLUSTER_NAME} --region ${AWS_REGION} \
      --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")

    cat > trust-policy.json <<EOF
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": {
            "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/${OIDC_PROVIDER}"
          },
          "Action": "sts:AssumeRoleWithWebIdentity",
          "Condition": {
            "StringEquals": {
              "${OIDC_PROVIDER}:sub": "system:serviceaccount:${NAMESPACE}:${SERVICE_ACCOUNT_NAME}",
              "${OIDC_PROVIDER}:aud": "sts.amazonaws.com"
            }
          }
        }
      ]
    }
    EOF

    aws iam create-role \
      --role-name SentinelVPCReaderRole \
      --assume-role-policy-document file://trust-policy.json

    aws iam attach-role-policy \
      --role-name SentinelVPCReaderRole \
      --policy-arn arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}

    export ROLE_ARN=$(aws iam get-role --role-name SentinelVPCReaderRole --query 'Role.Arn' --output text)
    echo "Role ARN: ${ROLE_ARN}"
    ```

    </Accordion>

    <Accordion title="Option B: Let eksctl create a new service account" icon="fa-cog">

    ```bash
    export CLUSTER_NAME="your-eks-cluster"
    export NAMESPACE="atlas-agent"
    export SERVICE_ACCOUNT_NAME="atlas-sentinel-controller"
    export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

    eksctl utils associate-iam-oidc-provider \
      --cluster ${CLUSTER_NAME} --region ${AWS_REGION} --approve

    eksctl create iamserviceaccount \
      --cluster=${CLUSTER_NAME} \
      --namespace=${NAMESPACE} \
      --name=${SERVICE_ACCOUNT_NAME} \
      --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME} \
      --region=${AWS_REGION} \
      --approve
    ```

    </Accordion>

    **Step 3: Configure Helm**

    For Option A (existing service account):

    ```yaml
    controller:
      extraArgs:
        cloud-provider: aws
        cloud-provider-vpc-sync-enabled: true
        cloud-provider-vpc-name: "vpc-0123456789abcdef0"

      serviceAccount:
        create: true
        annotations:
          eks.amazonaws.com/role-arn: "arn:aws:iam::YOUR_ACCOUNT_ID:role/SentinelVPCReaderRole"
    ```

    For Option B (eksctl-created service account), set `serviceAccount.create: false`.

    ### IAM user with access keys

    **Step 1: Create IAM user and policy**

    ```bash
    export USER_NAME="sentinel-vpc-reader"
    export POLICY_NAME="SentinelVPCReaderPolicy"
    export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

    cat > sentinel-vpc-policy.json <<'EOF'
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "ec2:DescribeVpcs",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcPeeringConnections",
            "ec2:DescribeTransitGatewayAttachments",
            "ec2:DescribeTransitGatewayPeeringAttachments",
            "ec2:DescribeTransitGatewayRouteTables",
            "ec2:SearchTransitGatewayRoutes"
          ],
          "Resource": "*"
        }
      ]
    }
    EOF

    aws iam create-policy --policy-name ${POLICY_NAME} --policy-document file://sentinel-vpc-policy.json
    aws iam create-user --user-name ${USER_NAME}
    aws iam attach-user-policy --user-name ${USER_NAME} \
      --policy-arn arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${POLICY_NAME}
    aws iam create-access-key --user-name ${USER_NAME} > access-key.json
    ```

    **Step 2: Create Kubernetes secret**

    ```bash
    export AWS_ACCESS_KEY_ID=$(jq -r '.AccessKey.AccessKeyId' access-key.json)
    export AWS_SECRET_ACCESS_KEY=$(jq -r '.AccessKey.SecretAccessKey' access-key.json)

    kubectl create secret generic aws-credentials \
      --from-literal=AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
      --from-literal=AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
      --namespace atlas-agent

    rm access-key.json
    ```

    **Step 3: Configure Helm**

    ```yaml
    controller:
      extraArgs:
        cloud-provider: aws
        cloud-provider-vpc-sync-enabled: true
        cloud-provider-vpc-name: "vpc-0123456789abcdef0"

      envFrom:
        - secretRef:
            name: aws-credentials
    ```

    > 📘 AWS zone IDs
    >
    > By default, Sentinel uses zone names (e.g. `us-east-1a`). AWS also exposes zone IDs (e.g. `use1-az1`) which are consistent across accounts — useful for correct cross-account traffic attribution. Enable with `--set controller.extraArgs.cloud-provider-aws-use-zone-id=true`.
    >
    > When this flag is enabled, the `zone` field in any [static CIDR mappings](#static-cidr-mappings) must also use zone IDs (e.g. `use1-az1`), not zone names.
  </Tab>
  <Tab title="GCP (GKE)">
    ### IAM permissions

    The Sentinel controller service account needs the following GCP IAM permissions:

    - `compute.networks.list`
    - `compute.networks.listPeeringRoutes`
    - `compute.subnetworks.list`

    **Recommended IAM role:** `roles/compute.networkViewer`

    ### Workload Identity (recommended)

    **Step 1: Create GCP service account**

    ```bash
    export PROJECT_ID="your-gcp-project-id"
    export SERVICE_ACCOUNT_NAME="sentinel-vpc-reader"

    gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME} \
      --project=${PROJECT_ID} \
      --display-name="Sentinel VPC State Reader"

    gcloud projects add-iam-policy-binding ${PROJECT_ID} \
      --project=${PROJECT_ID} \
      --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
      --role="roles/compute.networkViewer"
    ```

    **Step 2: Enable Workload Identity binding**

    ```bash
    export NAMESPACE="atlas-agent"
    export KSA_NAME="atlas-sentinel-controller"

    gcloud iam service-accounts add-iam-policy-binding \
      ${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com \
      --project=${PROJECT_ID} \
      --role roles/iam.workloadIdentityUser \
      --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]"
    ```

    **Step 3: Configure Helm**

    ```yaml
    controller:
      extraArgs:
        cloud-provider: gcp
        cloud-provider-gcp-project-id: "your-gcp-project-id"
        cloud-provider-vpc-sync-enabled: true
        cloud-provider-vpc-name: "your-vpc-name"

      nodeSelector:
        iam.gke.io/gke-metadata-server-enabled: "true"

      serviceAccount:
        create: true
        annotations:
          iam.gke.io/gcp-service-account: "sentinel-vpc-reader@your-gcp-project-id.iam.gserviceaccount.com"
    ```

    ### Service account key file

    **Step 1: Create GCP service account and key**

    ```bash
    export PROJECT_ID="your-gcp-project-id"
    export SERVICE_ACCOUNT_NAME="sentinel-vpc-reader"

    gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME} \
      --project=${PROJECT_ID} \
      --display-name="Sentinel VPC State Reader"

    gcloud projects add-iam-policy-binding ${PROJECT_ID} \
      --project=${PROJECT_ID} \
      --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
      --role="roles/compute.networkViewer"

    gcloud iam service-accounts keys create sentinel-sa-key.json \
      --project=${PROJECT_ID} \
      --iam-account=${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com
    ```

    **Step 2: Create Kubernetes secret**

    ```bash
    kubectl create secret generic gcp-credentials \
      --from-file=credentials.json=sentinel-sa-key.json \
      --namespace atlas-agent

    rm sentinel-sa-key.json
    ```

    **Step 3: Configure Helm**

    ```yaml
    controller:
      extraArgs:
        cloud-provider: gcp
        cloud-provider-vpc-sync-enabled: true
        cloud-provider-gcp-project-id: "your-gcp-project-id"
        cloud-provider-vpc-name: "your-vpc-name"

      env:
        - name: GCP_CREDENTIALS_FILE
          value: /var/secrets/gcp/credentials.json

      volumeMounts:
        - name: gcp-credentials
          mountPath: /var/secrets/gcp
          readOnly: true

      volumes:
        - name: gcp-credentials
          secret:
            secretName: gcp-credentials
    ```
  </Tab>
</Tabs>

> 📘 Azure support
>
> Azure cloud network context is coming soon.

## Enable VPC state discovery

After configuring authentication, enable VPC state discovery with Helm:

<Callout icon="📘" theme="info">
  ### Which upgrade method to use

  The Helm commands on this page use the **umbrella chart** (`atlas-helm/atlas`) by default. If you need to use a different method:

  - **[atlasctl](connect-with-atlasctl.md)**: To upgrade all Atlas components at once without managing Helm flags:
    ```shell
    atlasctl atlasware upgrade
    ```
    This preserves your existing configuration. See the [atlasctl documentation](connect-with-atlasctl.md) for installation and authentication instructions.
  - **Individual charts**: If you installed each component as a separate Helm release (e.g., for ArgoCD or custom GitOps), replace the release name and chart reference with the component-specific ones (e.g., `atlas-workload-autoscaler` and `atlas-helm/atlas-workload-autoscaler`) and remove the `autoscaler.atlas-workload-autoscaler.` value prefix.

  Not sure which method you used? Run `helm list -n atlas-agent`. A single release named `atlas` means umbrella chart; separate releases like `atlas-workload-autoscaler` mean individual charts.
</Callout>

<br />

<Tabs>
  <Tab title="AWS (EKS)">
    ```shell
    helm upgrade atlas atlas-helm/atlas -n atlas-agent \
      --reset-then-reuse-values \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider=aws \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-aws-region=<your-aws-region> \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-vpc-sync-enabled=true \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-vpc-name=<your-vpc-id>
    ```
  </Tab>
  <Tab title="GCP (GKE)">
    ```shell
    helm upgrade atlas atlas-helm/atlas -n atlas-agent \
      --reset-then-reuse-values \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider=gcp \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-vpc-sync-enabled=true \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-vpc-name=<your-vpc-name> \
      --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-gcp-project-id=<your-project-id>
    ```
  </Tab>
</Tabs>

### Helm flags reference

| Flag                                        | Description                                                                                                | Default |
| ------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ------- |
| `cloud-provider`                            | Cloud provider: `aws` or `gcp`                                                                             | -       |
| `cloud-provider-vpc-sync-enabled`           | Enable VPC state discovery                                                                                 | `false` |
| `cloud-provider-vpc-name`                   | VPC ID (AWS) or network name (GCP)                                                                         | -       |
| `cloud-provider-vpc-sync-interval`          | How often to refresh VPC state                                                                             | `1h`    |
| `cloud-provider-vpc-cache-size`             | LRU cache size for IP lookups                                                                              | `10000` |
| `cloud-provider-gcp-project-id`             | GCP project ID (GCP only)                                                                                  | -       |
| `cloud-provider-aws-region`                 | AWS region of the cluster. Recommended; required for correct TGW region attribution (AWS only)             | -       |
| `cloud-provider-aws-use-zone-id`            | Use zone IDs (`use1-az1`) instead of zone names (`us-east-1a`) (AWS only)                                  | `false` |
| `cloud-provider-aws-cross-account-role`     | Cross-account role ARN template (see [Transit Gateway discovery](#aws-transit-gateway-discovery-aws-only)) | -       |
| `cloud-provider-public-cidrs-sync-interval` | How often to refresh public cloud service IP ranges                                                        | `24h`   |
| `cloud-provider-static-cidrs-file`          | Path to static CIDR YAML (see [Static CIDR mappings](#static-cidr-mappings))                               | -       |

All flags are set under `controller.extraArgs` in the Helm values.

## Advanced configuration

### AWS Transit Gateway discovery (AWS only)

When your cluster VPC is connected to other VPCs via AWS Transit Gateway, Sentinel can automatically discover remote VPCs and their CIDR ranges, removing the need for manual static CIDR mappings for TGW-connected networks.

#### How it works

1. Sentinel discovers Transit Gateway attachments for the cluster VPC
2. For each TGW, it enumerates all VPC and peering attachments (including remote accounts)
3. It queries TGW route tables to discover destination CIDRs per attachment
4. If a cross-account role ARN template is configured, Sentinel assumes that role via STS to fetch subnet-level detail (zone, zone ID) from remote accounts
5. For TGW peering attachments, it resolves the peer TGW's region and discovers VPCs behind it

```text
                        ┌──────────────────────────────────────────────────────┐
                        │              Same Region (e.g. us-east-1)           │
                        │                                                      │
┌─────────────────┐     │     ┌──────────────────┐     ┌─────────────────┐    │
│  Cluster VPC    │     │     │  Transit Gateway  │     │  Remote VPC A   │    │
│  Account 111    │─────┼────▶│                   │◀────│  Account 222    │    │
│  10.0.0.0/16    │     │     │  TGW route table  │     │  10.1.0.0/16    │    │
└─────────────────┘     │     │  has CIDR routes  │     └─────────────────┘    │
                        │     │  for all attached  │                            │
                        │     │  VPCs              │     ┌─────────────────┐    │
                        │     │                   │◀────│  Remote VPC B   │    │
                        │     └──────────────────┘     │  Account 333    │    │
                        │                               │  10.2.0.0/16    │    │
                        │                               └─────────────────┘    │
                        └──────────────────────────────────────────────────────┘
```

#### Basic setup (route-level CIDRs only)

TGW discovery works automatically when VPC sync is enabled — no extra flags needed. Sentinel discovers TGW-attached VPCs and indexes their route-table CIDRs (region attribution, no per-AZ detail).

Requires the IAM permissions from the [AWS authentication section](#configure-cloud-provider-authentication) (the `DescribeTransitGateway*` and `SearchTransitGatewayRoutes` actions).

```yaml
controller:
  extraArgs:
    cloud-provider: aws
    cloud-provider-aws-region: "us-east-1"
    cloud-provider-vpc-sync-enabled: true
    cloud-provider-vpc-name: "vpc-0123456789abcdef0"
```

> ⚠️ Important
>
> Set `cloud-provider-aws-region` to your cluster's AWS region. This is used as the region for discovered TGW VPCs and for targeting cross-account API calls. Without it, discovered VPCs may have no region attribution in netflow records.

#### Cross-account subnet discovery (optional)

For full subnet-level detail (zone names and zone IDs), configure a cross-account IAM role that Sentinel can assume in each remote account.

<Accordion title="Step 1: Create IAM role in each remote account" icon="fa-key">

**Permissions policy** (minimal — only needs to read subnets and TGW attachments):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeSubnets",
        "ec2:DescribeTransitGatewayAttachments"
      ],
      "Resource": "*"
    }
  ]
}
```

**Trust policy** (allows the Sentinel role in the cluster account to assume this role):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::CLUSTER_ACCOUNT_ID:role/SentinelVPCReaderRole"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

Create the role using the AWS CLI (run once per remote account):

```bash
export CLUSTER_ACCOUNT_ID="123456789012"

cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::${CLUSTER_ACCOUNT_ID}:role/SentinelVPCReaderRole"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

cat > permissions-policy.json <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeSubnets",
        "ec2:DescribeTransitGatewayAttachments"
      ],
      "Resource": "*"
    }
  ]
}
EOF

aws iam create-role \
  --role-name SentinelCrossAccountReader \
  --assume-role-policy-document file://trust-policy.json \
  --description "Allows sentinel in cluster account to read subnet info for TGW discovery"

REMOTE_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

aws iam create-policy \
  --policy-name SentinelCrossAccountReaderPolicy \
  --policy-document file://permissions-policy.json

aws iam attach-role-policy \
  --role-name SentinelCrossAccountReader \
  --policy-arn arn:aws:iam::${REMOTE_ACCOUNT_ID}:policy/SentinelCrossAccountReaderPolicy

rm trust-policy.json permissions-policy.json
```

</Accordion>

<Accordion title="Step 2: Add sts:AssumeRole to the cluster account role" icon="fa-shield">

The Sentinel role in the cluster account needs permission to assume the remote roles:

```json
{
  "Effect": "Allow",
  "Action": "sts:AssumeRole",
  "Resource": "arn:aws:iam::*:role/SentinelCrossAccountReader"
}
```

Using the AWS CLI (run in the cluster account):

```bash
cat > assume-policy.json <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": "arn:aws:iam::*:role/SentinelCrossAccountReader"
    }
  ]
}
EOF

aws iam put-role-policy \
  --role-name SentinelVPCReaderRole \
  --policy-name SentinelCrossAccountAssumePolicy \
  --policy-document file://assume-policy.json

rm assume-policy.json
```

</Accordion>

**Step 3: Configure the cross-account role ARN template**

Add the `cloud-provider-aws-cross-account-role` flag. Sentinel replaces `{account-id}` with the actual AWS account ID of each remote VPC discovered via the Transit Gateway.

```yaml
controller:
  extraArgs:
    cloud-provider: aws
    cloud-provider-aws-region: "us-east-1"
    cloud-provider-vpc-sync-enabled: true
    cloud-provider-vpc-name: "vpc-0123456789abcdef0"
    cloud-provider-aws-cross-account-role: "arn:aws:iam::{account-id}:role/SentinelCrossAccountReader"
```

#### Fallback behavior

| Configuration                         | What you get                                                                  |
| ------------------------------------- | ----------------------------------------------------------------------------- |
| VPC sync only (no cross-account role) | TGW route CIDRs, region only. Auto-discovers connected VPCs but no AZ detail. |
| VPC sync + cross-account role         | Subnet CIDRs, zone name + zone ID. Full zone-aware network mapping.           |
| Role configured but STS fails         | Falls back to TGW route CIDRs (logged as warning, non-fatal).                 |

### Static CIDR mappings

Static CIDR mappings let you manually annotate IP ranges that are not automatically discovered, such as on-premises networks, VPN endpoints, or cross-account VPCs without Transit Gateway connectivity.

Static entries take **highest priority** in IP lookups. A static `/32` will override a broader cloud-discovered subnet.

#### When to use

* The destination is in another account or cloud and not reachable via the cloud API
* You use Direct Connect, Site-to-Site VPN, or other connectivity not automatically discovered
* You have managed services (RDS, Cloud SQL, etc.) with known private IPs
* Cloud provider API discovery is not enabled

#### Configuration via Helm values

Add `staticCIDRs` under `controller.netflow` in your `values.yaml`. Helm creates a ConfigMap and mounts it automatically when the list is non-empty:

<Tabs>
  <Tab title="AWS example">
    ```yaml
    controller:
      netflow:
        staticCIDRs:
          mappings:
            - cidr: "10.1.0.0/24"
              zone: "us-east-1a"
              region: "us-east-1"
              name: "production-vpc"
              kind: "VPC"
              connectivityMethod: "TransitGateway"

            - cidr: "10.0.0.13/32"
              zone: "us-east-1a"
              region: "us-east-1"
              name: "production-rds"
              kind: "RDS"
              connectivityMethod: "PrivateLink"

            - cidr: "10.2.0.0/16"
              region: "us-east-1"
              name: "peered-vpc"
              kind: "VPC"
              connectivityMethod: "VPCPeering"
    ```
  </Tab>
  <Tab title="GCP example">
    ```yaml
    controller:
      netflow:
        staticCIDRs:
          mappings:
            - cidr: "10.128.0.0/20"
              zone: "us-central1-a"
              region: "us-central1"
              name: "production-subnet"
              kind: "VPC"
              connectivityMethod: "VPCPeering"

            - cidr: "10.200.0.0/32"
              region: "us-central1"
              name: "cloud-sql-instance"
              kind: "CloudSQL"
              connectivityMethod: "PrivateLink"

            - cidr: "172.16.0.0/16"
              region: "europe-west1"
              name: "on-prem-datacenter"
              kind: "OnPrem"
              connectivityMethod: "CloudVPN"
    ```
  </Tab>
</Tabs>

#### Mapping fields

| Field                | Required | Description                                                                                                                                                                                               |
| -------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cidr`               | Yes      | IP range in CIDR notation (e.g. `10.1.0.0/24`) or single IP (`10.0.0.1/32`)                                                                                                                               |
| `region`             | Yes      | Cloud region (e.g. `us-east-1` for AWS, `us-central1` for GCP)                                                                                                                                            |
| `zone`               | No       | Availability zone (e.g. `us-east-1a` for AWS, `us-central1-a` for GCP). Leave empty for regional ranges. For AWS, when `cloud-provider-aws-use-zone-id: true` is set, use zone IDs here (e.g. `use1-az1`) |
| `name`               | No       | Human-readable name (e.g. `production-rds`, `cloud-sql-instance`)                                                                                                                                         |
| `kind`               | No       | Resource type (e.g. `VPC`, `RDS`, `CloudSQL`, `OnPrem`)                                                                                                                                                   |
| `connectivityMethod` | No       | Connectivity label — see values below                                                                                                                                                                     |

#### connectivityMethod values

This field is optional and free-form. Recommended values for common networking paths:

| Value               | Description                       | Provider |
| ------------------- | --------------------------------- | -------- |
| `VPCPeering`        | VPC Peering / VPC Network Peering | AWS, GCP |
| `TransitGateway`    | AWS Transit Gateway               | AWS      |
| `PrivateLink`       | AWS PrivateLink / VPC Endpoints   | AWS      |
| `DirectConnect`     | AWS Direct Connect                | AWS      |
| `SiteToSiteVPN`     | AWS Site-to-Site VPN              | AWS      |
| `CloudVPN`          | GCP Cloud VPN                     | GCP      |
| `CloudInterconnect` | GCP Cloud Interconnect            | GCP      |
| `NATGateway`        | NAT Gateway / Cloud NAT           | AWS, GCP |
| `IntraVPC`          | Same VPC / same VPC network       | AWS, GCP |

Custom values (e.g. `"on-prem-mpls"`, `"ExpressRoute"`) are passed through as-is in netflow records.

> 📘 Note
>
> Sentinel does not perform any cost calculations. The `connectivityMethod` field is purely a label attached to netflow records so that downstream systems (e.g. Atlas cost attribution) can distinguish traffic by connectivity type.

#### Using a pre-existing ConfigMap

If you manage your own ConfigMap (e.g. via GitOps):

```yaml
controller:
  extraArgs:
    cloud-provider-static-cidrs-file: /etc/sentinel/my-cidrs/static-cidrs.yaml

  extraVolumes:
    - name: my-static-cidrs
      configMap:
        name: my-existing-configmap

  extraVolumeMounts:
    - name: my-static-cidrs
      mountPath: /etc/sentinel/my-cidrs
      readOnly: true
```

The ConfigMap file must be valid YAML:

```yaml
staticCIDRMappings:
  - cidr: "10.1.0.0/24"
    zone: "us-east-1a"
    region: "us-east-1"
    name: "production-vpc"
    kind: "VPC"
    connectivityMethod: "TransitGateway"
  - cidr: "10.128.0.0/20"
    region: "us-central1"
    name: "gcp-subnet"
    kind: "VPC"
    connectivityMethod: "VPCPeering"
```

#### Combining with cloud discovery

Static mappings and cloud provider discovery work together. Static entries always win for the same IP:

<Tabs>
  <Tab title="AWS example">
    ```yaml
    controller:
      extraArgs:
        cloud-provider: aws
        cloud-provider-vpc-sync-enabled: true
        cloud-provider-vpc-name: "vpc-0123456789abcdef0"

      netflow:
        staticCIDRs:
          mappings:
            # This /32 overrides whatever the cloud API says about 10.0.0.13
            - cidr: "10.0.0.13/32"
              region: "us-east-1"
              name: "cross-account-rds"
              connectivityMethod: "TransitGateway"
    ```
  </Tab>
  <Tab title="GCP example">
    ```yaml
    controller:
      extraArgs:
        cloud-provider: gcp
        cloud-provider-vpc-sync-enabled: true
        cloud-provider-vpc-name: "my-vpc-network"
        cloud-provider-gcp-project-id: "my-project-id"

      netflow:
        staticCIDRs:
          mappings:
            # This /32 overrides whatever the cloud API says about 10.200.0.5
            - cidr: "10.200.0.5/32"
              region: "us-central1"
              name: "cloud-sql-primary"
              connectivityMethod: "PrivateLink"
    ```
  </Tab>
</Tabs>

## Verify the configuration

Check that VPC state sync is working:

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller -n atlas-agent | grep -i "vpc\|cloud"
```

Look for log messages indicating successful VPC state sync, such as:

* VPC state loaded with subnet/CIDR counts
* Cloud service IP ranges fetched successfully

If Transit Gateway discovery is enabled, you should also see logs about TGW attachments and route table entries being indexed.

## Troubleshooting

### VPC sync not starting

Verify the cloud provider flags are set:

```shell
kubectl get deployment atlas-sentinel-controller -n atlas-agent -o yaml | grep -A 20 "args"
```

Ensure `cloud-provider-vpc-sync-enabled` is `true` and `cloud-provider-vpc-name` is set.

### Authentication errors

Check controller logs for credential errors:

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller -n atlas-agent | grep -i "error\|auth\|credential\|forbidden"
```

<Tabs>
  <Tab title="AWS (EKS)">
    Common causes:
    - **IRSA**: Service account annotation missing or OIDC provider not configured
    - **Access keys**: Secret not created in the correct namespace (`atlas-agent`)
    - **Insufficient permissions**: IAM policy missing required `ec2:Describe*` actions
  </Tab>
  <Tab title="GCP (GKE)">
    Common causes:
    - **Workload Identity**: Binding not configured, or `nodeSelector` for `iam.gke.io/gke-metadata-server-enabled` missing
    - **Service account key**: Secret not mounted correctly or `GCP_CREDENTIALS_FILE` env var not set
    - **Insufficient permissions**: GCP service account missing `roles/compute.networkViewer` role
    - **Project ID**: `cloud-provider-gcp-project-id` not set or incorrect
  </Tab>
</Tabs>

### Cross-account role assumption failing (AWS only)

If TGW cross-account discovery falls back to route-level CIDRs, check for STS errors:

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller -n atlas-agent | grep -i "assume\|sts"
```

Verify the trust policy in the remote account allows the Sentinel role ARN and that the cluster account role has `sts:AssumeRole` permission.

<Cards columns={2}>
  <Card title="Configuring Sentinel" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Configure cloud network context settings.
  </Card>
  <Card title="Network observability" href="https://docs.atlas.example/docs/network-observability" icon="fa-diagram-project">
    Monitor network flows and identify service communication patterns.
  </Card>
</Cards>