# Applier configuration

> Configure the Orion Applier component


This topic describes the additional parameters available for configuring the Orion Applier. For parameters that apply to the Orion Agent, see the [Agent configuration](installation__cluster-agent-config.md) topic.

<div class="alert alert-info" role="alert">


The Applier is enabled by default. To disable it, add `--set applier.enabled=false` to your Helm install or upgrade command.
</div>


## Applying these settings

These are Helm values for the `orion` chart. Add them to a YAML values file (for example, `values.yaml`) and pass it to your Helm command.

At install time:

```shell
helm install orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  -f values.yaml
```

To apply changes after installation:

```shell
helm upgrade orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  --reset-then-reuse-values \
  -f values.yaml
```

---


**Applier settings:**

- [Continuous reconciliation](#continuous-rec)
- [Additional RBAC permissions](#rbac)
- [Applier alerts](#alerts)
- [Auto-deploy exempt namespaces](#auto-deploy-exempt)
- [Suspend the Applier](#suspend)
- [Pod priority and scheduling](#pod-priority)

---

## Continuous reconciliation {#continuous-rec}

Controls resource patching behavior based on CI/CD and cluster activity. [Auto-deploy](configure__settings__apply-settings.md#auto-deploy) must be enabled.

| Parameter                                    | Default value |
| -------------------------------------------- | ------------- |
| `applier.reconciliation.mode`                | `EventBased`  |
| `applier.reconciliation.conflictingManagers` | None          |

### Description {#continuous-rec-description}

The Orion Applier can be configured to continuously reconcile drift, ensuring that the latest recommended settings are always maintained on a workload.

The Applier logs whenever it reconciles a workload's resource settings to the recommended values.

### Valid values {#continuous-rec-valid-values}

- `EventBased` (default): Workload requests and limits are updated only by relevant Orion events that trigger an update: auto deployment of a new recommendation, clicking **Apply Now** in the UI, or running `orion apply`. Subsequent updates to requests and limits by GitOps tools will override the recommended settings.

- `Continuous`: Recommended requests and limits values are maintained and not overwritten during CI/CD or deployment activity on the cluster, except when changes are made by the declared `applier.reconciliation.conflictingManagers`.

    - `applier.reconciliation.conflictingManagers`: Optional, honored only when `applier.reconciliation.mode=Continuous`. Declares which field managers conflict with the Applier. The Applier will not reconcile the desired state for fields owned by conflicting managers. Use this if you cannot configure your existing field manager to accept or ignore changes from Orion.

<div class="alert alert-info" role="alert">


Continuous reconciliation doesn't run for workloads using a webhook-based apply method. See [Webhook configuration][webhook-config] for details.
</div>


---

## Additional RBAC permissions {#rbac}

Grants additional RBAC permissions to the Applier.

### Description {#rbac-description}

To optimize resources that are not standard Kubernetes workload types, the Applier needs additional RBAC permissions on the API group and resources.

### Example {#rbac-example}

Add the API groups and their resources to your values file as shown below.

```yaml
applier:
  rbac:
    additionalRules: |-
      - apiGroups:
        - "YOUR_API_GROUP"
        resources:
        - "YOUR_RESOURCES"
        verbs:
        - get
        - list
        - watch
        - patch
        - update
```

---

## Applier alerts {#alerts}

Optionally send notifications to external endpoints when the Applier applies recommendations.

| Parameter                        | Default value |
| -------------------------------- | ------------- |
| `applier.alerts.endpoints`       | None          |

### Description {#alerts-description}

When `applier.alerts.endpoints` is configured with one or more entries, the Applier sends an HTTP request to each endpoint whenever it applies a recommendation. Useful for audit logging, Slack notifications, or triggering downstream workflows.

If your endpoint requires credentials via a file, mount the file using the top-level `additionalVolumes` and `additionalVolumeMounts` parameters.

### Example {#alerts-example}

```yaml
applier:
  alerts:
    endpoints:
      - name: my-webhook
        url: https://hooks.example.com/orion
        method: POST
        contentType: application/json
        authorization:
          type: Bearer
          credentialsFile: /tmp/credentials
```

For full configuration details, see the [Configure applier alerts][applier-alerts] guide.

---

## Auto-deploy exempt namespaces {#auto-deploy-exempt}

Namespaces that inherit the cluster-default schedule but **not** the auto-deploy setting.

| Parameter                                     | Default value                       |
| --------------------------------------------- | ----------------------------------- |
| `clusterDefaultsAutoDeployExemptNamespaces`   | `["^kube-system$", "^openshift.*"]` |

### Description {#auto-deploy-exempt-description}

By default, namespaces matching `^kube-system$` or `^openshift.*` inherit the cluster-default recommendation schedule, but their auto-deploy setting remains `"Disabled"` regardless of the cluster-default auto-deploy value. Add entries to this list to protect additional namespaces from automatic deployment when auto-deploy is enabled at the cluster level. Namespace and workload annotations can still override this and enable auto-deploy for individual namespaces or workloads.

### Example {#auto-deploy-exempt-example}

```yaml
clusterDefaultsAutoDeployExemptNamespaces:
  - "^kube-system$"
  - "^openshift.*"
  - "^my-critical-namespace$"
```

---

## Suspend the Applier {#suspend}

Temporarily pauses Applier activity without uninstalling.

| Parameter         | Default value |
| ----------------- | ------------- |
| `applier.suspend` | `""`          |

### Description {#suspend-description}

Use `applier.suspend` during maintenance windows or change freezes to temporarily stop the Applier from processing new recommendations. The value is read at startup — changes require a Pod restart.

### Valid values {#suspend-valid-values}

- Unset (default): normal operation.
- `NewRecommendations`: change freeze — the Applier ignores new recommendations. The webhook continues to inject previously applied values into new pods, and continuous reconciliation (if enabled) continues to enforce them.
- `All`: fully disengaged — the Applier ignores new recommendations and the webhook stops injecting applied values.

### Example {#suspend-example}

```yaml
applier:
  suspend: "NewRecommendations"
```

---

## Pod priority and scheduling {#pod-priority}

Enables scheduling and priority configurations on Orion Applier pods.

| Helm parameter               | Default value | Scope                   |
| ---------------------------- | ------------- | ----------------------- |
| `priorityClassName`          | `""`          | Top-level               |
| `applier.priorityClassName`  | `""`          | Applier only (override) |

### Description {#pod-priority-description}

The top-level `priorityClassName` applies to all Orion components. Set `applier.priorityClassName` to override it for the Applier only. A non-empty per-component value **replaces** the top-level default.

`applier.priorityClassName` assigns an existing Pod `PriorityClass` to the `orion-applier` pod.

### Valid values {#pod-priority-valid-values}

A string that matches the `.metadata.name` of a `PriorityClass` defined in your cluster.

### Example {#pod-priority-example}

To apply a priority class to all Orion components:

```yaml
priorityClassName: "high-priority"
```

To override for the Applier only:

```yaml
applier:
  priorityClassName: "high-priority"
```

## Related topics
- [Orion Applier description][applier-desc]
- [Webhook configuration](installation__webhook-config.md)
- Kubernetes official documentation:
  - [PriorityClass](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#priorityclass-v1-scheduling-k8s-io)
  - [Pod Priority and Preemption](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)

[apply-method]: /docs/configure/settings/apply-settings/#apply-method
[cluster]: /docs/configure/cluster-defaults/
[namespaces]: /docs/configure/annotate-namespaces/
[workloads]: /docs/configure/annotate-workloads/
[applier-desc]: /docs/concepts/#applier
[cont-rec-applier]: /docs/installation/applier-config/#continuous-rec
[applier-alerts]: /docs/guides/applier-alerts/
[webhook-config]: /docs/installation/webhook-config/
