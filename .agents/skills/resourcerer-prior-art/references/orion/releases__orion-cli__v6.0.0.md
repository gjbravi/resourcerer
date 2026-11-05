# Version 6.0.0

Site [llms.txt](/llms.txt)

---

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
