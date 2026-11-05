---
updatedAt: 2026-10-27T09:40:57.000Z
---

# Sentinel

Configure, install, and upgrade the Atlas Sentinel cluster observability agent.

Sentinel is Atlas's cluster observability agent. It collects network traffic flows, resource usage statistics, GPU metrics, and storage data that powers Atlas's monitoring and cost features. When security insights are enabled, Sentinel additionally performs image vulnerability scanning and compliance checks.

## What is Sentinel?

Sentinel operates as both a Kubernetes controller (Deployment) and an agent (DaemonSet). The controller reconciles cluster state and schedules assessment jobs; the agent runs on each node to collect real-time telemetry.

The agent is lightweight and minimally invasive. By default it runs in a read-only observability mode — security features are opt-in additions to this baseline.

### Key capabilities

#### Network traffic monitoring

Sentinel collects Kubernetes network flows using eBPF. This data powers Atlas's [network cost monitoring](network-cost.md) feature, giving you visibility into pod-to-pod and pod-to-external traffic patterns across IPv4 and IPv6.

#### Resource metrics

Sentinel collects storage utilization, Pressure Stall Information (PSI) metrics, and CPU/memory/I/O usage statistics from containers and nodes. This data surfaces in the Atlas console's Node list and powers reliability metrics.

#### GPU metrics

On clusters with NVIDIA GPUs, Sentinel collects GPU utilization data via DCGM Exporter. This powers Atlas's [GPU utilization monitoring](gpu-utilization.md) feature.

#### Cluster proxy

When enabled, the Sentinel controller exposes a cluster proxy that lets Atlas query the Kubernetes API on demand for live cluster state — including container logs for [AtlasPilot Assistant](atlaspilot-assistant.md) troubleshooting. It runs under a dedicated ServiceAccount and ClusterRole (separate from the main Sentinel service account) with read-only (`get`, `list`, `watch`) access. See [Sentinel Cluster Proxy permissions](kubernetes-permissions.md).

#### Container image vulnerability scanning (optional)

When enabled, Sentinel scans container images running in your cluster against known vulnerability databases. It evaluates both public and private registries and provides severity assessments based on CVSS scores. See [Vulnerabilities](image-security.md) for details.

#### Compliance checks (optional)

When enabled, Sentinel evaluates your Kubernetes environment against security best practices, identifying misconfigurations and validating RBAC settings. See [Compliance](best-practices.md) for details.

## Architecture

Sentinel operates with a two-component architecture:

1. **Sentinel Controller** (Deployment): Communicates with the Atlas control plane, reconciles cluster state, and schedules scanning jobs.

2. **Sentinel Agent** (DaemonSet): Runs on every node to collect real-time telemetry. Required for netflow monitoring, resource metrics, and GPU metrics.

## Permissions

Sentinel uses the same base permission set as the standard Atlas agent, plus additional permissions for its cluster proxy component. For detailed information, see [Kubernetes permissions](kubernetes-permissions.md).

### Extended system permissions

Some optional features require additional system-level privileges.

#### Kube-bench configuration validation

The Sentinel Controller schedules `atlas-kube-bench` jobs to validate kubelet and other system components against security benchmarks. These jobs require elevated privileges to access system configuration files.

**Required privileges:**

* HostPID access

**Read-only host mounts:**

* `/var/lib/kubelet`
* `/etc/systemd`
* `/etc/kubernetes`
* `/home/kubernetes`

<Cards columns={2}>
  <Card title="Installing and upgrading Sentinel" href="https://docs.atlas.example/docs/sec-sentinel-installation-and-upgrading" icon="fa-download">
    Step-by-step installation and upgrade instructions.
  </Card>
  <Card title="Configuring Sentinel features" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Enable and configure monitoring, scanning, and GPU metrics.
  </Card>
</Cards>