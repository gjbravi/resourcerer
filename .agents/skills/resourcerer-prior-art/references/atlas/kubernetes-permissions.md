---
updatedAt: 2026-09-22T08:46:05.000Z
---

# Kubernetes permissions

Kubernetes Service Accounts and permissions used by Atlas components.

Atlas components running on your clusters use predefined service accounts and relevant permissions to perform functions such as sending data about the cluster state. This section discusses all required service accounts and permissions granted to Atlas components.

## Kubernetes Service Accounts used by Atlas components

Each [Atlas component](hosted-components.md) installed in your cluster uses a dedicated service account. Such a setup allows you to fine-tune permissions for each component:

```shell
» kubectl get serviceAccounts -n atlas-agent
NAME                                   SECRETS   AGE
atlas-agent                           0         97d
atlas-ai-optimizer-proxy              0         55d
atlas-aibrix-controller-manager       0         55d
atlas-aws-node                        0         57d
atlas-cluster-controller              0         97d
atlas-evictor                         0         97d
atlas-sentinel                          0         97d
atlas-sentinel-cluster-proxy            0         97d
atlas-sentinel-controller               0         97d
atlas-live-controller                 0         57d
atlas-live-daemon                     3         57d
atlas-live-patch-sa                   0         16d
atlas-pod-mutator                     0         85d
atlas-pod-pinner                      0         97d
atlas-spot-handler                    0         97d
atlas-workload-autoscaler             0         97d
atlasware-operator-controller-manager   0         3d16h
default                                0         97d
```

## The Atlas Operator's permissions

The [Atlas Operator](atlas-operator.md) manages the installation, configuration, and updating of Atlas components. The Operator requires permissions to install and manage components through Helm charts, which include both read permissions to understand the cluster state and write permissions to create and update resources.

### Privilege escalation prevention

The Operator follows Kubernetes privilege escalation prevention mechanisms. When creating or updating roles and cluster roles, the Operator's service account must already have all permissions contained in the resource being created or updated. This prevents the Operator from granting itself or other components permissions it doesn't already have.

Roles and cluster roles can only escalate privileges if the `escalate` verb is explicitly specified for the `roles` or `clusterroles` resource, which the Operator does not have.

### Phase 1 permissions

Phase 1 permissions allow the Operator to install and manage the `atlas-agent` component. These permissions include all the read-only permissions the agent needs, plus additional write permissions to manage the agent's lifecycle.

#### Namespace-wide permissions in `atlas-agent`

Namespace-level permissions in the `atlas-agent` namespace are lower risk than cluster-level permissions as they only affect Atlas resources.

**Permissions shared by both the agent and Operator:**

| API Group           | Resources   | Verbs                            | Description                                                                                                                              |
| ------------------- | :---------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| coordination.k8s.io | leases      | create, get, list, watch, update | Required for lease election                                                                                                              |
| core                | configmaps  | get, list, watch                 | Required for cost savings estimation features                                                                                            |
| apps                | deployments | patch                            | Required for proportional vertical cluster autoscaler to adjust `atlas-agent` requests/limits. Scoped to `atlas-agent` deployment only |

**Operator-only permissions in `atlas-agent` namespace:**

| API Group                 | Resources            | Verbs                                           | Description                                                                                              |
| ------------------------- | :------------------- | ----------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| apps                      | deployments          | create, delete, get, list, patch, update        | Required to create/update/delete agent deployment                                                        |
| apps                      | daemonsets           | create, delete, get, list, patch, update        | Required to create/update/delete spot-handler daemonsets                                                 |
| core                      | secrets              | create, delete, get, list, patch, update, watch | Required to create/update/delete secrets containing Helm release state and Operator webhook certificates |
| core                      | serviceaccounts      | create, delete, get, list, patch, update, watch | Required to create/update/delete agent service accounts                                                  |
| core                      | services             | create, delete, get, list, patch, update, watch | Required to create/update/delete Operator webhook services                                               |
| policy                    | poddisruptionbudgets | create, delete, get, list, patch, update, watch | Required to create/update/delete agent pod disruption budget                                             |
| rbac.authorization.k8s.io | rolebindings, roles  | create, delete, get, list, patch, update, watch | Required to create/update/delete agent role and role binding                                             |
| core                      | pods/log             | get                                             | Required by the Operator to read agent logs and extract cluster ID                                       |
| core                      | resourcequotas       | create, delete, get, list, patch, update, watch | Required to create/update/delete agent resource quota                                                    |
| core                      | configmaps           | create, update, patch, delete                   | Required to create/update/delete agent configmap and webhook CA configmap                                |
| batch                     | jobs                 | create, delete                                  | Required to run self-upgrade and post-uninstall jobs                                                     |
| coordination.k8s.io       | leases               | create, update, patch, delete                   | Required for Operator leader election                                                                    |
| core                      | events               | create, patch                                   | Required for event recording                                                                             |

**Operator-only permissions in `kube-system` namespace:**

| API Group                 | Resources           | Verbs  | Description                                     |
| ------------------------- | :------------------ | ------ | ----------------------------------------------- |
| rbac.authorization.k8s.io | roles, rolebindings | delete | Required by the Operator to uninstall the agent |

Resource names are scoped to `atlas-agent` only.

#### Namespace-wide permissions in `kube-system`

| API Group | Resources  | Verbs            | Description                               |
| --------- | :--------- | ---------------- | ----------------------------------------- |
| core      | configmaps | get, list, watch | Required by the agent only in kube-system |

#### Cluster-wide permissions

Cluster-level permissions are required by the agent to read resources in the cluster and create snapshots to send to the platform, and by the Operator to grant those permissions to the agent.

**Permissions shared by both the agent and Operator (read-only):**

| API Group                 | Resources                                                                                                                                 | Verbs            | Description                                                |
| :------------------------ | :---------------------------------------------------------------------------------------------------------------------------------------- | :--------------- | :--------------------------------------------------------- |
| core                      | pods, nodes, replicationcontrollers, persistentvolumeclaims, persistentvolumes, services, namespaces, events, limitranges, resourcequotas | get, list, watch | Required for cost savings estimation features              |
| core                      | namespaces                                                                                                                                | get              |                                                            |
| apps                      | deployments, replicasets, daemonsets, statefulsets                                                                                        | get, list, watch |                                                            |
| storage.k8s.io            | storageclasses, csinodes                                                                                                                  | get, list, watch |                                                            |
| batch                     | jobs, cronjobs                                                                                                                            | get, list, watch |                                                            |
| autoscaling               | horizontalpodautoscalers                                                                                                                  | get, list, watch |                                                            |
| metrics.k8s.io            | pods                                                                                                                                      | get, list        |                                                            |
| policy                    | poddisruptionbudgets                                                                                                                      | get, list, watch |                                                            |
| (non-resource URL)        | /version                                                                                                                                  | get              |                                                            |
| networking.k8s.io         | networkpolicies, ingresses                                                                                                                | get, list, watch | Required for security and k8s compliance reporting         |
| rbac.authorization.k8s.io | roles, rolebindings, clusterroles, clusterrolebindings                                                                                    | get, list, watch | Required for security and k8s compliance reporting         |
| karpenter.sh              | provisioners, machines, nodepools, nodeclaims, nodeoverlays                                                                               | get, list, watch | Required for Karpenter resource monitoring                 |
| karpenter.k8s.aws         | awsnodetemplates, ec2nodeclasses                                                                                                          | get, list, watch | Required for Karpenter AWS resource monitoring             |
| datadoghq.com             | extendeddaemonsetreplicasets                                                                                                              | get, list, watch | Required for Datadog resource monitoring                   |
| argoproj.io               | rollouts                                                                                                                                  | get, list, watch | Required for Argo Rollouts monitoring                      |
| autoscaling.atlas.example       | recommendations                                                                                                                           | get, list, watch | Required for Atlas autoscaling recommendations           |
| pod-mutations.atlas.example     | podmutations                                                                                                                              | get, list, watch | Required for Atlas pod mutation monitoring               |
| resource.k8s.io           | deviceclasses, resourceclaims, resourceclaimtemplates, resourceslices, devicetaintrules                                                   | get, list, watch | Required for Kubernetes resource management API monitoring |
| runbooks.atlas.example          | recommendationsyncs                                                                                                                       | get, list, watch | Required for Agentic AI workflows                          |

**Operator-only cluster permissions:**

| API Group                    | Resources                                                      | Verbs                                           | Description                                                                                                                                                                                                                                                                                                                      |
| ---------------------------- | :------------------------------------------------------------- | ----------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| atlasware.atlas.example             | clusters, components                                           | create, delete, get, list, patch, update, watch | Required to create/update/delete Operator custom resources                                                                                                                                                                                                                                                                       |
| atlasware.atlas.example             | clusters/status, components/status                             | get, patch, update                              | Required to update status of Operator custom resources                                                                                                                                                                                                                                                                           |
| atlasware.atlas.example             | clusters/finalizers, components/finalizers                     | update                                          | Required to add finalizers to Operator custom resources                                                                                                                                                                                                                                                                          |
| apiextensions.k8s.io         | customresourcedefinitions                                      | create, delete, get, list, patch, update        | Required to create/update/delete Operator CRDs during installation, runtime, and upgrades. Scoped to `clusters.atlasware.atlas.example` and `components.atlasware.atlas.example` only. The `create`, `update`, and `patch` permissions enable automatic CRD upgrades during Helm operations and allow the operator to self-manage CRD schemas. |
| rbac.authorization.k8s.io    | clusterrolebindings, clusterroles                              | delete                                          | Required to install/update/delete agent cluster role and cluster role binding if the agent is uninstalled. Scoped to `atlas-agent` only                                                                                                                                                                                         |
| rbac.authorization.k8s.io    | clusterrolebindings, clusterroles, rolebindings, roles         | create, patch                                   | Required to create RBAC resources for managed components                                                                                                                                                                                                                                                                         |
| admissionregistration.k8s.io | mutatingwebhookconfigurations, validatingwebhookconfigurations | get, list, patch, update, watch                 | Required to create/update/delete Operator webhooks                                                                                                                                                                                                                                                                               |

### Extended permissions (Phase 2)

Phase 2 permissions are optional and not required to install the agent. They're disabled by default and can be enabled by setting the Helm value `extendedPermissions` to `true`. These permissions are enabled automatically when running the install script for `cluster-controller` with the Operator.

Phase 2 permissions match those required by the Cluster Controller because the Operator cannot escalate privileges and needs the same permissions as the components it installs.

#### Namespace-wide permissions in `atlas-agent` (Phase 2)

When Phase 2 permissions are enabled, both the Cluster Controller and Operator become namespace administrators for the `atlas-agent` namespace:

| API Group | Resources | Verbs | Description                                                 |
| --------- | :-------- | ----- | ----------------------------------------------------------- |
| `*`       | `*`       | `*`   | Full administrative permissions in `atlas-agent` namespace |

#### Cluster-wide permissions (Phase 2)

Cluster-level Phase 2 permissions are granted to both the Cluster Controller and Operator:

| API Group             | Resources                           | Verbs                                           | Description                                                                                                                              |
| --------------------- | :---------------------------------- | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| core                  | nodes, pods                         | get, list, watch                                | Read-only permissions required for autoscaler                                                                                            |
| core                  | nodes                               | patch, update                                   | Write permissions required for node draining, patching, and deletion                                                                     |
| core                  | nodes, pods                         | delete                                          | Write permissions required for node draining, patching, and deletion                                                                     |
| core                  | nodes/status                        | patch                                           | Required to update node status                                                                                                           |
| core                  | pods/eviction                       | create                                          | Required for pod eviction                                                                                                                |
| certificates.k8s.io   | certificatesigningrequests/approval | patch, update                                   | Read/write permissions required for CSR approval                                                                                         |
| certificates.k8s.io   | signers                             | approve                                         | Required for approving kubelet certificates. Scoped to `kubernetes.io/kube-apiserver-client-kubelet` and `kubernetes.io/kubelet-serving` |
| certificates.k8s.io   | certificatesigningrequests          | create, delete, get, list, watch                | Required for certificate signing request management                                                                                      |
| core                  | namespaces                          | delete, get                                     | Required to install/update/uninstall Atlas components. Delete scoped to `atlas-llms` only                                             |
| core                  | namespaces                          | get                                             | Required for namespace access                                                                                                            |
| core                  | namespaces                          | delete                                          | Required to delete Atlas namespace. Scoped to `atlas-agent` only                                                                      |
| autoscaling.atlas.example   | `*`                                 | create, delete, get, list, patch, update, watch | Required for autoscaling CRD management                                                                                                  |
| live.atlas.example          | `*`                                 | create, delete, get, list, patch, update, watch | Required for Live Migration                                                                                                              |
| pod-mutations.atlas.example | `*`                                 | create, delete, get, list, patch, update, watch | Required for pod mutations related custom resources                                                                                      |
| storage.k8s.io        | volumeattachments                   | delete, get, list                               | Required to manipulate volume attachments                                                                                                |

## The Atlas agent's permissions (Read-only)

The Atlas agent collects cluster operational details (snapshots) and delivers them to the central platform to assess if there is room for optimization. That's why it must get cluster-wide permissions:

| API Group      | Resources                                                                                | Verbs            |
| -------------- | :--------------------------------------------------------------------------------------- | ---------------- |
| core           | pods, nodes, replicationcontrollers, persistentvolumeclaims, persistentvolumes, services | get, list, watch |
| core           | namespaces                                                                               | get              |
| apps           | deployments, replicasets, daemonsets, statefulsets                                       | get, list, watch |
| storage.k8s.io | storageClasses, csinodes                                                                 | get, list, watch |
| batch          | jobs                                                                                     | get, list, watch |

The Atlas agent's resource consumption vastly depends on the cluster size. The agent must be able to adjust resource limits proportionally to the size of your cluster. For that purpose, Cluster Proportional Vertical Autoscaler patches the Atlas agent's deployment with re-estimated limits, which requires the following permission:

| API Group | Resources   | Verbs | Description                                      |
| --------- | :---------- | ----- | ------------------------------------------------ |
| apps      | deployments | patch | Used only to patch the `atlas-agent` deployment |

## Sentinel permissions

[Sentinel](sentinel.md) is the Atlas cluster observability agent. It runs as a DaemonSet (`sentinel-agent`) on every node and a Deployment (`sentinel-controller`) in the `atlas-agent` namespace. The agent collects runtime and workload telemetry, while the controller reconciles cluster state and coordinates scanning jobs.

### Cluster-wide permissions used by Sentinel Agent

The Sentinel Agent requires cluster-wide read permissions to observe workloads and cluster state:

| API Group | Resources                                          | Verbs            | Description                                      |
| --------- | :------------------------------------------------- | ---------------- | ------------------------------------------------ |
| core      | pods, nodes, services                              | get, list, watch | Required to observe cluster workloads and nodes. |
| apps      | deployments, replicasets, daemonsets, statefulsets | get, list, watch | Required to observe workload controllers.        |
| batch     | jobs, cronjobs                                     | get, list, watch | Required to observe batch workloads.             |

When GPU monitoring is enabled (`sentinel.gpuEnabled=true`), the agent gets one additional permission:

| API Group | Resources  | Verbs | Description                                                    |
| --------- | :--------- | ----- | -------------------------------------------------------------- |
| core      | namespaces | list  | Required to discover namespaces when collecting GPU telemetry. |

### Cluster-wide permissions used by Sentinel Controller

The Sentinel Controller requires cluster-wide read permissions to reconcile security policies and cluster state:

| API Group                 | Resources                                                                                       | Verbs            | Description                                                     |
| ------------------------- | :---------------------------------------------------------------------------------------------- | ---------------- | --------------------------------------------------------------- |
| core                      | nodes, pods, events, namespaces, services, endpoints, persistentvolumeclaims, persistentvolumes | get, list, watch | Required to observe cluster state and security-relevant events. |
| core                      | nodes/proxy                                                                                     | get              | Required to collect node-level network and runtime information. |
| apps                      | deployments, replicasets, daemonsets, statefulsets                                              | get, list, watch | Required to observe workload controllers.                       |
| batch                     | jobs, cronjobs                                                                                  | get, list, watch | Required to observe batch workloads.                            |
| networking.k8s.io         | networkpolicies, ingresses                                                                      | get, list, watch | Required for network policy and ingress analysis.               |
| rbac.authorization.k8s.io | roles, rolebindings, clusterroles, clusterrolebindings                                          | get, list, watch | Required for RBAC posture analysis.                             |

When reliability metrics collection is enabled, the controller also gets read-only access to a few additional resources:

| API Group   | Resources                              | Verbs            | Description                                   |
| ----------- | :------------------------------------- | ---------------- | --------------------------------------------- |
| core        | resourcequotas, replicationcontrollers | get, list, watch | Required for the Kubernetes cluster receiver. |
| autoscaling | horizontalpodautoscalers               | get, list, watch | Required for the Kubernetes cluster receiver. |

When OBI memory autoscaling is enabled (`controller.reliabilityMetrics.obiVPA.enabled=true` or `--obi-mem-autoscaler-enabled=true`), the controller receives pod resize permissions:

| API Group | Resources   | Verbs | Description                            |
| --------- | :---------- | ----- | -------------------------------------- |
| core      | pods        | patch | Required to resize OBI collector pods. |
| core      | pods/resize | patch | Required to resize OBI collector pods. |

### Namespace-wide (`atlas-agent`) permissions used by Sentinel Controller

Within the `atlas-agent` namespace, the Sentinel Controller manages security jobs and coordinates its own operations:

| API Group           | Resources | Verbs                                    | Description                                        |
| ------------------- | :-------- | ---------------------------------------- | -------------------------------------------------- |
| batch               | jobs      | create, get, list, watch, delete         | Required to run security scanning jobs.            |
| core                | pods/log  | get                                      | Required to read logs from security scanning jobs. |
| core                | services  | create                                   | Required to create services for CIDR discovery.    |
| coordination.k8s.io | leases    | get, create, update, list, watch, delete | Required for controller leader election.           |
| core                | events    | create                                   | Required to record controller events.              |

When the cluster proxy is enabled (`controller.extraArgs.cluster-proxy-enabled=true`), the controller receives one additional permission:

| API Group | Resources             | Verbs  | Description                                                      |
| --------- | :-------------------- | ------ | ---------------------------------------------------------------- |
| core      | serviceaccounts/token | create | Required to create service account tokens for the cluster proxy. |

## Cluster Controller's permissions

Atlas's Cluster Controller component gets installed when your connected cluster moves to Phase 2, in which you can enable managed cost savings:

```shell
» kubectl get deployments -n atlas-agent
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
atlas-agent                1/1     1            1           43h
atlas-cluster-controller   2/2     2            2           64m
atlas-evictor              0/0     0            0           64m
```

### Cluster-wide permissions used by the Cluster Controller

The Cluster Controller operates mostly on the cluster level as it performs operations required to optimize its costs:

| API Group                 | Resources                                | Verbs                                | Description                                                            |
| ------------------------- | :--------------------------------------- | ------------------------------------ | ---------------------------------------------------------------------- |
| core                      | namespaces                               | get                                  |                                                                        |
| core                      | pods, nodes                              | get, list                            |                                                                        |
| core                      | nodes                                    | patch, update                        | Used for node draining and patching.                                   |
| core                      | pods, nodes                              | delete                               |                                                                        |
| core                      | pods/eviction                            | create                               |                                                                        |
| certificates.k8s.io       | certificatesigningrequests               | get, list, delete, create            | Used for creating a new certificate when adding a node to the cluster. |
| certificates.k8s.io       | certificatesigningrequests/approval      | patch, update                        | Used for creating a new certificate when adding a node to the cluster. |
| certificates.k8s.io       | signers                                  | approve                              | Applicable only for kubelet.                                           |
| core                      | events                                   | list, create, patch                  |                                                                        |
| rbac.authorization.k8s.io | roles, clusterroles, clusterrolebindings | get, patch, update, delete, escalate | Applicable to all Atlas components.                                  |
| core                      | namespaces                               | delete                               | Applicable only to the Atlas  agent.                                 |

### Namespace-wide (atlas-agent) permissions used by the Cluster Controller

One of the main tasks of the Cluster Controller is to update Atlas components. The Cluster Controller is granted with **all permissions in `atlas-agent` namespace** necessary for current and future changes.

Additionally, it includes two cluster-wide permissions for managing the RBAC of Atlas components and the ability to delete the Atlas namespace (see above).

## Evictor permissions

When onboarding the cluster to Phase 2 for automated cost optimization, Atlas installs more components than the Cluster Controller. One such component is Evictor, which minimizes the number of nodes your cluster uses.

### Cluster-wide permissions used by Evictor

When installed, Evictor handles non-Atlas pods, so it requires a set of cluster-wide permissions:

| API Group           | Resources     | Verbs                                           | Description                                                                          |
| ------------------- | :------------ | ----------------------------------------------- | ------------------------------------------------------------------------------------ |
| core                | events        | create, patch                                   |                                                                                      |
| core                | nodes         | get, list, watch, patch, update                 | Used to find a suitable node for eviction.                                           |
| core                | pods          | get, list, watch, patch, update, create, delete | List pods to find a suitable node for eviction and delete a stuck pod from the node. |
| apps                | replicaSets   | get                                             | Used to find out if it's safe to evict a pod (it belongs to RS and has replicas).    |
| core                | pods/eviction | create                                          | Used for pod eviction.                                                               |
| coordination.k8s.io | leases        | \*                                              | Used for leader election when there may be a single instance active.                 |

## Sentinel Cluster Proxy permissions

Sentinel's cluster proxy enables on-demand querying of the Kubernetes API to fetch live cluster state data. It uses a dedicated ServiceAccount and ClusterRole with read-only access:

| API Group                 | Resources                                                                                                                                                                                   | Verbs            |
| ------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | :--------------- |
| core                      | pods, pods/log, services, configmaps, namespaces, nodes, endpoints, events, persistentvolumeclaims, persistentvolumes, resourcequotas, serviceaccounts, limitranges, replicationcontrollers | get, list, watch |
| apps                      | deployments, replicasets, daemonsets, statefulsets                                                                                                                                          | get, list, watch |
| batch                     | jobs, cronjobs                                                                                                                                                                              | get, list, watch |
| networking.k8s.io         | ingresses, ingressclasses, networkpolicies                                                                                                                                                  | get, list, watch |
| rbac.authorization.k8s.io | roles, rolebindings, clusterroles, clusterrolebindings                                                                                                                                      | get, list, watch |
| storage.k8s.io            | storageclasses, volumeattachments                                                                                                                                                           | get, list, watch |
| autoscaling               | horizontalpodautoscalers                                                                                                                                                                    | get, list, watch |
| policy                    | poddisruptionbudgets                                                                                                                                                                        | get, list, watch |
| metrics.k8s.io            | pods, nodes                                                                                                                                                                                 | get, list        |
| gateway.networking.k8s.io | gateways, gatewayclasses, httproutes, grpcroutes, referencegrants                                                                                                                           | get, list, watch |

## Pod Pinner permissions

To function correctly, [Pod Pinner](pod-pinner.md) requires the following cluster-wide permissions:

| API Group                    | Resources                     | Verbs                     | Description                                                              |
| ---------------------------- | :---------------------------- | ------------------------- | ------------------------------------------------------------------------ |
| core                         | nodes                         | create, delete, get       | Required to execute pod pinning actions                                  |
| core                         | pods                          | delete, get, list         | Required to execute pod pinning actions                                  |
| core                         | pods/binding                  | create                    | Required to execute pod pinning actions                                  |
| admissionregistration.k8s.io | mutatingwebhookconfigurations | list, watch               | Required for Webhook functionality                                       |
| admissionregistration.k8s.io | mutatingwebhookconfigurations | watch, get, patch, update | Required for Webhook functionality. Scoped to `pod-pinner` resource only |

## Keeper permissions

Keeper is the Atlas controller for Karpenter-based clusters. It manages rebalancing, node lifecycle, and workload placement for clusters using Karpenter.

### Cluster-wide permissions used by Keeper

| API Group           | Resources                                                                                                                                  | Verbs                                           | Description                                                           |
| ------------------- | :----------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------------- |
| core                | namespaces, persistentvolumeclaims, persistentvolumes, pods                                                                                | get, list, watch                                | Required to read cluster state for scheduling and placement decisions |
| core                | nodes                                                                                                                                      | get, list, patch, update, watch                 | Required for node lifecycle management and draining                   |
| core, events.k8s.io | events                                                                                                                                     | create, list, patch, watch                      | Required for event recording                                          |
| apps                | daemonsets, deployments, replicasets, statefulsets                                                                                         | get, list, watch                                | Required to read workload state for placement decisions               |
| autoscaling.atlas.example | commitmentbudgets, rebalances, recommendations                                                                                             | get, list, watch                                | Required to read Atlas autoscaling custom resources                 |
| autoscaling.atlas.example | commitmentbudgets/status, rebalancemigrationplans/status, rebalanceplanclaims/status, rebalanceplans/status, rebalanceplanschedules/status | get, patch, update                              | Required to update status of Atlas autoscaling custom resources     |
| autoscaling.atlas.example | rebalancemigrationplans, rebalanceplanclaims, rebalanceplans, rebalanceplanschedules, unavailableofferings                                 | create, delete, get, list, patch, update, watch | Required to manage rebalancing custom resources                       |
| autoscaling.atlas.example | rebalancemigrationplans/finalizers, rebalanceplanclaims/finalizers, rebalanceplans/finalizers, rebalanceplanschedules/finalizers           | update                                          | Required to manage finalizers on rebalancing custom resources         |
| batch               | jobs                                                                                                                                       | get, list, watch                                | Required to read job state for placement decisions                    |
| karpenter.k8s.aws   | ec2nodeclasses                                                                                                                             | get, list, watch                                | Required to read Karpenter AWS node class configurations              |
| karpenter.sh        | nodeclaims                                                                                                                                 | create, delete, get, list, patch, watch         | Required to manage Karpenter node claims for rebalancing              |
| karpenter.sh        | nodeclaims/status                                                                                                                          | get, list, watch                                | Required to read Karpenter node claim status                          |
| karpenter.sh        | nodepools                                                                                                                                  | create, delete, get, list, update, watch        | Required to manage Karpenter node pools                               |
| live.atlas.example        | migrations                                                                                                                                 | create, get, list, watch                        | Required to create and track Live Migration resources                 |
| live.atlas.example        | migrations/status                                                                                                                          | get, list, watch                                | Required to read Live Migration status                                |
| policy              | poddisruptionbudgets                                                                                                                       | get, list, watch                                | Required to respect pod disruption budgets during rebalancing         |
| storage.k8s.io      | csinodes, storageclasses                                                                                                                   | get, list, watch                                | Required to read storage topology for placement decisions             |

### Namespace-wide permissions used by Keeper

The following permissions are scoped to the `atlas-agent` namespace:

| API Group           | Resources   | Verbs                                           | Description                                   |
| ------------------- | :---------- | ----------------------------------------------- | --------------------------------------------- |
| coordination.k8s.io | leases      | create, delete, get, list, patch, update, watch | Required for leader election                  |
| core                | events      | create, patch                                   | Required for event recording                  |
| apps                | deployments | create, delete, get, list, patch, update, watch | Required to manage Keeper deployments     |
| core                | configmaps  | get, list, patch, update, watch                 | Required for dynamic configuration management |

### Audit namespace permissions (optional)

When audit logging is enabled (`atlas.audit.enabled=true` in Helm values), Keeper requires the following additional permissions scoped to the audit namespace (default: `atlas-audit`, configurable via `atlas.audit.namespace`):

| API Group     | Resources | Verbs                                           | Description                        |
| ------------- | :-------- | ----------------------------------------------- | ---------------------------------- |
| events.k8s.io | events    | create, delete, get, list, patch, update, watch | Required to write audit event logs |