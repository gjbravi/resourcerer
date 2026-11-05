# Workload recommendation statuses

> Readiness, error, and warning statuses for workload recommendations


You can view workload recommendation status details in these locations: 
- In workload lists on the Overview and Workloads pages
- On the Workloads Summary page: Click  **View Summary** on the Workloads page and drill down by status, error, or warning
- In the **Recommendation** section of the workload details page header
&nbsp;

|Icon |Status|Description|
|-----|------|----------------------|
|![](/img/docs/ol-status-wl-learning.png)| Learning | Indicates one of the following: **Preliminary recommendations are available** based on data collected so far; complete recommendations will be available when the workload's [learning period][learning-period] is complete. **Waiting for Data**: A workload (typically a CronJob) has not run yet and Orion is waiting for data. No action required. |
|![](/img/docs/ol-status-wl-ready.png)| Ready | A recommendation is available to apply to this workload and will be applied if auto-deployment is enabled; otherwise, you can it apply manually.|
|![](/img/docs/ol-status-wl-info.png)| Informational | Provides additional details about a recommendation, for example, the recommendation was constrained by bounds, HPA is at min or max replicas. See the banner on the workload page for details. No action required.|
|![](/img/docs/ol-status-wl-unsup.png)| Unsupported | Recommendations are not provided in these scenarios: corrupt usage data; `batch/v1` Jobs that run for less than 10 minutes; when an HPA oscillates between min and max replicas, scales on >1 resource or containerResource metric, or has a `target.averageUtilization` >=100. Typically, no action is required. For details see the [Unsupported workloads](reference__troubleshooting.md#unsup) section of the Troubleshooting topic.|
|<img src="/img/docs/warning_amber.png" width="32">| Warning | Warning types include scaled to zero, invalid workload annotation, high resource usage variance, and missing patch path configuration. For details and troubleshooting steps, see the [Errors and warnings](reference__troubleshooting.md#errors-warnings) section of the Troubleshooting topic. |
|![](/img/docs/ol-status-wl-error.png)| Error | Error types include recommendation failure, missing data, expired recommendations, and applier failure. For details and troubleshooting steps, see the [Errors and warnings](reference__troubleshooting.md#errors-warnings) section of the Troubleshooting topic. |

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/view/">View recommendations</a>
                </h5>
                <p>View recommendations on the Workload Details page or from the command line</p>
            </div>
        </div>


[troubleshooting]: https://kubernetes.io/docs/tasks/debug/debug-application/
[learning-period]: /docs/configure/settings/apply-settings/#learning-period
