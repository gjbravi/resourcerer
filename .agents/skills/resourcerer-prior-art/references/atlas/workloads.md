---
updatedAt: 2026-08-26T11:20:13.000Z
---

# Workload costs

Learn how to use the Workloads report to analyze cluster costs, efficiency, workload grouping, and troubleshooting.

<Image align="center" src="https://files.readme.io/d889582-Screenshot_2023-09-19_at_12.19.06.png" />

The Workloads report allows you to see your cluster costs broken down by workload. You can use this data to analyze compute expenses per workload or selected group of workloads, identify inefficiencies, and discover potential savings opportunities.

The Workloads report provides four different views:

* List of workloads with cost details per selected period
* List of workloads with efficiency details
* Individual workload cost details with daily history
* Individual workload efficiency details with daily history

## List of workloads with cost details per selected period

<Image align="center" src="https://files.readme.io/22b166a-Screenshot_2023-09-19_at_12.19.51.png" />

This is a list of all workloads in your cluster, together with their cost details shown within the period you chose.

Each entry contains the following:

* workload name
* workload controller type
* namespace
* average number of pods
* average requested CPU per hour
* average requested memory per hour
* total cost of CPU
* total cost of memory
* the total cost of compute

You can filter workloads by their labels and namespaces.

<Image align="center" src="https://files.readme.io/43ae196-Screenshot_2023-09-19_at_12.20.27.png" />

To see the total cost of multiple workloads, select the relevant workloads by ticking the box on the left side of the table. You’ll get the total cost data for this group of workloads at the bottom.

<Image align="center" src="https://files.readme.io/25874eb-Screenshot_2023-09-19_at_12.21.39.png" />

## List of workloads with efficiency details

<Image align="center" src="https://files.readme.io/bcf5804-image2.png" />

This is a list of all cluster workloads with their efficiency details per selected period.\
Each entry contains the following:

* workload name
* workload controller type
* namespace
* CPU hours wasted
* memory hours wasted
* `$ wasted` (the amount of money wasted, in US dollars)

<Image align="center" src="https://files.readme.io/493ef56-image3.png" />

Each entry also contains the requested and used resource hours.

The metric we use here is resource hours, which corresponds to resources multiplied by hours of usage.

For example, if a workload with requests set to 2 CPUs runs for 48 hours during the selected period, its total requested CPU hours would be 96 CPU hours. If a workload's average CPU usage is 0.5 CPU during those 48 hours, its total usage is 24 CPU hours.

<Image align="center" src="https://files.readme.io/f03e689-image1.png" />

Atlas managed mode customers can also use a quick recommendations patch that they can apply to their workload and change workload resource requests in line with our recommendations.

**How are wasted CPUs, memory, and money calculated?**

To calculate the number of wasted CPU and memory resources, the workload’s used resources are subtracted from the total number of requested resources.

To follow the example above:

```text
96 CPU hours - 24 CPU hours = 72 CPU hours
```

> 📘 Note
>
> For better recommendations and information on efficiency, set the requests based on the workloads within your cluster.

## Individual workload cost details with daily history

<Image align="center" src="https://files.readme.io/bc8b73e-Screenshot_2023-09-19_at_12.22.44.png" />

This report provides cost details and the history of a single workload within the selected period.

You get the following data:

* total spend
* current month forecast
* average daily cost
* average daily cost per resource (CPU and memory)

## Daily compute spend

You can also check the daily chart for compute spending per resource and per resource offering.

<Image align="center" src="https://files.readme.io/92d0c85-Screenshot_2023-09-19_at_12.23.09.png" />

## Daily cost history table

Daily cost history table with average pod count, cost per pod, amount of requested resources, cost per resource, and total cost.

<Image align="center" src="https://files.readme.io/6250bc0-Screenshot_2023-09-19_at_12.23.47.png" />

## Individual workload efficiency details with daily history

This report provides efficiency details for a single workload: both current and within the selected period.

![](https://files.readme.io/0b98005-image.png)

In the top section, the report provides efficiency details for a workload:

* Efficiency score in %
* Wasted resource hours both for CPU and memory
* $ wasted in US dollars
* Pods currently running

## Current workload efficiency

The current efficiency part provides current efficiency details per container:

* resource requests
* resource usage
* rightsizing recommendation
* computed overall efficiency

  ![](https://files.readme.io/b903ebc-image.png)

Computed efficiency is calculated by comparing resource requests against the recommended rightsized resource values. CPU is more expensive and has a larger impact on efficiency ratings. Learn more about [how CPU and memory costs are split](#how-cpu-and-memory-costs-are-split).

For rightsizing recommendations, the Cost monitoring module analyzes the resource use of a container during the last 5 days and calculates the percentile value (95th percentile for CPU and 99th percentile for memory).

> 📘 Note
>
> Usage metrics are based on [metrics-server](https://github.com/kubernetes-sigs/metrics-server) which in some rare cases could lead to incorrect CPU usage data due to Kernel bug. In case you experience such issue please consider upgrading metrics-server to [v0.7.0](https://github.com/kubernetes-sigs/metrics-server/releases/tag/v0.7.0) or above.

## Workload grouping

Each line item in the workload report corresponds to a distinct workload running within the cluster. Typically, these workloads are represented by top-level Kubernetes objects, such as Deployments, DaemonSets, StatefulSets, etc.

Atlas performs an analysis of all the pods and their associated controllers, identifying the top-level controller for each pod. It then aggregates all pods under their respective top-level controllers into a single entry in the workload report. This approach ensures a clear and organized representation of cluster workloads.

Atlas supports default and well-known Kubernetes controllers. In instances where pods don't have an associated controller or if the controller is not recognized, it categorizes these pods under separate entries. This ensures comprehensive coverage and visibility of all Pods within the cluster, regardless of their controller status.

## Issues with dynamic pods

In instances where clusters utilize unsupported controllers or frameworks that dynamically create pods for batch jobs or other operations, the default grouping mechanism will not work. This can lead to a significant number of separate entries in the workload report, each corresponding to an individual pod.

These short-lived pods will lack historical data and efficiency calculations, making it challenging to gauge their performance and cost impact accurately.

### Add a label for custom workload grouping

To group pods into a single workload entry, add the `workloads.atlas.example/custom-workload` label to your pods. All pods sharing the same label value will be aggregated under a single workload row in the report.

This is the same label used by [Workload Autoscaling](workload-autoscaling-configuration.md#custom-workload-support) for workload identification.

```yaml
metadata:
  labels:
    workloads.atlas.example/custom-workload: "my-batch-job"
```

All pods with `workloads.atlas.example/custom-workload: "my-batch-job"` will appear as a single `my-batch-job` entry in the Workloads report.

## Troubleshooting

<Accordion title="Cannot access the Workloads Efficiency tab">
  It could be the case that the Atlas agent cannot discover and poll metrics from the Metrics Server.

  ![](https://files.readme.io/7099f13-Pasted_Graphic_3.png)

  Validate whether the Metrics Server is running and is accessible by running the following commands:

  ```shell
  kubectl get deploy,svc -n kube-system | egrep metrics-server
  ```

  If Metrics Server is installed, the output is similar to the following example:

  ```shell
  deployment.extensions/metrics-server   1/1     1            1           3d4h
  service/metrics-server   ClusterIP   198.51.100.0   <none>        443/TCP         3d4h
  ```

  If the Metrics Server is not running, follow the installation process [here](https://github.com/kubernetes-sigs/metrics-server?org=03d99e73-1c6a-4a1d-b87d-24f67b4e8102#installation).

  If the Metrics Server is running, verify that the Metrics Server is returning data for all pods by issuing the following command:

  ```shell
  kubectl get --raw “/apis/metrics.k8s.io/v1beta1/pods”
  ```

  The output should be similar to the one below:

  ```json
  {“kind”:”PodMetricsList”,”apiVersion”:”metrics.k8s.io/v1beta1”,”metadata”:{},
  “items”:[
    {“metadata”:{“name”:”atlas-agent-84d48f88c9-44tr8”,”namespace”:”atlas-agent”,”creationTimestamp”:”2023-03-15T12:28:12Z”,”labels”:{“app.kubernetes.io/instance”:”atlas-agent”,”app.kubernetes.io/name”:”atlas-agent”,”pod-template-hash”:”84d48f88c9”}},”timestamp”:”2023-03-15T12:27:36Z”,”window”:”27s”,”containers”:[{“name”:”agent”,”usage”:{“cpu”:”5619204n”,”memory”:”75820Ki”}}]}
  ]}
  ```

  If no output or erroneous output is returned, review the configurations of your Metrics Server and/or reinstall it.

  If those checks passed, look for any unavailable APIs with the following command:

  ```shell
  kubectl get apiservices
  ```

  All API services must show `True` in the available column. Any that show `False` will need to be repaired if it's still in use by one of your workloads. If it is not still in use, the API service can be deleted.

  If everything looks good, but you still cannot access the Workloads Efficiency tab, please contact our support.
</Accordion>

## How CPU and memory costs are split

<Accordion title="CPU vs. memory cost calculation" icon="fa-calculator">

Google Cloud Platform (GCP) is the only major cloud provider that divides compute pricing into separate CPU and memory costs. AWS and Azure do not provide this breakdown natively.

Atlas uses a formula derived from GCP's public pricing to calculate separate CPU and memory costs for AWS and Azure instances. This gives you a more precise cost breakdown across workloads, namespaces, and allocation groups — which is especially important if you run memory-optimized virtual machines.

### The 88:12 ratio

By comparing general-purpose and memory-optimized instance families across GCP and AWS, Atlas determined that CPU accounts for approximately **88%** and memory for approximately **12%** of the total compute cost. This ratio is consistent across on-demand and spot pricing tiers.

### The formula

```text
instancePricePerHour / ((cpu_count × cpu_ratio) + (ram_gib × ram_ratio)) × resource_ratio
```

For example, an AWS r5.large instance ($0.126/hr, 2 vCPUs, 16 GiB RAM):

```text
CPU: 0.126 / ((2 × 88) + (16 × 12)) × 88 = $0.0301/vCPU-hour
RAM: 0.126 / ((2 × 88) + (16 × 12)) × 12 = $0.0041/GiB-hour
```

> 📘 Custom pricing
>
> By default, cost reports use public cloud list prices. If your organization has negotiated rates (CUDs, Savings Plans, EDP discounts, or private pricing), configure [Price adjustments](discount-engine-overview.md) to align cost data with your actual cloud spend.

### Where to check in the dashboard

To see the CPU vs. memory cost breakdown, go to the cost monitoring module and open the cluster or workloads report. The workloads report shows average requests and costs per resource type.

![](https://files.readme.io/9d82d0b-image1.png "image1.png")

</Accordion>

## Include Idle Cost

By default, the Workloads report shows only what your workloads **request**. Any CPU or memory on a node that no workload requested is idle capacity. That idle cost stays at the cluster level and is not added to workload costs.

When you enable **Include Idle Cost**, Atlas splits that idle cost among the workloads on each node, based on how much each workload requested.

### How it works

1. Atlas finds the idle resources on a node:

   ```text
   idle resources = provisioned resources - requested resources
   ```

2. It splits the idle cost among workloads, using each workload’s share of the total requested resources:

   ```text
   workload idle cost share = idle cost × (workload requested resources ÷ total requested resources on the node)
   ```

3. It adds that share to the workload’s requested cost:

   ```text
   workload total cost = workload requested cost + workload idle cost share
   ```

### Example

A node has **4 CPUs** and costs **$100**. Two pods run on it:

* **Pod A** requests **1 CPU** → costs **$25**
* **Pod B** requests **2 CPUs** → costs **$50**
* **1 CPU** is idle → **$25** of unused cost

The pods requested **3 CPUs** in total. Atlas splits the idle $25 by each pod’s share:

* **Pod A** gets 1/3 of the idle cost → **$8.33**
* **Pod B** gets 2/3 of the idle cost → **$16.67**

With **Include Idle Cost** enabled:

```text
Workload A: $25 + $8.33 = $33.33
Workload B: $50 + $16.67 = $66.67
```

> 📘 Note
>
> Atlas calculates Include Idle Cost separately for CPU and memory.

<Cards columns={2}>
  <Card title="Namespace costs" href="https://docs.atlas.example/docs/namespaces" icon="fa-folder-tree">
    Break down cluster costs by namespace.
  </Card>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Overview of total compute spend for your cluster.
  </Card>
</Cards>