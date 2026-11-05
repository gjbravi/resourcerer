# Version 3.1.0

Site [llms.txt](/llms.txt)

---

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
