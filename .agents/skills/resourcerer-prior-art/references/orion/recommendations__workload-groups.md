# Workload groups

> Aggregate metrics from related workloads to improve recommendations for blue/green, canary, and other split-traffic deployment patterns


Workload groups let you aggregate usage metrics from related workloads, so Orion generates appropriate recommendations even when a workload has temporarily low or split traffic.

## Overview

Orion generates recommendations based on observed resource usage. When workloads share a logical traffic pattern but run as separate deployments — blue/green pairs, canary variants, or ephemeral workloads — the inactive variants can report near-zero usage. This can lead to undersized recommendations for a variant that should align with a busier workload.

Workload groups solve this problem by letting you define related workloads. Orion aggregates the group's metrics before computing a recommendation, so a workload's recommendation reflects the group's combined resource usage rather than just its own observed metrics.

Workload group configuration is per-workload: configuring a group for workload A doesn't affect workload B's recommendations unless B is also configured with a group.

## How it works

A workload group is configured on a **target workload** — the workload a recommendation is being generated for. Workloads that match the configuration become **members** of the group. The target workload is also a member by default. At recommendation time, Orion aggregates CPU, memory, replica counts, and OOM events across all members, then uses the aggregate to compute the target workload's recommendation.

Set `workload-group.mirror: true` to exclude the target workload from its own group — useful when the recommendation should be based entirely on another workload's metrics. In mirror mode, Orion computes the recommendation from the other members only. See [Mirror another workload's recommendations](#mirror-another-workloads-recommendations).

In the Orion UI, recommendation charts include a toggle between **Individual** (target workload only) and **Aggregated** (group-wide) views.

## Configuration

Configure workload groups using the `workload-group` settings in your workload configuration.

| Annotation | Description | Default |
| ---------- | ----------- | ------- |
| `live.orion.example/workload-group.selector` | Kubernetes label selector matching workloads to include | - |
| `live.orion.example/workload-group.expression` | [CEL expression](#cel-expressions) matching workloads to include | - |
| `live.orion.example/workload-group.name` | Name of a workload to include | Target workload's name |
| `live.orion.example/workload-group.cluster` | Cluster of a workload to include | Target workload's cluster |
| `live.orion.example/workload-group.namespace` | Namespace of a workload to include | Target workload's namespace |
| `live.orion.example/workload-group.resource` | Resource type of a workload to include | Target workload's resource type |
| `live.orion.example/workload-group.mirror` | Exclude the target workload from its own group ([mirror mode](#mirror-another-workloads-recommendations)) | `false` |

For most use cases — blue/green pairs, canary deployments, and ephemeral workloads — use `expression` or `selector` to identify related workloads:

- `expression` is the more flexible option. Because a [CEL expression](#cel-expressions) can reference the target workload via `target`, the same expression can be applied to many workloads — for example, through a [`ClusterOptimizationConfiguration`][cluster-opt-config] rule — and each resolves its own group independently.
- `selector` matches by label, so it works best when all workloads in the group share a common label.

The identity fields (`name`, `cluster`, `namespace`, `resource`) are most useful in [mirror mode](#mirror-another-workloads-recommendations), where a workload's recommendation should come from a specific other workload.

All configured fields combine to narrow the match. `mirror` is independent of matching and controls only whether the target workload is also a member.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Verify the resolved group</div>


To see which workloads will be aggregated for a given workload's next recommendation, run:

```sh
orion describe workload <workload>
```
</div>


## CEL expressions

The `workload-group.expression` field accepts a [CEL][cel] expression for matching workloads by their attributes, with two available variables:

- `target` — the target workload (the workload to generate the recommendation for)
- `candidate` — a candidate workload being evaluated for inclusion in the group

The expression should return `true` to include the candidate. Both variables expose the Orion representation of the workload, including the fields `name`, `namespace`, `cluster`, and `resource`.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">CEL expression overrides</div>


Referencing `candidate.cluster`, `candidate.namespace`, `candidate.resource`, or `candidate.name` in a CEL expression overrides the corresponding identity field. Where a CEL reference and an identity field would otherwise conflict, the CEL reference wins.
</div>


## Examples

### Blue/green deployments

A blue/green deployment runs two variants — typically `<app>-blue` and `<app>-green` — and shifts traffic between them. Ungrouped, the inactive variant would receive recommendations sized for near-zero traffic.

Use a [CEL expression](#cel-expressions) to match the other variant by name:

```yaml
live.orion.example/workload-group.expression: 'candidate.name.matches(target.name.find(".*-") + "(blue|green)")'
```

This expression works on `<app>-blue` and `<app>-green`, so a single `workload-group` configuration applies across every blue/green pair in a cluster.

If your blue/green pairs share a common label, use a `selector`:

```yaml
live.orion.example/workload-group.selector: "app=nginx"
```

Either option results in recommendations that reflect the combined traffic regardless of which variant is currently active.

### Canary deployments

Progressive delivery tools like [Flagger][flagger] create an `<app>-primary` deployment that serves live traffic, with an `<app>` deployment used as the canary during rollouts. The canary may run at near-zero replicas between rollouts, making its standalone recommendation undersized for production traffic.

Include the paired deployment in the group with a [CEL expression](#cel-expressions):

```yaml
live.orion.example/workload-group.expression: 'candidate.name + "-primary" == target.name || target.name + "-primary" == candidate.name'
```

This expression works whether the target workload is the canary (`podinfo`) or the primary (`podinfo-primary`), so the same configuration applies to every pair.

If the canary and primary share a label — Flagger preserves the canary's labels on the primary by default — you can use a `selector`:

```yaml
live.orion.example/workload-group.selector: "app=podinfo"
```

### Ephemeral workloads

Some deployment pipelines produce logically identical workloads with unique suffixes such as hashes, timestamps, or build IDs. Because names change on every deployment, matching by the target workload's literal name isn't possible.

When these workloads share a label, define a workload group using a `selector`. For example, with [Knative revisions][knative-revisions], each configuration change creates a new `Deployment` with a unique name, but all service revisions carry the label `serving.knative.dev/service=<service>`:

```yaml
live.orion.example/workload-group.selector: "serving.knative.dev/service=my-service"
```

When the workloads don't share a label, match on a stable name prefix with a [CEL expression](#cel-expressions). For workloads named `<app>-<hash>`, strip the suffix from the target workload's name to get the prefix, then match any workload sharing it:

```yaml
live.orion.example/workload-group.expression: 'candidate.name.startsWith(target.name.substring(0, target.name.lastIndexOf("-") + 1))'
```

All matching workloads contribute metrics to the group, so the target workload's recommendation remains stable across deployments.

### Mirror another workload's recommendations

To compute a workload's recommendation from another workload's metrics only, set `mirror: true` along with identity fields pointing to a single workload to mirror.

Use mirror mode for:

- A newly deployed workload that doesn't yet have representative usage data but resembles an existing workload.
- A low-traffic or staging workload that should match a busier or production workload.
- Mirroring recommendations from a primary cluster to a standby cluster to keep both environments consistently sized.

Mirror a workload in the same namespace:

```yaml
live.orion.example/workload-group.name: "production-api"
live.orion.example/workload-group.mirror: "true"
```

Mirror a workload with the same name in a different cluster:

```yaml
live.orion.example/workload-group.cluster: "prod-us-east"
live.orion.example/workload-group.mirror: "true"
```

Keep the target workload's other settings, such as bounds and goals, the same as the workload it mirrors. Orion still applies these settings, so differences can prevent the recommendations from matching.

<div class="alert alert-warning" role="alert"><div class="h4 alert-heading" role="heading">Mirror mode and scale point preservation</div>


Mirror mode doesn't support [`hpa.preserve-horizontal-scale-point`][preserve-scale-point]. Leave it set to `False` on both the target workload and the workload it mirrors.
</div>


[cel]: <https://cel.dev>
[cluster-opt-config]: /docs/configure/optimization-configurations/
[preserve-scale-point]: /docs/configure/settings/hpa/#preserve-scale-point
[flagger]: <https://flagger.app>
[knative-revisions]: <https://knative.dev/docs/serving/revisions/>
