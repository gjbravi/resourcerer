---
updatedAt: 2026-09-04T09:32:05.000Z
---

# Cluster insights

Get an overview of your cluster's key data, health score, and resource status.

<Image align="center" src="https://files.readme.io/9e69f90-dash1.png" />

Atlas's dashboard provides a quick overview of the cluster's most current details and relevant information.

The report starts with **cluster details** such as its status, K8s provider, region, ID, and the time it was connected to Atlas.

The nodes section presents **details of your cluster's nodes**. It includes the total number of nodes split by resource offering (on-demand, spot, and fallback) and the tools used to manage them.

The **resources section** displays how many CPUs and memory your cluster currently provisions, requests, and allocates.

The dashboard provides **the total pod count**, including the number of scheduled and unscheduled ones.

<Image align="center" src="https://files.readme.io/da85fab-dash2.png" />

The bottom part of the report displays **resources provisioned and requested** in the past 24 hours, seven days, and 30 days, helping you assess the overall state of the cluster.

<Cards columns={3}>
  <Card title="Cluster score" href="https://docs.atlas.example/docs/cluster-score" icon="fa-gauge-high">
    Comprehensive rating of your cluster's optimization and efficiency.
  </Card>
  <Card title="Optimization constraints" href="https://docs.atlas.example/docs/optimization-constraints" icon="fa-list-check">
    Workload configuration checks surfaced as part of the Cluster Score.
  </Card>
  <Card title="Reliability metrics" href="https://docs.atlas.example/docs/reliability-metrics" icon="fa-heartbeat">
    Monitor request rate, error rate, and latency for your services.
  </Card>
</Cards>