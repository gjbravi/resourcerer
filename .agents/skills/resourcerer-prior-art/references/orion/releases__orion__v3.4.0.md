# Version 3.4.0

Site [llms.txt](/llms.txt)

---

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
