# Atlas docs

Local copy of the product docs, one file per page. Grep this directory; open a page by file name.


## Introducing Atlas

- [Getting started](getting-started.md): Learn how to cut your Kubernetes expenses, monitor your costs, and improve container security all in one place.
  - [About the read-only agent](about-the-read-only-agent.md)
- [Connecting your cluster](connecting-your-cluster.md): Choose how to connect your Kubernetes cluster to Atlas: using atlasctl, the web console, or Terraform.
  - [Connect using the atlasctl CLI](connect-with-atlasctl.md): Connect or disconnect a Kubernetes cluster from Atlas using the atlasctl command-line tool.
  - [Connect using the Atlas console](connect-via-console.md): Connect your Kubernetes cluster to Atlas from the web console, or discover all clusters across your cloud account with Cloud Connect.
  - [Connect clusters with separate Kubernetes and cloud owners](connect-separate-kubernetes-cloud-owners.md): Connect a cluster when the Kubernetes team and the cloud team have separate access: the Kubernetes team runs `atlasctl cluster connect --k8s-only` first, then the cloud team registers cloud…
  - [Take a snapshot of your AWS account](cloud-snapshot.md): Run a one-time, agentless collection of AWS reliability data (EKS control plane, EC2, and ALB metrics, plus EKS, EC2, and EBS describe data) and upload it to Atlas without installing anything in…
  - [Cloud Connect](cloud-connect.md): Discover and inventory cloud resources across your entire cloud account without installing any agents.
  - [GCP Private Service Connect](gcp-private-service-connect.md)
  - [AWS PrivateLink](aws-privatelink.md)
  - [Legacy onboarding script reference](legacy-onboarding-script.md)
  - [Node Autoscaler preparation checklist](autoscaler-checklist.md)
  - [Troubleshooting cluster onboarding](cluster-onboarding-troubleshooting.md): Solutions for resolving issues when connecting clusters to Atlas, including connectivity problems, authentication errors, and cloud provider-specific onboarding issues.
- [Platform permissions & data privacy](platform-permissions-and-data-privacy.md)
  - [Kubernetes permissions](kubernetes-permissions.md): Kubernetes Service Accounts and permissions used by Atlas components.
  - [Cloud permissions](cloud-permissions.md): Permissions used by Cloud Providers (AWS/GCP/Azure/Oracle)
    - [AKS workload identity impersonation](sa-impersonation-aks.md)
    - [GKE service account impersonation](sa-impersonation-gke.md)
  - [Data collection and storage](data-collection-and-storage.md): Data collection and storage practices at Atlas.
  - [Communication requirements](communication-requirements.md): Traffic and communication requirements of Atlas.
  - [Overview](atlas-anywhere-overview.md)
  - [Getting started](atlas-anywhere-getting-started.md)
- [API access](api-access.md)
  - [Hosted components](hosted-components.md): Atlas components installed in your Kubernetes cluster and how they map to the umbrella chart operating modes.
    - [Cluster controller](cluster-controller.md)
    - [Spot Handler](spot-handler.md)
  - [Helm charts](helm-charts.md): Install and manage Atlas components in your Kubernetes cluster using the umbrella Helm chart or individual charts.
  - [Terraform provider](terraform-provider.md): Register your cluster and provision cloud IAM resources with the Atlas Terraform provider, then install Atlas components using the Helm chart.
    - [GKE via GitOps](terraform-provider-gke.md): Onboard a GKE cluster to Atlas using the umbrella Helm chart and Terraform. Start in read-only mode with Helm alone and upgrade to node autoscaling when ready.
    - [EKS via GitOps](terraform-provider-eks.md): Onboard an EKS cluster to Atlas using the umbrella Helm chart and Terraform. Choose the operating mode that fits your needs and install directly into it.
    - [AKS via GitOps](terraform-provider-aks.md): Onboard an AKS cluster to Atlas using the umbrella Helm chart and Terraform. Choose the operating mode that fits your needs and install directly into it.
    - [Terraform troubleshooting](terraform-troubleshooting.md): Solutions for common Terraform issues when setting up and managing Kubernetes clusters with Atlas.
    - [Import and apply cluster configuration with atlasctl](atlasctl-import-cluster-configuration.md): Discover existing node pool configurations from your cloud provider, then import or apply them into Atlas as node configurations and node templates using atlasctl.
    - [Migrate standalone Helm charts with atlasctl](atlasctl-migrate-standalone-helm-charts.md): Consolidate individually-installed Atlas Helm releases into the unified umbrella chart using atlasctl cluster migrate.
    - [Migrate GitOps and IaC installations with atlasctl](atlasctl-migrate-gitops.md): Move ArgoCD-, Flux-, or Terraform-managed Atlas standalone releases to the unified umbrella chart.
    - [atlasctl command reference](atlasctl-reference.md): Every atlasctl command and flag, in one place, for connecting and disconnecting clusters, migrating standalone charts to the umbrella chart, importing existing node configuration, and connecting AI…
  - [Component control](component-control.md)
  - [Atlas Operator](atlas-operator.md): Automated lifecycle management for Atlas components
  - [Open source components](open-source.md)
    - [Audit log exporter](audit-log-exporter.md)
  - [Troubleshooting Atlas components](component-troubleshooting.md): Solutions for resolving issues with Atlas agent, cluster controller, and other platform components.
- [Disconnect your cluster](disconnect-your-cluster.md): Learn how to disconnect your cluster from Atlas and remove all related data.

## AtlasPilot

- [AtlasPilot Assistant](atlaspilot-assistant.md): An AI-powered assistant that answers operational questions about your Kubernetes clusters, costs, and database optimization in plain English.
- [Connect AI tools to Atlas MCP](atlaspilot-mcp.md): Connect your AI tools to Atlas MCP to ask questions about your clusters, costs, efficiency, and health.

## Karpenter Enterprise suite

- [Overview](karpenter-enterprise.md)
- [Getting started](karpenter-enterprise-getting-started.md)
- [Feature reference](karpenter-enterprise-features.md)
- [Keeper](karpenter-enterprise-keeper.md)
- [Commitments in Karpenter Enterprise](karpenter-enterprise-commitments.md)
- [Scheduled rebalancing for Karpenter clusters](karpenter-enterprise-scheduled-rebalancing.md)
- [Continuous rebalancing](karpenter-enterprise-continuous-rebalancing.md)

## Application Performance Automation (APA)

- [Overview ](apa-overview.md)
- [Getting started](apa-getting-started.md): Enable Application Performance Automation and configure the required integrations.
  - [Fix container image vulnerabilities](apa-runbook-fix-image-vulnerabilities.md): Automatically remediate vulnerabilities in container images and generate pull requests for review.
  - [Synchronize Workload Autoscaler recommendations](apa-runbook-sync-workload-autoscaler-recommendations.md): Keep your Git manifests aligned with Workload Autoscaler recommendations to prevent configuration drift.

## Node Autoscaling

- [Autoscaling](autoscaler.md)
  - [Node templates](node-templates.md)
  - [Node configuration](node-configuration.md)
  - [Spot Instances](spot.md)
    - [Spot interruption prediction API](ml-spot-interruption-prediction-api.md)
  - [GPU Instances](gpu.md): Configure Atlas Autoscaler to scale your cluster using GPU-optimized instances across AWS EKS, GCP GKE, and Azure AKS with NVIDIA GPU support.
  - [Storage-optimized nodes](storage-optimized-nodes.md): Use storage-optimized nodes with locally attached NVMe/SSD disks for workloads that need high-throughput, low-latency local storage.
  - [TPU Instances (GKE)](tpu.md): Configure Atlas Autoscaler to scale your GKE cluster using TPU-optimized instances with Google Cloud TPU v5 support.
  - [AWS Neuron Instances (EKS)](neuron.md): Configure Atlas Autoscaler to scale your EKS cluster using AWS Neuron accelerator instances — including AWS Inferentia and AWS Trainium — for machine learning inference and training workloads.
  - [GPU sharing](gpu-sharing.md)
    - [Time-slicing](gpu-sharing-time-slicing.md)
    - [Multi-Instance GPU (MIG)](gpu-sharing-mig.md)
    - [Multi-Process Service (MPS)](gpu-sharing-mps.md)
    - [Fractional GPUs (AWS)](gpu-sharing-fractional.md)
  - [Dynamic Resource Allocation (DRA)](dra.md): Configure Atlas Autoscaler to use Kubernetes Dynamic Resource Allocation (DRA) for flexible resource management
  - [Pod placement](pod-placement.md): Learn how to place pods using labels and other Kubernetes scheduling features
  - [Pod Pinner](pod-pinner.md)
  - [Subnets](subnets.md)
  - [Network bandwidth](network-bandwidth.md)
  - [Commitments](commitments.md)
    - [Enterprise commitments](enterprise-commitments.md)
    - [AWS capacity reservations](aws-capacity-reservations.md): Learn how to use AWS On-Demand Capacity Reservations (ODCRs) and EC2 Capacity Blocks for ML with Atlas to guarantee instance availability for critical workloads.
    - [GCP capacity reservations](gcp-reservations.md): Learn how to use Google Cloud Compute Engine on-demand reservations with Atlas to guarantee VM availability for critical workloads.
  - [Autoscaler settings](autoscaler-settings.md)
  - [Autoscaler Node Labels and Taints](autoscaler-reference-node-labels-and-taints.md): Reference of all labels and taints that Atlas automatically applies to nodes it manages.
  - [Managing DaemonSets with Atlas](troubleshooting-daemonset-management.md): Learn about the implications of making changes to DaemonSets and its effects on existing nodes.
  - [Troubleshooting node autoscaling](autoscaler-troubleshooting.md): Solutions for resolving node provisioning, pod scheduling, and cloud provider-specific issues with the Atlas autoscaler.
- [Downscaling](downscaling.md)
  - [Evictor](evictor.md): Learn how to enable and configure Atlas's Evictor, a bin-packing component that continuously compacts pods into fewer nodes for cost savings.
  - [Evictor vs. Rebalancer](node-autoscaling-evictor-vs-rebalancer.md)
- [Rebalancing](rebalancing.md)
  - [Workload preparation](preparation.md)
  - [Scheduled rebalancing](scheduled-rebalancing.md)
  - [Paused drain configuration](rebalancing-paused-drain-configuration.md)
- [Cluster hibernation](autoscaling-cluster-hibernation.md)
  - [Cluster hibernation (Legacy)](pausing-a-cluster.md)
- [Migration from Karpenter](migration-from-karpenter.md)
- [Upgrading Kubernetes version](upgrading-kubernetes-version.md)
- [Cluster certificate rotation](troubleshooting-certificate-rotation.md): Learn how to handle Kubernetes certificate rotation in clusters managed by Atlas.
  - [Overview](clm-overview.md)
  - [Probe and lifecycle behavior](clm-concept-probe-lifecycle-behavior.md): Technical documentation explaining how Kubernetes probes and lifecycle hooks behave during container live migration, including what is skipped, what runs, and how to customize behavior for edge cases.
    - [Requirements and limitations](clm-requirements-and-limitations.md)
    - [Labels, Annotations, Taints, and Events](clm-reference-labels-and-annotations.md): Reference documentation for all labels, annotations, taints, and Kubernetes events used by Container Live Migration.
    - [TCP migration modes](clm-reference-tcp-migration-modes.md): Overview of the three TCP preservation paths available during container live migration: AWS VPC CNI, Traffic Control, and Calico as full CNI.
    - [Workload eligibility](clm-reference-workload-eligibility.md): How the CLM controller determines which pods are eligible for live migration — pod owner types, GPU and CPU constraints, PVC zones, pod state, and refusal reasons.
    - [Traffic Control](clm-reference-traffic-control.md): Reference for the Traffic Control (TC) TCP preservation path — configuration, kernel requirements, how it works, limitations, and comparison with the VPC CNI path.
    - [AWS VPC CNI](clm-reference-aws-vpc-cni.md): Reference for the AWS VPC CNI TCP preservation path — autodiscovery, configuration, external connections, limitations, and comparison with Traffic Control.
  - [Cloud providers](clm-cloud-providers.md)
    - [AWS EKS](clm-cloud-providers-eks.md)
    - [GCP GKE](clm-cloud-providers-gke.md)
    - [Azure AKS](clm-cloud-providers-aks.md)
    - [Using Container Live Migration with Evictor and Rebalancer](clm-getting-started.md)
    - [Using Container Live Migration with Spark](clm-spark.md): Tutorial: configure Apache Spark workloads for Container Live Migration — heartbeat tuning, memory overhead, headless service for TC, and testing migrations during a job.
  - [Quickstart](pod-mutations-quickstart.md)
  - [Overview](pod-mutations-overview.md)
    - [Enable Workload Autoscaler with pod mutations](enable-workload-autoscaler-with-pod-mutations.md)
  - [Reference](pod-mutations-reference.md)
- [Using ARM nodes with Atlas](guide-arm-and-graviton-support.md)
- [Business continuity](business-continuity.md): This article outlines how you can minimize the operational impact of your applications on the Kubernetes cluster if Atlas can't provide services.
- [Watchdog](watchdog.md)

## OMNI Edge Provisioning

- [Overview](omni-overview.md)
- [Getting started](omni-getting-started.md)
- [Custom edge locations](omni-custom-edge-location.md)
- [Edge configuration](omni-edge-configuration.md)
- [Custom MIG configurations](omni-custom-mig.md)
- [Manual edge provisioning](omni-manual-edge-provisioning.md)

## Workload Autoscaling

- [Overview](workload-autoscaling-overview.md)
- [Workload Autoscaler configuration](workload-autoscaling-configuration.md)
  - [Available settings](woop-configuration-settings.md)
  - [Annotations reference](workload-autoscaler-annotations-reference.md)
- [Scaling policies](woop-scaling-policies.md)
  - [How-to: Create a scaling policy](woop-scaling-policies-create.md)
  - [How-to: Manage scaling policies](woop-scaling-policies-manage.md)
  - [How-to: Generate scaling policies with AtlasPilot Assistant](woop-scaling-policies-atlaspilot-assistant.md)
- [Custom metrics](woop-custom-metrics.md)
- [JVM workload optimization](woop-jvm-optimization.md)
- [In-Place Pod Resizing](woop-in-place-resizing.md)
- [Node-aware DaemonSet sizing](woop-node-aware-daemonset-sizing.md)
- [Pod startup recommendations](woop-startup-recommendations.md)
- [Horizontal Pod Autoscaling](horizontal-pod-autoscaling.md)
    - [How-to: Configure HPA on a workload](woop-hpa-how-to-configure.md)
    - [How-to: HPA in scaling policies](woop-hpa-how-to-policies.md)
    - [How-to: Migrate from legacy horizontal scaling to HPA](woop-hpa-how-to-migrate.md)
    - [KEDA compatibility](woop-hpa-third-party-support.md)
    - [Vertical & horizontal workload autoscaling](workload-autoscaler-hpa-vpa.md)
- [Event log](event-log.md)

## Monitoring & Reporting

- [Overview](cost-management.md)
  - [Available savings](available-savings.md): Use this report to find out how much exactly you can save thanks to optimizing your cluster.
  - [Realized savings](savings-report.md): Understand how much money you have saved thanks to optimizing your clusters.
  - [Savings calculations](savings-baseline.md): Baseline methodology, calculation formulas, and pricing logic behind every number in the Savings Report.
  - [Cluster cost report](cluster-total-compute-cost.md)
  - [Organization cost report](organizational-cluster-cost-report.md)
  - [Workload costs](workloads.md): Learn how to use the Workloads report to analyze cluster costs, efficiency, workload grouping, and troubleshooting.
  - [Namespace costs](namespaces.md)
  - [Network cost](network-cost.md): The network cost feature enables you to delve deeper into the distribution of traffic across your cluster, namespaces, and workloads and provides insights into the associated costs.
  - [Storage cost](storage-cost-monitoring.md)
  - [GPU utilization](gpu-utilization.md)
  - [Cluster efficiency](cluster-efficiency-report.md)
  - [Cost comparison](cost-comparison.md): Quickly grasp how efficient your cluster is becoming over time.
  - [Idle resources](cost-idle-resources-report.md): Identify and eliminate cloud waste by tracking unused resources across your organization.
  - [Cluster allocation groups](cluster-allocation-groups.md)
  - [Organizational allocation groups](organizational-allocation-groups.md)
- [Cluster insights](cluster-insights.md): Get an overview of your cluster's key data, health score, and resource status.
  - [Cluster score](cluster-score.md)
  - [Optimization constraints](optimization-constraints.md): Advisory workload configuration checks surfaced as part of the Cluster Score — missing probes, Pod Disruption Budget issues, and topology spread constraints.
  - [Reliability metrics](reliability-metrics.md): Monitor request rate, error rate, and latency for every service in your cluster, with no changes to your application code.
    - [Reference](reliability-metrics-reference.md): atlasctl flags, auto-sizing behavior, and what gets deployed in your cluster when you enable reliability metrics.
  - [Prometheus and Grafana](prometheus-and-grafana.md): Learn how to scrape Atlas metrics and visualize them in Grafana.
  - [Integrating Prometheus Metrics with New Relic](observability-metrics-integrating-prometheus-metrics-with-new-relic.md)
- [Sentinel](sentinel.md): Configure, install, and upgrade the Atlas Sentinel cluster observability agent.
  - [Installation & upgrading](sec-sentinel-installation-and-upgrading.md)
  - [Configuring Sentinel features](sec-configuring-sentinel.md)
  - [Cloud network context](cloud-network-context.md): Enrich network flow data with cloud provider region, zone, and subnet information.
  - [Private image scanning](image-scanning.md): Set up access to private or managed container image registries.
  - [Network intelligence](network-observability.md)
- [Security insights](security-insights.md): Detect vulnerabilities in container images and assess cluster compliance with security best practices.
  - [Vulnerabilities](image-security.md): Identify container image vulnerabilities across your Kubernetes cluster and remediate them automatically.
  - [Compliance](best-practices.md): Assess your Kubernetes cluster configuration against Atlas security best practices.

## Storage Autoscaling

- [Introduction](storage-introduction.md): Automatically expand Kubernetes persistent volume capacity and rightsize IOPS and throughput with Atlas Storage Autoscaling.
- [How it works](storage-how-it-works.md): Understand how Atlas Storage Autoscaling expands persistent volume capacity and rightsizes IOPS and throughput performance.
- [Installation](storage-installation.md): Install Storage Autoscaling on GKE or EKS clusters with dry-run safety mode, cloud credentials, and Sentinel storage metrics.
- [Configuration](storage-configuration.md): Configure PVPolicy Custom Resources and PVC annotations for storage capacity, IOPS, and throughput autoscaling.
- [Troubleshooting](storage-troubleshooting.md): Troubleshoot common Storage Autoscaling issues including PV not growing, no IOPS changes, permission errors, and cooldown problems.

## Database Optimizer

- [Introduction](dbo-introduction.md)
- [Getting started](dbo-getting-started.md)
  - [AWS RDS & Aurora quick start](dbo-aws-rds-and-aurora-quick-start.md)
  - [Cloud SQL Proxy quick start](dbo-cloud-sql-proxy-quick-start.md)
  - [How does it work?](dbo-how-does-it-work.md)
  - [Access requirements and permissions](dbo-access-requirements-and-permissions.md)
  - [Security and compliance](dbo-security-and-compliance.md)
  - [Supported platforms](dbo-supported-platforms.md)
  - [Performance Advisor](dbo-performance-advisor.md)
    - [Installing the agent for PostgreSQL](dbo-installing-the-agent-for-postgresql.md)
    - [Installing the agent for MySQL](dbo-installing-the-agent-for-mysql.md)
  - [Caching](dbo-caching.md)
    - [Auto vs. Manual mode and override rules](dbo-auto-vs-manual-mode-and-override-rules.md)
    - [Performance estimation and cost savings](dbo-performance-estimation-and-cost-savings.md)
  - [Connection Pooling Overview](dbo-connection-pooling-overview.md)
    - [PostgreSQL](dbo-postgresql.md)
    - [MySQL](dbo-mysql.md)
  - [Organization overview](dbo-organization-overview.md)
  - [Instance dashboard](dbo-instance-dashboard.md)
  - [Database dashboard](dbo-database-dashboard.md)
  - [Connecting client applications](dbo-connecting-client-applications.md)
  - [Application failover configuration](dbo-application-failover-configuration.md)
  - [Pause DBO for troubleshooting](dbo-pause-dbo-for-troubleshooting.md)
- [Audit log](dbo-audit-log.md)
- [FAQ](dbo-faq.md)

## Administration

- [Single Sign-On (SSO)](sso.md): Single Sign-On (SSO) allows users easy access to Atlas through their existing SSO provider.
  - [IdP user group sync](idp-user-group-sync.md)
    - [Okta: Set up IdP user group sync](how-to-sso-group-sync-okta.md)
    - [Okta: Managing IdP user group sync](how-to-managing-sso-group-sync.md)
    - [Azure AD: Set up IdP user group sync](azure-ad-set-up-idp-user-group-sync.md)
- [Organization management](organization-management.md): Learn how to manage your organization settings, user roles, and access controls safely in Atlas.
  - [Enterprise organizations](enterprise-organizations.md): Centrally manage multiple child organizations, users, and access controls from a single Enterprise workspace in Atlas.
  - [Manage child organizations](how-to-manage-child-organizations.md): Create, edit, and navigate between child organizations within your Atlas Enterprise.
  - [Manage Enterprise users](how-to-manage-enterprise-users.md): Invite users, assign roles, manage access, and configure user groups across your Atlas Enterprise and its child organizations.
- [Role-Based Access Control (RBAC)](role-based-access-control-rbac.md)
  - [Users](users.md)
    - [How-to: Inviting users to organization](how-to-invite-users-to-organization.md)
    - [How-to: Changing user roles and access](how-to-changing-user-roles-and-access.md)
    - [How-to: Removing users](how-to-removing-users.md)
  - [User Groups](user-groups.md)
    - [How-to: Creating and configuring user groups](how-to-creating-and-configuring-user-groups.md)
    - [How-to: Managing existing user groups](how-to-managing-existing-user-groups.md)
  - [Service Accounts](service-accounts.md)
    - [How-to: Creating service accounts](how-to-creating-service-accounts.md): Learn how to create and configure service accounts for programmatic access to your resources on Atlas.
    - [How-to: Managing service accounts](how-to-managing-service-accounts.md): Learn how to modify service account settings, manage API keys, and control resource access.
- [IP-Based Access Control](ip-based-access-control.md): Restrict access to the Atlas platform based on source IP addresses and ranges.
- [AWS Marketplace subscription setup](aws-marketplace-subscription-setup.md)
  - [Overview](discount-engine-overview.md)
  - [Getting started](discount-engine-getting-started.md)
- [Notifications](notifications.md)
    - [Set up Slack notifications](observability-tutorial-set-up-slack-notifications.md): Learn how to connect your Slack workspace to Atlas and configure alerts to be delivered to your Slack channels.
    - [Set up webhook notifications](setup-notification-webhook.md): Check how to configure webhooks to send important notifications to external Ops systems when something happens with your clusters.
  - [Webhook integration examples](examples.md)
- [Audit log](audit-log.md): Check out this page to learn how Atlas's Audit log tracks and displays detailed records of all operations, policy actions, and changes made in your cluster.
- [Cluster and node status](cluster-node-status.md): Reference for Atlas cluster lifecycle status values and node health indicators, including API status values for building integrations.
- [Platform usage](administration-platform-usage.md): Track billable usage, credits, and consumption across Atlas products in your organization.

## FAQ

- [General](general.md)
- [API](api.md)
- [Arm and Graviton](arm-and-graviton.md)
- [Autoscaler](autoscaler-1.md)
- [Evictor](faq-evictor.md)
- [CUD, Savings Plans, and reservations](cud-savings-plans-and-reservations.md)
- [egressd, network and VPC](egressd-network-and-vpc.md)
- [Helm](helm.md)
- [Kubernetes](kubernetes.md)
- [Logs, alerts, and metrics](logs-alerts-and-metrics.md)
- [Node templates, node configuration, and labels](nodetemplates-nodeconfiguration-and-labels.md)
- [Permissions, users, Orgs, and SSO](permissions-users-orgs-and-sso.md)
- [Reports and UI/UX](reports-and-uiux.md)
- [Storage](storage.md)
- [Terraform](terraform.md)
- [Updates and images](updates-and-images.md)
