# Orion docs

Local copy of the product docs, one file per page. Grep this directory; open a page by file name.


## Get started

- [Get started](get-started.md): Get up and running with Orion
- [Quickstart](get-started__quickstart.md): Get recommendations in under 10 minutes

## Installation

- [Installation](installation.md): Install Orion on your clusters
- [Basic installation](installation__install.md): Install Orion on a cluster
- [Advanced installation](installation__install-adv.md): Installation behind a proxy server or from a private container registry
- [Agent configuration](installation__cluster-agent-config.md): Configure the Orion Agent component and shared top-level parameters
- [Applier configuration](installation__applier-config.md): Configure the Orion Applier component
- [Forwarder configuration](installation__forwarder-config.md): Configure the Orion Forwarder component
- [Node Agent configuration](installation__node-agent-config.md): Configure the Orion Node Agent component
- [Webhook configuration](installation__webhook-config.md): Configure the webhook used by the Applier to apply recommendations at admission time
- [Cluster defaults templates](installation__cluster-defaults-templates.md): ConfigMap templates for cluster-level optimization defaults
- [Install Orion as an EKS add-on](installation__install-eks-addon.md): Installation through the Orion EKS add-on in AWS Marketplace
- [Install on Red Hat OpenShift Container Platform](installation__install-openshift.md): Install Orion on Red Hat OpenShift Container Platform
- [Upgrade](installation__upgrade.md): Upgrade Orion to the latest version
- [Uninstall](installation__uninstall.md): Uninstall Orion from a cluster
- [Migrate from v2 to v3](installation__migration.md): Migrate from separate Agent and Applier Helm charts to a unified Orion chart

## Configure

- [Configure optimization settings](configure.md): Define the optimization policy for requests and limits at the cluster, namespace, or workload level
- [Getting started with configuring Orion](configure__configure-opt-settings.md): The five most important optimization settings and a starter template
- [Configure optimization defaults for the cluster](configure__cluster-defaults.md): Configure default optimization settings for the cluster
- [Configure optimization defaults for namespaces](configure__annotate-namespaces.md): Default optimization settings for every workload in a namespace, set with annotations
- [Configure optimization for individual workloads](configure__annotate-workloads.md): Optimization settings for a single workload, set with annotations
- [Configure optimization using CRDs](configure__optimization-configurations.md): Specify optimization configurations at the cluster or namespace level using Custom Resource Definitions

## Settings

- [Optimization settings and descriptions](configure__settings.md): Orion optimization settings, by category
- [Annotations](configure__settings__annotations.md): All available annotations for Orion configuration
- [Optimization goal](configure__settings__goal.md): Savings, Balanced, or Reliability tradeoff for recommendations
- [Schedule](configure__settings__schedule.md): When Orion generates recommendations and how long they stay valid
- [Apply settings](configure__settings__apply-settings.md): Enable automatic deployment, configure the learning period, set optional thresholds to skip low-impact recommendations, and choose an apply method
- [Reliability response](configure__settings__reliability.md): Respond to OOM and CPU pressure events with a temporary resource bump-up
- [CPU and memory](configure__settings__cpu-and-memory.md): Optimization policy, limit-to-request ratio, and bounds for CPU and memory
- [HPA](configure__settings__hpa.md): Configure how Orion interacts with HPA-managed workloads
- [Runtime](configure__settings__runtime.md): Max heap size settings for Java workloads
- [Pod scheduling](configure__settings__pod-scheduling.md): Scheduling optimization policy for node affinity recommendations

## Concepts

- [Concepts](concepts.md): Clusters, workloads, components, schedules, and other Orion terminology
- [Orion architecture](concepts__architecture.md): Orion components that run in your cluster and in the Orion cloud

## Recommendations

- [Recommendations](recommendations.md): How Orion generates, delivers, and applies recommendations
- [Requests and limits](recommendations__requests-and-limits.md): Recommendations for CPU and memory resource settings
- [Java heap size](recommendations__java.md): Max heap size recommendations for Java workloads
- [Pod scheduling](recommendations__pod-scheduling.md): Automated node affinity for proportional resource scheduling
- [GPU optimization](recommendations__gpu.md): GPU visibility for Kubernetes workloads
- [Workload groups](recommendations__workload-groups.md): Aggregate metrics from related workloads to improve recommendations for blue/green, canary, and other split-traffic deployment patterns
- [Generate recommendations on demand](recommendations__generate.md): Recommendations generated on demand, outside the regular schedule
- [View recommendations](recommendations__view.md): View recommendations on the Workload Details page or from the command line
- [Apply recommendations](recommendations__apply.md): Apply recommendations using the method that suits your environment
- [Custom workload kinds](recommendations__custom-workload-kinds.md): Configure which Kubernetes kinds Orion recognizes as workloads, including custom resources and operator-owned workloads

## Guides

- [Guides](guides.md): How-to guides for Orion
- [Try a sample application on your own cluster](guides__showcase-app.md): Load-generating sample app for trying Orion on your own cluster
- [Connect AI agents to Orion with MCP](guides__mcp-server.md): Hosted and local MCP server setup for AI agents
- [Try the Orion sandbox](guides__sandbox.md): View-only Orion environment with no installation required
- [Manage workloads](guides__workload-management.md): Add, delete, and restore workloads, or exclude them from optimization
- [Monitoring the health of Orion](guides__monitoring-health-optlive.md): Monitor Orion health with status metrics, logs, and component alerts
- [Configure applier alerts](guides__applier-alerts.md): Send HTTP webhook notifications when the Orion Applier applies a recommendation to a workload
- [How to use Orion and Argo CD together](guides__argo-cd.md): Apply method and ignoreDifferences settings for using Orion with Argo CD
- [How to use Orion and Flux together](guides__flux.md): Use a webhook-based apply method, and configure drift detection ignore rules for HPAs
- [Orion Kustomize export](guides__kustomize-export.md): Automate committing patches to Git and deploying through CI
- [Import recommendation data into APM tools](guides__apm_export.md): Prometheus scrape job for importing recommendation data into APM tools
- [Kubernetes cost allocation](guides__kubernetes-cost-allocation.md): Orion sends container usage metrics to our Costs Platform to attribute Kubernetes costs to namespaces and workloads
- [Transparent proxy configuration](guides__proxy-config.md): Orion configuration for a transparent proxy
- [Manage secrets with a third-party credential manager](guides__manage-secrets.md): Orion Agent secrets managed by a third-party credential manager
- [Subscribe through AWS Marketplace](guides__aws_marketplace.md): Set up Orion pay-as-you-go billing in AWS Marketplace

## Reference

- [Reference](reference.md): Events, statuses, troubleshooting, and other Orion reference material
- [Best practices](reference__best-practices.md): Tips to help you get the most from Orion
- [User roles](reference__user-roles.md): The four Orion user roles and what each can do
- [Workload recommendation statuses](reference__statuses.md): Readiness, error, and warning statuses for workload recommendations
- [Applier Kubernetes events](reference__applier-events.md): Kubernetes Events emitted by the Applier and how to monitor them
- [Security FAQ](reference__faq-security.md): SOC 2 compliance, data collection and retention, SSO, and allowlist URLs
- [Troubleshooting](reference__troubleshooting.md): Get help with common questions, errors, and warnings

## CLI

- [Orion CLI](orion-cli.md): Install and authenticate the Orion CLI
- [Common Orion CLI workflows](orion-cli__common-cli-workflows.md): Common Orion CLI tasks, from authentication to managing recommendations
- [Reference](orion-cli__reference.md): Commands and options for the Orion CLI

## Optional

- [Releases](releases.md): Orion release history and updates
- [What's new in Orion](releases__whats-new.md): Monthly summaries of Orion updates
- [Orion Agent releases](releases__orion.md): Release history for the Orion Agent
- [Orion CLI releases](releases__orion-cli.md): Release history for the Orion CLI

## Other pages

- [API reference](api.md)
- [Account not activated](api__auth_not_activated.md)
- [You are now authenticated](api__auth_success.md)
- [Orion Documentation](home.md)
- [Preview](preview.md)
- [HPA reconciliation](preview__hpa-reconciliation.md)
- [Configure optimization using CRDs](preview__optimization-configurations.md)
- [Version 5.0.0](releases__orion-cli__v5.0.0.md)
- [Version 5.0.1](releases__orion-cli__v5.0.1.md)
- [Version 5.0.2](releases__orion-cli__v5.0.2.md)
- [Version 5.0.3](releases__orion-cli__v5.0.3.md)
- [Version 5.0.4](releases__orion-cli__v5.0.4.md)
- [Version 5.0.5](releases__orion-cli__v5.0.5.md)
- [Version 5.0.6](releases__orion-cli__v5.0.6.md)
- [Version 5.1.0](releases__orion-cli__v5.1.0.md)
- [Version 5.1.1](releases__orion-cli__v5.1.1.md)
- [Version 5.1.10](releases__orion-cli__v5.1.10.md)
- [Version 5.1.11](releases__orion-cli__v5.1.11.md)
- [Version 5.1.12](releases__orion-cli__v5.1.12.md)
- [Version 5.1.13](releases__orion-cli__v5.1.13.md)
- [Version 5.1.2](releases__orion-cli__v5.1.2.md)
- [Version 5.1.3](releases__orion-cli__v5.1.3.md)
- [Version 5.1.4](releases__orion-cli__v5.1.4.md)
- [Version 5.1.5](releases__orion-cli__v5.1.5.md)
- [Version 5.1.6](releases__orion-cli__v5.1.6.md)
- [Version 5.1.7](releases__orion-cli__v5.1.7.md)
- [Version 5.1.8](releases__orion-cli__v5.1.8.md)
- [Version 5.1.9](releases__orion-cli__v5.1.9.md)
- [Version 6.0.0](releases__orion-cli__v6.0.0.md)
- [Version 6.0.1](releases__orion-cli__v6.0.1.md)
- [Version 6.0.2](releases__orion-cli__v6.0.2.md)
- [Version 6.1.0](releases__orion-cli__v6.1.0.md)
- [Version 6.1.1](releases__orion-cli__v6.1.1.md)
- [Version 6.2.0](releases__orion-cli__v6.2.0.md)
- [Version 6.2.1](releases__orion-cli__v6.2.1.md)
- [Version 6.2.10](releases__orion-cli__v6.2.10.md)
- [Version 6.2.2](releases__orion-cli__v6.2.2.md)
- [Version 6.2.3](releases__orion-cli__v6.2.3.md)
- [Version 6.2.4](releases__orion-cli__v6.2.4.md)
- [Version 6.2.5](releases__orion-cli__v6.2.5.md)
- [Version 6.2.6](releases__orion-cli__v6.2.6.md)
- [Version 6.2.7](releases__orion-cli__v6.2.7.md)
- [Version 6.2.8](releases__orion-cli__v6.2.8.md)
- [Version 6.2.9](releases__orion-cli__v6.2.9.md)
- [Version 3.0.0](releases__orion__v3.0.0-agent.md)
- [Version 3.1.0](releases__orion__v3.1.0-agent.md)
- [Version 3.1.1](releases__orion__v3.1.1.md)
- [Version 3.1.2](releases__orion__v3.1.2.md)
- [Version 3.1.3](releases__orion__v3.1.3.md)
- [Version 3.2.0](releases__orion__v3.2.0.md)
- [Version 3.3.0](releases__orion__v3.3.0.md)
- [Version 3.3.1](releases__orion__v3.3.1.md)
- [Version 3.4.0](releases__orion__v3.4.0.md)
- [Version 3.5.0](releases__orion__v3.5.0.md)
- [February 2026](releases__whats-new__2026-02.md)
- [March 2026](releases__whats-new__2026-03.md)
- [May 2026](releases__whats-new__2026-05.md)
- [June 2026](releases__whats-new__2026-06.md)
- [July 2026](releases__whats-new__2026-07.md)
- [August 2026](releases__whats-new__2026-08.md)
- [September 2026](releases__whats-new__2026-09.md)
