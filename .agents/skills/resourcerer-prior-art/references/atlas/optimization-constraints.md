---
updatedAt: 2026-09-04T11:39:21.000Z
---

# Optimization constraints

Advisory workload configuration checks surfaced as part of the Cluster Score — missing probes, Pod Disruption Budget issues, and topology spread constraints.

As part of the [Cluster Score](cluster-score.md), Atlas checks your workloads for configuration issues that affect reliability, availability, and cost optimization. These checks are called **optimization constraints**.

Each constraint flags a specific issue in the workload manifest and points to the location in the YAML where it was detected.

> 📘 Note
>
> Optimization constraints are advisory — Atlas does not modify your workloads to resolve them.

## Constraint types

### Missing health probes

Checks whether containers have **liveness** and **readiness** probes configured.

| Probe type          | Purpose                                                                   |
| ------------------- | ------------------------------------------------------------------------- |
| **Liveness probe**  | Detects hung or deadlocked containers so Kubernetes can restart them.     |
| **Readiness probe** | Prevents traffic from reaching containers that haven't finished starting. |

Readiness probe checks are skipped for **CronJob** and **Job** workloads.

### Pod Disruption Budget (PDB) issues

Checks whether PDBs matching your workloads could block voluntary disruptions such as node drains during rebalancing.

| Issue                             | Description                                                                             |
| --------------------------------- | --------------------------------------------------------------------------------------- |
| **Missing availability settings** | A PDB exists but has neither `minAvailable` nor `maxUnavailable` set.                   |
| **Too strict PDB**                | `maxUnavailable: 0` or `minAvailable` ≥ replica count blocks all voluntary disruptions. |

Applies to workload types that have replicas: Deployments, ReplicaSets, StatefulSets, and ReplicationControllers.

### Topology spread constraints

Checks whether workloads with **3 or more replicas** have topology spread constraints and whether existing constraints have high skew values.

| Issue                                   | Description                                                                                       |
| --------------------------------------- | ------------------------------------------------------------------------------------------------- |
| **Missing topology spread constraints** | No `topologySpreadConstraints` configured, risking uneven pod distribution across zones or nodes. |
| **High max skew**                       | `maxSkew` exceeds the recommended threshold (2 for zone/region, 3 for hostname).                  |

Skipped for **Pod**, **Job**, and **CronJob** workloads.

## Suppressing constraints

To suppress all optimization constraint checks for a specific workload, add this annotation to its **pod template** (`spec.template.metadata.annotations`):

| Annotation                                          | Value                                                            | Where to apply                                                                           | Description                                                                                                                                                     |
| --------------------------------------------------- | ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `reporting.atlas.example/ignore-optimization-constraints` | `"true"` (case-sensitive — `"True"`, `"1"`, `"yes"` are ignored) | Pod template (`spec.template.metadata.annotations`) or bare Pod (`metadata.annotations`) | Suppresses **all** optimization constraint checks; you cannot suppress individual types. Removing the annotation re-enables checks on the next reporting cycle. |

**Example:** the annotation in context, on a controller-managed workload and a bare Pod:

<Tabs>
  <Tab title="Controller (Deployment)">
    Add the annotation to the pod template at `spec.template.metadata.annotations`:

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: my-app
      namespace: default
    spec:
      replicas: 3
      selector:
        matchLabels:
          app: my-app
      template:
        metadata:
          labels:
            app: my-app
          annotations:
            reporting.atlas.example/ignore-optimization-constraints: "true"
        spec:
          containers:
            - name: app
              image: my-app:latest
    ```
  </Tab>

  <Tab title="Bare Pod (no controller)">
    Without a pod template, place the annotation directly on `metadata.annotations`:

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: my-app
      namespace: default
      labels:
        app: my-app
      annotations:
        reporting.atlas.example/ignore-optimization-constraints: "true"
    spec:
      containers:
        - name: app
          image: my-app:latest
    ```
  </Tab>
</Tabs>

To apply this annotation across many workloads at once, you can use [Pod mutations](https://docs.atlas.example/atlas/docs/pod-mutations) to inject it automatically at admission time — useful for enforcing suppression consistently or in GitOps workflows where manifests aren't edited by hand.