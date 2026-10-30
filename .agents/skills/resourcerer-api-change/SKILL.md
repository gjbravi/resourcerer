---
name: resourcerer-api-change
description: Add or change a field in the Recommendation CRD status or spec, or in the HTTP API the UI uses. Use when the controller needs to record something new, or the UI needs data it does not get yet.
---

# Changing the CRD or the API

## CRD (`api/v1alpha1/recommendation_types.go`)

1. Add the field with a doc comment and markers: `// +optional` plus `json:",omitempty"` for anything not always
   set; `+kubebuilder:validation:Enum=...` for closed sets.
2. Store numbers that need precision as decimal strings (as `Totals` and `Pricing.Factor` do) and percentages as
   `int32`. Keep the status small: it is written on every regeneration and read by every list call. Large or
   time-series data belongs in a server endpoint computed on request.
3. `make generate`, then commit `zz_generated.deepcopy.go` and `charts/resourcerer/crds/...` with the change;
   `make verify-generate` compares with git.
4. Fill the field in the controller (`internal/recommender`) or the applier (`internal/applier`) only; the server
   never writes status.
5. Breaking changes are acceptable in beta; say so in the PR description so the release notes carry it.

## HTTP API (`internal/server`)

1. Add the field to the response type in `internal/server/types.go`. Reuse CRD types (`*v1alpha1.Pricing`) instead of
   copying them when the shape is the same.
2. New endpoints go under `/api/v1/workloads/{ns}/{kind}/{name}/...` for per-workload data, registered in
   `Handler()`, using `s.get(w, r)` to load the Recommendation.
3. Live queries: cache with `s.cache.get(key, ttl, fn)`; include the window end, truncated to the cache period, in the
   key. Return a structured "unavailable, because …" response instead of an HTTP error when the feature does not
   apply (no Karpenter, no data).
4. Mirror the type in `web/src/api.ts` and add the call to `api` and a hook in `hooks.ts`.

## Tests

- Engine or pure logic: table tests.
- Server: handler tests where they exist; otherwise an e2e check that the endpoint answers with the new field
  (`curl -fs "$API/..." | jq -e`).
- UI: tests for the pure function that interprets the new field.
