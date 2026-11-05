---
updatedAt: 2026-09-17T08:22:04.000Z
---

# Notifications

This Atlas feature informs you via UI, [Slack](observability-tutorial-set-up-slack-notifications.md), or [webhook](setup-notification-webhook.md) about key issues affecting the cluster. This guide outlines all notification types and examples you may see in Atlas with relevant action points.

Once new items are ready for you to view, the bell icon in the top menu will show a count. You can view all items in the **Notifications** page.

Due to the dynamic nature of Kubernetes clusters, notifications are set to expire automatically in 24 hours.

![](https://files.readme.io/407305e-image.png)

# Notification severity types

Atlas uses several notification severity types to categorize the severity and importance of messages.

| Severity | Description                                                                                                 |
| :------- | :---------------------------------------------------------------------------------------------------------- |
| Critical | Indicates a severe issue that requires immediate attention and may significantly impact cluster operations. |
| Error    | Signifies a problem that is causing a malfunction or preventing expected behavior.                          |
| Warning  | Alerts about potential issues or situations that could lead to problems if not addressed.                   |
| Info     | Provides general information about cluster operations, updates, or status changes.                          |
| Success  | Confirms that an operation or process has completed successfully.                                           |

# Notification categories

Atlas organizes notifications into the following categories:

* **Workload Autoscaler** - Workload autoscaling event notifications
* **Node Autoscaler** - Node autoscaling and scheduling event notifications
* **Reporting anomalies** - Cost and performance anomaly detection notifications
* **Inventory** - New cloud provider instance availability notifications
* **Security** - Image security and best practice notifications
* **Other** - General cluster operations, system connectivity, and configuration issues

When you configure an alert to deliver notifications to [Slack](observability-tutorial-set-up-slack-notifications.md) or a [webhook](setup-notification-webhook.md), the **Category** dropdown includes an **All** option. **All** captures every notification category with a single selection, so one alert can cover all of the categories above instead of you selecting them individually. Select individual categories only when you want to narrow the alert's scope.

# Complete notifications reference

This section provides a comprehensive list of all notifications that Atlas can generate, organized by category and severity level.

## Critical notifications

### Other category

| Notification                                      | Description                                                                                                                                                                                                                                              |
| :------------------------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| The Atlas agent is unable to connect to the API | The Atlas agent cannot communicate with Atlas API endpoints and is unable to send data. Check the status of agent pods, network connectivity, firewall restrictions, or authentication.                                                              |
| Cluster controller not responding                 | The cluster controller component is unresponsive. This can prevent cluster management operations and autoscaling from functioning properly.                                                                                                              |
| IP Address quota exceeded                         | Your cloud provider's IP address quota has been exceeded, preventing new nodes from being created with proper network connectivity.                                                                                                                      |
| Node Configuration Validation Failed              | Node configuration is invalid. Check the notification details for specific configuration errors.                                                                                                                                                         |
| Node deletion failed                              | Atlas was unable to delete a node from the cluster. This may indicate permission issues or cloud provider API problems.                                                                                                                                |
| Operation failed                                  | A critical cluster operation has failed. Check notification details for which operation failed and why.                                                                                                                                                  |
| Spot Instance quota exceeded                      | Your cloud provider's Spot Instance quota has been exceeded, preventing Atlas from launching cost-effective Spot Instances for your workloads. Additionally, since Spot Fallback is not enabled, the Autoscaler might not be able to add any capacity. |

## Error notifications

### Other category

| Notification                                                     | Description                                                                                                   |
| :--------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------ |
| Missing permission when adding a node to a target group          | Atlas lacks the necessary IAM permissions to add nodes to target groups in your load balancer.              |
| Missing permission when adding a node to load balancer(s)        | Atlas lacks the necessary IAM permissions to add nodes to load balancers.                                   |
| Missing permission when adding a node to target groups           | Atlas lacks the necessary IAM permissions to add nodes to target groups.                                    |
| Missing permission when adding VMSS IP address to a backend pool | Atlas lacks the necessary Azure permissions to add Virtual Machine Scale Set IP addresses to backend pools. |
| Missing permission when deleting a node from target groups       | Atlas lacks the necessary IAM permissions to remove nodes from target groups.                               |
| Missing permission when removing a node from load balancer(s)    | Atlas lacks the necessary IAM permissions to remove nodes from load balancers.                              |
| SSO Connection problem                                           | There is an issue with your Single Sign-On configuration preventing proper user authentication.               |

## Warning notifications

### Workload Autoscaler category

| Notification                         | Description                                                                                                                                                                                                     |
| :----------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Continuous OOMKilled Events Detected | A continuous stream of out-of-memory (OOMKilled) events was detected for a workload. Workload automation is disabled and re-enabled automatically after a cooldown period. Check if the application is healthy. |

### Node Autoscaler category

| Notification         | Description                                                                                                                                                     |
| :------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Pending pod detected | A pod has been stuck in a pending (unscheduled) state. Atlas is monitoring the situation; this notification resolves automatically once the pod is scheduled. |

### Reporting anomalies category

| Notification                                                               | Description                                                                                                                                                                                                   |
| :------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Cost anomaly detected on the {metric} metric / Cost anomalies detected ... | Atlas monitors cost and efficiency metrics across your cluster (e.g. compute cost, CPU/RAM overprovisioning, network egress). The notification name and message identify which metrics triggered detection. |

### Other category

| Notification                                   | Description                                                                                                                                                                                                                                                   |
| :--------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Failed Helm Test of atlas-workload-autoscaler | The Helm test for the Atlas workload autoscaler component has failed, indicating potential deployment or configuration issues.                                                                                                                              |
| Failed to reconcile cluster                    | A non-critical reconciliation issue occurred. While not immediately severe, this should be monitored and addressed.                                                                                                                                           |
| GPU quota exceeded                             | Your cloud provider's GPU quota has been exceeded, preventing allocation of GPU resources for workloads that require them.                                                                                                                                    |
| Outdated cluster-controller                    | The cluster controller component is running an outdated version and should be updated for optimal functionality and security.                                                                                                                                 |
| Unable to update pool                          | Atlas was unable to update a node pool configuration, which may prevent scaling operations or configuration changes.                                                                                                                                        |
| Spot Instance quota exceeded                   | Your cloud provider's Spot Instance quota has been exceeded, preventing Atlas from launching cost-effective Spot Instances for your workloads. Fires as a warning when Spot Fallback is enabled, since the Autoscaler can fall back to on-demand instances. |

## Info notifications

### Inventory category

| Notification                    | Description                                                                                                     |
| :------------------------------ | :-------------------------------------------------------------------------------------------------------------- |
| New machines available in AWS   | New AWS instance types have become available and can now be used by Atlas for your clusters.                  |
| New machines available in Azure | New Azure virtual machine sizes have become available and can now be used by Atlas for your clusters.         |
| New machines available in GCP   | New Google Cloud Platform machine types have become available and can now be used by Atlas for your clusters. |

### Other category

| Notification               | Description                                                                                                                         |
| :------------------------- | :---------------------------------------------------------------------------------------------------------------------------------- |
| Read-Only access activated | Atlas has been activated in read-only mode, which means it can monitor your cluster but cannot make changes to it.                |
| Trial expires soon         | Your Atlas trial period is approaching its expiration date. Consider upgrading to a paid plan to continue using Atlas features. |
| Trial has expired          | Your Atlas trial period has expired. Upgrade to a paid plan to restore full functionality.                                        |

## Success notifications

### Other category

| Notification       | Description                                                                                                                                          |
| :----------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------- |
| Cluster reconciled | The cluster has been successfully reconciled, indicating that Atlas has successfully synchronized the desired cluster state with the actual state. |

## Taking action on notifications

When you receive notifications, consider the following general action steps:

1. **Critical notifications** - Address immediately as they can severely impact cluster operations
2. **Error notifications** - Investigate and resolve permission issues or configuration problems
3. **Warning notifications** - Review for potential cost savings or performance improvements
4. **Info notifications** - Stay informed about new capabilities and system status
5. **Success notifications** - Confirm that operations completed as expected

For specific troubleshooting steps related to individual notifications, consult the relevant Atlas documentation or contact support if the issue persists.