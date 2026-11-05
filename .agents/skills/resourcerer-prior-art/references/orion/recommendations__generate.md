# Generate recommendations on demand

> Recommendations generated on demand, outside the regular schedule


This feature is helpful when you are experimenting with settings and want to see the results as you go.

1. In the left navigation, click **Workloads**, then find and click the name of the workload you want to work with.
1. On the workload details page, in the **Recommendation** section of the page header, click **Regenerate**.
1. Optional: To apply the recommendation, complete one of the following steps: 
    - Click **Apply Now**. 
    - Apply the recommendation as patch: On the **Patches** tab of the workload details page, download the patch and then copy and run the generated `kubectl apply` command.
    - Run `orion apply --name WORKLOAD_NAME` from the command line, replacing WORKLOAD_NAME with the name of the workload. (Note: The optional Orion CLI tool must be [installed][cli-installed].)

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
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/view/">View recommendations</a>
                </h5>
                <p>View recommendations on the Workload Details page or from the command line</p>
            </div>
        </div>


[cli-installed]: /docs/installation/install-adv/#install-the-orion-cli-tool
