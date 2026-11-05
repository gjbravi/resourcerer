# Manage workloads

> Add, delete, and restore workloads, or exclude them from optimization


**Topics:**

- [Adding workloads](#add-workloads)
- [Deleting workloads](#delete-workloads)
    - [Restoring workloads](#restore-workloads)
- [Excluding a workload from optimization](#stop-optimization)

## Adding workloads {#add-workloads}

Orion automatically detects when the following types are added to a cluster:
- Any Kubernetes Workload object (CronJob, DaemonSet, Deployment, ReplicaSet, StatefulSet) and HPA
- Workloads scaled by a KEDA-owned HPA (ScaledObjects in `keda.sh` apiGroup)

When Orion detects a new workload:
- Within about 2 minutes, Orion starts generating preliminary recommendations for the workload and updating the UI with workload and recommendation details. Preliminary recommendations cannot be applied automatically, but can be applied on demand.
 Orion generates these preliminary recommendations for the duration of the workload's learning period.
- When a workload's learning period is complete, Orion generates recommendations that can be deployed either automatically or on demand.

Related information: 
- [Learning period][learning-period]
- [Deploying recommendations automatically][auto-deploy] 

## Deleting workloads {#delete-workloads}

After a workload is deleted (for example, manually or as part of a deployment process), it's considered _soft deleted_ for a period of 24 hours. The workload and its data are removed from the UI and can be accessed only by using the CLI. 

Within this period, the workload can be [restored](#restore-workloads) (re-created) if needed. 

<div class="alert alert-warning" role="alert"><div class="h4 alert-heading" role="heading">Warning:</div>


Orion _permanently_ deletes a workload, its data, and its recommendations at the end of the 24-hour soft-deletion period. These items cannot be restored.
</div>


After a workload is permanently deleted, creating a workload with the same name results in a net new workload with a new creation timestamp and new learning period, as described in [Adding workloads](#add-workloads) above.
 
### Restoring workloads {#restore-workloads}

If the workload is re-created (restored) within the 24-hour soft-deletion period:
- The workload's original creation date, most recent settings, and previous recommendations are restored. 
- The charts on the **Summary** and **Recommendation** tabs of a workload details page show a corresponding gap in usage and recommendations. 

Related information: 
- [Learning period][learning-period]
- [Deploying recommendations automatically][auto-deploy] 


## Excluding a workload from optimization (without deleting it) {#stop-optimization}

To exclude a workload from optimization without deleting it, use one of the methods below.

**Note**: To exclude entire _namespaces_ from optimization, use a namespace allow list or deny list. See the related link below.

- Set the workload's auto-deploy setting to `Disabled`, as shown below.\
Orion generates a recommendation for the workload but does not deploy it. However, you can apply it on demand.
    ```yaml
    live.orion.example/auto-deploy: "Disabled"
    ```

- Set the workload's schedule to `@never`, as shown below.\
Orion will not generate recommendations automatically for the workload, but you can still generate and apply them on demand.
    ```yaml
    live.orion.example/schedule: "@never"
    ```
- Set the workload's optimization policies to `DoNotOptimize`, as shown below.\
Orion generates recommendations for the workload according to the configured schedule, but the recommendations do not suggest any changes; and, when patches are generated, the patches do not include the workload.

    ```yaml
    ... 
    live.orion.example/containers.cpu.optimization-policy: "DoNotOptimize"
    live.orion.example/containers.memory.optimization-policy: "DoNotOptimize"
    ...
    ```
    **Note:** You can also set _container-level_ optimization policies: for example, you can set  `DoNotOptimize` on injected containers. For details, see the Optimization policy topic.

Related information: 
- [Schedule setting][schedule]
- [Learning period][learning-period]
- [Optimization policy][opt-policy] 
- [Generating recommendations on demand][generate]
- [Namespace allow and deny lists][namespace-lists]

[learning-period]: /docs/configure/settings/apply-settings/#learning-period
[auto-deploy]: /docs/configure/settings/apply-settings/#auto-deploy
[schedule]: /docs/configure/settings/schedule/#schedule
[opt-policy]: /docs/configure/settings/cpu-and-memory/#opt-policy
[generate]: /docs/recommendations/generate/
[namespace-lists]: /docs/installation/cluster-agent-config/#namespace-lists
