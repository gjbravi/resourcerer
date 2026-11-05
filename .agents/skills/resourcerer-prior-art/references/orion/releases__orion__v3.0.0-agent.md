# Version 3.0.0

Site [llms.txt](/llms.txt)

---

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
