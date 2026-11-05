# Version 3.1.1

Site [llms.txt](/llms.txt)

---

### Fixed

 * **OpenShift pod scheduling:** The consolidated v3 chart no longer applies a hardcoded pod `securityContext` (`fsGroup`/`runAsGroup`/`runAsUser: 65534`) when `openshift=true` and `commonPodSecurityContext` is unset.
