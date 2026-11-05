# Agent configuration

> Configure the Orion Agent component and shared top-level parameters


This topic describes the additional parameters available for configuring the Orion Agent, as well as shared top-level parameters that apply to all Orion components. For parameters that apply to the Orion Applier, see the [Applier configuration][applier-config] topic.

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


**Agent parameters:**

- [Third-party secret management](#manage-auth-secret)
- [Installing on Red Hat OpenShift](#openshift)
- [Namespace allowList and denyList](#namespace-lists)
- [Additional volumes and volume mounts](#pods-mounts-volumes)
- [Pod priority and scheduling](#pod-priority)
- [GKE Autopilot](#gke-autopilot)
- [Cluster labels and provider](#cluster-labels)
- [Network policies](#network-policies)

---

## Third-party secret management {#manage-auth-secret}

Controls whether the Orion Agent creates and manages its authorization secret.

| Helm parameter        | Default value |
| --------------------- | ------------- |
| `createAuthSecret`    | `true`        |

### Description {#manage-auth-secret-description}

Set to `false` to manage the Agent authorization secret with a third-party credential manager. For details and the Agent secret format, see [Manage secrets with a third-party credential manager](guides__manage-secrets.md).

### Valid values {#manage-auth-secret-valid-values}

- `true` (default)
- `false`

---

## Installing on Red Hat OpenShift {#openshift}

Required only when installing on Red Hat OpenShift.

| Helm parameter | Default value |
| -------------- | ------------- |
| `openshift`    | `false`       |

### Description {#openshift-description}

Set to `true` when installing Orion on Red Hat OpenShift Container Platform version 4.x and later. For detailed steps, see [Install on Red Hat OpenShift Container Platform](installation__install-openshift.md).

### Valid values {#openshift-valid-values}

- `true`
- `false`

---

## Namespace allowList and denyList {#namespace-lists}

Comma-separated list of namespaces that constrains metrics collection to workloads in specific namespaces.

| Helm parameter             | Default value |
| -------------------------- | ------------- |
| `clusterAgent.allowNamespaces`    | None          |
| `clusterAgent.denyNamespaces`     | None          |

### Description {#namespace-lists-description}

Set only one of these values, not both. If both lists are provided, `clusterAgent.denyNamespaces` is ignored.

- To allow metrics collection on specific namespaces, use `clusterAgent.allowNamespaces`.
- To deny metrics collection on specific namespaces, use `clusterAgent.denyNamespaces`.

### Valid values {#namespace-lists-valid-values}

An array containing any string or regex expressions (`[]*regexp.Regexp`) to match against namespace names.

Examples:

```yaml
clusterAgent:
  denyNamespaces:
  - ".*-system"
```

or

```yaml
clusterAgent:
  denyNamespaces: null
  allowNamespaces:
  - "monitoring"
  - "frontend"
  - ".ecommerce-*"
```

To explicitly unset either of these values, use `null`.

```yaml
clusterAgent:
  denyNamespaces: null
```

---

## Additional volumes and volume mounts {#pods-mounts-volumes}

These parameters enable additional persistent pod storage and data sharing.

| Helm parameter           | Default value |
| ------------------------ | ------------- |
| `additionalVolumes`      | None          |
| `additionalVolumeMounts` | None          |

### Description {#pods-mounts-volumes--description}

Volumes defined in `additionalVolumes` are added to every Orion pod except the Node Agent, and mounts defined in `additionalVolumeMounts` are added to their containers. More information about configuring volume storage for Pods can be found in the [Kubernetes documentation][configure-volume-storage].

One possible use case for these parameters is mounting custom CA bundles as part of using a transparent proxy.

### Valid values {#pods-mounts-volumes-valid-values}

The `additionalVolumes` Helm parameter is an array of [`core/v1 Volumes`][v1-volumes].

The `additionalVolumeMounts` Helm parameter is an array of [`core/v1 VolumeMounts`][v1-volume-mounts].

Example:

You can copy the following excerpt into a Helm values file. The example shows replacing the default CA certificate in /etc/ssl/certs with an internal corporate CA bundle.

```yaml
---
# Configures additional volumes to add to all pods
additionalVolumes:
- name: corporate-root-ca
  configMap:
    name: corporate-root-ca
    items:
    - key: ca.crt
      path: ca.crt

# Configures additional volumeMounts to add to all containers
additionalVolumeMounts:
- name: corporate-root-ca
  mountPath: /etc/ssl/certs
```

Related information:
- [Volumes][volumes] (Kubernetes official documentation)

---

## Pod priority and scheduling {#pod-priority}

Enables scheduling and priority configurations on Orion pods.

| Helm parameter             | Default value | Scope                 |
| -------------------------- | ------------- | --------------------- |
| `priorityClassName`        | `""`          | Top-level             |
| `clusterAgent.priorityClassName`  | `""`          | Agent only (override) |
| `nodeSelector`             | `{}`          | Top-level             |
| `tolerations`              | `[]`          | Top-level             |
| `affinity`                 | `{}`          | Top-level             |

### Description {#pod-priority-description}

The top-level `priorityClassName`, `nodeSelector`, `tolerations`, and `affinity` values apply to all Orion components. To override scheduling for one component, set the corresponding `clusterAgent.*`, `applier.*`, `forwarder.*`, or `nodeAgent.*` value. A non-empty per-component value **replaces** the top-level default.

The Agent, Applier, and Forwarder set no scheduling defaults of their own, so they use the top-level values unless you override them. The Node Agent ships with its own `nodeSelector` and `tolerations` that place it on GPU nodes; setting `nodeAgent.nodeSelector` merges with that default rather than replacing it. See [Node selection](installation__node-agent-config.md#node-selection).

`clusterAgent.priorityClassName` assigns an existing Pod `PriorityClass` to the `orion-cluster-agent` pod.

### Valid values {#pod-priority-valid-values}

A string that matches the `.metadata.name` of a `PriorityClass` defined in your cluster.

### Example {#pod-priority-example}

To apply a priority class to all Orion components:

```yaml
priorityClassName: "high-priority"
```

To override for the Agent only:

```yaml
clusterAgent:
  priorityClassName: "high-priority"
```

For `nodeSelector`, `tolerations`, and `affinity` format, see the [Kubernetes documentation on assigning pods to nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/).

---

## GKE Autopilot {#gke-autopilot}

Required only when deploying on GKE Autopilot.

| Helm parameter | Default value |
| -------------- | ------------- |
| `gkeAutopilot` | `false`       |

### Description {#gke-autopilot-description}

Set to `true` when installing Orion on a Google Kubernetes Engine Autopilot cluster.

### Valid values {#gke-autopilot-valid-values}

- `true`
- `false`

---

## Cluster labels and provider {#cluster-labels}

Optional metadata that associates additional context with the cluster.

| Helm parameter    | Default value |
| ----------------- | ------------- |
| `clusterLabels`   | `{}`          |
| `clusterProvider` | `""`          |

### Description {#cluster-labels-description}

`clusterLabels` specifies key-value labels associated with the cluster in Orion.

`clusterProvider` specifies the cloud provider (for example, `aws`, `gcp`, `azure`, `vsphere`). If not set, the provider is auto-detected from node `providerID` fields.

### Example {#cluster-labels-example}

```yaml
clusterLabels:
  env: production
  team: platform
clusterProvider: "aws"
```

---

## Network policies {#network-policies}

Optional Kubernetes NetworkPolicy resources for Orion components.

| Helm parameter                           | Default value |
| ---------------------------------------- | ------------- |
| `networkPolicies.create`                 | `false`       |
| `networkPolicies.additionalIngressRules` | `[]`          |
| `networkPolicies.additionalEgressRules`  | `[]`          |

### Description {#network-policies-description}

Set `networkPolicies.create: true` to create NetworkPolicy resources restricting traffic to Orion components. Use `additionalIngressRules` and `additionalEgressRules` to extend the default policies.

### Valid values {#network-policies-valid-values}

- `networkPolicies.create`: `true` or `false`
- `additionalIngressRules` / `additionalEgressRules`: arrays of Kubernetes NetworkPolicy ingress/egress rule objects

---

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Deprecated agent parameters</div>


The `agent` parameter group is deprecated and will be removed in a future major
version; use `clusterAgent` instead. Until then, `agent.*` values still work
and override equivalent `clusterAgent.*` values.
</div>


## Related topics

- [Orion Agent description][agent-desc]
- Kubernetes official documentation:
  - [PriorityClass](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#priorityclass-v1-scheduling-k8s-io)
  - [Pod Priority and Preemption](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)

[agent-desc]: /docs/concepts/#agent
[applier-config]: /docs/installation/applier-config/
[v1-volumes]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#volume-v1-core
[v1-volume-mounts]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#volumemount-v1-core
[volumes]: https://kubernetes.io/docs/concepts/storage/volumes/
[configure-volume-storage]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-volume-storage/
