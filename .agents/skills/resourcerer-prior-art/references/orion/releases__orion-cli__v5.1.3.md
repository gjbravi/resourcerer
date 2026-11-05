# Version 5.1.3

Site [llms.txt](/llms.txt)

---

### Added

* When creating auto-deployed recommendations, you can now include a `--wait` flag and corresponding `--timeout` flag to delay applying the recommendation

### Changed

* Workload CSV numeric values are rendered at a fixed precision

### Fixed

* The `orion apply` command now correctly reapplies recommendations
* Sort keys are not silently ignored
