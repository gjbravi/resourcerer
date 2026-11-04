---
title: Permissions
description: What each component's ServiceAccount may read and write in the cluster.
---

The chart creates one ServiceAccount, ClusterRole and ClusterRoleBinding per component (`rbac.create: true`), so each
component can do only its part. The full rules are in
[`charts/resourcerer/templates/rbac.yaml`](https://github.com/gjbravi/resourcerer/blob/main/charts/resourcerer/templates/rbac.yaml).

## controller

Reads the cluster and writes `Recommendation`s. It never changes a workload.

| Resources | Verbs |
|---|---|
| Deployments, StatefulSets, DaemonSets, Argo Rollouts | get, list, watch |
| Namespaces, Pods, ConfigMaps | get, list, watch |
| HorizontalPodAutoscalers | get, list, watch |
| `recommendations` | all |
| `recommendations/status` | get, update, patch |
| Events | create, patch |
| Leases (leader election) | all |

## applier

The only component that changes workloads, and only when it is enabled (`applier.enabled`).

| Resources | Verbs |
|---|---|
| Deployments, StatefulSets, DaemonSets, Argo Rollouts | get, list, watch, **patch** |
| HorizontalPodAutoscalers | get, list, watch, **patch** |
| Namespaces | get, list, watch |
| `recommendations` | get, list, watch |
| `recommendations/status` | get, update, patch |
| `applypolicies`, `clusterapplypolicies` | get, list, watch |
| `applypolicies/status`, `clusterapplypolicies/status` | get, update, patch |
| Events | create, patch |
| Leases (leader election) | all |

Its patches only touch the `resources` of the pod template's containers (server-side apply, field manager
`resourcerer`) and an HPA's CPU target utilization.

## server

Serves the UI and the API. With `readOnly: true` the **patch** verbs below are not granted.

| Resources | Verbs |
|---|---|
| Deployments, StatefulSets, DaemonSets, Argo Rollouts | get, list, watch, **patch** (settings annotations) |
| Namespaces | get, list, watch, **patch** (settings annotations) |
| `recommendations` | get, list, watch, **patch** (apply, revert and regenerate requests) |
| ConfigMaps, Events | get, list, watch |
| Pods, HorizontalPodAutoscalers | get, list |
| `decisionmodels` | get |
| Secrets, in resourcerer's namespace only (a `Role`) | get: the token of a managed [decision model](../guides/decision-models.md) |

## Who can do what, in practice

Without authentication in the UI (see [Exposing the UI](../guides/exposing-the-ui.md)), anyone who reaches the server
acts with its permissions. The server cannot change workload resources itself: it records requests that the applier
carries out after checking the hash, the phase and the apply policies.

Kubernetes RBAC still governs everyone else: a team that can edit `ApplyPolicy` objects in its namespace can make its
own applies stricter, never looser than the `ClusterApplyPolicy` objects the platform team owns. Grant `applypolicies`
to the namespace roles that should manage them, for instance by aggregating into `edit`:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: resourcerer-applypolicy-edit
  labels:
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
rules:
  - apiGroups: ["resourcerer.io"]
    resources: ["applypolicies"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```
