---
updatedAt: 2026-10-25T20:35:32.000Z
---

# Troubleshooting

Troubleshoot common Storage Autoscaling issues including PV not growing, no IOPS changes, permission errors, and cooldown problems.

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

## Persistent volume is not growing despite being over the threshold

* Confirm the StorageClass used by the PV has `allowVolumeExpansion: true`:

```shell
kubectl get storageclass -o yaml | grep allowVolumeExpansion
```

* Check whether a cooldown is active:

```shell
kubectl get pvc <name> -o jsonpath='{.metadata.annotations.storage\.atlas\.example/last-capacity-modification}'
```

* Check which PVPolicy matched the PVC and whether it is in dry-run:

```shell
kubectl get pvc <name> -o jsonpath='{.metadata.annotations.storage\.atlas\.example/matched-pvpolicy}'
kubectl get pvc <name> -o jsonpath='{.metadata.annotations.storage\.atlas\.example/dry-run}'
```

If the annotation is empty, no PVPolicy matched. Verify your policy selector and namespace. If the matched policy has `dryRun: true`, planned actions are logged but not executed.

* Ensure the `atlas-storage-autoscaler` service account has `nodes/proxy` RBAC access.
* Check whether the PVC was explicitly excluded. Look for `storage.atlas.example/storage-enabled: "false"` on the PVC, or a restrictive `config.pvc.excludeLabels` / `config.namespaces.watch` scoping. Storage management is enabled by default.

## No IOPS or throughput changes being made

* Verify Sentinel is installed with `agent.extraArgs.storage-stats-enabled=true` and the cloud provider is configured. See [Configuring Sentinel: Kubernetes storage metrics](sec-configuring-sentinel.md#kubernetes-storage-metrics).
* Verify GCP Workload Identity Federation grants access to Cloud Monitoring and Compute APIs (`roles/compute.storageAdmin`, `roles/monitoring.viewer`), or that the AWS IRSA role is correctly attached.
* Check `storage.atlas.example/disk-type`. Only Hyperdisk types, PD Extreme, and AWS gp3/io1/io2 support independent IOPS tuning.
* Check whether the PVC was explicitly opted out via `storage.atlas.example/iops-enabled: "false"` or `storage.atlas.example/throughput-enabled: "false"`.
* Check whether the matched PVPolicy has `iopsEnabled` or `throughputEnabled` set to false:

```shell
kubectl get pvpolicy <policy-name> -n atlas-agent -o jsonpath='{.spec.config.dimensions}'
```

* Check the 4-hour GCP cooldown:

```shell
kubectl get pvc <name> -o jsonpath='{.metadata.annotations.storage\.atlas\.example/last-iops-modification}'
```

* On AWS, check `storage.atlas.example/modifications-history`. If the array already has 4 recent entries, the volume has used its full 24-hour ModifyVolume budget.
* Verify that the minimum warmup window has elapsed. Storage Autoscaling needs at least `minWarmupWindow` (default: 30 minutes) of accumulated metric data before acting.

## No PVPolicy matching a PVC

Symptom: a PVC has no `storage.atlas.example/matched-pvpolicy` annotation, or logs show no policy matched.

Causes and fixes:

* Verify PVPolicy CRs exist in the `atlas-agent` namespace:

```shell
kubectl get pvpolicy -n atlas-agent
```

* Verify the policy's namespace selector includes the PVC's namespace. An empty `selector.namespaces` list matches all namespaces; a non-empty list restricts matching to the listed namespaces.
* Verify the PVC matcher is not too restrictive. Check `selector.pvc.names`, `selector.pvc.labels`, and `selector.pvc.storageClasses`. All non-empty fields are AND-ed, so the PVC must satisfy every condition.
* If no policy matches, Storage Autoscaling falls back to global Helm config but does not set the `matched-pvpolicy` annotation.

> 📘 Note
> Workload-based selectors are not yet supported. If your policy relies solely on `selector.workload`, it will not match any PVCs. Use `selector.pvc` instead. See the [PVC selector](storage-configuration.md#pvc-selector) reference for details.

## Check for errors on a specific PVC

If a resize or performance change fails, Storage Autoscaling records the error in a PVC annotation. Check it with:

```shell
kubectl get pvc <pvc-name> -o jsonpath='{.metadata.annotations.storage\.atlas\.example/last-error}'
```

An empty result means no error has been recorded. If the annotation contains an error message, it typically points to a permissions issue, a cloud-enforced cooldown, or an unsupported disk type.

## Permission errors on AWS

* Check the `atlas-storage-autoscaler` logs for access errors:

```shell
kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-storage-autoscaler | grep -E "AccessDenied|Unauthorized"
```

* The most common miss is `cloudwatch:GetMetricData`. Without it, observed IOPS/throughput stays at 0 and there are never enough datapoints to act on.
* Confirm the pod's service account has the `eks.amazonaws.com/role-arn` annotation set, and that the trust policy's `Federated` principal matches the cluster's actual OIDC provider ARN. The `StringEquals` condition must match `system:serviceaccount:atlas-agent:atlas-storage-autoscaler`.
* Verify the pod has AWS credential environment variables injected (`AWS_ROLE_ARN`, `AWS_WEB_IDENTITY_TOKEN_FILE`). The IRSA webhook injects these when the service account annotation is present. Their absence usually means the annotation was not applied to the service account or the webhook is not active on the cluster.

## Re-enable dry-run

You can switch back to dry-run mode at any time by patching the PVPolicy:

```shell
kubectl patch pvpolicy default -n atlas-agent --type=merge -p '{"spec":{"dryRun":true}}'
```

For a complete list of configuration options, see [Configuration](storage-configuration.md).

## Related resources

<Cards columns={2}>
  <Card title="Configuration" href="https://docs.atlas.example/docs/storage-configuration" icon="fa-sliders">
    PVPolicy field reference, PVC annotations, and Helm values.
  </Card>
  <Card title="How it works" href="https://docs.atlas.example/docs/storage-how-it-works" icon="fa-cogs">
    Learn how capacity expansion and IOPS/throughput rightsizing work.
  </Card>
</Cards>