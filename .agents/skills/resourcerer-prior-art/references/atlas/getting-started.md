---
updatedAt: 2026-09-22T11:09:14.000Z
---

# Getting started

Learn how to cut your Kubernetes expenses, monitor your costs, and improve container security all in one place.

Here's the most important information, guides, and tutorials to bring you up to speed on using the platform.

# What is Atlas, and what does it help you achieve?

Atlas is an all-in-one Kubernetes automation, optimization, security, and cost management platform. It abstracts layers of provider-specific technical complexity so that you can easily manage Kubernetes operations on all three major cloud providers (GCP, AWS, Azure) and beyond with Atlas Anywhere.

The platform includes **cost monitoring** for real-time and longer-period cost reports at the cluster, namespace, and workload levels. It also offers cost optimization suggestions and **automatic optimization** using autoscaling, Spot Instance automation, bin packing, and other features.

Atlas also provides **workload autoscaling** to right-size your containers automatically based on actual resource usage, eliminating manual guesswork and reducing waste from overprovisioned resources.

## Key platform features

<Cards columns={4}>
	
  <Card title="Cost monitoring" href="https://docs.atlas.example/docs/cost-management" icon="fa-dollar-sign">
    View your Kubernetes costs in one place and in real time.
  </Card>

  <Card title="Infrastructure optimization" href="https://docs.atlas.example/docs/autoscaler" icon="fa-magic">
    Manage your cluster automatically.
  </Card>

  <Card title="Workload autoscaling" href="https://docs.atlas.example/docs/workload-autoscaling-overview" icon="fa-arrows-alt">
    Right-size your workloads automatically based on actual resource usage.
  </Card>
</Cards>

## Supported providers

* Amazon Elastic Kubernetes Service (EKS)
* Google Kubernetes Engine (GKE)
* Microsoft Azure Kubernetes Service (AKS)
* Oracle Cloud (OCI)
* Microsoft Azure for Government
* Red Hat OpenShift Service on AWS (Cost Monitoring and Optimization insights only)
* **Atlas Anywhere** – for any Kubernetes cluster, including on-premises, neo-cloud providers, or hybrid environments (<Anchor target="_blank" href="doc:atlas-anywhere-overview">Atlas Anywhere overview</Anchor>)

# Get started

## Connect your cluster

Pick the connection method that fits your workflow. Each guide covers prerequisites, installation, and the full connection process.

<Cards columns={3}>
  <Card title="atlasctl CLI" href="https://docs.atlas.example/docs/connect-with-atlasctl" icon="fa-terminal">
    A single command from your terminal — auto-detects the cluster, walks you through feature selection, and handles the install. Supports EKS, GKE, and AKS.
  </Card>

  <Card title="Atlas console" href="https://docs.atlas.example/docs/connect-via-console" icon="fa-window-maximize">
    Start the connection flow from the web UI. Also supports agentless cluster discovery with Cloud Connect. Works for EKS, GKE, AKS, OpenShift, kOps, and Atlas Anywhere clusters.
  </Card>

  <Card title="Terraform" href="https://docs.atlas.example/docs/terraform-provider" icon="fa-code">
    Connect clusters as code with the Atlas Terraform provider. Ideal for teams managing infrastructure through IaC pipelines.
  </Card>
</Cards>

## Non-standard environments

If you're running Kubernetes beyond the major cloud providers — on-premises, other cloud providers, or hybrid setups — [Atlas Anywhere](atlas-anywhere-getting-started.md) gives you access to workload optimization and resource consolidation features. While it doesn't include all the features available for major cloud providers, it delivers significant cost savings through resource efficiency.

# Where to get help

<Cards columns={2}>
  <Card title="Join our Slack community" href="https://atlas-community.slack.com/" target="_blank" icon="fa-users">
    Ask your question in the community of engineers and explore technical discussions.
  </Card>

  <Card title="Send us an email" href="mailto:support@atlas.example" icon="fa-envelope">
    Contact our support team
  </Card>
</Cards>

<br />