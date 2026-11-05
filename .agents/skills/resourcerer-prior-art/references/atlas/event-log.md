---
updatedAt: 2026-10-27T09:40:57.000Z
---

# Event log

The Event log provides a comprehensive view of all Workload Autoscaler actions in your cluster. You can access it from the left-hand menu under the Workload Autoscaler category.

![](https://files.readme.io/3eb488d1e39eb60336fbf7571973f72ade6ad41d4b75a0806eeb3d8622aafba5-image.png)

> 📘 Ask AtlasPilot Assistant to explain a recommendation
>
> If you want to understand why a specific workload received a particular recommendation, including the events that shaped it (such as traffic surges or OOM kills), you can ask [AtlasPilot Assistant](atlaspilot-assistant.md) instead of correlating event-log entries manually. See [Understanding workload recommendations](atlaspilot-assistant.md#understanding-workload-recommendations).

# Understanding the Event log

The Event log captures all types of events that help you monitor and troubleshoot automated workload optimization on Atlas.

## Workload optimization events

1. **Recommendation generated events**: Records when the system generates new resource adjustment recommendations, including:
   * Immediate and deferred recommendation modes
   * CPU and memory request recommendations
   * Container-specific resource calculations

2. **Request updated events**: Track changes to resource requests, including:
   * When the change was applied
   * Which workload was affected
   * Previous and new request values

3. **Pod count recommendation updated events**: Records when horizontal scaling recommendations are updated, showing:
   * Current and recommended pod counts
   * Scaling decisions and thresholds

4. **Pods reverted events**: Track when pods are reverted to previous configurations, including:
   * Pod count changes
   * Reasons for reversion
   * User or system-initiated reversions

## Resource and performance events

5. **OOM kills events**: Out-of-memory incident tracking with:
   * Occurrence timestamp
   * Affected workload and containers
   * Memory usage data at time of termination

6. **Resource surge events**: Notifications when the platform detects unusual spikes in resource usage compared to established patterns, including:
   * CPU and memory surge detection
   * Container-specific surge information
   * Surge magnitude and duration

7. **CPU stall detected events**: Notifications when CPU contention exceeds configured thresholds, including:
   * Stall percentage per container
   * Number and percentage of affected pods
   * Threshold that was breached
   * Resulting recommendation adjustment

8. **Startup failure events**: Notifications when readiness or startup probe failures exceed the detection threshold, indicating the workload may not have enough resources to initialize. See [Startup failure detection](workload-autoscaling-overview.md#startup-failure-detection) for details.

## System and configuration events

9. **Configuration changed events**: Records of workload configuration changes, including:
   * Timestamp and affected workload
   * The user who made the change
   * Detailed before/after configuration comparisons
   * Changes to optimization settings, constraints, and scaling policies

10. **Failed Helm test events**: System alerts when Helm test validations fail, including:

* Test failure details and error messages
* Links to troubleshooting resources
* Component health status

## Scaling policy events

11. **Scaling policy created events**: Records when new scaling policies are created, including:

* Policy configuration details
* Target rules and workload assignments
* User who initiated the creation

12. **Scaling policy assigned events**: Track when workloads are assigned to scaling policies, including:
    * Number of workloads affected
    * Policy assignment details
    * User who made the assignment

13. **Scaling policy edited events**: Records modifications to existing scaling policies, including:
    * Configuration changes (optimization settings, overhead, percentiles)
    * Target rule modifications
    * Impact on assigned workloads

14. **Scaling policy deleted events**: Track when scaling policies are removed, including:
    * Policy details before deletion
    * Impact on previously assigned workloads
    * User who initiated the deletion

15. **Scaling policy order changed events**: Records when the priority order of scaling policies is modified, including:
    * Previous and new policy ordering
    * Number of policies affected
    * Impact on workload assignments based on rule evaluation priority

# Analyzing events

Each event offers two ways to analyze what happened.

## Context view

Click on an event in the list to view it in context on a timeline:

![](https://files.readme.io/082d0483d5769a32dfaf91a9e714369bf5acaadbb3993b2af9e458e6111ffca2-image.png)

You can adjust the history window to view the desired amount of context for the workload by adjusting the period in the top right:

* Last hour
* Last 24 hours
* Last 7 days

![](https://files.readme.io/12b43a9052abbd5d4677d7ec0ed8fe67111471e30d89a5bbf067a89f13d5bcd6-image.png)

## JSON view

You can examine the raw event data in JSON format for detailed troubleshooting. This provides complete transparency into the changes made and the reasoning behind them.

To enter the raw JSON view:

1. Expand the chosen event and click on the JSON icon in the top-right of the expanded cell:

   ![](https://files.readme.io/7fa32d44a0db303566389d5d278518c7ead31a3982043eaef9dc423684b47d99-image.png)

2. This will switch the event view to raw JSON that you can analyze or easily copy into an editor or tool of your choosing:

   ![](https://files.readme.io/b02e1e33104a4359ac2bafe25527115342fe74c4913f5ff00c5acc72f18cc2f4-image.png)

# Event filtering and search

The Event log provides filtering capabilities to help you focus on specific types of events or time periods:

* **Event type filters**: Filter by specific event categories (recommendations, configuration changes, scaling policy events, etc.)
* **Time range selection**: Focus on events within particular time windows
* **Workload filtering**: View events related to particular workloads or namespaces
* **User activity tracking**: Filter events by the user who initiated changes

![](https://files.readme.io/70bb15e78d3552aa3f037946f50b0a8d698b938af96573a8d5efdda404d0ff4a-image.png)

# Event relationships

Many events in the log are interconnected and represent different stages of the optimization process or changes made by the user:

* **Policy changes** often trigger **configuration changed events** for affected workloads
* **Resource surge events** typically lead to **recommendation generated events** with adjusted parameters
* **Scaling policy order changes** may result in **workload reassignment events** as assignment rules are re-evaluated
* **OOM kill events** automatically trigger **recommendation generated events** with increased memory allocations
* **CPU stall detected events** automatically trigger **recommendation generated events** with increased CPU allocations
* **Startup failure events** automatically trigger **recommendation generated events** with increased CPU and memory allocations