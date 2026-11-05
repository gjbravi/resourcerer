# Cluster defaults templates

> ConfigMap templates for cluster-level optimization defaults


The Orion Agent reads cluster-level optimization defaults from a ConfigMap named `cluster-defaults` in the `orion-system` namespace. You can create and manage this ConfigMap using `kubectl` or your GitOps tool. For full management details, see [Configure optimization defaults for the cluster](configure__cluster-defaults.md).

## Check current cluster defaults

To check whether any defaults are already set:

```shell
kubectl describe configmap cluster-defaults -n orion-system
```

If you see `Error from server (NotFound): configmaps "cluster-defaults" not found`, no cluster-level defaults are set.

## Optimization templates

The following templates cover the most common cluster-level configurations. Copy the relevant template, review the example values and adjust for your environment, and remove any settings you don't need.

### Balanced optimization (CPU and memory)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-defaults
  namespace: orion-system
data:
  cluster-defaults.yaml: |
    live.orion.example/schedule: "P1D"
    live.orion.example/auto-deploy: "Disabled"
    live.orion.example/cpu.optimization-goal: "Balanced"
    live.orion.example/memory.optimization-goal: "Balanced"
    live.orion.example/containers.cpu.optimization-policy: "RequestsAndLimits"
    live.orion.example/containers.memory.optimization-policy: "RequestsOnly"
```

### Savings-focused

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-defaults
  namespace: orion-system
data:
  cluster-defaults.yaml: |
    live.orion.example/schedule: "P1D"
    live.orion.example/auto-deploy: "Disabled"
    live.orion.example/cpu.optimization-goal: "Savings"
    live.orion.example/memory.optimization-goal: "Savings"
    live.orion.example/containers.cpu.optimization-policy: "RequestsAndLimits"
    live.orion.example/containers.memory.optimization-policy: "RequestsAndLimits"
    live.orion.example/containers.cpu.requests.min: "20m"
    live.orion.example/containers.memory.requests.min: "32Mi"
```

### Reliability-focused

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-defaults
  namespace: orion-system
data:
  cluster-defaults.yaml: |
    live.orion.example/schedule: "P1D"
    live.orion.example/auto-deploy: "Disabled"
    live.orion.example/cpu.optimization-goal: "Reliability"
    live.orion.example/memory.optimization-goal: "Reliability"
    live.orion.example/containers.cpu.optimization-policy: "RequestsOnly"
    live.orion.example/containers.memory.optimization-policy: "RequestsOnly"
```

### Automatic deployment enabled

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-defaults
  namespace: orion-system
data:
  cluster-defaults.yaml: |
    live.orion.example/schedule: "PT16H"
    live.orion.example/auto-deploy: "Enabled"
    live.orion.example/cpu.optimization-goal: "Savings"
    live.orion.example/memory.optimization-goal: "Savings"
    live.orion.example/containers.cpu.optimization-policy: "RequestsAndLimits"
    live.orion.example/containers.memory.optimization-policy: "RequestsAndLimits"
    live.orion.example/containers.cpu.requests.min: "20m"
    live.orion.example/containers.memory.requests.min: "32Mi"
    # Maximum bounds: align with your organization's maximums
    # live.orion.example/containers.cpu.requests.max: "16000m"
    # live.orion.example/containers.memory.requests.max: "32Gi"
```

## Apply the ConfigMap

```shell
kubectl apply -f FILENAME
```

Then restart the Agent to pick up the changes:

```shell
kubectl rollout restart deployment orion-cluster-agent -n orion-system
```

## Confirm your changes

```shell
kubectl describe configmap cluster-defaults -n orion-system
```

Review the `cluster-defaults.yaml` item in the **Data** section of the output.

## Related topics

- [Configure optimization defaults for the cluster](configure__cluster-defaults.md) — manage defaults using `kubectl` after install
- [Optimization settings and descriptions](configure__settings.md) — complete list of annotation keys
