# Best practices

> Tips to help you get the most from Orion


Get the most value out of Orion and minimize unexpected results by following these best practices.


<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">If you're getting started... </div>


Check out the [Getting started with configuring Orion](configure__configure-opt-settings.md) guide, which discusses the five most important settings and whether you need to consider adjusting them to suit your organization's requirements.
</div>


## Workload settings

### Generate recommendations once daily

A once-daily recommendation schedule is considered a best practice in stable or production environments (this is the default schedule when you roll out Orion). For details about this setting, see [Recommendation schedule][schedule] in the Schedule topic.

### Apply recommendations automatically

Before configuring Orion to automatically apply recommendations, you should review and selectively apply the first few recommendations to canary workloads to see what happens. Then, as you trust recommendations more, you can enable them to be applied automatically on more workloads.

The [Orion Applier][applier] is installed by default. When you're ready to have Orion apply scheduled recommendations automatically, configure Orion to [deploy recommendations automatically][auto-deploy].

[schedule]: /docs/configure/settings/schedule/
[applier]: /docs/installation/install/#orion-applier
[auto-deploy]: /docs/configure/settings/apply-settings/
