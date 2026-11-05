---
updatedAt: 2026-09-25T08:23:12.000Z
---

# Webhook integration examples

# Example of OpsGenie Notification webhook setup

This configuration enables the integration of OpsGenie with the Atlas console, ensuring that alerts are received in OpsGenie.

## OpsGenie pre-requisite

### Add an API integration

1. Go to "Settings" → "Integrations".
2. Run a search and select “API”.
3. On the next screen, enter a name for the integration.
4. Optional: Select a team in the "Assignee team" field if you want a specific team to receive alerts from the integration. [Learn about the alert notifications flow in Opsgenie.](https://support.atlassian.com/opsgenie/docs/alert-notifications-flow/)
5. Select "Continue". The integration is saved at this point.
6. Expand the "Steps to configure the integration" section and copy the API key for use in your API requests.
7. Select "Turn on integration." The rules you create for the integration will only work if you turn it on.

<Image src="https://files.readme.io/9cf85bf-image.png" alt="API key integration example" align="center" caption="API key integration example" border={true} />

### Define rules for creating and processing alerts

The configuration of an API integration involves defining many different alert scenarios. These scenarios specify how and when alerts can be created, closed, acknowledged, etc. The integration comes with a default set of alert creation and processing rules; you can customize them and add as many rules of your own as you like. For example, you can add three rules to create the alert, which means the incoming data is evaluated against these three scenarios in order, and if one of them matches, an alert is created.

<Image src="https://files.readme.io/270eb7c-image.png" alt="Alert rules example" align="center" caption="Alert rules example" border={true} />

## OpsGenie Integration

### Add a Webhook integration

1. Go to "Settings" > "Integrations".
2. Select "Add integration".
3. Run a search and select “Webhook”.
4. On the next screen, enter a name for the integration.
5. Optional: Select a team in the "Assignee team" if you want a specific team to receive alerts from the integration.
6. Select "Continue". The integration is saved at this point.
7. Select "Turn on integration." The rules you create for the integration will only work if you turn it on.

Sending alert descriptions and details is only available for "Create" and "Custom" actions.

<Image src="https://files.readme.io/9f9d4dd-image.png" alt="Webhook integration example" align="center" caption="Webhook integration example" border={true} />

### Webhook configuration

Utilize the provided API integration key to configure your webhook integration, ensuring it is used for authentication.

* Webhook URL should be `https://api.opsgenie.com/v2/alerts`
* Customer Header should include:
  * \_Authorization: GenieKey `<API Key>`
  * \_Content-Type: `application/json`

### Webhook integration validation

Before entering the essential details into the Atlas webhook configuration, we recommend executing the following `curl` command. This action will verify whether the configured integration receives alerts and notifications as anticipated:

```shell
curl -X POST https://api.opsgenie.com/v2/alerts \
    -H "Content-Type: application/json" \
    -H "Authorization: GenieKey <API Key>" \  
    -d '{"message": "An example alert message"}'
```

<Image src="https://files.readme.io/2fd51c1-image.png" alt="OpsGenie Alert" align="center" caption="OpsGenie Alert" border={true} />

## Atlas Integration

### Add webhook via Atlas console

* Navigate to **Notifications** > **Manage alerts** > **Create alert** and select **Webhook** as the delivery method.

<Image src="https://files.readme.io/c12a7a6-image.png" align="center" border={true} />

* Set up the fields below using the information collected from the Webhook and API integration

  * *Name of Webhook*
  * *Callback URL*
  * *Category* (select **All** to capture every notification category with a single selection)
  * *Severity*
  * *Request template*
  * *Authorization keys (Headers)*

  <Image src="https://files.readme.io/5c1b9c1-Screenshot_2024-02-26_at_12.10.19_PM.png" align="center" border={true} />

<Image src="https://files.readme.io/5d24479-image.png" alt="Webhook “Can connect” example" align="center" caption="Webhook “Can connect” example" border={true} />

# Example of MS Teams Notification webhook setup

This configuration enables the integration of MS Teams with the Atlas console, ensuring that alerts are received in MS Teams.

## Method 1: MS Teams Incoming Webhook Integration

### Create an Incoming Webhook

1. In the New Teams client, select Teams and navigate to the channel where you want to add an Incoming Webhook.

2. Select "More options" (•••) on the right side of the channel name.

3. Select "Manage channel".

   <Image src="https://files.readme.io/18b7b23-50cf2fcc-8a26-4fcc-8b79-ced7f4e83db7.png" align="center" width="35%" />

4. Select "Edit".

   <Image src="https://files.readme.io/7606dbc-584f064a-050c-40be-b6d5-58fb5909220d.png" align="center" width="70%" border={true} />

5. Search for "Incoming Webhook" and select "Add."

   <Image src="https://files.readme.io/f8ebd13-6085b7ee-a574-43e5-9655-8c5b7c60d6ac.png" align="center" width="50%" border={true} />

6. Select "Add".

   <Image src="https://files.readme.io/84c0645-e0a6609d-9646-4a68-acad-d89d4bd25c93.png" align="center" width="50%" border={true} />

<Callout icon="📘" theme="info">
  ### Note

  If you’ve already added an Incoming Webhook before, the **Configure** option appears instead of **Add**. Select **Configure** to create another Incoming Webhook.
</Callout>

7. Provide a name for the webhook and upload an image if necessary.

8. Select "Create".

   <Image src="https://files.readme.io/7df13f5-7d761cad-7a15-4b43-b0a1-c749a66ca60d.png" align="center" width="60%" border={true} />

9. Copy and save the unique webhook URL present in the dialog. The URL maps to the channel, and you can use it to send information to Teams.

10. Select "Done". The webhook is now available in the Teams channel.

<Image src="https://files.readme.io/0cd530b-f6399868-0353-4424-82db-8886d215c249.png" align="center" width="60%" border={true} />

### Atlas Integration

#### Add webhook via Atlas console

* Navigate to **Notifications** > **Manage alerts** > **Create alert** and select **Webhook** as the delivery method.

<Image src="https://files.readme.io/c12a7a6-image.png" align="center" border={true} />

* Set up the fields below using the information collected from the Webhook and API integration
  * *Name of Webhook*
  * *Callback URL*
  * *Category* (select **All** to capture every notification category with a single selection)
  * *Severity*
  * *Request template*
  * *Authorization keys (Headers)*

<Image src="https://files.readme.io/a4aa337-Screenshot_2024-02-26_at_12.12.46_PM.png" align="center" width="60%" border={true} />

<Image src="https://files.readme.io/c44f1c3-1eb8283f-1b8a-4293-8734-feab2b98c0f6.png" align="center" border={true} />

After setting up the webhook, you should see notifications in your designated Teams channel when Atlas triggers it. The notifications will appear as cards in the channel, displaying relevant information from Atlas.

#### Test Notification

<Image src="https://files.readme.io/da10b80-9d64ff1b-491d-49ee-9fcd-57527ab1975c.png" align="center" border={true} />

<Image src="https://files.readme.io/9ea9f00-fdc32492-9c06-401b-8d15-049f031bab07.png" align="center" border={true} />

## Method 2: MS Teams Workflows Integration

This method uses the [Workflows](https://support.microsoft.com/en-us/office/browse-and-add-workflows-in-microsoft-teams-4998095c-8b72-4b0e-984c-f2ad39e6ba9a) feature in MS Teams to set up notifications from Atlas.

### Set up the Teams workflow

<Columns layout="auto">
  <Column>
    1. In Microsoft Teams, select the channel to set up the notification workflow.
    2. Click the "More options" (•••) button next to the channel name.
    3. Select "Workflows" from the dropdown menu.
  </Column>

  <Column>
    ![More options menu in a Teams channel with Workflows highlighted](https://files.readme.io/37775db649771e72390ad98fb910df32c98e7bab547395578f2805a2cac29263-image.png)
  </Column>
</Columns>

4. In the "Workflows" panel, scroll down and select "Send webhook alerts to a channel".

![](https://files.readme.io/f5586772a95d452b0f92ad481fda7b94caa8b8e2a071353a4f9752fb7e1ddad3-image.png)

5. Select the team and channel you want to use, then click "Add workflow."

![](https://files.readme.io/d32171f9e105f61a9596ad0f318df3d6df18474e914f0fee99d09e45ef4d4c92-image.png)

6. Once added, you'll receive a Webhook URL. Copy this URL, as you'll need it for the Atlas configuration described below.

![](https://files.readme.io/3735ef5d1ffe713224dc38d40256d71cd201eaf8aac0201d58fca89fa5d3f1f7-image.png)

### Configure Atlas to use the Teams workflow

1. In the Atlas console, navigate to **Notifications** > **Manage alerts** > **Create alert** and select **Webhook** as the delivery method.

<Image src="https://files.readme.io/83f87d5110a4e647121f60a8a41c3bea5d91e1549c1e44a496fa4054093efd44-image.png" align="center" width="80%" border={true} />

2. Fill in the following details:

   * *Name of Webhook*
   * *Callback URL* (use the webhook URL copied from Teams)
   * *Category*
   * *Severity*
   * *Request template* (you can use the sample payload template provided below)

   ```json
   {
     "type": "message",
     "attachments": [
       {
         "contentType": "application/vnd.microsoft.card.adaptive",
         "content": {
           "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
           "type": "AdaptiveCard",
           "version": "1.0",
           "body": [
             {
               "type": "TextBlock",
               "text": "{{ .Cluster.Name }}",
               "size": "Large",
               "weight": "Bolder",
               "color": "Attention",
               "horizontalAlignment": "Center"
             },
             {
               "type": "TextBlock",
               "text": "{{ .Message }}",
               "wrap": true,
               "spacing": "Medium"
             }
           ]
         }
       }
     ]
   }
   ```

   * *Authorization keys* (Headers): Set to `application/json`

<Image src="https://files.readme.io/1b44c36315950b492868edc47d968a471ac143e633cfb5723aa8ccbacbd8ef8a-image.png" align="center" width="75%" />

3. Click "Add" to create the webhook in Atlas.

<Columns layout="auto">
  <Column>
    After setting up the webhook, you should see notifications in your designated Teams channel when Atlas triggers it. The notifications will appear as cards in the channel, displaying relevant information from Atlas.
  </Column>

  <Column>
    <Image src="https://files.readme.io/488293904cbd04c0e77abf5e310873cc5c016d49739e0b0872ad155447d3301c-image.png" align="center" width="100%" border={true} />
  </Column>
</Columns>

<Columns layout="auto">
  <Column>
    You can also view these notifications in the Atlas console. Click the **Notifications** icon in the top navigation to see them alongside the ones delivered to Teams.
  </Column>

  <Column>
    ![Atlas console Notifications panel showing recent cluster alerts](https://files.readme.io/7ece1c16c852c687b03deca609a0c09687fa6e203baf9bd8d6097a3aa107e9b1-image.png)
  </Column>
</Columns>

# Example of Cost Anomaly Notifications in Slack

<Callout icon="📘" theme="info">
  ### Note

  For most use cases, we recommend using the [native Slack integration](observability-tutorial-set-up-slack-notifications.md), which requires no webhook configuration. The webhook-based example below is available for customization of cost anomaly notification templates.
</Callout>

<Columns layout="auto">
  <Column>
    Atlas supports sending cost-anomaly detection notifications via Slack, providing enhanced visualization and faster alerts. This integration lets your team receive real-time notifications with graphical data when cost anomalies are detected in your clusters.

    Cost anomaly notifications in Slack include detailed information and visualizations:
  </Column>

  <Column>
    <Image src="https://files.readme.io/875c8e60d3b380d13c937c1d36a4bc87477a1acc0f1d428d7e054a12c4ee82be-image.png" align="center" width="100%" />
  </Column>
</Columns>

## Setting up Slack notifications

To configure Slack notifications for cost anomalies:

<Columns layout="auto">
  <Column>
    1. **Create a Slack Webhook**:
       * In Slack, go to your workspace settings and create an incoming webhook
       * Follow Slack's guide at [api.slack.com/messaging/webhooks](https://api.slack.com/messaging/webhooks) for detailed instructions.

    2. **Configure the Webhook in Atlas**:
       * Navigate to **Notifications** > **Manage alerts** > **Create alert** in the Atlas console and select **Webhook** as the delivery method
       * Set the following fields:
         * **Name**: Provide a descriptive name (e.g., "Slack Cost Anomalies")
         * **Callback URL**: Paste your Slack webhook URL
         * **Category**: Select "Reporting anomalies" for cost anomaly notifications
         * **Severity**: Choose which severity levels trigger notifications
         * **Request template**: When you select "Reporting anomalies," Atlas automatically provides a pre-filled Slack-optimized template for instant setup
  </Column>

  <Column>
    <Image src="https://files.readme.io/8006ae3d2b338965240a3bc7d6b98cae3c2c0342cb1ad36718ee35ea880ae3e3-image.png" align="center" width="100%" />
  </Column>
</Columns>

<Callout icon="📘" theme="info">
  ### Note

  Consider creating a dedicated Slack channel for cost anomaly notifications.
</Callout>

### Template customization (optional)

While the pre-filled template works out of the box, you can still customize it if needed. The default Slack template structure looks like this:

```json
{{- $details := fromJSON .Details -}}
{
    "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Cluster Name:* {{ .Cluster.Name }}\n {{ $details.slackMessage }}"
            }
        }
        {{- if $details.chartUrl }},
        {
            "type": "image",
            "image_url": "{{$details.chartUrl}}",
            "alt_text": "Anomaly Detection Chart"
        }
        {{- end }}
    ]
}
```

You can modify this template to include additional information or formatting based on your requirements. The template uses the Go templating syntax and can access all the data available in the notification payload.

Available variables include:

| Variable                | Description                                 |
| ----------------------- | ------------------------------------------- |
| `.Cluster.Name`         | The name of the cluster                     |
| `.Cluster.ID`           | The unique identifier of the cluster        |
| `.Details`              | Contains all the anomaly-specific data      |
| `$details.slackMessage` | A formatted message about the anomaly       |
| `$details.chartUrl`     | URL to a chart visualization (if available) |

For more advanced webhook configurations and examples, see our [Notification Webhook Setup guide](setup-notification-webhook.md).