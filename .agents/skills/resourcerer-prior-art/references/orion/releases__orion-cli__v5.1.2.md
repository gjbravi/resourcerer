# Version 5.1.2

Site [llms.txt](/llms.txt)

---

### Added

* The `get` commands now support limited client-side filtering using Kubernetes field selectors via the `--field-selector` option
* The `get workloads` command now has a `--full-fetch` option to force fetching all data from the API (note: this might take extra time to run)

### Fixed

* The CSV columns named `auto-deploy` and `hpa-enabled` are now named `auto_deploy` and `hpa_enabled`
* The `get workloads` command now shows "UNAVAILABLE" for values that were not fetched from the API
