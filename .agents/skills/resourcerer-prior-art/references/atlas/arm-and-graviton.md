---
updatedAt: 2026-05-12T13:39:45.000Z
---

# Arm and Graviton

### What's the process for enabling Graviton for an organization, and what are the usage limitations (around `nodeSelector`)?

Graviton is enabled if you use the node selector. You can modify deployment/StatefulSet as below:

You can use the `nodeSelector`:

```yaml
nodeSelector:  
  kubernetes.io/arch: arm64
```

For GKE, you will also need the toleration:

```yaml
tolerations:
  - key: "kubernetes.io/arch" 
  	value: "arm64"
  	operator: "Equal"
  	effect: "NoSchedule"
```

***

### Does Atlas support ARM with the `atlas-evictor` Helm chart image?

Yes, Evictor can run on ARM nodes.

***

### Can Atlas provision different resources per namespace? For example, only ARM Graviton-based nodes for the app namespace.

To achieve this, use the`nodeTemplates` and implement `nodeSelector` or toleration for workloads within the "app" namespace. This approach would enable the assignment of pods to designated nodes, specifically ARM Graviton nodes.

As of now, this process entails applying these changes to all workloads within the "app" namespace to ensure distinct resources are generated for each namespace.

To learn more about `nodeTemplates`, please refer to [this page](node-templates.md).

Here's an example configuration:

<Image align="center" src="https://files.readme.io/98e2087-image-20230814-191549.png" />

***

### Can I run workloads on multi-architecture nodes, both `arm64` and `amd64` nodes?

You can do this by enabling both architectures in a Atlas node template. Then, you can use the `nodeAffinity` that we provide in the node template UI on your multi-architecture workloads.

***

### When using multi-architecture, is Atlas defaulting to ARM or AMD?

If there are no `nodeSelectors` or affinities on the workloads, and both architectures  on the node template are selected, the following conditions will be checked:

* Users can set their preference - ARM or x86\_64 (AMD).
* If preference is not set, the node template picks the cheapest option.