# Version 5.1.0

Site [llms.txt](/llms.txt)

---

### Changed

* Orion Pro commands are now hidden
* The `orion patch` command supports multiple recommendations
* Workload CSV output includes `replicas` and `optimization_score`

### Removed

* The `orion run` command has been removed
* The `orion edit workload label` command has been removed. Use `orion edit workload --set-label` instead.

### Fixed

* The `Deleted` value is no longer suggested for `--state` flags
* Create recommendation failures are properly reported
* Allow recommendations to be sorted by timestamp
* The `apply` command now prevents empty filters

### Security

* Switched container image base from Alpine to Distroless
