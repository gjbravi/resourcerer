# Orion Agent releases

> Release history for the Orion Agent

---

To install the latest version, see the [Upgrade] topic.

For past releases, see [Previous Orion Agent and Applier releases].

[Upgrade]: /docs/installation/upgrade/
[Previous Orion Agent and Applier releases]: /docs/releases/orion/previous/

---

## Version 3.5.0

_October 28, 2026_

### Added

* **`optimize.orion.example/v2` OptimizationConfiguration CRDs**: `OptimizationConfiguration` and `ClusterOptimizationConfiguration` now have a v2 schema and storage version. Workload matching moves from the flat `selector`, `namespaceSelector`, and `matchConditions` fields into a single `match` object that can be set per resource or rule. Selectors are now label-selector strings, for example `app=web`. A workload matches when:
  * It matches every criterion that is set: `objectSelector`, `expression` (CEL), and `namespaceSelector` (`ClusterOptimizationConfiguration` only).
  * It matches at least one term in `anyOf`, if set.
* **v2 validation**: The validating webhook now covers v2, rejecting v2 selectors that cannot be represented in v1.

### Changed

* **Webhook required for CRDs**: The OptimizationConfiguration CRDs convert between v1 and v2 through the webhook. When the chart installs the CRDs (`crds.enabled=true`, the default), installing or upgrading fails if `webhook.enabled` is `false` or if `webhook.tlsSecretName` is set without `webhook.caBundle`. If you manage CRDs yourself (`crds.enabled=false`), keep `webhook.enabled=true` and manage the CRDs' conversion settings and CA bundle.
* Dependency updates

### Deprecated

* **`optimize.orion.example/v1` OptimizationConfiguration APIs**: v1 is still served but deprecated. Clients get the warning `optimize.orion.example/v1 is deprecated; use v2`. Migrate manifests to v2.

---

## Version 3.4.0

_October 22, 2026_

### Added

* **Webhook component**: The webhook component now runs as a stand-alone Deployment rather than in the Applier. It serves both pod admission mutation (container resource recommendations) and config/Custom Resource Definition (CRD) validation.
* **CRD validation**: Orion now runs a validating admission webhook for ClusterOptimizationConfiguration and OptimizationConfiguration CRDs, to help catch mistakes in setting names or values.
* **Pod mutation outcome metrics**: Each pod admission mutation outcome is now recorded in the `orion_admission_pod_mutations_total` metric, and mutation failures have additional logging.
* **Network cost allocation**: The Agent now reports whether each pod uses the host network, to improve the accuracy of network cost allocation.

### Changed

* **Faster node sync**: The Agent syncs nodes through a new incremental node API, which speeds up startup on large clusters.

### Fixed

* **Stale recommendation URLs**: A pod whose `recommendation-url` annotation points to a deleted recommendation is now counted as skipped, not failed.

### Security

* **Dependency updates**: Updated dependencies to address [CVE-2026-113470](https://nvd.nist.gov/vuln/detail/cve-2026-113470).

---

## Version 3.3.1

_October 9, 2026_

### Changed

* **Dependency updates**

### Fixed

* **LimitRange violations**: The Agent now detects configured LimitRanges, so recommendations that violate them are not deployed.

### Security

* **Go 1.27.2**: Updated to Go 1.27.2, which includes security fixes to the `crypto/tls` and `net/http` packages.

---

## Version 3.3.0

_October 1, 2026_

### Added

* **Node CPU pressure detection**: The Agent detects Kubernetes nodes under sustained CPU pressure and identifies the workloads responsible. It reads each node's Pressure Stall Information (PSI) from cAdvisor. When a node's CPU pressure exceeds the detection threshold, the Agent finds the containers on that node using more CPU than they request, maps them to their workloads, and creates a recommendation with the `CPUPressureDetected` reason. Each recommendation lists the node, pod, and container, and how far CPU usage exceeded the request. Detection is always on. See [CPU pressure response](configure__settings__reliability.md#cpu-pressure).

  The Agent reports these metrics for CPU pressure detection:

  * `orion_container_cpu_pressure_waiting_seconds_total`
  * `orion_container_cpu_pressure_stalled_seconds_total`
  * `orion_container_cpu_pressure_waiting_rate`
  * `orion_container_cpu_pressure_stalled_rate`
  * `orion_workload_event_last_timestamp_seconds`

### Changed

* **Dependency updates**

### Fixed

* **Node metrics refresh on provider ID change**: Node metrics now pick up provider IDs that are assigned late or change. Previously, node updates that changed only the provider ID were ignored.
* **Provider ID sync to Orion API**: When cluster data is enabled, the Agent sends a node's provider ID to the Orion API whenever it changes. Previously, it was sent only when a node was added or its labels changed.

---

## Version 3.2.0

_September 21, 2026_

### Added

* **GPU optimization (early access)**: Orion can now collect GPU metrics for workloads that use NVIDIA GPUs and show per-workload GPU memory usage and GPU allocation alongside device-level GPU activity. See [GPU optimization](recommendations__gpu.md).
* **Orion Node Agent**: A new optional DaemonSet, `orion-node-agent`, collects the GPU metrics from GPU nodes. It is disabled by default; enable it with `--set nodeAgent.enabled=true`. See [Node Agent configuration](installation__node-agent-config.md).
* **GPU workload detection**: Workloads that request a GPU, whether through a GPU extended resource such as `nvidia.com/gpu` or through a Dynamic Resource Allocation resource claim, are identified by Orion. Detection needs no configuration and does not require the Node Agent.

### Changed

* **Dependency updates**

---

## Version 3.1.3

_September 17, 2026_

### Changed

* **Dependency updates**

### Fixed

* **Recommendation status update conflict**: Resolved a conflict between the pod admission webhook and the Applier when updating recommendation status.

---

## Version 3.1.2

_September 9, 2026_

### Added

 * **Applier alerts change diffs**: Applier alerts sent to external endpoints when the applier applies recommendations now include full before-and-after diffs for resources that were changed.

 ### Changed

 * **Performance improvements**
 * **Dependency updates**

 ### Fixed

 * **Missing RBAC for Events**: The Applier now includes RBAC to allow creating and patching Events when use of the Edit role is disabled.
 * **Stale cost metric retention**: The Agent now correctly removes cost-related metrics for deleted pods not owned by workload types Orion recognizes. Previously, some of these metrics were retained until the Agent restarted.

---

## Version 3.1.1

_July 24, 2026_

### Fixed

 * **OpenShift pod scheduling:** The consolidated v3 chart no longer applies a hardcoded pod `securityContext` (`fsGroup`/`runAsGroup`/`runAsUser: 65534`) when `openshift=true` and `commonPodSecurityContext` is unset.

---

## Version 3.1.0

_July 23, 2026_

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Renamed chart parameters</div>


The `agent` parameter group is renamed to `clusterAgent`, and `exporter` to `forwarder`. The deprecated `agent.*` and `exporter.*` values still work, overriding their `clusterAgent.*` and `forwarder.*` equivalents, and will be removed in a future major version. Update your values when convenient.
</div>


### Added

* **JVM old-gen live-set metric**: The forwarder (formerly `exporter`) now collects `jvm_memory_pool_collection_used_bytes` for live data size transforms.
* **Chart `deepOverride` helper**: Explicit-zero-preserving merge helper so intentional `0` values in overrides are no longer dropped.

### Changed

* **`exporter` component renamed to `forwarder`**: The `exporter` Helm values key is renamed to `forwarder`. The old key is deprecated but still works (see the note above).
* **`agent` component renamed to `cluster-agent`**: The `agent` Helm values key is renamed to `clusterAgent`, and the agent config file to `cluster-agent.yml`. The old key is deprecated but still works (see the note above); the config file is passed via `--config`, so no action is needed for it.
* **Webhook injection scope narrowed to the applier's own pods**: The mutating webhook exclusion no longer covers the entire release namespace. It now skips only pods labeled `webhook.orion.example/skip` (which the applier carries). As a result:
    * Orion's own workloads in the release namespace become eligible for webhook-based apply.
    * The release namespace may now appear in `allowNamespaces`.
    * Helm 4 server-side-apply upgrades work without `--force-conflicts`.
* **Quieter logging**: Pod spec generation failures now log at debug instead of error.
* Dependency updates


### Fixed

* **Applier init container fails behind a corporate CA/proxy**: The applier init container (`applier license-check`, which calls the Orion API over TLS) now mounts `additionalVolumeMounts`. Previously it mounted no volumes, so customers injecting a corporate CA cert hit `x509: certificate signed by unknown authority` during init and the pod never started. Restores the 2.x chart behavior and matches the `additionalVolumeMounts` docs ("added to all containers").

* **`SuspendState` event silently skipped**: The applier Deployment now projects `POD_NAME` / `POD_NAMESPACE` from the downward API, so the applier can resolve its own Deployment and emit the `SuspendState` event on startup (for example, when setting `suspend=All`). Previously this failed with `ErrSelfPodEnvUnset` and only logged a warning.

---

## Version 3.0.0

_June 10, 2026_

<div class="alert alert-warning" role="alert"><div class="h4 alert-heading" role="heading">Breaking change</div>


Version 3.0.0 unifies all Orion components into a single Helm chart. Upgrading from the v2 Agent and Applier charts requires migration — see [Migrate v2 → v3][migration].
</div>


### Added

- **Unified `orion` Helm chart**: New chart that bundles the workload-agent and the applier in a single installation, replacing the separate `orion-agent` chart. Required values are the same (`clusterName`, `authorization.clientID`, `authorization.clientSecret`). **Breaking change** — replaces `orion-agent`.

- **`workloadResourceKinds` model**: A single list replaces the two parallel v2 parameters (`workloadResourceTypes` + `ownerResourceTypes`). Each entry carries an explicit `role` field — `workload`, `conditionalWorkload`, `owner`, `ignoredOwner` — eliminating the ambiguity of separate lists. `patchTargetDefaults` (renamed from `patchTargetTypeDefaults`) is set per entry; the chart injects the required `live.orion.example/` prefix internally. **Breaking change** — see [Migrate v2 → v3][migration].

- **`includeCommonAddOnResourceKinds`**: Replaces `includeCommonAddOnResourceTypes`, enabled by default. Prepends curated entries for Argo Rollouts, Prometheus operator, OpenTelemetry collector, KEDA, and Dynatrace.


### Changed

- **CRDs are now Helm-managed**: CRDs moved from `crds/` to `templates/crds/` so they upgrade in lockstep with the chart. `crds.enabled` (default `true`) opts out for out-of-band CRD management; `crds.keep` (default `true`) prevents cascade-delete on uninstall. **Breaking change** — upgrade note below.

- **cAdvisor scrape method auto-detected**: The `CadvisorDirectScrape` config flag is removed. The scrape method is now inferred automatically from the `instance` label on incoming PSI metrics.

- **PSI waiting-rate uses a time-weighted window**: The fixed-sample-count ring buffer is replaced with timestamped samples averaged over a configurable `WaitingAverageWindow`. The `SampleInterval` config field is removed; set `WaitingAverageWindow` (default `10m`) instead. Stale samples outside the window are excluded regardless of scrape cadence changes.

---

### Upgrade note — CRD ownership transfer

CRDs previously installed from the `crds/` directory carry no Helm ownership metadata. The first `helm upgrade` onto v3.0.0 may fail with "invalid ownership metadata." Remediation options:

- **Helm ≥ 3.17**: run the upgrade once with `--take-ownership`
- **GitOps / older Helm**: label and annotate the existing CRDs before upgrading:
  ```sh
  kubectl label crd <name> app.kubernetes.io/managed-by=Helm --overwrite
  kubectl annotate crd <name> \
    meta.helm.sh/release-name=<release> \
    meta.helm.sh/release-namespace=<namespace> --overwrite
  ```

[migration]: /docs/installation/migration/

---

Section pages:

- [Previous Orion Agent and Applier releases](/docs/releases/orion/previous/)
