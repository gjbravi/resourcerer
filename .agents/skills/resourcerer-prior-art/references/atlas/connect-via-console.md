---
updatedAt: 2026-10-27T10:52:49.000Z
---

# Connect using the Atlas console

Connect your Kubernetes cluster to Atlas from the web console, or discover all clusters across your cloud account with Cloud Connect.

The Atlas console guides you through connecting a cluster. When you start the connection flow in the console, you can choose between two options: **atlasctl commands** (recommended) or the **legacy onboarding script**.

The console also offers **Cloud Connect** — an agentless way to discover all your clusters across a cloud account before deciding which ones to connect.

For a one-time, read-only collection of reliability data from your AWS account, see [Take a snapshot of your AWS account](cloud-snapshot.md).

## Connect with atlasctl

The console shows the `atlasctl` commands to run in your terminal. This is the fastest path and the same flow described in detail on the [Connect using the atlasctl CLI](connect-with-atlasctl.md) page.

In summary:

1. Log in to the [Atlas console](https://console.atlas.example/) and click **Connect cluster**.

2. Select your operating system.

3. Copy and run the commands shown in the console:

   ```shell
   atlasctl auth login
   atlasctl cluster connect
   ```

4. Follow the interactive prompts in your terminal.

5. Once the connection completes, your cluster appears in the console.

For the full walkthrough — including installation, non-interactive mode, and all available options — see [Connect using the atlasctl CLI](connect-with-atlasctl.md).

## Connect with the onboarding script

> ⚠️ Deprecated
>
> The script-based connection flow is deprecated. Use [atlasctl](connect-with-atlasctl.md) for new cluster connections. The script remains available for environments where atlasctl is not yet supported.

### Prerequisites

Before connecting your cluster with the script, ensure your environment meets the following requirements:

#### General requirements

* `curl` command line utility installed
* `kubectl` command line utility installed
* `jq` command line utility installed
* `yq` command line utility installed
* `helm` command line utility installed
* Appropriate permissions to create resources in your cluster

#### EKS-specific requirements

The atlas-agent receives cluster metadata (`EKS_ACCOUNT_ID`, `EKS_REGION`, `EKS_CLUSTER_NAME`) as environment variables via the Helm chart's `additionalEnv` values. atlasctl and the Terraform module set these automatically.

> 📘 Manual Helm installation
>
> If you deploy the atlas-agent directly via Helm (outside atlasctl or Terraform), you must set these environment variables yourself.
>
> ```shell
> helm upgrade --install atlas-agent atlas-helm/atlas-agent \
>   --namespace atlas-agent --create-namespace \
>   --set provider=eks \
>   --set additionalEnv.EKS_ACCOUNT_ID="<your-aws-account-id>" \
>   --set additionalEnv.EKS_REGION="<your-eks-region>" \
>   --set additionalEnv.EKS_CLUSTER_NAME="<your-cluster-name>" \
>   --set apiKey="<your-atlas-api-key>"
> ```

### Step 1: Log in to the Atlas console

Once logged in, you can see a button to connect a new Kubernetes cluster in the upper-right corner. Click on this button.

On the bottom, click **Switch to legacy script**.

### Step 2: Select your provider

Pick your managed Kubernetes service provider. You can choose from the following options:

* Amazon Elastic Kubernetes Service (EKS)
* Azure Kubernetes Service (AKS)
* Google Kubernetes Engine (GKE)
* Red Hat OpenShift service on AWS
* All Alternative Clouds (Atlas Anywhere)

Pick your provider and click **Next**.

### Step 3: Run the script

Select the **Script** tab, then open your cloud shell or terminal and run the provided script. Make sure that `kubectl` is installed and can access your cluster.

Copy the script and run it in your cloud shell or terminal.

> 📘 Note
>
> Atlas's read-only agent has no access to your sensitive data and doesn't change your cluster configuration. Read more about the [security of our agent](about-the-read-only-agent.md).

Once you're done, click **I ran the script**.

### Step 4: Automated cluster connection

Atlas will automatically connect your cluster to the platform.

Once that happens, you're ready to explore the platform and learn more about your cluster's performance and cost.

## Discover clusters with Cloud Connect

If you want to see all of your clusters and their optimization potential before connecting individual ones, use **Cloud Connect**. Rather than connecting a specific cluster, Cloud Connect performs agentless, read-only discovery across your entire cloud account.

This is useful when you manage many clusters and want to identify which ones would benefit most from Atlas before installing any components.

### How to discover clusters

1. Log in to the [Atlas console](https://console.atlas.example/) and click **Connect cloud** in the cluster list.

2. Select your cloud provider (AWS, Azure, or GCP).

3. Choose your **permission scope** based on your security requirements:

   * **Default** — Read-only access using AWS-managed ReadOnlyAccess policy, Azure Reader role, or GCP Reader role.
   * **Minimal** — Targeted permissions for core services (EC2, EKS, RDS, SageMaker, commitments for AWS; VMs, AKS for Azure; Compute, GKE for GCP).
   * **AWS AI services** — Additional List\*, Describe\*, and Get\* permissions for AWS AI/ML services.

   For details on what each scope includes, see [Cloud Connect permissions](cloud-permissions.md#permissions-for-cloud-connect).

4. Copy and run the generated script in your cloud shell or terminal. This gives Atlas read-only access to discover your existing clusters.

5. All your clusters appear as **Discovered** in the console. You can view the optimization potential for each cluster without installing any components.

> 📘 Note
>
> Discovered resources are synchronized every hour for all supported providers. For more details on what each provider syncs, see [Cloud Connect](cloud-connect.md#supported-providers).

### Connect a discovered cluster

Select any discovered cluster in the console and click **Connect** to start the connection flow. From there, the process follows the same steps as [connecting with atlasctl](connect-with-atlasctl.md) or the onboarding script described above.

## See also

* [Connect using the atlasctl CLI](connect-with-atlasctl.md) — the recommended connection method with the full walkthrough.
* [Legacy onboarding script reference](legacy-onboarding-script.md) — the permissions and components the legacy script creates to manage nodes.
* [Cloud permissions](cloud-permissions.md) — the IAM resources created during the connection.

<br />