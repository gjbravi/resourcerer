---
title: HTTP API
description: The endpoints the UI uses, under /api/v1, for scripts and integrations.
---

The UI is a client of the server's HTTP API, served on the same Service under `/api/v1`. Scripts can use it too: it is
the simplest way to apply a reviewed recommendation from a pipeline.

!!! warning
    The API is in beta like the rest of resourcerer: responses can gain and lose fields between minor versions.
    `web/src/api.ts` mirrors the response types.

## Authentication

`auth.strategy` decides who may call, and the API follows the UI: see
[Exposing the UI](../guides/exposing-the-ui.md) for both setups.

- `anonymous` (the default) identifies nobody: every call is allowed, with the server's permissions. Keep the
  Service internal, or put an authenticating proxy in front.
- `oidc` needs a session. A browser gets one at `/auth/login`, which sends it to the provider and back to
  `/auth/callback`; `/auth/logout` clears it and lands on `/?signed-out`. Scripts sign in the same way, or reuse the
  `resourcerer_session` cookie the UI holds.

Under `oidc` every route but `me` answers `401` with `"reason": "unauthenticated"` without a session, and `403` with
`"reason": "forbidden"` to a user the rules leave with no role. Mutations also need `X-Resourcerer-Request: 1` and a
body typed as JSON, or they are refused with `403` and `"reason": "ajax-required"`: that is the guard against a page
on another origin riding the session. `anonymous` needs no guard, and sending the header there is harmless.

`{kind}` is the workload kind as in the manifest: `Deployment`, `StatefulSet`, `DaemonSet` or `Rollout`.

## Read

| Method and path | Returns |
|---|---|
| `GET /api/v1/me` | Who is signed in: `authenticated`, `user`, `groups`, `role`, `strategy` and `configStale`, plus `expiresAt` under `oidc`. Always `200`, and never cached. |
| `GET /api/v1/info` | Cluster name, version, settings defaults, cost and reliability configuration, Prometheus status, whether the server is read-only. |
| `GET /api/v1/workloads` | Every managed workload with its phase, score, costs, auto-deploy and reliability summary. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}` | The workload's recommendation: containers, values and basis, impact, evidence, settings with sources, history, apply policy decision. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/series?range=…&container=…&resource=…` | Usage, requests and recommendation history for a chart. `range` is `learning` or a fixed zoom. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/impact` | Cost and requests over time, before and after. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/reliability?window=…` | Live reliability checks over a window: `15m`, `1h`, `6h`, `1d` or `learning`. |
| `GET /api/v1/workloads/{ns}/{kind}/{name}/placement` | Where the workload ran: capacity and instance types with their cost. Cached for 5 minutes. |
| `GET /api/v1/cluster/series` | Cluster-wide requests, usage and recommendations over time. |

## Write

Refused with `403` when the server is read-only (`"reason": "read-only"`), and under `oidc` to anyone who is not an
editor (`"reason": "editor-required"`).

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
curl -s -X POST "$API/workloads/$W/apply" \
  -H 'Content-Type: application/json' -H 'X-Resourcerer-Request: 1' \
  -d "{\"hash\": \"$hash\"}"
```

Passing the hash is what makes an apply safe to automate: if the recommendation was regenerated between the read and
the apply, the applier refuses it instead of applying values nobody looked at. Under `oidc` the script also needs a
session: pass the browser's `resourcerer_session` cookie with `-b`, or sign in from the script and keep the cookie.
