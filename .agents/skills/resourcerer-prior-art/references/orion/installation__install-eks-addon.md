# Install Orion as an EKS add-on

> Installation through the Orion EKS add-on in AWS Marketplace


If you're running an Amazon EKS cluster, you can install Orion as an EKS add-on in one of two ways:
- [Using the EKS UI and Orion UI](#eks-ui)
- [Using the EKS CLI and Orion UI](#eks-cli)

## Prerequisites

- An Amazon EKS cluster
- A Orion account — [sign up for a free trial](https://orion.example/signup)

## Install using the EKS UI and Orion UI {#eks-ui}

### Install the add-on

1. Open two browser windows side by side.
1. In one window, navigate to and log in to your Amazon EKS cluster.
1. On the cluster page, add the **Add-ons** tab, and then click **Get more add-ons**.
1. On the **Select add-ons** page, scroll to the **AWS Marketplace add-ons** section.
1. From the **Any category** menu, select **cost-management**.
1. On the Orion add-on card, select the check box in the upper right (see below); then, click **Next**.

    [![EKS cost management card](/img/docs/ol-eks-install-costmg-card.png)](/img/docs/ol-eks-install-costmg-card.png)

1. On the **Configure selected add-ons** settings page, expand the **Optional configuration settings** section.
1. In the **Add-on configuration schema** section, you can customize your installation the same way you can customize a Helm install.
1. Copy the following placeholder text into the **Configuration values** text area. In a later step, you'll replace this text with the values generated on the Orion Access Credentials page.
    ```yaml
    authorization:
      issuer:
      clientID:
      clientSecret:
    standbyMode: false
    ```

### Get the cluster credentials and copy them to EKS configuration values

1. In the second browser window, go to [app.orion.example](https://app.orion.example/).
1. In the left navigation, click **Settings** > **Access Credentials**, and press **Add Credential**.
1. In the **Create Credential** window, provide an identifier for the credential and click **Create**.
1. Copy the **issuer**, **clientID**, and **clientSecret** from the new credential.

    [![Image showing Orion credential](/img/docs/ol-sf-cred.png)](/img/docs/ol-sf-cred.png)

1. In your EKS browser window, copy the new credential into the **Configuration values** text area, replacing the empty placeholder. The **Advanced configuration** panel will look like this:

    [![EKS Configuration panel](/img/docs/ol-eks-config-vals-creds.png)](/img/docs/ol-eks-config-vals-creds.png)
1. In the **Conflict resolution** method, select **Override**.
1. Click **Next**, and then click **Next** again.

### Validate the installation

In the Orion left navigation, click **Settings** > **Cluster Management** and confirm the new cluster was added.

## Install using the EKS CLI and Orion UI {#eks-cli}

This method installs Orion in standby mode.

### Install the add-on

Run the following command, replacing `CLUSTER_NAME` with the name of your cluster and `REGION` with your region.

```sh
eksctl create addon --cluster CLUSTER_NAME --name orion_optimize-live --version "v2.8.0-eksbuild.1" --region REGION --force
```

This installs the add-on with the default values provided by Orion. The Agent pod is in standby mode, and the Prometheus agent is not started yet.

### Validate the installation

To validate the installation, run:

```sh
aws eks describe-addon --cluster-name CLUSTER_NAME --region REGION --addon-name orion_optimize-live
```

The output resembles the following:

[![EKS add-on health check](/img/docs/ol-eks-cli-installcomplete.png)](/img/docs/ol-eks-cli-validate-health.png)

In the output, look for `"status": "ACTIVE"` and confirm that there are no issues in the `"health"` section:

[![EKS add-on health check](/img/docs/ol-eks-cli-validate-health.png)](/img/docs/ol-eks-cli-installcomplete.png)

### Manage your Orion subscription

In Amazon EKS, go to the Add-ons page. To get the **orion_optimize-live** add-on, click the **Manage subscription link**.

### Create a `config.yaml` file

This `config.yaml` file updates an existing cluster with the EKS add-ons. Update the placeholder values and apply:

  ```yaml
  apiVersion: eksctl.io/v1alpha5
  kind: ClusterConfig
  metadata:
    name: "NAME"
    region: "REGION"
  addons:
  - name: orion_optimize-live
    version: v2.8.0-eksbuild.1
    configurationValues: |-
      authorization:
        issuer: https://api.orion.example/
        clientID: "CLIENT_ID"
        clientSecret: "CLIENT_SECRET"
      standbyMode: false
    resolveConflicts: overwrite
  ```

### Get the cluster credentials

1. In a browser window, go to [app.orion.example](https://app.orion.example/).
1. In the left navigation, click **Settings** > **Access Credentials**, and press **Add Credential**.
1. In the **Create Credential** window, provide an identifier for the credential and click **Create**.
1. Copy the **issuer**, **clientID**, and **clientSecret** lines from the new credential.

    [![Image showing Orion credential](/img/docs/ol-sf-cred.png)](/img/docs/ol-sf-cred.png)

1. Replace the **issuer**, **clientID**, and **clientSecret** in the EKS `config.yaml` file with the new credential information you just copied. Be sure to copy it to the `addons.configurationValues.authorization` section.

### Apply the YAML config and activate the add-on

Run the following command to apply the config file and remove it from standby mode:

```sh
eksctl update addon --config-file config.yaml --force
```

After a minute or two, you should see output similar to the following:

[![EKS eksctl update output](/img/docs/ol-eks-update-output.png)](/img/docs/ol-eks-update-output.png)

### Subscribe

You must subscribe to use this add-on.

1. Sign in to [aws.amazon.com](https://aws.amazon.com), and in the AWS Marketplace Buyer website, search for **Orion**.
1. On the Orion product page, click **Continue to Subscribe** in the upper right.
1. Review the agreement and then click **Create Contract**.
