# Version 3.5.0

Site [llms.txt](/llms.txt)

---

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
