---
updatedAt: 2026-05-12T14:12:37.000Z
---

# Migration from Karpenter

Atlas automatically detects the presence of Karpenter autoscaler in your cluster and lets you migrate from Karpenter to Atlas autoscaling with one click. You can easily migrate your Karpenter setup by importing Node Pools (or Provisioners) and EC2 Node classes (or AWS Node templates) as Atlas's equivalent configurations: Node Templates and Node Configurations.

# How to migrate from Karpenter to Atlas

Select your cluster and go to the Autoscaling settings page. Atlas will automatically detect the presence of Karpenter in your cluster and display the import option on the UI.

![](https://files.readme.io/fdd18b3-image.png)

<br />

After clicking *Begin import* button, a drawer on the right side will appear. You will see a list of Karpenter objects to be imported.

![](https://files.readme.io/ed95956-image.png)

<br />

Objects that can't be imported will be skipped.

![](https://files.readme.io/958aecc-image.png)

<br />

Click the *Import* button and the migration will start. Once the import has been completed, you will see the following status:

![](https://files.readme.io/9d2a8d0-image.png)

Node templates and node configurations that were created from Karpenter objects are now turned on and ready for you to use.