---
updatedAt: 2026-09-17T08:22:04.000Z
---

# Set up webhook notifications

Check how to configure webhooks to send important notifications to external Ops systems when something happens with your clusters.

Select the organization you want to configure the webhook to send notifications from the Atlas console to an external system.

1. Click the **Notifications** icon in the top navigation, then click **View all**.

2. Click **Manage alerts**, then click **Create alert**.

3. Under **Delivery method**, select **Webhook**.

4. Configure the webhook with the fields below, then click **Save**.

| Field             | Description                                                                                                                         |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Callback Url      | The callback URL to send the requests to                                                                                            |
| Name              | The name of the webhook configuration                                                                                               |
| Category          | The notification categories that trigger this webhook. Select **All** to capture every category with a single selection.            |
| Clusters          | The clusters that trigger this webhook. The **Include non-cluster alerts** toggle includes alerts not tied to any specific cluster. |
| Severity Triggers | The severity levels that will trigger that notification                                                                             |
| Template          | The template of the request that will be sent to the callback URL                                                                   |
| Headers           | The authorization keys or other headers sent with the request, such as `Content-Type: application/json`                             |

The Request Template should be a valid `JSON`. We provide a better overview of how to customize the payloads in the next section.

## Request Template Configuration

We allow users to fully customize the request sent to external systems, in that way, we can support almost any application out there. The Request Template is the payload sent within the webhook call. The following variables from notifications are available:

| Variable                  | Description                                                                                                                                                                                                         | Output type | Usage                            |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | -------------------------------- |
| NotificationID            | The UUID of the notification, it is unique                                                                                                                                                                          | String      | \{\{ .NotificationID }}          |
| OrganizationID            | The organization that owns the notification                                                                                                                                                                         | String      | \{\{ .OrganizationID }}          |
| Severity                  | Indicates the severity of the impact on the affected system.                                                                                                                                                        | String      | \{\{ .Severity }}                |
| Name                      | Name of the notification                                                                                                                                                                                            | String      | \{\{ .Name }}                    |
| Message                   | A high-level text summary message of the event.                                                                                                                                                                     | String      | \{\{ .Message}}                  |
| Details as JSON           | Free-form details from the event, rendered as a JSON object. Use in fields that accept objects.                                                                                                                     | JSON object | \{\{ toJSON .Details }}          |
| Details as Escaped String | Free-form details from the event, rendered as an escaped JSON string. Use in fields that require a string value (e.g., `description`).                                                                              | String      | \{\{ toEscapedString .Details }} |
| Timestamp                 | When the Notification was created by Atlas                                                                                                                                                                        | String      | \{\{ toISO8601 .Timestamp }}     |
| Cluster                   | Cluster information, might be empty, if the notification isn't specific                                                                                                                                             | JSON object | \{\{ toJSON .Cluster }}          |
| Cluster.ID                | The unique identifier of the cluster on Atlas                                                                                                                                                                     | String      | \{\{ .Cluster.ID }}              |
| Cluster.Name              | Name of the cluster on Atlas                                                                                                                                                                                      | String      | \{\{ .Cluster.Name }}            |
| Cluster.ProjectName       | The cloud provider's resource grouping scope where the cluster is hosted (e.g., GCP Project ID, AWS Account ID, Azure Subscription ID). Useful for routing or tagging alerts by project across a fleet of clusters. | String      | \{\{ .Cluster.ProjectName }}     |
| Cluster.ProviderType      | Cloud provider of the cluster (`eks`, `gke`, `aks`, `kops`)                                                                                                                                                         | String      | \{\{ .Cluster.ProviderType }}    |

> 📘 Note
>
> `Cluster.ProjectName` and `Cluster.ProjectNamespaceID` return the same value. `ProjectNamespaceID` is kept for backward compatibility with existing templates — prefer `ProjectName` in new webhook configurations.

As you can see, the variables are in Go template style, and you can mix them anywhere you want in your Request Template.

> 🚧 String-typed fields require string output
>
> The rendered template must produce valid JSON that conforms to the Atlas webhook request schema. Fields such as `title`, `description`, and `deduplication_key` are typed as **strings**. If a template function produces a JSON object in a string-typed field — for example, `"description": {{ toJSON .Details }}` — the Atlas API rejects the request with a `422 validation_error`.
>
> Use `{{ toEscapedString .Details }}` to pass structured notification details into a string-typed field. Use `{{ toJSON .Details }}` only in fields that accept a JSON object value (for example, PagerDuty's `custom_details`).
>
> To extract individual values from `.Details`, create a local variable with `fromJSON`:
>
> ```
> {{- $details := fromJSON .Details -}}
> ```
>
> You can then reference specific fields like `{{ $details.fieldName }}` anywhere in your template.

## Example of Request Template Slack

> 📘 Note
>
> Atlas offers a [native Slack integration](observability-tutorial-set-up-slack-notifications.md) that requires no webhook setup. Use the webhook approach below only if you need custom message formatting.

To send a Slack notification, we need a simple JSON request with a payload in the body.

```json
{
    "text": "Atlas - {{ .Name }}",
    "blocks": [
     {
      "type": "section",
      "text": {
       "type": "mrkdwn",
       "text": "{{ .Cluster.Name }}<br> {{ .Message}}"
      }
     }
    ]
}
```

Creating the webhook URL isn't within the scope of this how-to. You can find more information in the [following link](https://api.slack.com/messaging/webhooks).

## Example of Request Template PagerDuty

PagerDuty accepts Alerts in the endpoint `https://events.pagerduty.com/v2/enqueue`. The content is a simple JSON request with the payload in the body. You can find below an example of a request template with the available variables:

```json
{
    "payload": {
        "summary": "{{ .Message }}",
        "timestamp": "{{ toISO8601 .Timestamp }}",
        "severity": "critical",
        "source": "Atlas",
        "component": "{{ .Cluster.Name}}-{{ .Cluster.ProviderType}}-{{ .Cluster.ProjectNamespaceID }}",
        "group": "{{ .Name }}",
        "class": "kubernetes",
        "custom_details": {
            "details": {{ toJSON .Details }}
        }
    },
    "routing_key": "--routing_key--",
    "dedup_key": "{{ .NotificationID }}",
    "event_action": "trigger",
    "client": "Atlas",
    "client_url": "https://console.atlas.example/external-clusters/{{ .Cluster.ID}}?org={{ .OrganizationID }}"
}
```

Note that `dedup_key` was set as the `NotificationID`. This field is unique in Atlas and will ensure you won't produce an alert with the same content more than once.

How to create the `routing_key` isn't in the scope of this how-to. You can find more information at <https://developer.pagerduty.com/docs/>

## Example of Request Template OpsGenie

OpsGenie accepts Alerts in the endpoint `https://api.opsgenie.com/v2/alerts`. The content is a simple JSON request with the payload in the body. You can find below an example of a request template with the available variables:

```json
{
    "message": "{{ .Message }}",
    "alias": "{{ .NotificationID }}",
    "description":{{ toEscapedString .Details }},
    "details": {{ toJSON .Details }},
    "tags": ["{{ .Cluster.ProjectName }}"],
    "priority":"P3"
}
```

Note that `alias` was set as the `NotificationID`. This field is unique in Atlas and will ensure you won't produce an alert with the same content more than once. The `tags` array uses `Cluster.ProjectName` to route and group alerts by the cluster's cloud provider project across your fleet — for example, into a single OpsGenie team or dashboard filtered by project.

See more details for all supported fields <https://docs.opsgenie.com/docs/alert-api#create-alert>

**NOTE:** You must set **Content-Type: application/json** in Atlas UI Webhook headers section.

## Example of Request Template incident.io

incident.io accepts alert events at `https://api.incident.io/v2/alert_events/http/<your-source-id>`. The `description` field must be a string: use `toEscapedString` to pass notification details as a readable string value.

```json
{
    "title": "{{ .Message }}",
    "description": {{ toEscapedString .Details }},
    "status": "firing",
    "deduplication_key": "{{ .NotificationID }}",
    "metadata": {
        "cluster": "{{ .Cluster.Name }}",
        "severity": "{{ .Severity }}"
    }
}
```

Note that `deduplication_key` was set as the `NotificationID`. This field is unique in Atlas and prevents duplicate alerts for the same event.

Replace `<your-source-id>` in the callback URL with the source ID from your incident.io HTTP alert source configuration. Set the **Authorization** header in the Atlas webhook headers section to the key provided by incident.io.

See more details for supported fields at <https://api-docs.incident.io/>

## Anomaly Detection Webhook

It is possible to configure a Webhook to receive notifications about newly detected Anomalies. For this select category `Security` and operation `Anomalies`. Provide an URL to your endpoint and configure the JSON request template like

```json JSON
{
    "details": {{ toJSON .Details }},
    "cluster": {{ toJSON .Cluster }}
}
```

The structure of the `.Details` JSON object is as follows:

```json JSON
{
  "anomaly_id": "<UUID of the detected anomaly>",
  "status: "<Anomaly status. One of open/acked/closed>",
  "rule_metadata": {
    "id": "<ID of the rule>",
    "name": "<Name of the rule>",
    "type": "<Type of the rule>",
    "category": "<Category defined for the rule>",
    "labels": {
      // Labels defined for the rule that detected the anomaly
      "custom": "label"
    }
  },
  "events": [
    {
      "timestamp": "<Event timestamp in RFC3339 format>",
      "type": "<Type of the event>",
      "cluster": {
        "id": "<ID of the cluster the event was recorded in>",
        "name": "<Name of the cluster the event was recorded in>",
        "organization_id": "<ID of the organization the event was recorded in>",
      },
      "resource": {
        "namespace": "<Kubernetes namespace the event was recorded in>",
        "pod": "<Name of the pod the event was recorded in>",
        "container": "<Name of the container the event was recorded in>",
        "workload_id": "<ID of the workload the event was recorded in>",
      	"workload_name": "<Name of the workload the event was recorded in>",
      	"workload_kind": "<Kind of the workload, e.g., Deployment>",
        "workload_id": "<ID of the workload>",
        "pod_labels": {
          // labels set on the pod
        },
        "pod_annotations": {
          // annotations set on the pod
        },
      },
      
      "process": "<Name of the process the event was recorded for>",
      "host_pid": "<PID on the host of the process the event was recorded for>",
      "payload_digest": "<Unique key used to group related events>",
      
      // only one of the following top level fields will be set
      
      "exec": {
        "path": "<Path to the executed file>",
        "args": ["<Arguments passed to the process>"],
        "sha256": "<SHA256 hash of the executed file>",
        "file_details": { // optional value that will not always be present
          "category": "<Category the file falls in e.g., crypto>",
          "malware_name": "<Name of the malware, if file has been categorized to be related to malware>",
          "malware_version": "<Version of the malware the file was identified to be>"
        },
      },
      
      "file": {
        "path": "<Path to the file (e.g., in magic write events, path to file)>",
      },
      
      "tcp": {
        "destination": {
          "ip": "<IP address the event connected to>",
          "port": <Port to which the process connected to>
        },
        "ip_details": { // optional value that will not always be present; see ipDetails section below for an example
          "ip_address": "<Same as destination.ip>",
          "ip_version": <version of the IP used>,
          "country_code": "<Country code of the IP address (e.g., US)>",
          "isp": "<Name of the Internet Service Provider that owns the IP address>",
          "domain": "<Correlated domain name to the IP>",
          "hostnames": [
          	"<Additional hostnames for this IP>"
          ],
        	"is_tor": <Flag indicating if the given IP is a Tor node>,
        	"abuse_confidence_score": <Score from 0-100 to indicate how confidently the IP was marked as malicious>
        },
        "network_details": { // optional value that will not always be present
          "category": "<Category the IP falls in e.g., crypto>"
        }
      },
      
      "dns": {
        "question": "<Domain ask to be resolved>",
        "answers": [
    			{
    				"type": "<Type of the DNS answer. One of PUBLIC/PRIVATE/CNAME>",
    				"ip": "<Answer IP, only set if type is either PUBLIC/PRIVATE>",
    				"cname": "<Answer CNAME, only set if the type is CNAME>"
    			}
  			],
        "flow_direction": "<Direction of the network flow. One of INGRESS/EGRESS/UNKNOWN>",
        "network_details": { // optional value that will not always be present
          "category": "<Category the IP falls in e.g., crypto>"
        }
      },
      
      "socks5": {
        "flow_direction": "<Direction of the network flow. One of INGRESS/EGRESS/UNKNOWN>",
        "role": "<Role the process takes in the SOCKS5 proxy. One of UNKNOWN/SERVER/CLIENT>",
        "command_or_reply": <Command or reply from the SOCKS5 client/server, see RFC1928 for more details>,
        "address_type": "<Type of the address. One of UNKNOWN/IPv4/DOMAIN_NAME/IPv6>",
        "destination": { // only set for address type IPv4/IPv6
          "ip": "<IP address the event connected to>",
          "port": <Port to which the process connected to>
        },
        "destination_domain": "<If addressType is DOMAIN_NAME, the domain the SOCKS5 proxy should connect to>"
      },
      
      "stdio_via_socket": {
        "destination": {
          "ip": "<IP address the event connected to>",
          "port": <Port to which the process connected to>
        },
        "fd": "<File descriptor identified to be hooked up to a socket (either 0,1,2)>"
      }
    },
    <Up to 10 events related to the anomaly>
  ],
}
```

### Rule Types

The following rule types are currently supported:

| Type                                     | Name                                 | Description                                                                                                                                                       |
| :--------------------------------------- | :----------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| crypto\_mining:binary\_executed          | Crypto mining command line arguments | Checks for EXEC events and tries to identify if they are related to crypto miners. The check is based on matching the binary file name, as well as the arguments. |
| crypto\_mining:dns\_lookup               | DNS to crypto mining                 | Checks for DNS events that try resolve a well known crypto related domains.                                                                                       |
| crypto\_mining:tcp\_connect              | TCP connection to crypto mining      | Checks for TCP connections to crypto related IPs.                                                                                                                 |
| network:tcp\_public\_non\_standard\_port | Suspicious Internet connection       | Checks for TCP connections to public IPs on non HTTP related ports (neither 80 nor 443).                                                                          |
| network:suspicious\_destination\_ip      | Suspicious Destination IP            | Checks for network related events, that have a suspicious IP as destination.                                                                                      |
| suspicious\_binary:nezha\_server         | Process related to Nezha server      | Checks EXEC events for execution of the Nezha Monitoring Tool.                                                                                                    |
| suspicious\_binary:vnc\_server           | Process related to VNC server        | Checks EXEC events for execution of VNC servers.                                                                                                                  |
| general:dropped\_binary\_executed        | Dropped new binary (container drift) | Checks for MAGIC\_WRITE events (fires if ELF headers are written to any filesystem).                                                                              |
| general:oom\_killed                      | Process OOM killed                   | Checks if a process was OOM killed.                                                                                                                               |
| ml:suspicious\_container\_stats          | Suspicious container stats           | Leverages Machine Learning to detect suspicious resource usage patterns of containers.                                                                            |

#### ipDetails

Example:

```json
{
  "ip_address": "118.25.6.39",
  "ip_version": 4,
  "country_code": "CN",
  "isp": "Tencent Cloud Computing (Beijing) Co. Ltd",
  "domain": "tencent.com",
  "hostnames": [],
  "is_tor": false
}
```

### Event Types

| Type                 | Description                                                              |
| :------------------- | :----------------------------------------------------------------------- |
| exec                 | Triggered by any executed processes in a pod.                            |
| dns                  | Triggered by any DNS lookup in a pod.                                    |
| file\_change         | Triggered by any write to a file in a pod.                               |
| tcp\_connect         | Triggered by any TCP connection.                                         |
| tcp\_listen          | Triggered by any TCP socket listening in a pod.                          |
| tcp\_connect\_error  | Triggered by any connection errors when trying to open a TCP conenction. |
| process\_oom\_killed | Triggered by any process that got OOM killed.                            |
| magic\_write         | Triggered by any write event that writes an ELF binary header.           |

## Anomaly Detection Webhook for OpsGenie

You can access .Details object inside template by creating local $details variable.

This allows to access anomaly related fields.

```json JSON
{{- $details := fromJSON .Details -}}
{
    "message": "Runtime: {{$details.rule_metadata.name}}",
    "alias": "{{$details.anomaly_id}}",
    "description": "View anomaly details in Console UI https://console.atlas.example/organization/security/runtime/anomalies/{{$details.anomaly_id}}",
    "details": {
       "cluster": "{{.Cluster.Name}}"
    },
    "tags": ["sec"],
    "priority":"P3"
}
```