# Orion architecture

> Orion components that run in your cluster and in the Orion cloud


The following components run in a dedicated namespace (by default, `orion-system`) **on your Kubernetes cluster**: 

- Orion Agent
- Orion Applier
- Orion Forwarder
- Orion Node Agent (optional)

 The following components exist **outside of your cluster**: 
- The Orion cloud
- The `orion` [CLI tool](orion-cli.md) (optional)
- The Orion UI (web application)

Or, shown diagrammatically: \
[![Orion architecture](/img/docs/orion-architecture.png)](/img/docs/orion-architecture.png)

The Orion Agent is a Kubernetes Pod with an application written in Go. The Agent discovers and reports on workloads in the namespaces in the cluster. You can restrict access to namespaces by using an allowNamespaces or a denyNamespaces list. 

The **Orion Forwarder** collects resource usage, performance metrics, and kube-state-metrics data and sends them to the **Orion cloud**.

The **Orion Node Agent** is an optional DaemonSet that runs on GPU nodes and collects per-workload and device-level GPU metrics, which the Forwarder sends on with the rest of the metrics.

The Orion machine learning algorithm analyzes workload usage patterns to provide CPU and memory recommendations on the schedule that you define.

The **Orion Applier** is installed by default and applies recommendations automatically or on demand. You can also download recommendations as patches to apply as part of a CI/CD workflow.

The **Orion web application** (UI) enables you to visualize the projected impact of recommendations, configure and apply recommendations, and view overall cluster and workload health. You can interact with workloads and recommendations by using the optional Orion CLI.

Related information: 
- [Namespace allowList and denyList][namespace-lists]
- [Schedule setting][schedule]
- [Applier configuration][applier-config]
- [Node Agent configuration][node-agent-config]
- [GPU optimization][gpu]
- [View][view] and [apply recommendations][apply]
- [CLI reference][cli]

[namespace-lists]: /docs/installation/cluster-agent-config/#namespace-lists
[schedule]: /docs/configure/settings/schedule/#schedule
[applier-config]: /docs/installation/applier-config/
[node-agent-config]: /docs/installation/node-agent-config/
[gpu]: /docs/recommendations/gpu/
[view]: /docs/recommendations/view/
[apply]: /docs/recommendations/apply/
[cli]: /docs/orion-cli/
