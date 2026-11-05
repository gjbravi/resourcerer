---
updatedAt: 2026-09-29T10:37:55.000Z
---

# Custom edge locations

Custom edge locations let you add any Linux VM or bare metal instance to your OMNI-enabled cluster — regardless of cloud provider. Instead of Atlas managing cloud resources on your behalf (as with AWS, GCP, or OCI edge locations), you run a single install script on each instance. The `atlas-edge-initd` agent handles joining the instance to your cluster automatically.

This is useful when you want to use smaller cloud providers, on-premise hardware, or any compute that Atlas does not natively integrate with.

# How it works

1. You create a **Custom** edge location in the Atlas console or via API. This registers a logical edge location that skips cloud provider reconciliation.
2. For each compute instance you want to add, you fetch the edge init script from the API and run it as root.
3. The script installs the `atlas-edge-initd` binary and starts it as a systemd service. The agent joins the instance to your cluster as an edge node.

# Before you begin

## Requirements

* OMNI must be enabled on your cluster. See [Getting started with OMNI](omni-getting-started.md).
* Each compute instance must meet the following requirements:

| Requirement  | Detail                                                      |
| ------------ | ----------------------------------------------------------- |
| OS           | Ubuntu Linux (recommended); other Linux distros (see below) |
| Kernel       | 6.0+ recommended; 5.10+ minimum required                    |
| Architecture | `x86_64` (amd64) or `aarch64` (arm64)                       |
| Access       | Root (or sudo) access to run the install script             |

### Using a non-Ubuntu distribution

Other Linux distributions are supported under two conditions:

* **Drivers pre-installed**: If NVIDIA drivers are already installed on the node, the GPU Operator detects them and skips installation. This works regardless of OS or kernel version.
* **Drivers not pre-installed**: The GPU Operator can automatically install drivers, but only for OS and kernel combinations for which it has published container images. If your OS or kernel version is not in the [supported list](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html#supported-operating-systems-and-kubernetes-platforms), automatic installation will likely fail.

If your distribution is not on the supported list, pre-install NVIDIA drivers before running the OMNI install script to ensure compatibility.

## Network requirements

Configure firewall rules on each instance to allow the following traffic.

**Egress:**

| Destination | Protocol | Port  | Purpose                                                           |
| ----------- | -------- | ----- | ----------------------------------------------------------------- |
| 0.0.0.0/0   | TCP      | 443   | Atlas API (see [API access](api-access.md)) and HTTPS services |
| 0.0.0.0/0   | TCP      | 80    | HTTP services                                                     |
| 0.0.0.0/0   | TCP      | 8443  | k0smotron control plane                                           |
| 0.0.0.0/0   | TCP      | 8132  | k0smotron konnectivity endpoint                                   |
| 0.0.0.0/0   | UDP      | 51840 | WireGuard VPN                                                     |

If you have less strict requirements, you can allow all ports and protocols in egress.

If your network policy does not allow `0.0.0.0/0` as a destination, you can scope TCP 443 to the Atlas API endpoint [IP allowlist](communication-requirements.md#ip-allowlisting). That allowlist covers only the Atlas API. The k0smotron (TCP 8443 and 8132) and WireGuard (UDP 51840) destinations live in the main cluster where the edge nodes will be connecting to, so you can scope them to the address of the Load Balancer that was created as part of the Edge Location onboarding.

**Ingress:**

The following ports and protocols must be open between all OMNI instances in the network:

| Destination | Protocol | Port | Purpose                                                               |
| ----------- | -------- | ---- | --------------------------------------------------------------------- |
| VPC CIDR    | TCP      | 179  | BGP — required for route exchange between nodes                       |
| VPC CIDR    | UDP      | 5555 | FOU encapsulation  - required for node-to-node encapsulation protocol |

If you have less strict requirements, you can allow ingress for all ports and protocols in the VPC. An alternative is to use tag-based filtering if available: allow all required ports and protocols only between instances tagged for OMNI.

> 📘 Note
>
> If you are using IPIP instead of FOU as encapsulation protocol, you need to allow Protocol 4 (IPIP) for all ports and you can block UDP 5555

# Limitations

## Networking

All nodes joining the same custom edge location must be on the same private network.
If private networking between instances is not possible, each node must join a separate custom edge location instead of sharing one. In such case traffic between edge locations will be routed via main cluster.

# Step 1: Create a custom edge location

<Tabs>
  <Tab title="Atlas console">
    1. Navigate to **Automation** → **Node autoscaler** → **Configuration** → **Edge Locations**
    2. Click **Create edge location**
    3. Set **Provider** to **Custom**
    4. Enter a **Name** for the location (for example, `nebius-gpu` or `on-prem-dc1`)
    5. Click **Next** and copy the edge location ID — you'll need it in the next step
  </Tab>

  <Tab title="API">
    ```shell
    #!/bin/bash
    set -e
    API_HOST="$ATLAS_API_URL"
    API_KEY="<key>"
    ORG_ID="<org>"
    CLUSTER_ID="<cluster>"

    curl --request POST \
         --url "$API_HOST/omni-provisioner/v1beta/organizations/$ORG_ID/clusters/$CLUSTER_ID/edge-locations" \
         --header "accept: application/json" \
         --header "X-API-Key: $API_KEY" \
         --header "content-type: application/json" \
         --data '{
           "custom": {},
           "name": "my-custom-location",
           "control_plane_mode": "SHARED"
         }'
    ```

    Note the `id` field in the response — this is your `EDGE_LOCATION_ID`.
  </Tab>
</Tabs>

After creation, the edge location appears in the list with **Ready** status.

# Step 2: Run the init script on each instance

For each compute instance you want to join to the edge location, fetch and run the init script as root. The script installs `atlas-edge-initd` and registers the instance as an edge node.

```shell
#!/bin/bash
set -e
API_HOST="$ATLAS_API_URL"
API_KEY="<key>"
ORG_ID="<org>"
CLUSTER_ID="<cluster>"
EDGE_LOCATION_ID="<edge-location-id>"

curl --url "$API_HOST/omni-provisioner/v1beta/organizations/$ORG_ID/clusters/$CLUSTER_ID/edge-locations/$EDGE_LOCATION_ID:edgeInitdScript" \
     --header "X-API-Key: $API_KEY" \
     --header "content-type: application/json" | bash
```

> ⚠️ Warning
>
> The script must be run as root on the instance. It installs a binary to `/usr/local/bin/atlas-edge-initd` and creates a systemd service.

Repeat this for every instance you want to add to the edge location.

## Optional: labels, taints, and GPU configuration

To configure node labels, taints, and GPU settings, create an env file on the instance and pass its path via `INITD_EXTRA_ENV_FILE` before piping to bash:

```shell
#!/bin/bash
set -e
API_HOST="$ATLAS_API_URL"
API_KEY="<key>"
ORG_ID="<org>"
CLUSTER_ID="<cluster>"
EDGE_LOCATION_ID="<edge-location-id>"

cat > ./edge.env <<EOF
INITD_KUBERNETES_LABELS=env=prod,team=infra
INITD_KUBERNETES_TAINTS=dedicated=gpu:NoSchedule,spot:NoExecute
INITD_GPU_CONFIG={"mig": {"partitionSizes": ["1g.5gb"]}}
EOF

curl --url "$API_HOST/omni-provisioner/v1beta/organizations/$ORG_ID/clusters/$CLUSTER_ID/edge-locations/$EDGE_LOCATION_ID:edgeInitdScript" \
     --header "X-API-Key: $API_KEY" \
     --header "content-type: application/json" | \
     INITD_EXTRA_ENV_FILE=$(pwd)/edge.env \
     bash
```

The env file supports the following variables, among others:

| Variable                  | Description                                                                                                                                                                                              | Example                                   |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| `INITD_KUBERNETES_LABELS` | Comma-separated key=value labels applied to the node                                                                                                                                                     | `env=prod,team=infra`                     |
| `INITD_KUBERNETES_TAINTS` | Comma-separated key=value:effect taints applied to the node                                                                                                                                              | `dedicated=gpu:NoSchedule,spot:NoExecute` |
| `INITD_GPU_CONFIG`        | JSON GPU configuration (for MIG partitioning)                                                                                                                                                            | `{"mig": {"partitionSizes": ["1g.5gb"]}}` |
| `HTTP_PROXY`              | HTTP proxy for outbound connections                                                                                                                                                                      | `http://proxy.example.com:3128`           |
| `HTTPS_PROXY`             | HTTPS proxy for outbound connections                                                                                                                                                                     | `https://proxy.example.com:3128`          |
| `NVME_DEVICES`            | Space-separated list of NVMe block devices to use as ephemeral storage for k0s and kubelet. If unset, all non-root NVMe devices are used automatically. Multiple devices are striped together using LVM. | `/dev/nvme0n1 /dev/nvme1n1`               |

> 📘 Note
>
> `INITD_GPU_CONFIG` is for NVIDIA MIG partitioning. Partition sizes must match the physical GPU in the instance. See the [NVIDIA MIG documentation](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/) for valid partition sizes per GPU model.

<Accordion title="Verify the agent is running" icon="fa-check-circle">
  After running the script, verify the `atlas-edge-initd` service started successfully:

  ```shell
  systemctl status atlas-edge-initd
  ```

  View agent logs:

  ```shell
  journalctl -f -u atlas-edge-initd
  ```

  View the init script that was downloaded and executed:

  ```shell
  cat /var/lib/atlas-edge-initd/init_script
  ```

  View the systemd service file:

  ```shell
  cat /etc/systemd/system/atlas-edge-initd.service
  ```

  View k0sworker logs:

  ```shell
  journalctl -f -u k0sworker
  ```
</Accordion>

# Scheduling workloads on edge nodes

To schedule workloads on a custom edge node, label the namespace with `omni.atlas.example/enable-scheduling=true` and use a `nodeSelector` targeting your edge location by name.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    omni.atlas.example/enable-scheduling: "true"
  name: omni-test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: omni-test
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
      nodeSelector:
        omni.atlas.example/edge-location-name: "my-custom-location"
      containers:
        - name: nginx
          image: nginx:latest
          ports:
            - containerPort: 80
              name: http
```

The namespace label enables Liqo's offloading mechanism and causes a mutating webhook to automatically inject the required edge node toleration into pods. Replace `my-custom-location` with the name you gave your edge location in Step 1.

# Removing an edge node

To remove a specific edge node from the cluster, delete it via the API:

```shell
#!/bin/bash
set -e
API_HOST="$ATLAS_API_URL"
API_KEY="<key>"
ORG_ID="<org>"
CLUSTER_ID="<cluster>"
EDGE_LOCATION_ID="<edge-location-id>"
EDGE_ID="<edge-id>"

curl --request DELETE \
     --url "$API_HOST/omni-provisioner/v2beta/organizations/$ORG_ID/clusters/$CLUSTER_ID/edge-locations/$EDGE_LOCATION_ID/edges/$EDGE_ID" \
     --header "X-API-Key: $API_KEY" \
     --header "content-type: application/json"
```

To find `EDGE_ID`, list the edges for your edge location in the Atlas console or via the API.

# Troubleshooting

## Common issues and solutions

### `atlas-edge-initd` service fails to start

* **Possible cause:** Script not run as root.

  **Solution:** Re-run the script with `sudo` or as root.

### Service starts but node doesn't appear in cluster

* **Possible cause:** Network connectivity issue.

  **Solution:** Verify that egress ports 443, 8443, 8132, and UDP 51840 are open.

### Node appears but pods won't schedule

* **Possible cause:** Namespace not labeled for offloading.

  **Solution:** Label the namespace:

  ```shell
  kubectl label ns <namespace> omni.atlas.example/enable-scheduling=true
  ```

### MIG partitioning not applied

* **Possible cause:** Invalid partition size.

  **Solution:** Check the GPU model and use a valid partition size from the [NVIDIA documentation](https://docs.nvidia.com/datacenter/tesla/mig-user-guide/).

### Init script stuck in "CREATING" state for 10-15 minutes

If the init script runs for 10-15 minutes without completing, you may see a log message like this:

```text
msg="waiting for edge to become READY, current phase is CREATING" version=v1.13.2-13-g504f98c89af-dirty cluster_id=a1b2c3d4-e5f6-7890-abcd-ef1234567890 organization_id=12ab34cd-56ef-7890-abcd-ef1234567890 edge_location_id=34cd56ef-7890-abcd-ef12-34567890abcd component=edge-controller
```

* **Possible cause:** System pods on the remote edge cluster are failing to start — a common issue is the GPU Operator failing to initialize.

  **Solution:** Connect to the remote edge cluster and check for failing pods. The edge cluster's kubeconfig is stored as a secret in the `atlas-omni` namespace.

  First, find the kubeconfig secret name:

  ```shell
  kubectl get secrets -n atlas-omni
  ```

  Then extract the kubeconfig and list all pods. Replace `<secret-name>` with the secret name from the previous step:

  ```shell
  KUBECONFIG_FILE=$(mktemp)
  kubectl get secret <secret-name> -n atlas-omni -o jsonpath='{.data.value}' | base64 --decode > "$KUBECONFIG_FILE"
  KUBECONFIG="$KUBECONFIG_FILE" kubectl get pods -A -o wide
  ```

  Look for pods in `CrashLoopBackOff`, `Error`, or `Pending` states — particularly GPU Operator-related pods. Run any additional diagnostic commands you need before cleaning up:

  ```shell
  rm "$KUBECONFIG_FILE"
  ```

* **Possible cause:** Kubelet on the edge instance is failing to start.

  **Solution:** Check the kubelet (k0s worker) logs directly on the instance:

  ```shell
  journalctl -u k0sworker --no-pager -n 200
  ```

  Look for errors related to container runtime, network setup, or node registration that would prevent the kubelet from joining the cluster. To follow logs in real time during a restart attempt:

  ```shell
  journalctl -f -u k0sworker
  ```

<br />