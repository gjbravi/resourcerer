# February 2026

Site [llms.txt](/llms.txt)

---

- **Recommendations applied without restarting Pods**\
The webhook apply methods now use Kubernetes in-place Pod resizing by default, updating requests and limits on running Pods instead of triggering a rollout. Requires Applier v2.12.0 or later and Kubernetes v1.33 or later; on earlier Kubernetes versions these methods fall back to a rollout or defer the change until the next restart. See [Apply settings].

[Apply settings]: /docs/configure/settings/apply-settings/
