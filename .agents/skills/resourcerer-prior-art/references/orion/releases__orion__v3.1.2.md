# Version 3.1.2

Site [llms.txt](/llms.txt)

---

### Added

 * **Applier alerts change diffs**: Applier alerts sent to external endpoints when the applier applies recommendations now include full before-and-after diffs for resources that were changed.

 ### Changed

 * **Performance improvements**
 * **Dependency updates**

 ### Fixed

 * **Missing RBAC for Events**: The Applier now includes RBAC to allow creating and patching Events when use of the Edit role is disabled.
 * **Stale cost metric retention**: The Agent now correctly removes cost-related metrics for deleted pods not owned by workload types Orion recognizes. Previously, some of these metrics were retained until the Agent restarted.
