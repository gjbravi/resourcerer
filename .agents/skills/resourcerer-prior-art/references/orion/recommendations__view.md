# View recommendations

> View recommendations on the Workload Details page or from the command line


A recommendation includes the CPU and memory requests and limits values (and recommended HPA configuration, if an HPA is running) that the machine learning has determined to be optimal for each container in a workload.

By default: 
- Recommendations are **not applied automatically**, giving you the opportunity to review their projected impact on your workloads. You can configure Orion to deploy recommendations automatically, as described later in this topic.
- Recommendations are applied as patches. You can choose to configure Orion to use a mutating admission webhook to apply recommendations, as described later in this topic.

## Viewing the overall projected impact of a recommendation

To understand how a recommendation will right-size a workload, review the workload details page header:  

[![Orion web application shows current and optimized request values](/img/docs/ol-rec-header.png)](/img/docs/ol-rec-header.png)

- **Optimization Score** measures how well your current request settings align with Orion's recommendations. A score of 100 indicates perfect alignment.

- **Recommendation Summary**
  - **Current Requests** shows the current requests multiplied by the average number of replicas over the last 7 days. 

  - **Recommended** shows the recommended requests multiplied by the average number of replicas over the last 7 days.

  - **Net Impact** shows the *difference* between the current and recommended total CPU and memory requests.

  **Note:** To change the CPU and memory cost/hr values used to calculate the estimated monthly costs, click **Settings** in the upper-right corner of the page.

## Viewing recommendation details in the Orion web application

To view the recommended CPU and memory settings for each container in a workload, navigate to a workload details page, click **Recommendation**, and then click **View** beside a container name in the Container Values section. 

The container-level charts show the projected impact of using the recommended settings, averaged across the time period shown on the graph. Hover at any point on the graph to see details for a point in time.

[![Orion web application shows current and optimized request values](/img/docs/ol-rec-cont.png)](/img/docs/ol-rec-cont.png)


## Viewing recommendations from the command line

Run the following command, replacing WORKLOAD_NAME with your workload name:
```sh
$ orion get recommendations WORKLOAD_NAME -o json
```

Here’s an excerpt from the recommendation: 
````json
"goal": "savings",
"values": {
  "containerResources": [
    {
      "containerName": "respod",
      "requests": {
        "cpu": "0",
        "memory": "1099366"
      }
    }
  ],
  "autoScaling": {
    "metrics": [
      {
        "resource": {
          "name": "cpu",
          "target": {
            "averageUtilization": 90
          }
        }
      }
    ]
  }
}
````

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Installation<span class="related-topics-divider">/</span><a href="/docs/installation/install-adv/">Advanced installation</a>
                </h5>
                <p>Installation behind a proxy server or from a private container registry</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/apply/">Apply recommendations</a>
                </h5>
                <p>Apply recommendations using the method that suits your environment</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/generate/">Generate recommendations</a>
                </h5>
                <p>Recommendations generated on demand, outside the regular schedule</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Reference<span class="related-topics-divider">/</span><a href="/docs/reference/statuses/">Workload recommendation statuses</a>
                </h5>
                <p>Readiness, error, and warning statuses for workload recommendations</p>
            </div>
        </div>


[Orion Applier]: /docs/installation/install/#orion-applier
[Auto-deploy]: /docs/configure/settings/apply-settings/#auto-deploy 
[deployment thresholds]: /docs/configure/settings/apply-settings/#thresholds
[recommendation schedule]: /docs/configure/settings/schedule/
[Continuous reconciliation]: /docs/installation/applier-config/#continuous-rec
[apply-method]: /docs/configure/settings/apply-settings/#apply-method 
[reliability]: /docs/configure/settings/reliability/
[orion-apply-ref]: /docs/orion-cli/reference/#orion-apply
[schedule]: /docs/configure/settings/schedule/#schedule
