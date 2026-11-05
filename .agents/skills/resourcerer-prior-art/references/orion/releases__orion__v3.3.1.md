# Version 3.3.1

Site [llms.txt](/llms.txt)

---

### Changed

* **Dependency updates**

### Fixed

* **LimitRange violations**: The Agent now detects configured LimitRanges, so recommendations that violate them are not deployed.

### Security

* **Go 1.27.2**: Updated to Go 1.27.2, which includes security fixes to the `crypto/tls` and `net/http` packages.
