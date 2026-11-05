---
updatedAt: 2026-10-14T08:53:00.000Z
---

# Set up Slack notifications

Learn how to connect your Slack workspace to Atlas and configure alerts to be delivered to your Slack channels.

## Overview

In this tutorial, you'll learn how to connect your Slack workspace to Atlas and configure alerts to be delivered to your Slack channels. This tutorial is intended for those who want to receive Atlas notifications directly in Slack without manually configuring webhooks.

By the end of this tutorial, you'll be able to:

* Connect a Slack workspace to your Atlas organization
* Create an alert that delivers notifications to the Slack channels you add
* Test the connection to verify alerts are working
* Manage connected workspaces and troubleshoot connection issues

## Background

Atlas's native Slack integration is the recommended method for receiving cluster notifications in Slack. Unlike the webhook-based approach, which requires you to manually create and manage Slack incoming webhooks, this integration uses OAuth to connect your Slack workspace directly to Atlas. This means you can set up notifications in a few clicks without any technical configuration in Slack.

Slack notifications include key information about cluster events, such as severity, a summary message, and a link to the Atlas console. Full notification details remain available in the console.

For details on how Atlas handles data collected through this integration, see the [Atlas Privacy Policy](https://atlas.example/privacy-policy/).

<Callout icon="⚠️" theme="warn">
  ### Important

  You must initiate the Slack connection from the Atlas console. The setup will not work if you attempt to start it from within Slack.
</Callout>

## Before you start

Before you begin, ensure you have:

* Access to the Atlas console with permissions to manage notifications (Organization Owner or Enterprise Owner role)
* A Slack workspace where you want to receive notifications
* Permission to install apps in your Slack workspace, or access to a Slack workspace administrator who can approve app installations

<Callout icon="⚠️" theme="warn">
  ### Important

  Each Atlas organization must connect to its own Slack workspace. If the same Slack workspace is already connected to a different Atlas organization, the connection attempt will fail during the **Connect** step. To connect this organization, use a different Slack workspace or disconnect the workspace from the other organization first.
</Callout>

## Create a Slack alert

Creating a Slack alert involves configuring the alert details and connecting your Slack workspace. If you haven't connected a workspace yet, you'll do so as part of this process.

1. In the Atlas console, click the **Notifications** icon in the top navigation, then click **View all**.

   ![](https://files.readme.io/96bc5859d5506970d5bfbcab19f8f6f7f7cab05ead5ec5495650e3e5a06203cb-image.png)

2. Click **Manage alerts**, then click **Create alert**.

   <Image src="https://files.readme.io/e3e1942a65d6b85581e8e35e2118cae9ca5bddb44c6cf124e35ace3224cc5120-2026-01-20_20-29-37.gif" align="center" border={true} />

3. Under **Delivery method**, verify that **Slack** is selected. Slack is the default delivery method.

4. Click **Connect** to link your Slack workspace.

   This opens a Slack authorization page in a new window.

5. On the Slack authorization page, select the workspace you want to connect from the **Workspace** dropdown.

   <Callout icon="📘" theme="info">
     ### Note

     The dropdown lists Slack workspaces you belong to. Each Atlas organization must use its own Slack workspace — connecting the same workspace to another Atlas organization will fail.
   </Callout>

   The page displays the permissions Atlas is requesting:

   * **Content and info about channels & conversations** — Allows Atlas to see which channels are available for posting notifications
   * **Content and info about your workspace** — Allows Atlas to identify the workspace connection
   * **Perform actions in channels & conversations** — Allows Atlas to post notification messages to channels you select

6. Click **Allow** to grant permissions and complete the connection.

   ![](https://files.readme.io/8a16bb5174265dc03f1e45f57de8c46d129618824f2cd77911f4914d48f43d22-image.png)

   You're redirected back to the Atlas console. What happens next depends on your Slack workspace settings:

   * **If the app is automatically approved**: The workspace name appears in the Workspace section, and you can proceed to add channels.

   * **If admin approval is required**: The Workspace section shows a **Pending** status with the message "Your Slack workspace admin must approve this connection. The request expires in 7 days."

   <Callout icon="📘" theme="info">
     ### Note

     If you see the Pending status, contact your Slack workspace administrator to approve the Atlas app. Click **Retry** to check if the approval has been granted.
   </Callout>

7. Once the workspace is connected, click the **Channels** field and paste the ID of a Slack channel, then press **Enter**. Repeat for each channel where you want notifications delivered.

   You can add up to 5 channels per alert. Added channels appear as tags below the field. To remove one, click the tag.

   To find a channel's ID, open the channel in Slack, click the channel name, and copy the **Channel ID** from the **About** tab.

   <Callout icon="📘" theme="info">
     ### Note

     Atlas does not list your Slack channels, so you add each channel by its ID. The **Add** action verifies only that the ID is valid. It does not confirm that the channel exists or that the Atlas app can post to it. Use the [Test button](#test-the-connection) to verify delivery.

     For a private channel, you must also invite the Atlas app so it can post. Open the private channel in Slack and run `/invite @Atlas`, or go to the channel's settings and add Atlas from the **Integrations** tab. Private channels are never listed by name; the alert shows them as **Private channel** with the channel ID.
   </Callout>

8. Under **Details**, enter a **Name** for your alert.

9. Under **Category**, specify which type of notifications trigger this alert. Select **All** to capture every notification category with a single selection, or select individual categories to narrow the alert's scope. Categories include:

   * All
   * Workload Autoscaler
   * Node Autoscaler
   * Reporting anomalies
   * Inventory
   * Security
   * Other

   For descriptions of each category, see [Notifications](notifications.md#notification-categories).

10. Under **Clusters**, select which clusters trigger this alert. Leave **All clusters** selected to trigger on every cluster in your organization, or select specific clusters to limit the alert's scope.

    The **Include non-cluster alerts** toggle is on by default and includes alerts that are not associated with any specific cluster, such as organization-level notifications. Turn it off to receive only alerts tied to the selected clusters.

11. Under **Notification severity**, select which severity levels should trigger notifications. Options include:
    * Critical
    * Error
    * Warning
    * Success
    * Info

12. Click **Save** to create the alert.

    ![](https://files.readme.io/7118ecb4cad10aa83de9125470b2f2498763f652eae6473fc98a97a7249b6631-image.png)

## Test the connection

After saving an alert, you can send a test notification to verify that messages are delivered to your configured Slack channels.

1. In the **Alerts** list, click on the alert you want to test.

2. Click the **Test** button in the bottom-left corner of the alert details panel.

   ![](https://files.readme.io/5b59f79896d404efaaeaec91fbc826a939f466ad554c0525a0835e4fa1b8eacc-image.png)

   Atlas sends a test notification to all configured Slack channels for this alert.

3. Check your Slack channels to confirm the test message arrived.

If the test message doesn't appear, see [Troubleshooting](#troubleshooting) for common issues and solutions.

## Manage connected workspaces

You can disconnect a Slack workspace from Atlas at any time. Disconnecting a workspace removes the connection but doesn't uninstall the Atlas app from Slack.

To disconnect a workspace:

1. Open an existing alert that uses the workspace, or create a new alert.

2. Under **Delivery method**, select **Slack**.

3. In the **Workspace** section, click **Disconnect** next to the workspace name.

The workspace is removed from your Atlas organization. Any alerts configured to use this workspace will no longer deliver notifications to Slack until you reconnect a workspace and update the alert configuration.

## Next steps

Consider exploring these related topics:

* [Notifications](notifications.md) — Learn about all notification types and severity levels available in Atlas
* [How to set up notification webhooks](setup-notification-webhook.md) — Configure webhooks for external systems like PagerDuty or OpsGenie
* [Webhook integration examples](examples.md) — See example configurations for various notification destinations

## Troubleshooting

### Connection shows Pending status

If your connection shows "Pending" after authorizing the Atlas app, your Slack workspace requires administrator approval for new app installations. Contact your Slack workspace administrator and ask them to approve the Atlas app. The approval request expires after 7 days—if it expires, you'll need to initiate the connection process again.

### Test notification not received

If the test notification doesn't reach a channel:

* Verify the channel ID is correct. Open the channel in Slack, click the channel name, and compare the **Channel ID** in the **About** tab with the ID on the alert.
* For a private channel, invite the Atlas app so it can post. Open the channel in Slack and run `/invite @Atlas`, or add Atlas from **Settings** > **Integrations** > **Add apps**.
* Verify the channel still exists and hasn't been archived.
* Verify the Atlas app is still installed in your Slack workspace.
* Try disconnecting and reconnecting the workspace.

### Connection fails when initiated from Slack

The Slack integration must be set up from within the Atlas console. If you attempt to add Atlas from Slack's app directory or workspace settings, the connection will not link properly to your Atlas organization. Always start the connection process by clicking **Connect** in the Atlas console.

### Slack workspace already linked to another organization

A failure during the **Connect** step, after you select a Slack workspace and click **Allow**, usually means the Slack workspace you're trying to connect is already linked to a different Atlas organization. Atlas does not support sharing one Slack workspace across multiple organizations.

To resolve this:

* Connect a different Slack workspace to this Atlas organization, or
* Have an admin of the other Atlas organization disconnect the workspace first, then retry the connection from this organization.