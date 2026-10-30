---
name: resourcerer-design-issue
description: Write a design issue for a resourcerer feature before building it, such as a new custom resource, a policy, a cost or reliability capability. Use when a change is larger than a small fix or touches the CRDs, the chart values or how decisions are made.
---

# Design issues

Larger features start as a GitHub issue that others can review. The roadmap issues (#26 to #31) follow this shape;
read the ones next to your feature first and keep their conventions.

## Structure

1. **Summary:** one paragraph: what and why.
2. **Motivation:** what happens today, with the concrete gaps. Point to the current code or setting by name.
3. **Background**, when the domain is not obvious (for example how Savings Plans apply, how a model is called).
4. **Proposed design:** a full YAML example of the resource or values, then the rules: how it is evaluated, how
   several objects combine, what is stored in status, what the UI shows.
5. **Simple examples:** a handful of minimal configurations for common cases, each a few lines, plus what an
   omitted field means.
6. **Related:** links to the other design issues it interacts with.
7. **Phases:** the order to ship it in, each phase useful on its own.
8. **Open questions:** what still needs a decision.

## Conventions for custom resources

- A cluster-scoped kind for the platform (`Cluster...`) and a namespaced kind for teams, which only affects its own
  namespace.
- `namespaceSelector` on the cluster kind, `workloadSelector` and `kinds` on both.
- State how several matching objects combine: guardrails take the **most restrictive** value (windows intersect,
  limits take the minimum); preferences take the **most specific** level, with `priority` and then the name breaking
  ties, and conflicts reported in a condition.
- Status with conditions (`Valid`, `Ready`, or `Effective` when combination matters) and counts of matched workloads.
- Validation in the CRD with CEL rules where possible, so no webhook is needed.
- RBAC: a ClusterRole for editing the namespaced kind, optionally aggregated into `edit`; overrides that bypass a
  guardrail are a separate permission.
- Helm values that configure the same thing render the cluster-scoped objects, so values and resources do not drift.
- Credentials never come from Helm values: Kubernetes Secrets synced from a secrets manager, or IRSA / workload
  identity for cloud APIs. Say what data leaves the cluster.
- A feature that fails degrades to today's behaviour and says why.

## Writing

English, concrete, fictional examples (`payments/checkout`, `demo/api`), nothing specific to a company's
environment. Label `enhancement`.
