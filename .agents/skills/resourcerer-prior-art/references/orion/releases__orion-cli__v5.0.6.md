# Version 5.0.6

Site [llms.txt](/llms.txt)

---

### Added

* The new `orion edit cluster` command can be used to manage cluster resources

### Changed

* The `orion apply` command no longer applies recommendations that have insufficient data unless `--force` is used

### Deprecated

* The `orion edit workload label` command is now deprecated. Use  `orion edit workload --set-label` instead.
* Deprecation notice: Starting in the *next* release, in an effort to reduce dependencies, the `orion-cli` container image will no longer include a shell.  

### Fixed

* The `orion get workload` table view honors `--show-impact`
* The `orion get recommendations` command now fetches all recommendations correctly when you specify JSON, YAML, or CSV output format

### Security

* Upgrades dependencies with known vulnerabilities
