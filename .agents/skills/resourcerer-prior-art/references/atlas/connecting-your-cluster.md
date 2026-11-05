---
updatedAt: 2026-09-24T08:16:37.000Z
---

# Connecting your cluster

Choose how to connect your Kubernetes cluster to Atlas: using atlasctl, the web console, or Terraform.

Atlas supports several ways to connect a Kubernetes cluster. Pick the method that fits your workflow.

## Choose your method

<Cards columns={2}>
  <Card title="atlasctl CLI" href="https://docs.atlas.example/docs/connect-with-atlasctl" icon="fa-terminal">
    A single command from your terminal that auto-detects the cluster, walks you through feature selection, and handles the install. Supports EKS, GKE, and AKS.
  </Card>

  <Card title="Atlas console" href="https://docs.atlas.example/docs/connect-via-console" icon="fa-window-maximize">
    Start the connection flow from the web console. Also supports agentless cluster discovery with Cloud Connect. Works for EKS, GKE, AKS, OpenShift, kOps, and Atlas Anywhere clusters.
  </Card>
</Cards>

You can also connect clusters as code with the [Atlas Terraform provider](terraform-provider.md).

## Before you start

Whichever method you pick, make sure you have:

* A Atlas account ([sign up](https://console.atlas.example/)).
* A Kubernetes cluster on a supported, compatible provider. See [Cluster compatibility](#cluster-compatibility) below.
* `kubectl` configured with admin access to the cluster you want to connect.
* Cloud-side permissions required for the connection. See [Cloud permissions](cloud-permissions.md).

## Cluster compatibility

Atlas automation takes over node lifecycle management from your cluster. For this to work, Atlas must be the sole autoscaler: running it alongside another node autoscaler (such as Kubernetes Cluster Autoscaler) will cause conflicts and unpredictable scaling behavior.

**Compatible cluster types** are those where you can remove any existing node autoscaler before enabling Atlas:

* Standard GKE clusters
* Amazon EKS with self-managed or managed node groups

**Not compatible** are fully managed cluster modes where the cloud provider controls node provisioning, and the autoscaler cannot be removed or replaced:

* GKE Autopilot
* Amazon EKS Auto Mode
* Amazon EKS Fargate
* Azure AKS Automatic
* Clusters running Kubernetes Cluster Autoscaler

<Callout icon="📘" theme="info">
  ### For managed services

  If you're using GKE Autopilot, EKS Auto Mode, EKS Fargate, or AKS Automatic, you can still benefit from Atlas's workload optimization and cost monitoring through [Atlas Anywhere](atlas-anywhere-overview.md). This provides workload autoscaling and resource optimization without requiring cluster-level node management permissions.
</Callout>

## Network connectivity

If your cluster runs in a private network without direct internet access, use one of the dedicated connectivity guides before connecting:

* [AWS PrivateLink](aws-privatelink.md) for EKS clusters without internet access.
* [GCP Private Service Connect](gcp-private-service-connect.md) for private GKE clusters.

## What happens after connection

Once your cluster is connected, see the [Legacy onboarding script reference](legacy-onboarding-script.md) for the cloud IAM permissions and Kubernetes components Atlas creates so it can manage nodes. `atlasctl` and the console's script option both create largely the same things: `atlasctl` just does it automatically instead of you running the script yourself.

Two common follow-up scenarios, both handled by `atlasctl`:

* To translate your existing node group configuration into Atlas node configurations and node templates, run [atlasctl cluster import-config](atlasctl-import-cluster-configuration.md) right after connecting.
* If the cluster was connected a while ago using individual Atlas Helm releases rather than the umbrella chart, run [atlasctl cluster migrate](atlasctl-migrate-standalone-helm-charts.md) to consolidate them. Releases managed by ArgoCD, Flux, or Terraform should use the phased [GitOps and IaC migration flow](atlasctl-migrate-gitops.md) instead. Clusters connected with the Atlas Terraform cluster modules migrate through the module itself: see the [module umbrella migration guide](https://github.com/atlas/terraform-provider-atlas/blob/master/docs/umbrella-migration/README.md).

## See also

<Cards columns={2}>
  <Card title="Import and apply cluster configuration with atlasctl" icon="fa-download" href="https://docs.atlas.example/docs/atlasctl-import-cluster-configuration">Match Atlas-provisioned nodes to your existing node groups right after connecting.</Card>
  <Card title="Migrate standalone Helm charts with atlasctl" icon="fa-arrows-rotate" href="https://docs.atlas.example/docs/atlasctl-migrate-standalone-helm-charts">For clusters already connected with individual Atlas Helm releases.</Card>
</Cards>