# Configure applier alerts

> Send HTTP webhook notifications when the Orion Applier applies a recommendation to a workload


The Orion Applier can send an HTTP webhook notification each time it successfully applies a recommendation to a workload. You can configure one or more endpoints to receive these alerts, with optional custom payload formatting and bearer token authentication. Use this to connect Orion to your existing observability or incident management workflows.

## How alerts work

When the Applier applies a patch, it sends an HTTP POST to each configured endpoint. The request includes the header `X-Event-Type: recommendation applied`. The body contains either the full JSON recommendation object (default) or a custom payload that you can define using a Go template.

## Alert payload

### Default payload

If you don't set `payloadTemplate`, the Applier sends the full recommendation object.

<details>
<summary>View an example payload with illustrative values</summary>

```json
{
  "patches": [
    {
      "patch": {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "metadata": {
          "name": "cartservice"
        },
        "spec": {
          "template": {
            "spec": {
              "containers": [
                {
                  "name": "server",
                  "resources": {
                    "requests": {
                      "cpu": "10m",
                      "memory": "88Mi"
                    }
                  }
                }
              ]
            }
          }
        }
      },
      "patchTarget": {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "name": "cartservice"
      },
      "patchType": "application/strategic-merge-patch+json"
    }
  ],
  "recommendation": "https://api.orion.example/v3/workloads/my-cluster/hipster/deployments/cartservice/recommendations/1234567890",
  "workload": {
    "apiVersion": "apps/v1",
    "cluster": "my-cluster",
    "kind": "Deployment",
    "name": "cartservice",
    "namespace": "hipster",
    "resource": "deployments",
    "type": "deployment.apps"
  }
}
```

</details>

### Custom payload template

The `payloadTemplate` field accepts a Go template string with the following variables:

| Variable | Description |
| -------- | ----------- |
| `.recommendation` | Recommendation address |
| `.workload.name` | Workload name |
| `.workload.namespace` | Workload namespace |
| `.workload.cluster` | Cluster name |
| `.workload.kind` | Resource type |
| `.patches` | List of applied patches |
| `.patches[].patchTarget.name` | Patched resource name |
| `.patches[].patchTarget.kind` | Patched resource type |
| `.patches[].patch` | Full patch object |

You can also use the `toPrettyJson` helper function to render `.patches[].patch` as formatted JSON.

Example template:

```yaml
payloadTemplate: |
  {
    "recommendation": "{{ .recommendation }}",
    "workload": "{{ .workload.name }}",
    "namespace": "{{ .workload.namespace }}",
    {{- if .patches }}
    "patches": [
      {{- range $index, $patch := .patches }}
      {
        "target": "{{ $patch.patchTarget.name }}",
        "kind": "{{ $patch.patchTarget.kind }}",
        "patch": "{{ toPrettyJson $patch.patch }}"
      },
      {{- end }}
    ]
    {{- end }}
  }
```

## Configure the endpoint

Add an `applier.alerts.endpoints` block to your `values.yaml` with at least one endpoint. Omit the `authorization` block if the endpoint accepts unauthenticated requests.

Example:

```yaml
applier:
  alerts:
    endpoints:
      - name: my-webhook           # Identifier for this endpoint
        url: https://example.com/webhook
        method: POST
        contentType: application/json
        payloadTemplate: |         # Omit to send the full recommendation object
          { ... }
        authorization:
          type: Bearer
          credentialsFile: /etc/secrets/my-token  # Path to token value file
```

## Store credentials securely

To use bearer token authentication, mount a Kubernetes secret into the Applier pod.

1. Create a Kubernetes secret in the `orion-system` namespace:

    ```sh
    kubectl -n orion-system create secret generic orion-alerts-token \
      --from-literal=token='YOUR_BEARER_TOKEN'
    ```

1. Add the following volume and volume mount to your `values.yaml`:

    ```yaml
    additionalVolumes:
      - name: alerts-token
        secret:
          secretName: orion-alerts-token
          items:
            - key: token
              path: alerts-token

    additionalVolumeMounts:
      - name: alerts-token
        mountPath: /etc/secrets
        readOnly: true
    ```

1. Set `credentialsFile` in the endpoint config to `/etc/secrets/alerts-token`.

To manage the secret with a third-party credential manager, see [Manage secrets with a third-party credential manager][secrets].

## Integration examples

### Slack

Use a [Slack Incoming Webhook][slack-webhook] URL as the endpoint `url`. Omit the `authorization` block, as Slack embeds authentication in the URL.

The Slack payload format uses a `text` field or [Slack's Block Kit UI framework][slack-block-kit].

Example:

```yaml
applier:
  alerts:
    endpoints:
      - name: slack
        url: https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
        method: POST
        contentType: application/json
        payloadTemplate: |
          {
            "text": "Orion applied a recommendation to *{{ .workload.name }}* ({{ .workload.namespace }})"
          }
```

### ServiceNow

Send a Change Request to ServiceNow each time the Applier applies a recommendation. The Applier posts directly to the [ServiceNow Table API][snow-table-api] using Basic Authentication — no intermediate infrastructure required.

#### Prerequisites

Before configuring the Applier, you need a ServiceNow service account with permission to create Change Requests via the API.

1. In ServiceNow, go to **User Administration → Users** and create a new user (for example, `orion-svc`). Uncheck **Password needs reset**.
1. On the user record, open the **Roles** tab and add the **`itil`** role.
1. Confirm the REST API is enabled on your instance (**System Web Services → REST API Explorer**).
1. Test that the account can create Change Requests:

    ```sh
    curl -X POST \
      https://YOUR_INSTANCE.service-now.com/api/now/table/change_request \
      -H "Authorization: Basic $(echo -n 'orion-svc:YOUR_PASSWORD' | base64)" \
      -H "Content-Type: application/json" \
      -d '{
        "short_description": "Test from Orion",
        "type": "standard",
        "risk": "low",
        "impact": "3"
      }'
    ```

    A `201 Created` response confirms correct credentials and permissions.

#### Create the Kubernetes secret

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Don't Base64-encode the credentials</div>


ServiceNow requires `type: Basic` Authentication. The Applier expects the credentials file to contain a plain-text `username:password` string and handles Base64 encoding itself — don't encode the value before storing it in the secret.
</div>


Create a secret in the `orion-system` namespace with your ServiceNow credentials:

```sh
kubectl -n orion-system create secret generic orion-snow-alerts-token \
  --from-literal=token='orion-svc:YOUR_PASSWORD'
```

#### Configure the Applier

Add the following to your `values.yaml`, replacing the `url` with your ServiceNow instance:

```yaml
applier:
  alerts:
    endpoints:
      - name: servicenow-change-request
        url: https://YOUR_INSTANCE.service-now.com/api/now/table/change_request
        method: POST
        contentType: application/json
        payloadTemplate: |
          {
            "short_description": "Orion: {{ .workload.name }} ({{ .workload.namespace }}) recommendation applied",
            "description": "Orion applied a recommendation automatically.\n\nCluster: {{ .workload.cluster -}}
              \nWorkload: {{ .workload.name -}}
              \nNamespace: {{ .workload.namespace -}}
              \nKind: {{ .workload.kind -}}
              \nRecommendation: https://app.orion.example/workloads/{{ .workload.cluster -}}
              /{{ .workload.namespace }}/{{ .workload.resource }}/{{ .workload.name -}}
              /recommendations\n\nPatches applied:
              {{- range .patches -}}
              {{- range .patch.spec.template.spec.containers -}}
              \n  - Container: {{ .name -}}
              \n    CPU Request: {{ .resources.requests.cpu -}}
              \n    Memory Request: {{ .resources.requests.memory -}}
              \n    CPU Limit: {{ or .resources.limits.cpu "none" -}}
              \n    Memory Limit: {{ or .resources.limits.memory "none" -}}
              {{- end -}}
              {{- end -}}
              ",
            "type": "standard",
            "risk": "low",
            "impact": "3",
            "category": "Software"
          }
        authorization:
          type: Basic
          credentialsFile: /etc/secrets/alerts-token

additionalVolumes:
  - name: alerts-token
    secret:
      secretName: orion-snow-alerts-token
      items:
        - key: token
          path: alerts-token

additionalVolumeMounts:
  - name: alerts-token
    mountPath: /etc/secrets
    readOnly: true
```

This produces a Change Request description like:

```
Orion applied a recommendation automatically.

Cluster: my-cluster
Workload: cartservice
Namespace: hipster
Kind: Deployment
Recommendation: https://app.orion.example/workloads/my-cluster/hipster/deployments/cartservice/recommendations

Patches applied:
  - Container: server
    CPU Request: 10m
    Memory Request: 88Mi
    CPU Limit: none
    Memory Limit: none
```

Adjust these payload fields for your ServiceNow instance:

| Field | Description | Example values |
| ----- | ----------- | -------------- |
| `type` | Change type | `standard`, `normal`, `emergency` |
| `risk` | Risk level | `low`, `medium`, `high` |
| `impact` | Impact level (numeric) | `1` (High), `2` (Medium), `3` (Low) |
| `category` | Change category | `Software`, `Infrastructure` |
| `assignment_group` | Auto-assign to a team (optional) | Use the group's `sys_id` |

### Amazon Simple Notification Service

[Amazon Simple Notification Service (SNS)][aws-sns-sigv4] requires AWS Signature Version 4 (SigV4)-signed API calls and doesn't accept arbitrary HTTP POSTs directly. To bridge this, deploy an AWS API Gateway and a Lambda function as a webhook receiver. The Lambda function validates the bearer token and publishes the payload to an SNS topic.

**Workflow:** Applier → HTTPS (Bearer token) → API Gateway → Lambda → SNS topic → subscribers (for example, Amazon Simple Queue Service (SQS), email)

#### Provision AWS components

Create the following AWS resources:

- An **SNS topic** to receive messages
- A **Lambda function** that validates the bearer token and publishes to the SNS topic
- An **API Gateway HTTP API** with a `POST /webhook` route fronting the Lambda function
- (Optional) An **SQS queue** subscribed to the SNS topic, for consuming messages

Use the following settings as a starting point when provisioning these resources:

```sh
export AWS_REGION=us-east-1
export TOPIC_NAME=orion-alerts        # SNS topic name
export API_NAME=orion-sns-webhook     # API Gateway name
export FN_NAME=orion-sns-publisher    # Lambda function name
export TOKEN=$(openssl rand -hex 24)       # Bearer token (auto-generated)
export QUEUE_NAME=orion-alerts-queue  # Optional: SQS queue for consuming messages
```

After provisioning, record the **Webhook URL** (the API Gateway endpoint) and the **Bearer token**. You'll need both to configure the Applier.

#### Create the Kubernetes secret

Create a secret in the `orion-system` namespace using the bearer token from the provisioning step:

```sh
kubectl -n orion-system create secret generic orion-alerts-token \
  --from-literal=token='YOUR_BEARER_TOKEN'
```

#### Configure the Applier

Add the following to your `values.yaml`, replacing the `url` with your API Gateway endpoint:

```yaml
applier:
  alerts:
    endpoints:
      - name: sns-webhook
        url: https://YOUR_API_ID.execute-api.YOUR_REGION.amazonaws.com/prod/webhook
        method: POST
        contentType: application/json
        payloadTemplate: |
          {
            "recommendation": "{{ .recommendation }}",
            "workload": "{{ .workload.name }}",
            "namespace": "{{ .workload.namespace }}",
            {{- if .patches }}
            "patches": [
              {{- range $index, $patch := .patches }}
              {
                "target": "{{ $patch.patchTarget.name }}",
                "kind": "{{ $patch.patchTarget.kind }}"
              },
              {{- end }}
            ]
            {{- end }}
          }
        authorization:
          type: Bearer
          credentialsFile: /etc/secrets/alerts-token

additionalVolumes:
  - name: alerts-token
    secret:
      secretName: orion-alerts-token
      items:
        - key: token
          path: alerts-token

additionalVolumeMounts:
  - name: alerts-token
    mountPath: /etc/secrets
    readOnly: true
```

#### Test by consuming from SQS (optional)

If you set up the SQS queue, poll it to verify messages arrive after the Applier applies a recommendation:

```sh
aws sqs receive-message \
  --region YOUR_REGION \
  --queue-url https://sqs.YOUR_REGION.amazonaws.com/YOUR_ACCOUNT_ID/orion-alerts-queue \
  --max-number-of-messages 10 \
  --wait-time-seconds 20
```

## Apply the configuration

```sh
helm upgrade --install orion \
  oci://registry.orion.example/library/orion \
  --namespace orion-system \
  -f values.yaml \
  --reset-then-reuse-values
```

## Verify

1. Enable debug mode in your `values.yaml` (`applier.debug: true`) and redeploy:

   ```sh
   helm upgrade --install orion \
   oci://registry.orion.example/library/orion \
   --namespace orion-system \
   --set applier.debug=true \
   -f values.yaml \
   --reset-then-reuse-values
   ```

1. Trigger an apply (via [**Apply Now** in the UI][apply-ui] or [`orion apply` in the CLI][apply-cli]).

1. Check the Applier logs:

    ```sh
    kubectl logs -n orion-system -l component=applier --tail=-1
    ```

    Look for this log entry:

    ```json
    {"level":"debug","message":"alerting endpoints of recommendation applied"}
    ```

    Followed by:

    ```json
    {"level":"info","message":"workload is healthy after patching"}
    ```
<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Reference<span class="related-topics-divider">/</span><a href="/docs/reference/applier-events/">Applier Kubernetes events</a>
                </h5>
                <p>Kubernetes Events emitted by the Applier and how to monitor them</p>
            </div>
        </div>


[apply-cli]: /docs/orion-cli/reference/#orion-apply
[apply-ui]: /docs/recommendations/apply/#deploy-recommendations-on-demand
[aws-sns-sigv4]: <https://docs.aws.amazon.com/sns/latest/dg/sns-using-temporary-credentials.html>
[secrets]: /docs/guides/manage-secrets/
[slack-block-kit]: <https://docs.slack.dev/block-kit/>
[slack-webhook]: <https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks/>
[snow-table-api]: <https://developer.servicenow.com/dev.do#!/reference/api/latest/rest/c_TableAPI>
