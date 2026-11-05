---
updatedAt: 2026-06-04T13:33:07.000Z
---

# GCP Private Service Connect

# Connecting to Atlas through GCP Private Service Connect

Atlas supports connection through GCP Private Service Connect for GKE clusters without direct internet access. This enables secure communication between your private GKE cluster and Atlas without exposing traffic to the public internet.

> 📘 Reference Implementation
>
> For Terraform examples and code samples, refer to the <Anchor label="GKE Private Service Connect example" target="_blank" href="https://github.com/atlas/terraform-provider-atlas/tree/master/examples/gke/gke_private_service_connect">GKE Private Service Connect example</Anchor> in the Atlas Terraform provider repository.

## Prerequisites

* A GKE cluster in a private VPC
* GCP project ID for creating the Private Service Connect endpoint
* Appropriate GCP permissions to create VPC endpoints and configure DNS
* Choice of Atlas region: prod-master (US) or prod-eu (Europe). Contact Atlas Customer Success for India region support.

## Architecture

Private Service Connect uses a single endpoint in your GCP network to route all Atlas traffic privately. Unlike AWS PrivateLink, which requires multiple VPC endpoints, GCP Private Service Connect uses a single endpoint with wildcard DNS routing to serve all Atlas services.

<Image align="center" border={false} caption="Private Service Connect connection between customer GKE cluster and Atlas platform (diagram source: GCP documentation)" src="https://files.readme.io/c9951c85629f4c6922cb88bdd2d70156317922bfa585963eb26b120241074ca5-psc-endpoint-overview.svg" />

In this architecture diagram:

* **Consumer VPC**: Your GCP environment where your GKE cluster runs
* **Endpoint**: The Private Service Connect endpoint you create in your VPC
* **Clients**: Atlas components running in your GKE cluster
* **Service attachment**: Atlas's Service Attachment that your endpoint targets
* **Producer VPC**: Atlas's infrastructure hosting the platform services
* **Service**: Atlas platform APIs (see [communication requirements](communication-requirements.md))

The Private Service Connect connection enables your cluster to communicate with all Atlas services through a single endpoint using private IP addresses, without traffic traversing the public internet.

## Setup Process

### Step 1: Request access from Atlas

Contact Atlas Customer Success and provide:

* Your GCP project ID that will host the Private Service Connect endpoint
* Your preferred Atlas region (prod-master or prod-eu)

Atlas will allowlist your project and provide the Service Attachment ID needed for the next step.

### Step 2: Create a Private Service Connect endpoint

Create a Private Service Connect endpoint in your GCP project. The endpoint must be created in the same region as the Atlas Service Attachment.

**Region Requirements:**

| Atlas Region | GCP Region   | Hosts                       |
| -------------- | ------------ | --------------------------- |
| prod-master    | us-east4     | `*.psc.prod-master.atlas.example` |
| prod-eu        | europe-west3 | `*.psc.prod-eu.atlas.example`     |

<Tabs>
  <Tab title="GCP Console">
    1. Navigate to **Network Services > Private Service Connect**
    2. Click **Connect Endpoint**
    3. Configure the endpoint:
       * **Target**: Paste the Service Attachment ID provided by Atlas
       * **Endpoint name**: Choose a descriptive name (e.g., `atlas-psc-endpoint`)
       * **Region**: Select `us-east4` (for prod-master) or `europe-west3` (for prod-eu)
       * **Network**: Select your VPC
       * **Subnetwork**: Select an appropriate subnet
       * **(Optional)** Enable **Global access** if you need to access Atlas from multiple regions
    4. Click **Add Endpoint**
    5. Note the assigned IP address

    For detailed console instructions, see <Anchor label="Access published services through endpoints" target="_blank" href="https://cloud.google.com/vpc/docs/configure-private-service-connect-services#create-endpoint">Access published services through endpoints</Anchor> in the GCP documentation.
  </Tab>

  <Tab title="gcloud CLI">
    ```shell
    # Reserve an IP address for the endpoint
    gcloud compute addresses create atlas-psc-ip \
      --region=us-east4 \
      --subnet=your-subnet

    # Create the forwarding rule (PSC endpoint)
    gcloud compute forwarding-rules create atlas-psc-endpoint \
      --region=us-east4 \
      --network=your-vpc \
      --address=atlas-psc-ip \
      --target-service-attachment=SERVICE_ATTACHMENT_ID_FROM_ATLAS
    ```

    Replace `us-east4` with `europe-west3` if connecting to prod-eu.

    For gcloud command reference, see <Anchor label="gcloud compute forwarding-rules create" target="_blank" href="https://cloud.google.com/sdk/gcloud/reference/compute/forwarding-rules/create">gcloud compute forwarding-rules create</Anchor> in the GCP documentation.
  </Tab>

  <Tab title="Terraform">
    Refer to the [GKE Private Service Connect Terraform example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/gke/gke_private_service_connect) for complete infrastructure code that handles both PSC setup and cluster onboarding.
  </Tab>
</Tabs>

### Step 3: Configure DNS

Configure DNS to route Atlas hostnames through your Private Service Connect endpoint.

Create a wildcard DNS record that maps `*.psc.prod-master.atlas.example` (for prod-master) or `*.psc.prod-eu.atlas.example` (for prod-eu) to your PSC endpoint IP address.

#### Using Cloud DNS

1. Create a private DNS zone:

```shell
gcloud dns managed-zones create atlas-psc-zone \
  --description="Atlas PSC" \
  --dns-name=psc.prod-master.atlas.example \
  --networks=your-vpc \
  --visibility=private
```

2. Add a wildcard A record:

```bash
gcloud dns record-sets create "*.psc.prod-master.atlas.example." \
  --zone=atlas-psc-zone \
  --type=A \
  --ttl=300 \
  --rrdatas=PSC_ENDPOINT_IP
```

Replace `prod-master` with `prod-eu` if connecting to the EU region.

Ensure the DNS zone is accessible from the network where your GKE cluster runs.

For more information on Cloud DNS private zones, see <Anchor label="Create a private managed zone" target="_blank" href="https://cloud.google.com/dns/docs/zones#create-private-zone">Create a private managed zone</Anchor> in the GCP documentation.

### Step 4: Verify the connection

Test connectivity from within your Kubernetes cluster before proceeding with the full cluster onboarding.

Deploy a test pod:

```shell
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: atlas-psc-test
spec:
  containers:
  - name: netshoot
    image: nicolaka/netshoot:latest
    command: ["sleep", "infinity"]
    securityContext:
      capabilities:
        add:
        - NET_ADMIN
        - NET_RAW
EOF
```

Run verification commands:

```shell
# Set the hostname (use prod-eu for EU region)
ATLAS_HOST="api.psc.prod-master.atlas.example"

# Verify DNS resolution
# Expected: Your PSC endpoint IP address
kubectl exec atlas-psc-test -- dig +short $ATLAS_HOST

# Verify connectivity
# Expected: Connection succeeded
kubectl exec atlas-psc-test -- nc -zv $ATLAS_HOST 443

# Verify SSL certificate
# Expected: Certificate with CN=*.psc.atlas.example from Let's Encrypt
kubectl exec atlas-psc-test -- openssl s_client -showcerts \
  -servername $ATLAS_HOST \
  -connect $ATLAS_HOST:443 </dev/null
```

**Troubleshooting:**

* **DNS resolution fails**: Verify your DNS zone configuration and confirm it's associated with the correct VPC. <Anchor label="See Troubleshooting Cloud DNS" target="_blank" href="https://cloud.google.com/dns/docs/troubleshooting">See Troubleshooting Cloud DNS</Anchor> for additional guidance.
* **Connection fails**: Check firewall rules allow traffic to the PSC endpoint. See <Anchor label="VPC firewall rules overview" target="_blank" href="https://cloud.google.com/firewall/docs/firewalls">VPC firewall rules overview</Anchor> for configuring firewall rules.
* **SSL certificate fails**: Contact Atlas Customer Success, as this indicates an issue with the Atlas infrastructure

### Step 5: Connect your cluster

Once connectivity is verified, connect your cluster using one of these methods:

<Tabs>
  <Tab title="Terraform">
    Use the complete Terraform example that handles both PSC setup and cluster onboarding:

    [GKE Private Service Connect Terraform example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/gke/gke_private_service_connect)
  </Tab>

  <Tab title="Atlas Console">
    #### Using Console scripts (Phase 1)

    1. Generate the Phase 1 script from the Atlas Console
    2. Add the `gke.psc=true` query parameter to the API URL
    3. Run the modified script:

    ```curl
    curl -H "Authorization: Token YOUR_TOKEN" \
      "$ATLAS_API_URL/v1/agent.yaml?provider=gke&gke.psc=true" | kubectl apply -f -
    ```

    #### Using Console scripts (Phase 2)

    1. Generate the Phase 2 script from the Atlas Console
    2. Modify the environment variables at the beginning of the script:

    **For prod-master:**

    ```shell
    # Add this new variable
    ATLASWARE_ATLAS_API_URL=https://api.psc.prod-master.atlas.example

    # Modify these existing variables
    ATLAS_API_GRPC_URL=api-grpc.psc.prod-master.atlas.example:443
    ATLAS_GRPC_URL=grpc.psc.prod-master.atlas.example:443
    ATLAS_SENTINEL_GRPC_URL=sentinel.psc.prod-master.atlas.example:443
    ```

    **For prod-eu:**

    ```shell
    # Add this new variable
    ATLASWARE_ATLAS_API_URL=https://api.psc.prod-eu.atlas.example

    # Modify these existing variables
    ATLAS_API_GRPC_URL=api-grpc.psc.prod-eu.atlas.example:443
    ATLAS_GRPC_URL=grpc.psc.prod-eu.atlas.example:443
    ATLAS_SENTINEL_GRPC_URL=sentinel.psc.prod-eu.atlas.example:443
    ```

    3. Run the modified script
  </Tab>
</Tabs>

Your cluster should now be connected to Atlas through Private Service Connect.

# Additional Resources

* <Anchor label="GKE Private Service Connect Terraform example" target="_blank" href="https://github.com/atlas/terraform-provider-atlas/tree/master/examples/gke/gke_private_service_connect">GKE Private Service Connect Terraform example</Anchor>
* <Anchor label="GCP Private Service Connect documentation" target="_blank" href="https://cloud.google.com/vpc/docs/private-service-connect">GCP Private Service Connect documentation</Anchor>