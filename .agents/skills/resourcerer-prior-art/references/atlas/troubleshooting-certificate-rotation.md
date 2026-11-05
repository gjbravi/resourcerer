---
updatedAt: 2026-07-20T07:37:57.000Z
---

# Cluster certificate rotation

Learn how to handle Kubernetes certificate rotation in clusters managed by Atlas.

Certificate rotation is a regular maintenance task that helps maintain cluster security. This guide explains how to handle the process in clusters managed by Atlas.

The process involves:

1. Rotating certificates at your cloud provider level
2. Ensuring Atlas components properly recognize and use the new certificates

# Before you begin

* Ensure you have access to perform node rebalancing in the Atlas console
* Review your cloud provider's certificate rotation documentation:
  * [GCP - Rotating Certificates](https://cloud.google.com/kubernetes-engine/docs/how-to/credential-rotation)

> 🚧 Important
>
> Incorrect certificate rotation steps may cause Atlas nodes to lose connection with the control plane. Follow the steps in exact order to prevent connectivity issues.
>
> Additionally, rotating Kubernetes certificates may result in your cluster being temporarily unavailable as components are restarted. For production environments, perform this action during a maintenance window.

# Rotate certificates in GKE

1. **Start certificate rotation**
   * Access your GKE console or use the `gcloud` CLI to initiate the rotation process
   * **Do not complete the rotation yet** - this happens in step 5

2. **Delete Atlas node pools**
   * Delete all GKE node pools used by Atlas. Including `atlas-pool`:\
     ![](https://files.readme.io/95a8fb4ee6eb9429a08961eff15f04a058db1d52c00b9143bb4a98a35160ad3d-image.png)

* If you have the `kube-gke-pools-per-node-configuration-enabled` feature flag enabled for your organization, delete all Atlas-created node pools.

> 📘 Note
>
> Review your CSP's documentation for instructions on deleting node pools:
>
> * [GCP - Rotating Certificates](https://cloud.google.com/kubernetes-engine/multi-cloud/docs/aws/how-to/delete-node-pool)

3. **Trigger cluster reconciliation**
   * In the Atlas console, navigate to the cluster list
   * In the row of your cluster, click "Actions" ("**⋮**") and select "Trigger reconcile":\
     ![](https://files.readme.io/2d40f4f3b1fb36f68bc87dfb004fadf078ac68d8266889cf42934bde17a33b59-image.png)\
     This step ensures Atlas GKE Node Pools use the updated certificates.
4. **Replace existing nodes**\
   Choose one of these methods to replace nodes with the new certificates:
   * Use [Scheduled rebalancing](scheduled-rebalancing.md) to gradually replace nodes. Set "Minimum Node Age" to zero to ensure all nodes get rotated.
   * Manually [rebalance](rebalancing.md) specific node groups to control the process.

> 📘 Note
>
> For large clusters, consider rebalancing in smaller batches to avoid hitting cloud provider quotas.

5. **Complete certificate rotation**\
   Return to GKE to complete the rotation process. Follow the [GCP documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/credential-rotation#completing_the_rotation) for detailed steps.

If you need assistance during the certificate rotation process, contact Atlas support via our [community Slack channel](https://join.slack.com/t/atlas-community/shared_invite/zt-nwtbsbqq-dWOBiUgLtIGMsUSup7Xksw).