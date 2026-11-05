---
updatedAt: 2026-10-27T10:11:43.000Z
---

# Communication requirements

Traffic and communication requirements of Atlas.

This guide explains the network requirements for Atlas components to communicate with our services.

# Required endpoints

## Core services

All Atlas components must be able to reach these endpoints:

### US region

* `api.atlas.example:443`
* `grpc.atlas.example:443` (Pod pinner only)
* `sentinel.prod-master.atlas.example:443` (Sentinel)
* `files.atlas.example:443`

### EU region

* `api.eu.atlas.example:443`
* `grpc.eu.atlas.example:443` (Pod pinner only)
* `sentinel.prod-eu.atlas.example:443` (Sentinel)
* `files.eu.atlas.example:443`

### India region

* `api.india.atlas.example:443`
* `grpc.india.atlas.example:443` (Pod pinner only)
* `api-grpc.india.atlas.example:443` (Sentinel, APA, and newer components)
* `files.india.atlas.example:443`

## Supporting services

Atlas also requires access to these endpoints:

* Container registries (`ghcr.io:443` for GHCR tokens and manifests, plus `pkg-containers.githubusercontent.com:443` for image blobs):

  | Registry                  | Host                                                      | Image repository                       | Role     |
  | ------------------------- | --------------------------------------------------------- | -------------------------------------- | -------- |
  | GitHub Container Registry | `ghcr.io:443`, `pkg-containers.githubusercontent.com:443` | `ghcr.io/atlas/images`                | Default  |
  | Google Artifact Registry  | `us-docker.pkg.dev:443`                                   | `us-docker.pkg.dev/atlas-hub/library` | Fallback |

  Both registries host the same images at the same repository path; append `/<image>`.

* Helm charts:
  * `atlas.github.io:443`
  * `objects.githubusercontent.com:443`
  * `release-assets.githubusercontent.com:443`
  * `releaseassetproduction.blob.core.windows.net:443`

* Node binaries: `storage.googleapis.com/atlas-node-components/`

* Node logs: `storage.googleapis.com/atlas-node-logs-sender/`

The `files.*.atlas.example` URLs listed under Core services represent static IP proxies for node binaries and node logs.

### Image registry override

In the `atlas` umbrella chart, each component selects its image through `image.repository`, at `autoscaler.<component>.image.repository`. To route a component to the fallback registry, point that value at the fallback path. Only the registry host changes; the image path is identical. For example, on the agent and Sentinel:

```yaml
autoscaler:
  atlas-agent:
    image:
      repository: us-docker.pkg.dev/atlas-hub/library/agent
  atlas-sentinel:
    image:
      repository: us-docker.pkg.dev/atlas-hub/library/sentinel
```

Components not overridden keep their default (GHCR) repository.

# Port requirements

## Webhook ports

Some Atlas components operate as admission webhooks and require the Kubernetes API server to reach them on specific ports:

* **Workload Autoscaler**: Port `9443` (default). This port can be changed via the `webhook.port` Helm value when resolving port conflicts with other services. See [Host network and port conflicts](workload-autoscaling-overview.md#host-network-and-port-conflicts) for details. For general troubleshooting, see [Workload Autoscaler Troubleshooting](workload-autoscaling-overview.md#troubleshooting).

For webhook functionality to work properly, these ports must be accessible from the Kubernetes control plane to the Atlas pods.

# Node startup logs upload

During node initialization, Atlas uploads startup logs to assist with troubleshooting and cluster setup. This process uses a scoped-down, temporary API key with the following characteristics:

* **Scope**: The API key provides read-only access and is restricted to log ingestion only
* **Lifecycle**: A unique key is generated per node and is automatically deleted when the node is terminated
* **Usage**: Used for up to 1 hour after the initialization AddNode operation to send logs to `storage.googleapis.com/atlas-node-logs-sender/`
* **Expiration**: The API key automatically expires 1 hour after creation (node add operation initiation)
* **Security**: The API key is not reused across nodes and does not grant access to other API functionalities

The API key appears in EC2 User Data during node startup, but it **cannot be used in any other way or for any other purpose.**

# Network configuration

## IP allowlisting

If DNS allowlisting is not possible in your network infrastructure (firewall, NAT, Security Group, etc.), you can allowlist the IPs below instead.

These are the static IPs of the Atlas API endpoints. Your components initiate outbound connections to them to deliver information to the Atlas SaaS. Within a region, the API, gRPC, and `files` endpoints share the same set.

### US region

* `104.16.81.56`
* `104.16.82.56`

### EU region

* `141.101.90.96`
* `141.101.90.97`
* `141.101.90.98`
* `141.101.90.99`

### India region

* `172.65.90.8`
* `172.65.90.9`
* `172.65.90.10`
* `172.65.90.11`

> 📘 Note
>
> To limit which source IP addresses can access the Atlas platform, API, and console, see [IP-Based Access Control](ip-based-access-control.md).

## Proxy configuration

To use Atlas components behind a proxy, add these environment variables to your deployments:

```yaml
env:
- name: HTTP_PROXY
  value: "http://<proxy-address>:<port>"
- name: HTTPS_PROXY
  value: "https://<proxy-address>:<port>"
- name: NO_PROXY
  value: "localhost,<pod-cidr>,<svc-cidr>,*.cluster.local,googleapis.com,metadata.google.internal"
```

Example manifest for the `atlas-agent` deployment on a GKE cluster:

```yaml
containers:
- env:
  - name: API_URL
    value: $ATLAS_API_URL  # Replace with your regional API URL. See API access doc for regional endpoints.
  - name: PROVIDER
    value: gke
  - name: MONITOR_METADATA
    value: /agent-metadata/metadata
  - name: PPROF_PORT
    value: "6060"
  - name: HTTP_PROXY
    value: "http://<proxy-address>:<port>"
  - name: HTTPS_PROXY
    value: "https://<proxy-address>:<port>"
  - name: NO_PROXY
    value: "localhost,<pod-cidr>,<svc-cidr>,*.cluster.local,googleapis.com,metadata.google.internal"
  envFrom:
  - secretRef:
      name: atlas-agent
  image: ghcr.io/atlas/images/agent:v0.48.1
```

> 📘 Note
>
> Configure `NO_PROXY` to match your environment to prevent internal Kubernetes traffic from being sent to the external proxy.

> ⚠️ TLS-intercepting (MITM) proxies
>
> If your proxy intercepts TLS traffic, it re-signs certificates with its own CA. Atlas components must trust that CA or their outbound calls fail TLS verification. Configure a custom CA certificate for both layers:
>
> * **Pod-level** components (`atlas-agent`, `atlas-cluster-controller`, etc.): see [Inject a custom CA certificate](helm-charts.md#inject-a-custom-ca-certificate).
> * **Node-level** components on AKS: see [Custom CA certificate for MITM proxies](terraform-provider-aks.md#custom-ca-certificate-for-mitm-proxies).

## Atlas egress IPs

Atlas services use these IP addresses to communicate with third-party services and cloud providers (for example, when managing your cluster resources). If your infrastructure restricts such traffic, allowlist these IPs.

### US region (prod-master)

* `34.48.199.96`
* `35.221.25.73`
* `34.48.86.252`
* `34.86.37.231`

### EU region (prod-eu)

* `34.141.71.171`
* `34.89.180.238`
* `35.246.207.216`
* `34.89.186.65`

### India region (prod-india)

* `34.14.155.44`
* `34.47.196.217`
* `35.200.215.233`
* `34.100.217.237`

## GKE with Istio requirements

When using Istio on GKE, configure your cluster with:

* Port `15017` in `firewall_inbound_ports`
* `add_master_webhook_firewall_rules` set to `true`

Example Terraform configuration:

```yaml
add_master_webhook_firewall_rules = true
firewall_inbound_ports            = ["15017"]
```