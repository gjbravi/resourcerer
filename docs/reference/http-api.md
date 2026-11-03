---
title: HTTP API
description: The endpoints the UI uses, under /api/v1, for scripts and integrations.
---

The UI is a client of the server's HTTP API, served on the same Service under `/api/v1`. Scripts can use it too: it is
the simplest way to apply a reviewed recommendation from a pipeline.

!!! warning
    The API has the same authentication as the UI: none, for now. Anyone who reaches it can apply and revert. See
    [Exposing the UI](../guides/exposing-the-ui.md). The API is in beta like the rest of resourcerer: responses can gain
    and lose fields between minor versions. `web/src/api.ts` mirrors the response types.


`{kind}` is the workload kind as in the manifest: `Deployment`, `StatefulSet`, `DaemonSet` or `Rollout`.

## Read

| Method and path | Returns |
|---|---|
| `GET /api/v1/info` | Cluster name, version, settings defaults, cost and reliability configuration, Prometheus status, whether the server is read-only. |
| `GET /api/v1/workloads` | Every managed workload with its phase, score, costs, auto-deploy and reliability summary. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}` | The workload's recommendation: containers, values and basis, impact, evidence, settings with sources, history, apply policy decision. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/series?range=…&container=…&resource=…` | Usage, requests and recommendation history for a chart. `range` is `learning` or a fixed zoom. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/impact` | Cost and requests over time, before and after. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/reliability?window=…` | Live reliability checks over a window: `15m`, `1h`, `6h`, `1d` or `learning`. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/placement` | Where the workload ran: capacity and instance types with their cost. Cached for 5 minutes. |
| `GET /api/v1/cluster/series` | Cluster-wide requests, usage and recommendations over time. |

## Write

Refused with `403` when the server is read-only.

| Method and path | Body | Effect |
|---|---|---|
| `POST /api/v1/workloads/{ns}/{kind}/{name}/apply` | `{"hash": "<status.hash>"}` | Requests an apply of the recommendation with that hash. `409` with the reason when an apply policy holds manual applies now. |
| `POST /api/v1/workloads/{ns}/{kind}/{name}/revert` | none | Requests a revert of the last apply. |
| `POST /api/v1/workloads/{ns}/{kind}/{name}/regenerate` | none | Regenerates the recommendation now. |
| `PUT /api/v1/workloads/{ns}/{kind}/{name}/settings` | `{"key": "<setting>", "value": "<value>"}`, or `"value": null` to remove | Sets or removes the `resourcerer.io/<key>` annotation on the workload. `400` for an unknown key or invalid value. |
| `PUT /api/v1/namespaces/{ns}/settings` | same | The same on the namespace. |

An apply or revert is **asynchronous**: the call records the request and returns; the applier carries it out within
seconds. Read `lastApplyResult` from the workload's detail, or the `Recommendation`'s status, to know how it ended.

## Example: apply what was reviewed

```bash
API=http://localhost:8080/api/v1
W=payments/Deployment/checkout

hash=$(curl -s "$API/workloads/$W" | jq -r .hash)
curl -s -X POST "$API/workloads/$W/apply" -H 'Content-Type: application/json' -d "{\"hash\": \"$hash\"}"
```

Passing the hash is what makes an apply safe to automate: if the recommendation was regenerated between the read and
the apply, the applier refuses it instead of applying values nobody looked at.
