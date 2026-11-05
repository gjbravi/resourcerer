# HPA reconciliation

Site [llms.txt](/llms.txt)

---

The applier has a reconciliation mode in which it will watch objects associated with workloads that have recommendations applied, and automatically reconcile changes to keep workload settings in sync with the applied recommendations after events such as `helm upgrade`.

When using the webhook to apply recommendations, the only resource type that needs reconciliation is the HPA. The latest applier supports enabling reconciliation for individual resource types.

To enable reconciliation for HPA resources only, use the following Helm values when installing or upgrading the applier:

```yaml
reconciliation:
  controllers:
  - resource: hpa
    enabled: true
```

For more information on the base reconciliation functionality, see the [product documentation](installation__applier-config.md#continuous-rec).
