# Applier Kubernetes events

> Kubernetes Events emitted by the Applier and how to monitor them


The Orion Applier emits standard Kubernetes Events. If your team uses `kubectl`, Lens, k9s, or a cluster event exporter to track cluster activity, you can track Applier activity without any additional configuration.

<div class="alert alert-info" role="alert">


For push-based notifications when the Applier applies a recommendation, see [Configure applier alerts][alerts].
</div>


## View events

Use `kubectl` to query events on any Applier-managed workload:

```sh
# Events on a specific workload
kubectl get events -n <namespace> --field-selector involvedObject.name=<workload-name>

# Filter by event reason (substitute any reason from the reference below)
kubectl get events -n <namespace> --field-selector reason=RecommendationApplied

# Watch for a specific reason in real time
kubectl get events -n <namespace> -w --field-selector reason=RecommendationApplied
```

Events also appear in `kubectl describe deployment <name>` (or `statefulset`, `daemonset`, etc.) under the `Events:` section.

### Event output

When [auto-apply][auto-deploy] applies a recommendation to a Deployment:

```
$ kubectl get events -n hipster \
    --field-selector involvedObject.name=cartservice,reason=RecommendationApplied

LAST SEEN   TYPE     REASON                  OBJECT                   MESSAGE
2m          Normal   RecommendationApplied   Deployment/cartservice   Orion successfully applied recommendation patches
```

Each event includes:

| Field | Description |
| ----- | ----------- |
| `type` | `Normal` or `Warning` |
| `reason` | A stable string you can key on for dashboards and alerts (see reference below) |
| `message` | A human-readable description |
| `involvedObject` | The workload (Deployment, StatefulSet, etc.) or HPA the event describes |

## Event reference

### Recommendation lifecycle

Events are attached to the affected workload.

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `RecommendationApplied` | Normal | Orion successfully applied recommendation patches | Patches applied to the workload |
| `RecommendationApplyFailed` | Warning | Orion failed to apply recommendation patches | Patch application encountered an error |
| `RecommendationExpired` | Warning | Orion timed out waiting to process this recommendation | Processing timeout exceeded before patches could be applied |

### Resource changes

Events are attached to the patched object, for example, the updated Deployment or StatefulSet.

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `WorkloadResourcesChanged` | Normal | Orion applied resource changes *(: summary if available)* | Resource changes recorded by the Applier; the optional summary lists changed fields, for example: `server cpu-requests 100m→50m, server memory-requests 256Mi→128Mi` |

`WorkloadResourcesChanged` fires once per patched object. `RecommendationApplied` fires once per recommendation after all objects are patched. Both fire for the same apply operation.

### Rollout validation

After applying patches, the Applier validates that the workload becomes healthy. These events fire on the affected workload when validation does not succeed.

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `RolloutValidationFailed` | Warning | Orion rollout validation failed: *details* | Workload unhealthy after patches applied; rollback follows unless disabled |
| `RolloutValidationTimedOut` | Warning | Orion timed out waiting for the rollout to become healthy | Validation window exceeded; patches are left in place |
| `RolloutValidationAbandoned` | Warning | Orion rollout validation failed with an unrecoverable error | Unrecoverable error during validation |

### Rollback

Events are attached to the affected workload.

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `RecommendationRolledBack` | Warning | Orion rolled back recommendation patch: *details* | Patches reverted after failed rollout validation |
| `RecommendationRollbackFailed` | Warning | Orion failed to roll back patches after validation failure | Rollback attempt also failed |

### In-place pod resize

Events are attached to the affected workload. These fire only when the Applier attempts to [resize pods without triggering a rollout][apply-method].

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `InPlaceResizeSucceeded` | Normal | Orion successfully resized all pods in place | All pods in the workload resized without a rollout |
| `InPlaceResizeFailed` | Warning | Orion in-place pod resize failed, falling back to regular patches | In-place resize failed; Applier retries with a standard patch |

### Drift reconciliation

The Applier watches workloads and HPAs it manages. If an external controller modifies fields the Applier owns, the Applier detects the change and re-applies the last recommendation. Workload and HPA targets share the same reason strings but produce distinct messages.

**Workload**

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `RecommendationDriftDetected` | Warning | Orion detected drift: fields modified by *manager* | An external manager modified workload fields owned by the Applier |
| `RecommendationDriftReconciled` | Normal | Orion reconciled drift to match the applied recommendation | Applied recommendation re-applied to the workload after drift |

**HPA**

| Reason | Type | Message | When it fires |
| ------ | ---- | ------- | ------------- |
| `RecommendationDriftDetected` | Warning | Orion detected HPA drift: fields modified by *manager* | An external manager modified HPA fields owned by the Applier |
| `RecommendationDriftReconciled` | Normal | Orion reconciled HPA drift to match the applied recommendation | Applied recommendation re-applied to the HPA after drift |

### Operational

| Reason | Type | Message | Target | When it fires |
| ------ | ---- | ------- | ------ | ------------- |
| `SuspendState` | Warning | applier started with suspend=*value* | Applier Deployment | Applier started with a non-default [suspend configuration][suspend-config] |

## Use events with your existing tooling

These events work with any standard Kubernetes tooling: dashboards such as Lens and k9s, cluster event exporters (Datadog, New Relic, Dynatrace), Prometheus via an event exporter, or custom automation keyed on the stable `reason` field.

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Guides<span class="related-topics-divider">/</span><a href="/docs/guides/applier-alerts/">Configure applier alerts</a>
                </h5>
                <p>Send HTTP webhook notifications when the Orion Applier applies a recommendation to a workload</p>
            </div>
        </div>


[alerts]: /docs/guides/applier-alerts/
[auto-deploy]: /docs/configure/settings/apply-settings/#auto-deploy
[apply-method]: /docs/configure/settings/apply-settings/#apply-method
[suspend-config]: /docs/installation/applier-config/#suspend
