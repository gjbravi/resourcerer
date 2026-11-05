---
updatedAt: 2026-05-12T13:39:45.000Z
---

# Updates and images

### What is the update frequency for your AKS images?

Our AKS images are updated every 30 days, or whenever we detect that the AKS control plane has been upgraded.

<br />

***

### Why does Atlas use custom images for AKS?

Microsoft suddenly stopped providing new machine images for third parties in 2022, so we developed our image flow for AKS clusters. We start with the same Ubuntu base image Microsoft used internally, update the image and its packages, install all required components to it, and run it on all Atlas-managed AKS nodes.

<br />

***

### Are some images on Docker and some of them are on the atlas-hub? Do we need to allow both? What's the difference?

All of our core pieces are on the `atlas-hub`, whereas add-on pieces are on Docker (for instance, our open-source `rebalancer/hibernate`). `atlas/pod-node-lifecycle is` the mutating webhook piece which is hosted on Docker rather than the `atlas-hub`.

<br />

***

### Is there a document somewhere that lists the correct procedure regarding the steps for getting the node pools upgraded after we update the control plane?

Atlas should update the version in the node pool within 10-15 minutes. You can also perform a "Trigger Reconcile" in the Atlas console UI. Once the node pool is updated you can rebalance to get all new nodes.

<br />

***

### What’s the best practice to update our EKS version with minimal downtime while we are using it in Atlas?

Once you upgrade the control plane version, Atlas will synchronize with it within 10-15 minutes. If you prefer, you can also manually reconcile it through the UI. After that, you can perform a rebalance.

Atlas will always create nodes with the latest available Kubernetes version. Additionally, when replacing, it first creates a new node and then drains the old one, ensuring that multi-replica applications experience minimal or no downtime.

You can also set a Node-TTL in a scheduled rebalance job that will automatically rotate out old nodes to pick up the upgrade without the need for a full rebalance.

This shouldn't be a problem unless you're upgrading to more than two versions where the APIs have changed significantly.

<br />

***

### To roll out Atlas's automated management on our AKS clusters, do we need to upgrade them? One of our clusters is in version 1.19 and the other in 1.22.

When creating the required node pools, the versions must align with what Azure currently supports. To clarify this requirement, please refer to the updated documentation provided here: [Troubleshooting](cluster-onboarding-troubleshooting.md#aks-versions-not-supported).

<br />

***

### Can we upgrade in a fixed window for maintenance with a fixed disruption ratio as we do for node groups?

You can set the execution time and minimum node age for the scheduled rebalance to run and set a maximum number of nodes. For example, you can set the following: "I want 3 nodes to rebalance if they are older than 7 days every hour between 10 pm and 2 am EST on Saturday and Sunday."

This would mean every hour from 10 pm to 2 am, the rebalancer would check for 3 nodes that are over 7 days old and swap them with new nodes. If all the nodes got swapped and no nodes were over 7 days old, it would do nothing.

<Image align="center" src="https://files.readme.io/b7c7fbe-Screenshot_2023-12-21_at_5.03.10_PM.png" />

<Image align="center" src="https://files.readme.io/18bbaaa-Screenshot_2023-12-21_at_5.02.24_PM.png" />

<br />

***

### How does Atlas deal with Azure auto-update of node OS image support?

Currently, AKS images used by Atlas are upgraded after every AKS control plane upgrade or every 30 days. We don't have a way of updating the OS image without a K8s upgrade at the moment.

<br />

***

### For a custom AKS image, can you provide any information on security updates, patch schedules, etc.?

The patch schedule is as follows: Atlas AKS images are re-created after every AKS control plane upgrade or every 30 days.

When a cluster is connected, we take a snapshot of the boot disk from an existing AKS-managed node that is running the official AKS image and use that to create a custom image. This ensures the custom image we build is up to date with the latest patches. This was implemented due to Microsoft stopping to share their machine images publicly.

<br />

***

### Does Atlas open IP forward by default?

Atlas uses the bootstrap script from EKS worker nodes and by default, it adds IP forwarding ([learn more here](https://github.com/awslabs/amazon-eks-ami/blob/569c6b146cbf28af71485a9f30171241643174c3/scripts/install-worker.sh#L213)). You can also do that via the nodeConfig via the init Script section if needed.