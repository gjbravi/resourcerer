# Orion CLI releases

> Release history for the Orion CLI

---

## Version 6.2.10

_October 26, 2026_

###  Added

* The `view` command can filter by cluster labels

###  Changed

* Improved error messages
* The `edit --config` flag has been renamed to `edit --managed`
* Dependency updates

###  Fixed

* Homebrew installations no longer use deprecated stanzas

---

## Version 6.2.9

_August 24, 2026_

###  Changed

* Upgrade Go to 1.26.7

###  Fixed

* Addressed an issue where `orion create recommendation --auto-deploy` reported a timeout instead of a failure

---

## Version 6.2.8

_July 23, 2026_

###  Added

* Support for GPU workload filtering

###  Changed

* Dependency updates
* Switched Homebrew tap to use a cask instead of a formula

###  Deprecated

* The `orion mcp` command. The stdio MCP server is deprecated in favor of the hosted Orion MCP server.

###  Fixed

* Various fixes for the Agent v2 to v3 migration tool

---

## Version 6.2.7

_June 15, 2026_

###  Changed

* Dependency updates

---

## Version 6.2.6

_June 2, 2026_

###  Changed

* The fields shown by `orion view workloads` are now more consistent with the UI

###  Fixed

* Reliability rate filters are usable to 4 decimal places

---

## Version 6.2.5

_May 19, 2026_

###  Changed

* Dependency updates

###  Fixed

* Fixed an issue where authorization tokens failed to refresh

---

## Version 6.2.4

_May 6, 2026_

###  Added

* The `view workload` command shows CPU throttling rate (when available)

###  Changed

* Made the `login --org` flag optional
* Dependency updates

###  Fixed

* Reliability filters on `view workload` accept decimal numbers

---

## Version 6.2.3

_April 28, 2026_

###  Fixed

* Fixed a compatibility issue with v6.2.2 authorization tokens

---

## Version 6.2.2

_April 27, 2026_

###  Added

* Added a container level `orion get recommendations --container-resources` alternate view
* Added a metric level `orion get recommendations --auto-scaling` alternate view

###  Changed

* Update dependencies
* The estate-wide configuration editor no longer defaults to the internal YAML representation

---

## Version 6.2.1

_March 25, 2026_

###  Fixed

* Fixes an issue where Orion tokens may not be properly decoded

---

## Version 6.2.0

_March 23, 2026_

###  Added

* Add an MCP Server to allow AI assistants to interact with the Orion platform
* Added support for editing global configuration

###  Changed

* Dependency updates

---

## Version 6.1.1

_December 3, 2025_

###  Changed

* Organization is required for logins when using `--url`
* Dependency updates

---

## Version 6.1.0

_October 31, 2025_

###  Added

* The validity period can be specified when creating recommendations

###  Changed

* Dependency updates

###  Deprecated

* The `--url` option of the `orion login` command will require an organization identifier in the next release

---

## Version 6.0.2

_September 15, 2025_

###  Changed

* Dependency updates
* Changed the version columns shown by `get clusters`

###  Fixed

* Corrected the URL emitted by the `check version` command
* The patch command no longer ignores the target subresource

---

## Version 6.0.1

_June 13, 2025_

### Added

* Restored the v5 `edit cluster --reset-uid` option
* Restored the v5 `--recommendation` option for specifying recommendation names

### Changed

* Dependency updates

### Removed

* The macOS binaries are no longer notarized

### Fixed

* Fixed an issue where get recommendations got too many results in YAML output
* Fixed an issue where the edit command would not report bad patch errors

---

## Version 6.0.0

_March 31, 2025_

### Added

* There is a new `orion view` command for fetching output comparable to the web UI
* There is a new `orion describe` command for fetching a more detailed view of a resource
* There is a new `orion manage` command for manually configuring resources

### Changed

* Commands accept a slash delimited name argument instead of (or in addition to) the filtering flags
* The `orion get` command shows different fields
* The `orion edit` command now launches an interactive editor

### Removed

* Orion Pro functionality has been removed
* Performance Test functionality has been removed

### Fixed

* The `--selector` argument is not ignored for tab completions
* Some commands mistakenly bypassed filtering requirements
* Flag and argument completions are more consistent

---

## Version 5.1.13

_February 6, 2025_

### Fixed

* Fixes an issue where `-o name` was selected when a TTY isn't available

---

## Version 5.1.12

_January 13, 2025_

### Changed

* Dependency updates

---

## Version 5.1.11

_December 18, 2024_

### Changed

* Dependency updates

### Fixed

* Fix for empty values showing up as "<nil>"

---

## Version 5.1.10

_December 16, 2024_

### Changed

* Dependency updates 
* The Go Template output has improved support for Kubernetes quantities
* Recommendation output includes JVM max heap (if applicable)

---

## Version 5.1.9

_October 17, 2024_

### Changed

* Dependency updates

### Fixed

* The workload patches can be read from stdin

---

## Version 5.1.8

_September 18, 2024_

### Changed

* Dependency upgrades
* Fixed an issue where some authorization scopes were not being requested

---

## Version 5.1.7

_July 16, 2024_

### Changed

* Dependency updates

### Fixed

* Fixes an issue where the `orion auth list` results could be truncated

---

## Version 5.1.6

_June 26, 2024_

### Changed

* Dependency upgrades

---

## Version 5.1.5

_May 24, 2024_

### Changed

* Dependency upgrades

### Security

* Built using Go 1.22.3 to address [CVE-2024-24788](https://nvd.nist.gov/vuln/detail/CVE-2024-24788)

---

## Version 5.1.4

_May 10, 2024_

### Changed

* Added missing workload impact data to the `orion get workload` output

---

## Version 5.1.3

_April 22, 2024_

### Added

* When creating auto-deployed recommendations, you can now include a `--wait` flag and corresponding `--timeout` flag to delay applying the recommendation

### Changed

* Workload CSV numeric values are rendered at a fixed precision

### Fixed

* The `orion apply` command now correctly reapplies recommendations
* Sort keys are not silently ignored

---

## Version 5.1.2

_March 14, 2024_

### Added

* The `get` commands now support limited client-side filtering using Kubernetes field selectors via the `--field-selector` option
* The `get workloads` command now has a `--full-fetch` option to force fetching all data from the API (note: this might take extra time to run)

### Fixed

* The CSV columns named `auto-deploy` and `hpa-enabled` are now named `auto_deploy` and `hpa_enabled`
* The `get workloads` command now shows "UNAVAILABLE" for values that were not fetched from the API

---

## Version 5.1.1

_March 12, 2024_

### Changed

* Added `schedule`, `auto-deploy`, and `hpa-enabled` columns to the workload CSV output

### Security

* Updated go-jose to address [CVE-2024-28180](https://nvd.nist.gov/vuln/detail/CVE-2024-28180)

---

## Version 5.1.0

_March 1, 2024_

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

---

## Version 5.0.6

_January 9, 2024_

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

---

## Version 5.0.5

_November 30, 2023_

### Changed

* Dependency upgrades

### Fixed

* JSON and YAML output should have matching field ordering

---

## Version 5.0.4

_November 2, 2023_

### Changed

* The `edit workload handlers --auto-deploy` flag is no longer implied by the thresholds
* The `edit workload container-resources` command accepts the `RequestsAndLimits` policy

### Fixed

* The `edit workload container-resources` command can unset values

---

## Version 5.0.3

_October 26, 2023_

### Added

* Support for workload auto-deploy thresholds. See the `orion edit workload handlers` command in the CLI reference.
* Workloads can be filtered by over- or under-provisioned state

### Changed

* Removed references to workload schedule names

### Fixed

* The `auth revoke` command was not accepting client identifiers

---

## Version 5.0.2

_October 12, 2023_

### Changed

* The `get cluster` output now includes the Orion Live and Orion Applier version numbers

### Fixed

* The `check connect` works correctly when testing client credentials
* Generating tokens no longer requests the `offline_access` scope

---

## Version 5.0.1

_September 8, 2023_

### Changed

* Updated recommendation state names

### Fixed

* The authorization issuer was not configured properly in some cases

---

## Version 5.0.0

_August 18, 2023_

### Changed

* Workload and recommendation commands reflect consolidation to a single schedule
* The `orion get workloads` command hides impact details unless `--show-impact` is used

### Removed

* The `install` and `uninstall` commands have been removed

### Fixed

* The `orion edit workload` commands ignore workloads using in-cluster configuration
