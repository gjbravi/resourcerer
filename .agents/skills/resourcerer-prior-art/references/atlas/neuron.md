---
updatedAt: 2026-07-08T11:10:22.000Z
---

# AWS Neuron Instances (EKS)

Configure Atlas Autoscaler to scale your EKS cluster using AWS Neuron accelerator instances — including AWS Inferentia and AWS Trainium — for machine learning inference and training workloads.

# Autoscaling using Neuron instances

The Atlas Autoscaler can scale EKS clusters using AWS Neuron-powered instances. This guide describes what is required for Neuron nodes to join the cluster and for Neuron workloads to be scheduled correctly.

## Supported providers

| Provider  | Supported |
| --------- | --------- |
| AWS EKS   | Yes       |
| GCP GKE   | No        |
| Azure AKS | No        |

## Supported instance families

Atlas supports all AWS EC2 instance families equipped with Neuron devices:

| Instance family  | Hardware        | Primary use          |
| ---------------- | --------------- | -------------------- |
| `inf1`           | AWS Inferentia  | Inference            |
| `inf2`           | AWS Inferentia2 | Inference            |
| `trn1` / `trn1n` | AWS Trainium    | Training             |
| `trn2`           | AWS Trainium2   | Training & inference |
| `trn3`           | AWS Trainium3   | Training & inference |

## How does it work?

Once a pending pod requests `aws.amazon.com/neuron`, `aws.amazon.com/neuroncore`, or `aws.amazon.com/neurondevice` resources, Atlas detects the workload and provisions a Neuron-capable node.

To enable provisioning of Neuron nodes, you need:

* A pod that requests Neuron resources in its container limits
* A toleration for the `aws.amazon.com/neuron=true:NoSchedule` taint
* The **Neuron device plugin** DaemonSet installed and running in the cluster

Atlas validates that the Neuron device plugin DaemonSet is present before provisioning any Neuron node. If it is not detected, Atlas emits a pod event with the message `Neuron daemon set driver is required` and does not attempt to scale.

### Neuron node resources

Atlas exposes the following extended resources on Neuron nodes:

| Resource name                 | Description                                                                    |
| ----------------------------- | ------------------------------------------------------------------------------ |
| `aws.amazon.com/neuron`       | Number of Neuron devices on the node                                           |
| `aws.amazon.com/neurondevice` | Alias for the number of Neuron devices (equivalent to `aws.amazon.com/neuron`) |
| `aws.amazon.com/neuroncore`   | Number of individual NeuronCores on the node                                   |

Use `aws.amazon.com/neuron` or `aws.amazon.com/neurondevice` to request whole devices. Use `aws.amazon.com/neuroncore` to request individual NeuronCores for finer-grained allocation.

### Tainting of Neuron nodes

Neuron nodes have the `aws.amazon.com/neuron=true:NoSchedule` taint applied automatically. Your Neuron workloads must include a matching toleration; without it, pods will not be scheduled onto Neuron nodes and Atlas will not provision one.

## Neuron device plugin

The Neuron device plugin is required for Neuron resources to be exposed to Kubernetes. Atlas checks for a DaemonSet whose pod name contains `neuron-device-plugin` before performing any Neuron node autoscaling.

Install the Neuron Helm Chart, which includes the device plugin:

```shell
helm upgrade --install neuron-helm-chart oci://public.ecr.aws/neuron/neuron-helm-chart
```

To install the device plugin only (without the Node Problem Detector):

```shell
helm upgrade --install neuron-helm-chart oci://public.ecr.aws/neuron/neuron-helm-chart \
    --set "npd.enabled=false"
```

Verify the device plugin DaemonSet is running:

```shell
kubectl get ds neuron-device-plugin -n kube-system
```

Verify that Neuron resources are visible on nodes:

```shell
kubectl get nodes "-o=custom-columns=NAME:.metadata.name,NeuronDevice:.status.allocatable.aws\.amazon\.com/neuron"
kubectl get nodes "-o=custom-columns=NAME:.metadata.name,NeuronCore:.status.allocatable.aws\.amazon\.com/neuroncore"
```

## Workload configuration examples

### Request whole Neuron devices (`aws.amazon.com/neuron`)

```yaml
spec:
  tolerations:
    - key: "aws.amazon.com/neuron"
      operator: Exists
      effect: NoSchedule
  containers:
    - image: my-neuron-image
      name: neuron-workload
      resources:
        requests:
          cpu: 4
          memory: 8Gi
          aws.amazon.com/neuron: 1
        limits:
          cpu: 4
          memory: 8Gi
          aws.amazon.com/neuron: 1
```

### Request individual NeuronCores (`aws.amazon.com/neuroncore`)

```yaml
spec:
  tolerations:
    - key: "aws.amazon.com/neuron"
      operator: Exists
      effect: NoSchedule
  containers:
    - image: my-neuron-image
      name: neuron-workload
      resources:
        requests:
          cpu: 4
          memory: 8Gi
          aws.amazon.com/neuroncore: 2
        limits:
          cpu: 4
          memory: 8Gi
          aws.amazon.com/neuroncore: 2
```

### Request multiple devices

```yaml
spec:
  tolerations:
    - key: "aws.amazon.com/neuron"
      operator: Exists
      effect: NoSchedule
  containers:
    - image: my-neuron-image
      name: neuron-workload
      resources:
        requests:
          cpu: 8
          memory: 16Gi
          aws.amazon.com/neuroncore: 8
        limits:
          cpu: 8
          memory: 16Gi
          aws.amazon.com/neuroncore: 8
```

## Required pod fields

| Field                                                                                                               | Required           | Description                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------- | ------------------ | -------------------------------------------------------------------------------------------------------------- |
| `spec.tolerations` with key `aws.amazon.com/neuron`                                                                 | Yes                | Required to allow scheduling onto Neuron nodes, which carry the `aws.amazon.com/neuron=true:NoSchedule` taint. |
| `resources.limits["aws.amazon.com/neuron"]` or `["aws.amazon.com/neuroncore"]` or `["aws.amazon.com/neurondevice"]` | Yes (at least one) | Amount of Neuron resources to allocate. Must be set in both `requests` and `limits` with equal values.         |

> 📘 Note
>
> AWS requires that `requests` and `limits` are equal for Neuron resources. Atlas enforces this and will emit an error event if they differ.

## Further reading

* [AWS Neuron EKS documentation](https://docs.aws.amazon.com/eks/latest/userguide/device-management-neuron.html)
* [Neuron Helm Chart on GitHub](https://github.com/aws-neuron/neuron-helm-charts/)