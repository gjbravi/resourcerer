---
updatedAt: 2026-05-12T14:01:40.000Z
---

# Network bandwidth

# Autoscaling using network bandwidth resources

The Atlas Autoscaler enables workloads to select nodes with suitable network bandwidth specifications. For example, if a workload specifies a requirement of up to 1000 Megabits per second (Mbps) in bandwidth, the Autoscaler will pick an instance type that provides at least 1000 Mbps for inbound and at least 1000 Mbps for outbound traffic.

## Supported providers

| Provider | Supported   |
| -------- | ----------- |
| AWS      | Supported   |
| Azure    | Coming soon |
| GCP      | Coming soon |

## Limitations

Network bandwidth requests are supported on OMNI nodes using atlas-omni-agent v1.12.5 and upwards.

## How it works

Workloads can request network bandwidth by specifying it in the container's `resources` stanza. The Autoscaler will respect the specified value and select a suitable instance type that guarantees sustainable bandwidth as specified by the cloud service provider.

To specify the required network bandwidth, add the `scheduling.atlas.example/network-bandwidth` annotation to both the `requests` and `limits` sections of the container spec. The value must be expressed in Megabits per second (Mbps).

### Workload configuration examples

```yaml Resources
resources:
  requests:
    scheduling.atlas.example/network-bandwidth: "750" # Requested network bandwidth in Mbps
  limits:
    scheduling.atlas.example/network-bandwidth: "750" # Maximum network bandwidth limit in Mbps
```

```yaml Full deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: nginx
  name: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - image: nginx
          name: nginx
          resources:
            requests:
              scheduling.atlas.example/network-bandwidth: "750"
            limits:
              scheduling.atlas.example/network-bandwidth: "750"
```