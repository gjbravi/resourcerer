---
title: Argo Rollouts
description: How resourcerer recommends and applies for Argo Rollouts, with an inline template or a workloadRef.
---

Rollouts are picked up automatically when the `rollouts.argoproj.io` CRD is installed. resourcerer reads them as
unstructured objects, so it does not depend on a particular Argo Rollouts version.

## A Rollout with its own template

A Rollout with `spec.template` is handled like a Deployment: it gets a `Recommendation` named `rollout-<name>`, and
applying patches its template.

Applying starts a **progressive rollout** that follows the Rollout's strategy: the canary steps, analysis and pauses, or
the blue-green preview and promotion. It is not an immediate restart of every pod. A paused or aborted rollout stops a
resource change like any other change; an analysis that fails on it aborts it, which brings back the previous
ReplicaSet with the previous requests.

## A Rollout with `workloadRef`

A Rollout that references a Deployment with `spec.workloadRef` gets the recommendation; the referenced Deployment does
not get one of its own, since its pods are the Rollout's. Applying patches the **referenced Deployment's template**
(`status.templateRef` on the `Recommendation` names it), which Argo Rollouts then rolls out with the Rollout's strategy.

## HPAs

An HPA targeting a Rollout is handled like one targeting a Deployment: with CPU utilization, its target is adjusted with
the requests.

## Usage history across revisions

Usage is joined to the Rollout through its ReplicaSets, so pods of earlier revisions, canaries and blue-green previews
count in the history of the same workload.

## Apply policies and rollouts in progress

A workload is never applied on top of a rollout in progress, and an apply counts as in progress until its rollout
finishes. A Rollout paused at a canary step is in progress, so rate limits keep counting it and the next automatic
apply for it waits. Promote or abort it to free the slot.
