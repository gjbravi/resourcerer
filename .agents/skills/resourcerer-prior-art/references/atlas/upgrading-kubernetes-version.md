---
updatedAt: 2026-05-12T14:12:37.000Z
---

# Upgrading Kubernetes version

# Upgrading Kubernetes version

This guide outlines the recommended steps for upgrading your Kubernetes cluster managed by Atlas. Following these steps ensures a smooth upgrade process and minimizes potential disruptions to your workloads.

## Before you begin

Before deciding to upgrade the Kubernetes version in your clusters, review the Kubernetes release notes for any new features or deprecations that may affect your workloads.

> 🚧 Important: Amazon Linux 2 (AL2) deprecation in EKS
>
> AWS is ending support for Amazon Linux 2 (AL2) with Kubernetes v1.32. Starting with v1.33, only AL2023 and Bottlerocket AMIs will be supported. When planning your Kubernetes upgrade:
>
> * If upgrading to v1.32: Consider using this opportunity to migrate from AL2 to AL2023 or Bottlerocket
> * If upgrading to v1.33 or later: You **must** use AL2023 or Bottlerocket for node images
>
> For node configuration updates related to this change, refer to our [Node Configuration documentation](node-configuration.md#eks-specific-configuration-options).

Before starting the upgrade process, ensure that:

* You have access to your cloud provider's console (AWS, Azure, or GCP).
* You have the necessary permissions to modify your Kubernetes cluster.
* You have access to the Atlas console.
* For EKS clusters on v1.32 or planning to upgrade to v1.33+, you've reviewed your node image requirements.

## Upgrade process

### Step 1: Upgrade the control plane

1. Log in to your cloud provider's console.
2. Navigate to your Kubernetes cluster management section.
3. Initiate the control plane upgrade process.

> 📘 Note
>
> Make sure to upgrade the control plane separately from the node pools. Atlas will automatically handle node pool upgrades after the control plane upgrade. This approach helps avoid scaling delays and ensures seamless updates.

4. Wait for the control plane upgrade to complete successfully.

> 📘 Note
>
> For exact steps on how to perform a control plane upgrade using your cloud provider's console, refer to your cloud provider's documentation:
>
> * [Azure Kubernetes Service (AKS)](https://learn.microsoft.com/en-us/azure/aks/upgrade-aks-cluster?tabs=azure-cli)
> * [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-upgrades)
> * [Amazon Elastic Kubernetes Service (EKS)](https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html)

<Image alt="Example of upgrading the Kubernetes version for AKS" align="center" src="https://files.readme.io/1ea95e6c6583d3734bb28493d5ad91f85f4668efe3ad6ed0270ef811beb0132c-image.png">
  Example of upgrading the Kubernetes version for AKS
</Image>

### Step 2: Reconcile in Atlas

After the control plane upgrade is finished:

1. Navigate to the [Atlas console](https://console.atlas.example).
2. Locate your cluster and click the "Trigger reconcile" button.

<Image alt="Triggering reconciliation from the cluster dashboard" align="center" src="https://files.readme.io/b8f76c826f485481d2b839e29db380c092897fe7281de3471cf40b149fa3a05d-image.png">
  Triggering reconciliation from the cluster dashboard
</Image>

> 📘 Note
>
> Alternatively, you can wait for the auto-reconcile, which occurs every 30 minutes.

The reconciliation process will:

* Initiate the creation of new images for node pools (this can take 20-30 minutes).
* Ensure that all new nodes added by Atlas use the upgraded control plane version.

<Callout icon="💡" theme="default">
  ### Tip

  For AKS clusters, you can confirm the image update by checking the Azure Compute Gallery for new images with the current date.

  For EKS clusters, if you're upgrading to v1.32 or later, this is a good time to update your node configuration to use AL2023 or Bottlerocket image families, as AL2 will be deprecated. See [Node Configuration](node-configuration.md#eks-specific-configuration-options) for details.
</Callout>

### Step 3: (Optional) Full node replacement

To replace all existing nodes with new ones using the upgraded version:

1. Set up a scheduled rebalance with the following parameters:

   * **Minimum node age**: 0

   <Image align="left" width="75% " src="https://files.readme.io/7062106803f937c935fa992aefb5883788a28be2a4473568ebf8870bd68fe04e-image.png" />

   * **Savings threshold**: Disabled

   <Image align="left" width="75% " src="https://files.readme.io/38ae23c5b2a5a65342d4a1cb9822dd765df84b348637f45f82bbdaaca40d5b64-image.png" />

This ensures all old nodes are replaced regardless of any potential cost changes.

2. Generate and run a rebalancing plan.

For detailed instructions on scheduled rebalancing, see our [Scheduled rebalancing guide](scheduled-rebalancing.md).

# Troubleshooting

If you encounter issues during the upgrade process, follow these steps:

## Nodes failing to be added after rebalance

If you initiated a rebalance without reconciling first, and nodes are timing out or failing to be added:

1. For AKS clusters:

   * Delete the `atlaspool` from the Azure cloud console.

   ![Delete atlaspool](https://files.readme.io/6802b10f817b80b7b98f76ebaf3d70f738e79398b198ca6bf9c3cb624496cca9-image.png)

2. Trigger a reconciliation from the Atlas console.

3. After the reconciliation is complete, re-run the rebalance to add nodes successfully.

## AL2 deprecation issues (EKS)

If upgrading to v1.32 or v1.33 and experiencing issues with AL2 deprecation:

1. Check that your node configuration has been updated to use either `FAMILY_AL2023` or `FAMILY_BOTTLEROCKET`
2. Ensure a full reconciliation has been completed after updating the node configuration
3. If nodes still fail to provision with the new OS family, contact Atlas support for assistance

# Next steps

After successfully upgrading your Kubernetes cluster:

* Verify that all your workloads are running correctly.
* Update any client applications or tools that interact with your cluster to ensure compatibility with the new Kubernetes version.