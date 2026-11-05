---
updatedAt: 2026-09-29T08:33:20.000Z
---

# About the read-only agent

The Atlas read-only agent is a crucial component that connects your Kubernetes cluster to our platform, enabling cost monitoring features. This guide explains what the agent is, how it works, and why it's essential for reaping the benefits of Atlas.

## What is the Atlas agent?

The Atlas agent is a lightweight component that securely connects your Kubernetes cluster to our platform. Its primary functions are:

* Collecting essential cluster data for analysis
* Enabling real-time cost monitoring
* Facilitating automated optimization recommendations

Key features of the agent:

* **Read-only access**: The agent follows the principle of least privilege, ensuring it can't modify your cluster configuration without explicit permission.
* **Regularly updated**: Our team continuously improves the agent, releasing updates to enhance performance and security.
* **Easily removable**: You can uninstall the agent and remove all associated resources at any time.

## Agentless discovery

For AWS, GCP, and Azure customers, Atlas offers an agentless discovery option called **[Cloud Connect](cloud-connect.md)** as an initial step. This allows you to:

* Discover all your clusters without installing any components
* View basic cluster information before deciding to install the agent

## What data can the Atlas read-only agent access?

**The agent requires minimal cluster access to deliver meaningful insights.** Once connected, it gathers information on how much storage, memory, and CPU units your cluster needs to run efficiently.

**Here are the things the agent can access:**

* Main resources, such as Nodes, Pods, and Deployments, are required for running the Available Savings Report.
* Environment Variables: Pods, Deployments, StatefulSets, and DaemonSets.

## How Atlas handles sensitive data

Atlas is committed to maintaining the highest standards of data security:

* **No access to secrets**: The agent cannot access secrets, ConfigMaps, or sensitive environment variables.
* **Automatic filtering**: Before analysis, we automatically remove environment variables that may contain sensitive information (e.g., passwords, tokens, keys).
* **Encrypted transmission**: All data is transmitted using secure, encrypted connections.
* **Compliance certifications**: Atlas is ISO 27001-certified and holds SOC 2 Type II certification, demonstrating our commitment to robust security practices.

## How the agent removes sensitive environment variables

The agent scrubbing is designed to keep literal secret values from ever reaching Atlas.

**Where scrubbing happens**

All filtering runs in your cluster, inside the agent, before any data is sent to Atlas (pre-egress). The backend and the console only ever receive data that has already been scrubbed. Sensitive environment variable values never leave your cluster.

**What is treated as sensitive**

An environment variable is removed only when both of the following are true:

1. Its name matches a sensitive pattern, and
2. It holds a literal value.

The matching is case-insensitive and checks the variable name for any of these fragments:

```
passwd | pass | password | pwd | secret | token | key | creds | credential
```

A name matches if it contains any fragment. For example, `DB_PASSWORD`, `api_key`, `MY_SECRET_TOKEN`, and `GCP_CREDENTIALS` are all sensitive.

**What is removed**

* If an environment variable both matches the pattern and has a literal value, the whole key/value pair is removed. Neither the name nor the value is sent.
* An environment variable that matches the pattern but has no literal value is retained. This covers variables that reference a Secret instead of embedding the value, for example a `valueFrom` and `secretKeyRef` reference such as a `TOKEN` pointing to a Secret. In that case only the reference metadata is kept (the variable name plus the Secret and key it points to), never the resolved value.

So a literal `PASSWORD=hunter2` is dropped entirely, while a `TOKEN` that references a Secret through `secretKeyRef` is kept as a reference only.

**How to verify from the console**

Scrubbing happens before any data leaves the cluster, so workload and snapshot views in the console show only the environment variables that were retained. You will not see a removed variable marked as scrubbed. From the console, a scrubbed variable is indistinguishable from one that was never set.

**Example: canary check**

If you deploy a test workload with one literal sensitive-looking value and one plain value:

```yaml
env:
  - name: PASSWORD
    value: "fake-dont-use"
  - name: APP_COLOR
    value: "blue"
```

`PASSWORD` matches the pattern and has a literal value, so it is removed before sending and Atlas never receives it. `APP_COLOR` does not match, so it is retained. In the console, you would see only `APP_COLOR=blue` for that container, confirming that the value scrubbed as sensitive never left your cluster.

## How the agent works and how to use it

### Step 1: Connect your cluster to Atlas

First, you need to deploy the agent in your cluster. Start by connecting to the Atlas Console via HTTPS. This process uses Auth0.com as a secure authentication method and CloudFlare for DDoS protection.

The platform uses Identity Aware Proxy to establish a central authorization layer for all applications accessed by HTTPS. This allows you to use an application-level access control model instead of network-level firewalls.

Finally, (JWT) (JSON Web Token) enables an identity provider to pass the identities of authenticated users to Atlas.

#### Agent deployment

When you connect your cluster, the agent creates several Kubernetes resources:

* **Namespace**: `atlas-agent`
* Service account, roles, and role bindings
* **Deployments**: `atlas-agent` and `atlas-agent-cpvpa`

<Callout icon="💡" theme="default">
  ### Tip

  You can review the YAML file contents before executing the script to understand exactly what will be deployed in your cluster for full transparency.
</Callout>

This is the full list of resources that will be created when you connect the agent to your cluster for the first time:

* `namespace/atlas-agent `
* `serviceaccount/atlas-agent`
* `clusterrole.rbac.authorization.k8s.io/atlas-agent `
* `clusterrolebinding.rbac.authorization.k8s.io/atlas-agent `
* `role.rbac.authorization.k8s.io/atlas-agent `
* `rolebinding.rbac.authorization.k8s.io/atlas-agent `
* `secret/atlas-agent `
* `configmap/atlas-agent-autoscaler `
* `deployment.apps/atlas-agent `
* `deployment.apps/atlas-agent-cpvpa`

Only metadata about the node workload configuration from the Kubernetes scheduler is sent during the process. The agent also accesses YAML files from a Node Configuration called Snapshots.

### Step 2: Run the Available Savings Report

Once you connect your cluster to the Atlas Console, the platform can analyze your setup and provide insights – check the Available Saving report.

If you are happy with the recommendations, you can implement them automatically. Find out more about [Rebalancing](rebalancing.md).

### Step 3: View your cost monitoring and security insights

Once you connect your cluster, you also get free access to the cost monitoring feature and container security report. These tools provide valuable insights into your cluster's performance and security posture.

## Troubleshooting

### Cluster metadata configuration (EKS)

The atlas-agent requires the AWS account ID, cluster region, and cluster name to identify the EKS cluster. These are provided as environment variables (`EKS_ACCOUNT_ID`, `EKS_REGION`, `EKS_CLUSTER_NAME`) via the Helm chart's `additionalEnv` values. atlasctl and the Terraform module set these automatically.

#### Missing environment variables

If the env vars are missing — for example, a manual Helm install without them, or an upgrade that dropped them — the agent may log errors like:

```shell
EC2MetadataRequestError: failed to get EC2 instance identity document caused by: RequestError: send request failed caused by: Get "http://169.254.169.254/latest/dynamic/instance-identity/document": context deadline exceeded
```

This error indicates the agent is missing the `EKS_ACCOUNT_ID`, `EKS_REGION`, and `EKS_CLUSTER_NAME` environment variables.

The fix is to ensure the env vars are set in the Helm release:

```shell
helm upgrade --install atlas-agent atlas-helm/atlas-agent \
  --namespace atlas-agent --create-namespace \
  --set provider=eks \
  --set additionalEnv.EKS_ACCOUNT_ID="<your-aws-account-id>" \
  --set additionalEnv.EKS_REGION="<your-eks-region>" \
  --set additionalEnv.EKS_CLUSTER_NAME="<your-cluster-name>" \
  --set apiKey="<your-atlas-api-key>"
```

For more troubleshooting guidance, see [Atlas components troubleshooting](hosted-components.md).

#### Temporary fix for a running deployment

If you have an already-running deployment that is missing the env vars and need an immediate fix before running `helm upgrade`, you can patch the deployment directly:

> ⚠️ Temporary measure
>
> This `kubectl patch` is a stopgap. The proper fix is a `helm upgrade` with the `additionalEnv` values shown above, so the env vars persist across future Helm upgrades.

```shell
kubectl patch deployment atlas-agent -n atlas-agent --type=strategic -p='{
  "spec": {
    "template": {
      "spec": {
        "containers": [
          {
            "name": "agent",
            "env": [
              {"name": "EKS_ACCOUNT_ID", "value": "<your-aws-account-id>"},
              {"name": "EKS_REGION", "value": "<your-eks-region>"},
              {"name": "EKS_CLUSTER_NAME", "value": "<your-cluster-name>"}
            ]
          }
        ]
      }
    }
  }
}'
```

## Frequently asked questions

<Accordion title="Can Atlas components be installed in a namespace other than `atlas-agent`?" icon="fa-info-circle">
  At this time, only the `atlas-agent` namespace is supported.
</Accordion>

<Accordion title="How often does the API key for the Atlas agent change?" icon="fa-info-circle">
  During the onboarding phase, the agent's API key is generated and set once for a cluster. Rerunning atlasctl will generate a new key, and pods for existing Atlas components may need to be deleted if 401 authorization errors prevent them from running.

  If you're using Terraform, you may need to invalidate the existing token and reapply your Terraform to get the Atlas components working again.
</Accordion>

<Accordion title="Do the agent and Helm chart need to be updated when a new release is available?" icon="fa-info-circle">
  Yes, the agent and Helm chart must both be updated accordingly. For upgrade instructions, see [Atlas hosted components](hosted-components.md).
</Accordion>

<Accordion title="Does the Atlas read-only agent interact with eBPF?" icon="fa-info-circle">
  The Atlas read-only agent doesn't interact with eBPF in any way. It only collects data using Kubernetes informers.
</Accordion>

<Accordion title="Will deploying the agent via the Marketplace remove existing configurations?" icon="fa-info-circle">
  No, our team has conducted thorough testing, and reconfiguring the clusters is not required. Agent deployment via the Marketplace will not affect your existing cluster configurations.
</Accordion>