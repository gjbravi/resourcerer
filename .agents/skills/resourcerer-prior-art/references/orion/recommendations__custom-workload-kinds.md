# Custom workload kinds

> Configure which Kubernetes kinds Orion recognizes as workloads, including custom resources and operator-owned workloads


Orion optimizes Kubernetes workloads, and the Agent recognizes standard workload kinds by default — declare `workloadResourceKinds` to tell Orion about custom resources or operators.

## When to use

Orion recognizes the built-in Kubernetes workload kinds (`Deployment`, `StatefulSet`, `DaemonSet`, `ReplicaSet`) automatically, while the default-enabled `includeCommonAddOnResourceKinds` flag adds support for common operators like Argo Rollouts, Prometheus operator, Dynatrace, and KEDA.

Configure custom workload kinds if:

- You run a **custom workload resource** that isn't supported by default, such as the GitHub Actions Runner Controller (`AutoscalingRunnerSet`) or the Spark operator (`SparkApplication`).
- You run an **operator that reconciles a standard workload** and reverts direct patches — Orion must patch the operator's custom resource instead.
- You run an **in-house platform operator** that owns your workloads via `ownerReferences` without managing their resource fields.

## Minimal configuration

For a custom workload resource that owns its own pods, add this to your Helm values:

```yaml
workloadResourceKinds:
- apiGroup: example.com
  kind: MyWorkload
  resource: myworkloads
```

For other cases (operator-owned workloads, in-house platform operators, custom resources with nonstandard field locations), see [Examples](#examples).

## Workload model

Orion's workload model has three concepts:

* **Pods** — the workload's replicas and the Agent's starting point.
* **Workload** — the resource that owns the pods, visible in the UI.
* **Patch target** — where Orion applies changes, usually the workload but can be an owner.

## How Orion identifies workloads

To identify a workload, the Agent walks up the pod's `ownerReferences` chain to find the workload and the patch target. At each kind it encounters, it asks:

> *Is this the workload? The patch target? Do I keep climbing?*

Declaring kinds in `workloadResourceKinds` and assigning each a **role** is how you answer those questions — the role determines how the Agent behaves when it encounters that kind during the climb.

<div class="alert alert-info" role="alert">


Climbing requires Kubernetes RBAC, and Orion only has RBAC for declared kinds. If the Agent encounters an undeclared kind, the climb stops with an RBAC failure.
</div>


## Roles

| Role | Use when… | Example kinds |
| --- | --- | --- |
| `workload` (default) | The kind is a scalable workload that owns pods. | `Deployment`, `AutoscalingRunnerSet` |
| `conditionalWorkload` | The kind acts as a workload when standalone but defers to a higher owner when one exists. Use for kinds like `ReplicaSet` that can be either. | `ReplicaSet`, `SparkApplication` |
| `owner` | An operator or other link above the workload in the ownership chain. Orion climbs through owners and patches the topmost one. | `Prometheus`, `OpenTelemetryCollector` |
| `ignoredOwner` | The kind owns workloads but doesn't reconcile their resource fields. Orion stops the climb below it and patches the workload directly. **Note:** if the operator does reconcile resources, those patches will silently fail. | in-house `AppStack`, `PlatformBundle`-style custom resources |

### Role interactions

**The top of the chain is the workload**

The simplest case: the climb reaches the top, and the top is what gets patched. On the left: a `StatefulSet` declared `workload`. On the right: a `ReplicaSet` declared `conditionalWorkload` with nothing above it (so the conditional resolves to "act as workload"). Same outcome from different role declarations.

![Workload identification: two cases where the climb's top tier acts as both workload and patch target](/img/docs/recommendations/role-pairs/role-pair-workload.svg)

**Adding an `owner` shifts the patch target up**

On the left: a `Deployment` with no owner is the workload and the patch target. On the right: the same `Pod → ReplicaSet → Deployment` chain, with an `OpenTelemetryCollector` declared `owner` above the Deployment. The Deployment is still *identified* as the workload, but the climb continues past it and the patch lands on the owner — otherwise Deployment patches would be reverted on the next reconcile.

![Adding an owner above a workload shifts the patch target up](/img/docs/recommendations/role-pairs/role-pair-operator.svg)

**Two ways the climb terminates**

On the left: ARC's `AutoscalingRunnerSet` is declared `workload`, so the climb reaches it and patches there. On the right: Prometheus's `StatefulSet` is the workload (built-in), `Prometheus` is the `owner` that gets patched, and the in-house `AppStack` above is declared `ignoredOwner`, so the climb stops below it.

![Two custom resources at the top of long chains: a workload that receives patches, and an ignoredOwner that blocks the climb](/img/docs/recommendations/role-pairs/role-pair-tops.svg)

## Examples

### Custom resource is a scalable workload

If your custom resource has a scale subresource and owns its pods, declare it as a `workload`. The default role is `workload`, so you can omit `role:`:

```yaml
workloadResourceKinds:
- apiGroup: example.com
  kind: MyWorkload
  resource: myworkloads
```

If the chain from pod to your custom resource goes through custom intermediate types (for example, `pod → MyReplicaSet → MyWorkload`), declare those as `owner` — the Agent will climb through them on its way to your workload:

```yaml
- apiGroup: example.com
  kind: MyReplicaSet
  resource: myreplicasets
  role: owner
```

### Workload is owned by an operator that reverts patches

This is the Prometheus-operator pattern: a `Prometheus` custom resource owns a `StatefulSet`, and patching the `StatefulSet` directly gets reverted by the operator. Tell Orion to patch the operator custom resource instead:

```yaml
workloadResourceKinds:
- apiGroup: monitoring.coreos.com
  kind: Prometheus
  resource: prometheuses
  role: owner
  patchTargetDefaults:
    containers.cpu.requests.patch-path: /spec/resources/requests/cpu
    containers.memory.requests.patch-path: /spec/resources/requests/memory
    containers.cpu.limits.patch-path: /spec/resources/limits/cpu
    containers.memory.limits.patch-path: /spec/resources/limits/memory
```

### Custom resource may be a workload or intermediate

The `SparkApplication`/`ReplicaSet` pattern: standalone, the resource is a workload; when owned by something higher (a `ScheduledSparkApplication`, a `Deployment`), it defers. Use `conditionalWorkload`:

```yaml
- apiGroup: sparkoperator.k8s.io
  kind: SparkApplication
  resource: sparkapplications
  role: conditionalWorkload
```

### In-house operator owns workloads but doesn't reconcile resources

Use this when a platform operator owns workloads but doesn't manage their resource fields. Without configuration, Orion climbs past the workloads to the operator custom resource, which isn't a useful patch target. Declare the operator as `ignoredOwner` so the climb stops below it:

```yaml
- apiGroup: platform.acme.io
  kind: AppStack
  resource: appstacks
  role: ignoredOwner
```

<div class="alert alert-info" role="alert">


Use `ignoredOwner` only when the operator doesn't manage the workload's resource fields — if it does, patches will silently fail.
</div>


## Fields

| Field | Required | Description |
|---|---|---|
| `apiGroup` | Yes | API group (for example, `actions.github.com`). Use `""` for core kinds. |
| `kind` | Yes | Kind, exactly as it appears in `kubectl get` (PascalCase). |
| `resource` | Yes | Plural resource name, lowercase — what `kubectl api-resources` shows and what RBAC rules use. |
| `role` | No | Which role the kind plays. Defaults to `workload`. |
| `patchTargetDefaults` | No | Where to write resource fields, when not at the standard pod-template location. |

### `role`

Optional. Defaults to `workload`. See [Roles](#roles) for descriptions of each value.

### `patchTargetDefaults`

Optional. Defines where Orion writes resource fields when not at the standard pod-template location (`/spec/template/spec/containers/.../resources`).

Usually omitted for `workload` and `conditionalWorkload` entries; usually required on `owner` entries that are the topmost in their chain. Values are either JSON paths or Go templates. Settings here are defaults and can be overridden by per-workload annotations.

Example for an `OpenTelemetryCollector`:

```yaml
- apiGroup: opentelemetry.io
  kind: OpenTelemetryCollector
  resource: opentelemetrycollectors
  role: owner
  patchTargetDefaults:
    containers.cpu.requests.patch-path: /spec/resources/requests/cpu
    containers.memory.requests.patch-path: /spec/resources/requests/memory
    containers.cpu.limits.patch-path: /spec/resources/limits/cpu
    containers.memory.limits.patch-path: /spec/resources/limits/memory
    pod-template.metadata.patch-path: /spec/podAnnotations
    pod-template.metadata.patch-format: '{ "orion.example/recommendation-url": "{{ .RecommendationURL }}" }'
```

`pod-template.metadata.patch-path` tells Orion where to write annotations that trigger a rolling restart. `pod-template.metadata.patch-format` is a Go template for the annotation value. See [Pod template metadata](configure__settings__apply-settings.md#pod-template-metadata).

For operator custom resources with resource fields in nonstandard locations (for example, `Prometheus` uses `/spec/resources/...`), `patchTargetDefaults` tells Orion where to write. Intermediate `owner` entries that are never the topmost in their chain don't need `patchTargetDefaults` — they're climbed past, not patched.

## Troubleshooting

**Orion isn't optimizing my workload**

- Is your workload resource declared with `role: workload`?
- If the pod is owned through intermediate custom resources, are those declared as `owner`? Orion climbs through declared owners — any undeclared link stops the climb.

**My patches are being reverted**

- If an operator owns the workload and reconciles it back, declare the operator with `role: owner` and `patchTargetDefaults`. Orion will then patch the operator's custom resource.

**Orion patches an object that's too high in the chain**

- A broad owner is being treated as the workload. Declare it `ignoredOwner` so Orion stops below it. If the owner does reconcile resource fields, use `role: owner` instead.

**Permission errors in the Agent log about a custom resource**

- The kind needs to be declared. Declaring it grants the Agent the RBAC it needs to read or patch it.
