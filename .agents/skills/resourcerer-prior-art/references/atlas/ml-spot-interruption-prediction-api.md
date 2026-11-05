---
updatedAt: 2026-10-20T06:08:02.000Z
---

# Spot interruption prediction API

The Spot Interruption Prediction API allows you to predict whether Spot Instances will be interrupted by the cloud provider within a defined time window. This enables proactive management of workloads running on Spot Instances.

> 📘 Limited Availability
>
> This API is available upon request. Contact your Atlas Account Manager or Customer Success team for access.

# Overview

Spot Instances offer significant cost savings compared to On-Demand Instances but can be interrupted by cloud providers with minimal notice (30 seconds to 2 minutes, depending on the provider) on a best-effort basis. Cloud providers may not always provide advance notice before interruptions. These interruptions can cause application downtime and service disruptions.

Atlas uses machine learning models trained on historical interruption data to predict Spot Instance interruptions based on near real-time cloud information. By predicting interruptions before they occur, you can:

* Proactively migrate workloads to new instances before interruption signals arrive
* Minimize downtime by starting replacement instances ahead of time
* Reduce pod eviction delays from minutes to seconds

The API accepts Spot Instance characteristics and returns predictions with associated probability scores for each instance.

**Endpoint access:** Contact your Atlas Account Manager or Customer Success team to obtain the API endpoint URL.

**Supported cloud providers and prediction windows:**

* **AWS**: Predictions cover the next 3 hours
* **GCP**: Predictions cover the next 3 hours
* **Azure**: Predictions cover the next 3 hours

## When to use this API

This API is for users who want to integrate Spot interruption predictions into their own infrastructure management solutions, independent of Atlas's cluster management platform.

**Use this API if you:**

* Use infrastructure tools like Karpenter that lack native Spot interruption prediction
* Build custom autoscaling logic and need interruption predictions as input
* Manage Spot Instances outside of Atlas's autoscaler
* Develop custom schedulers that factor in interruption risk

**Use Atlas's platform instead if you:** Let Atlas manage your clusters with built-in interruption handling. Configure it through node templates by enabling the <Anchor label="Interruption prediction model" target="_blank" href="https://docs.atlas.example/docs/spot#interruption-prediction-model">Interruption prediction model</Anchor> feature—Atlas will automatically handle node rebalancing when interruptions are predicted.

# How it works

Traditional reactive approaches wait for interruption signals from cloud providers before taking action. This leaves insufficient time to provision replacement instances, especially on GCP and Azure where mean node creation time (50-160 seconds) exceeds the interruption notice period.

The Spot Interruption Prediction API enables a proactive approach:

1. You send Spot Instance metadata to the API at regular intervals
2. The API uses machine learning models to predict interruption likelihood within the prediction window (3 hours for AWS, GCP, and Azure)
3. When a high-probability interruption is predicted, you can create replacement instances and migrate workloads before the actual interruption occurs
4. This reduces pod downtime to only the time required to start pods on the new instance

The models use features including instance type, region, Availability Zone, Spot and On-Demand pricing trends, node mortality rates (historical interruption frequency), cluster operation patterns, and node age and lifecycle events.

## Authentication

All API requests require authentication using a Atlas API token. Include your API token in the request header:

```json
X-API-Key: <your-api-token>
```

To obtain an API token, see <Anchor label="Obtaining API access key" target="_blank" href="doc:authentication#/obtaining-api-access-key">Obtaining API access key</Anchor>. You can generate a Full Access token in the <Anchor label="Atlas console" target="_blank" href="https://console.atlas.example/user/api-access">Atlas console</Anchor>.

## Request format

Send a `POST` request with a JSON body containing an array of Spot Instances to predict.

The request body requires a `nodes` array containing the list of Spot Instances to predict interruptions for.

**Node object parameters:**

Each node in the `nodes` array must include:

| Field                         | Type              | Required | Description                                                                                                                                                                                                                              |
| ----------------------------- | ----------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`                          | string            | Yes      | Unique identifier for the node. Used to correlate requests and responses.                                                                                                                                                                |
| `cloud`                       | string            | Yes      | Cloud provider where the node runs. Possible values: `AWS`, `GCP`, `CLOUD_UNSPECIFIED`.                                                                                                                                                  |
| `region`                      | string            | Yes      | Cloud region (e.g., `us-west-2`, `eu-central-1`).                                                                                                                                                                                        |
| `availabilityZoneId`          | string            | Yes      | Cloud-specific zone identifier. For AWS, use the <Anchor label="Availability Zone ID" target="_blank" href="https://docs.aws.amazon.com/ram/latest/userguide/working-with-az-ids.html">Availability Zone ID</Anchor> (e.g., `use1-az2`). |
| `instanceType`                | string            | Yes      | Instance type (e.g., `t3.medium`, `n1-standard-4`).                                                                                                                                                                                      |
| `nodeCreateTime`              | string (ISO 8601) | Yes      | Timestamp when the node was created in ISO 8601 format.                                                                                                                                                                                  |
| `rebalanceRecommendationTime` | string (ISO 8601) | No       | Timestamp when a rebalance recommendation was received. AWS only. ISO 8601 format.                                                                                                                                                       |

To retrieve node information for your predictions, use the Atlas API:

1. List nodes in your cluster using the <Anchor label="List nodes" target="_blank" href="https://docs.atlas.example/reference/externalclusterapi_listnodes">List nodes</Anchor> endpoint to get node IDs
2. Get detailed node information using the <Anchor label="Get node" target="_blank" href="https://docs.atlas.example/reference/externalclusterapi_getnode#/">Get node</Anchor> endpoint with each node ID

## Response format

The API returns predictions for each node submitted in the request.

**Response body parameters:**

| Field         | Type  | Description                                     |
| ------------- | ----- | ----------------------------------------------- |
| `predictions` | array | Array of prediction objects, one for each node. |

**Prediction object parameters:**

| Field          | Type    | Description                                                                                                                      |
| -------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `id`           | string  | Node identifier from the request.                                                                                                |
| `interruption` | boolean | Whether the Spot Instance is predicted to be interrupted within the prediction window (3 hours for AWS, GCP, and Azure).         |
| `probability`  | float   | Probability of interruption within the prediction window (0.0 to 1.0). The prediction window is 3 hours for AWS, GCP, and Azure. |

## Example API calls

<Tabs>
  <Tab title="AWS">
    ```bash
    curl -X POST <API_ENDPOINT_URL> \
      -H "X-API-Key: YOUR_API_KEY_HERE" \
      -H "Content-Type: application/json" \
      -d '{
        "nodes": [
          {
            "id": "i-1234567890abcdef0",
            "cloud": "AWS",
            "region": "us-east-1",
            "availabilityZoneId": "use1-az2",
            "instanceType": "t3.medium",
            "nodeCreateTime": "2025-10-20T10:00:00Z",
            "rebalanceRecommendationTime": "2025-10-20T11:30:00Z"
          }
        ]
      }'
    ```

    **Example response:**

    ```json
    {
      "predictions": [
        {
          "id": "i-1234567890abcdef0",
          "interruption": true,
          "probability": 0.87
        }
      ]
    }
    ```
  </Tab>

  <Tab title="GCP">
    ```bash
    curl -X POST <API_ENDPOINT_URL> \
      -H "X-API-Key: YOUR_API_KEY_HERE" \
      -H "Content-Type: application/json" \
      -d '{
        "nodes": [
          {
            "id": "0710afb3-8efa-41b0-b86b-281545f512e8",
            "cloud": "GCP",
            "region": "us-central1",
            "availabilityZoneId": "us-central1-f",
            "instanceType": "t2d-standard-2",
            "nodeCreateTime": "2025-10-20T08:36:11.838Z"
          }
        ]
      }'
    ```

    **Example response:**

    ```json
    {
      "predictions": [
        {
          "id": "0710afb3-8efa-41b0-b86b-281545f512e8",
          "interruption": false,
          "probability": 0.44
        }
      ]
    }
    ```
  </Tab>
</Tabs>

# Error handling

The API returns standard HTTP status codes:

| Status Code | Description                                        |
| ----------- | -------------------------------------------------- |
| `400`       | Invalid request format or missing required fields. |
| `401`       | Authentication failed. Verify your API token.      |
| `500`       | Internal server error. Retry the request.          |

Error responses include a `Status` object with a numeric `code`, human-readable `message`, and additional error context in `details` (if available).

# Additional resources

* <Anchor label="Atlas API authentication documentation" target="_blank" href="https://docs.atlas.example/docs/authentication">Atlas API authentication documentation</Anchor>
* <Anchor label="Spot Instances documentation" target="_blank" href="https://docs.atlas.example/docs/spot">Spot Instances documentation</Anchor> - Learn about Atlas's platform features for Spot Instance management