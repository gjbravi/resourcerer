# Try the Orion sandbox

> View-only Orion environment with no installation required


If you don't have a cluster of your own, you can check out Orion running in a view-only sandbox.

The sandbox contains two clusters, each running the same simulated e-commerce application.

This guide walks you through a comparison of optimized and non-optimized workloads, showing you how Orion machine learning rightsizes your workloads based on the metrics it collects.

## Log in to the sandbox

To preview Orion (with view-only privileges) without having to install anything, go to the [**sandbox**](https://app.orion.example/sandbox/).

## Applications in the sandbox

Each cluster is running a simulated e-commerce application that has several workloads. One Kubernetes workload is running for each service (shopping cart, ad service, and so on).

 The Orion Agent is installed on both clusters. All workloads in this estate are configured to have a recommendation generated once daily.
- In the `dev2-us-west2-optimized` cluster, recommendations are **applied automatically** as soon as they're generated.
- In the `dev1-us-east1-not-optimized` cluster, recommendations are **generated but not applied**.

Let's start with the Overview page.

 - **Note: The sandbox is a live environment, and its data is always changing**. Therefore, the data in the sandbox won't match what you see in this guide because **this guide captures a point in time**. Although the data differs, the principles are still the same.

## Get the estate-level overview

The Overview page shows an estate-level summary similar to the following:

[![Overview page showing the estate's optimization score, recommendation summary, recommendation impact, and waste and performance risk](/img/docs/sandbox-overview.png)](/img/docs/sandbox-overview.png)

- The **Optimization Score** measures how well your current request settings align with Orion's recommendations. A score of 100 indicates perfect alignment.
- The **Recommendation Summary** section shows the current and proposed rightsizing values for workloads across the estate (in this scenario, two clusters):
    - In the **Recommended** column, the requests value is calculated by aggregating per-container recommended request values multiplied by observed replica count, and averaging over the last 7 days.
- The **Recommendation Impact** shows the difference between current and recommended totals.

Let's break down the numbers.

At this point in time in the sandbox, across both clusters, you'd realize a net savings of $201.10 a month by rightsizing your workloads with the recommended settings. That might not sound like much, but imagine an estate with many thousands of workloads.

The **Waste** and **Performance Risk** cards break the net figure down. For CPU resources:
- Rightsizing overprovisioned workloads (**Waste**) will potentially save you 8.487 cores.
- Rightsizing underprovisioned workloads (**Performance Risk**) will consume an additional 0.401 cores.
- Therefore, the net impact of rightsizing CPU resources is a savings of 8.086 cores (8.487-0.401).

For memory resources:
- Rightsizing overprovisioned workloads will potentially save you 16.90 GiB, thus reducing costs.
- Rightsizing underprovisioned workloads will consume an additional 821 MiB.
- Therefore, the net impact of rightsizing memory requests is a savings of 16.09 GiB (approximately 16.90-0.80).

**Where do the dollar amounts come from?**\
They're calculated using the values on the Cost Estimates page. In the left navigation, click **Settings** > **Cost Estimates**. You can change them to be more reflective of your environment, or you can hide them if they're not important to you.

### Estate view: At-a-glance cluster, namespace, and workload health

The Top Namespaces and Top Clusters sections, and the workload lists linked from the **Waste** and **Performance Risk** cards, help you to understand where you'll get the greatest benefit from applying recommendations.

Typically, you'll just **scan these sections to decide where you need to investigate further**, perhaps due to traffic spikes, deployment of new applications, or if total requests somewhere are higher or lower than expected.

From the Overview page, you can either:
- **Drill down** to the next level of detail by clicking an item. Each section lists up to 5 items.
- **Skip directly** to an item type by using the left navigation.

This guide walks you through each section so that when you start using Orion in your environment, you'll know where to go based on what you want to see.

In the **Top Namespaces** and **Top Clusters** sections, the **Net Impact** column shows the estimated change in monthly cost from applying the recommended settings.

[![Top Namespaces and Top Clusters sections showing workload counts, optimization scores, and net cost impact](/img/docs/sandbox-top-clusters.png)](/img/docs/sandbox-top-clusters.png)

Notice how the `dev1-us-east1-not-optimized` cluster is listed first: Its overall optimization score is much lower (49%), and its net impact is far greater than that of the `dev2-us-west2-optimized` cluster. Optimizing the workloads in this cluster will have the greatest impact on savings.

The Top Namespaces section narrows this down. Most of the potential savings in `dev1-us-east1-not-optimized` come from its `microservices-demo-1` namespace, while the same namespace in `dev2-us-west2-optimized`, where recommendations are applied automatically, scores 99%.

In a real scenario, you might not know why there's a difference between current and recommended requests. **If the namespace and cluster views are too high-level, click into a namespace or cluster, or look at individual workloads.**

To see the workloads with the most waste, click **View Overprovisioned Workloads** on the **Waste** card. The **Workloads** page opens, filtered to overprovisioned workloads. Find and click the `sf-hipster-shop-lg-loadgenerator` workload in the `dev1-us-east1-not-optimized` cluster to view its details.

The workload details page opens on the **Summary** tab:

[![Workload details page, showing the Optimization Score, Potential Savings, and Recommendation cards and the Summary tab's requests table](/img/docs/sandbox-workload.png)](/img/docs/sandbox-workload.png)

- The **Attributes** card shows that this workload averages 8 replicas and has an HPA.
- The **Potential Savings** card shows the estimated cost savings and the percentage change in CPU and memory requests.
- In the table, the **Total Impact** column shows the difference between the current and recommended total requests, and the estimated change in cost.
- You already know that recommendations aren't applied: In the **Recommendation** card, **Automatic Deployment** is **Off**. Recommendations are generated once daily, but not applied.

##### View how recommendations perform over time
Now look at how the recommendations would perform over time in the **Usage vs. Requests** charts below the table. Select **Total Impact** to show totals across all replicas, starting with the **Total CPU** chart:

[![Total CPU chart, comparing usage with current and recommended total requests](/img/docs/sandbox-total-cpu.png)](/img/docs/sandbox-total-cpu.png)

Notice how much higher the current total requests (the purple line) are compared to the actual usage (the teal shaded area) and the recommended total requests (the orange line). The red **Waste** band shows wasted resources. **This workload is overprovisioned with respect to CPU resources, and applying recommendations would rightsize it.**

- **Tip:** To show or hide a series on the chart, click its name in the chart legend. For example, you might deselect **Waste** to reduce visual clutter.

Hover over the right end of the chart, which shows the latest recommendation. The **Recommended** value in the tooltip matches the recommended value in the table's **Total Requests** column: 1.752 cores in this example. Earlier points show previous daily recommendations, which is why the orange line steps up and down.

Similarly, in the **Total Memory** chart, the current requests are much higher than total usage, and the recommended requests are much closer to the actual usage. **Applying recommendations will rightsize this workload.**

Now let's look at a container in this workload.

## Reviewing the recommendation details: Container-level recommendations

Before we look at the recommendation details and the container's usage charts, let's look at how this workload and its container are configured. On the workload details page, click the **Settings** tab, and in the **CPU and Memory** section, review the container's optimization policy.

Remember, **a recommendation includes the proposed optimized requests and limits values for each container in a workload.** But the **_optimization policy_** for the container defines what is applied: requests only, requests and limits, or nothing.
- Note: If an HPA is present and scales on one resource metric (CPU or memory), the recommendation includes a target HPA value in order to preserve the scaling behavior after rightsizing the workload. This workload has an HPA, so check the **HPA** section of the **Settings** tab as well.

Go to the **Recommendation** tab, which shows the current and recommended requests and limits values for each container. This workload has one container, so its per-pod values match the **Per-Pod Requests** column on the **Summary** tab. Orion recommends the following adjustments to requests:

- CPU: Reduce requests from 250m to 219m
- Memory: Reduce requests from 700MiB to 74MiB

### View recommendation performance at the container level

Expand the container's row to view its CPU and memory usage history charts. In both, the recommended requests and limits sit much closer to actual usage than the current requests and limits. Hover over the right end of a chart, and the recommended values in the tooltip match the values in the table.

**The potential rightsizing you see in the charts should make a compelling argument for enabling recommendations on this workload.**

## Deploying recommendations automatically 

The [Orion Applier](concepts.md#applier) is installed as part of Orion and is ready to apply recommendations automatically.

When you're comfortable with the recommendations Orion generates, enable automatic deployment to keep resource settings current without manual review:

- Enable automatic deployment by setting the `live.orion.example/auto-deploy` annotation to `"Enabled"` (see [Auto-deploy]( /docs/configure/settings/apply-settings/) for details).

By default, **a new recommendation is applied once daily**. You can [change the schedule](configure__settings__schedule.md) based on how closely you want to track CPU and memory utilization and your tolerance for churn.

To reduce pod churn, you can define _thresholds_ to ensure only impactful recommendations are applied. Learn more in the [Auto-deploy](configure__settings__apply-settings.md#thresholds) topic.

## Configuring recommendations {#configure}

You don't need to configure much in Orion because the machine learning does the work. But you can adjust several optimization settings, including:
- **Recommendation schedule**: This defines how often Orion generates (and optionally applies) recommendations.
- **Optimization goals for CPU and memory**: Do you want more "aggressive" recommendations that optimize cost savings, less aggressive recommendations that ensure reliability, or a balance of the two?
- **Optimization policy**: For each container, you can choose what to optimize: requests, requests and limits, or neither.
- **HPA target utilization**: You can set this when an HPA is detected.
- **Container-level thresholds**: As mentioned above, thresholds ensure that only the most impactful recommendations are applied, which can help to reduce pod churn.

To learn more about setting these values at the cluster, namespace, and workload levels, as well as details about the settings themselves, see the topics under [Configure optimization settings](configure.md).

## Recap

The Orion UI gives you the flexibility to review the net impact of recommendations at the estate, cluster, namespace, workload, and container level.

Typically, **scanning the Overview page** gives you an idea of where to investigate further when you see an unusually large impact or unexpected results. You can **drill down to the level you want to view** and decide if it's necessary to change any of the [available optimization settings](configure__settings.md).

If you decide to change optimization settings at the [cluster](configure__cluster-defaults.md), [namespace](configure__annotate-namespaces.md), or [workload](configure__annotate-workloads.md) level, a best practice is to use annotations, however, you can also use the optional [Orion CLI](orion-cli.md).

**Key points to remember:**
- **Rightsizing is more than just cost savings: It's about making sure workloads run optimally.**
- **You'll get the most out of Orion when you let Orion apply recommendations automatically.**

By letting Orion machine learning rightsize the workloads in your estate, you'll see fast time to value and remove the toil of thinking about — and setting — the optimal Kubernetes requests values.

Installing Orion on your cluster takes less than one minute - and with a 30-day free trial, you can try it out for yourself. Sign up for free at **[app.orion.example/signup](https://app.orion.example/signup)**. To preview the installation steps, see [Install Orion][install].

[install]: /docs/installation/install/#install-orion
