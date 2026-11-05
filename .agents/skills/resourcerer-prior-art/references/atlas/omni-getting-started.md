---
updatedAt: 2026-10-30T12:30:58.000Z
---

# Getting started

This guide walks you through setting up OMNI for your cluster, from initial onboarding through creating your first edge location, configuring node templates, and provisioning compute capacity in the edge.

# Before you begin

Ensure you meet the following requirements before setting up OMNI.

## Cluster requirements

* **Phase 2 cluster** — A cluster with Atlas automation enabled. Phase 2 clusters have the Cluster Controller, Evictor, and other automation components installed, allowing Atlas to provision and manage nodes. If your cluster is in Phase 1 (read-only), see <Anchor target="_blank" href="doc:legacy-onboarding-script">Legacy onboarding script reference</Anchor> for what gets created when adding automation.
* **Supported cluster type** — EKS, GKE, or AKS

## Required tools

The following tools must be installed on your local machine:

| Tool      | Purpose                                        | Version            |
| --------- | ---------------------------------------------- | ------------------ |
| `kubectl` | Kubernetes CLI, configured for your cluster    | v1.24+             |
| Helm      | Package manager for Kubernetes                 | v3.10+             |
| `curl`    | HTTP client for downloading scripts            | Any recent version |
| `jq`      | JSON processor for script operations           | v1.6+              |
| Cloud CLI | Authenticated for your target edge location(s) | -                  |

## Cloud permissions

You'll need sufficient permissions in your cloud account(s) to create networking resources (VPCs, subnets, security groups) and compute instances in the regions where you want to create edge locations. See [Required cloud permissions](omni-overview.md#required-cloud-permissions) for details.

## Storage & networking requirements for OMNI

OMNI edge clusters need a storage class for PVC support and, on EKS, a LoadBalancer provider. The onboarding script configures both automatically. When using Terraform, set them as module inputs.

### Storage provider

OMNI requires a storage class to provision persistent volumes for the agent on edge clusters. An SSD-backed class is recommended for performance.

#### How the script selects a storage class

The script applies this resolution order:

1. `OVERRIDE_STORAGE_PROVIDER`**&#x20;environment variable** — takes precedence when set (optional).
2. **OMNI-preferred storage class** — `gp3` for EKS, `premium-rwo` for GKE, `managed-csi-premium` for AKS. Used if the class exists in the cluster.
3. **Cluster default storage class** — used when no OMNI-preferred class is found.
4. **Failure** — the script exits with an error: install the preferred class or set `OVERRIDE_STORAGE_PROVIDER`.

> ⚠️ Warning
>
> EKS clusters don't ship with volume provisioner nor storage class installed. To provision PVCs you need to install <Anchor target="_blank" href="https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html">AWS EBS CSI driver</Anchor> and a gp3 based storage class. We recommend using `gp3` storageclass for the OMNI agent due to performance benefits. If you have a different storage class or it has a different name, set `OVERRIDE_STORAGE_PROVIDER` to point to it.

#### Terraform

The <Anchor target="_blank" href="https://github.com/atlas/terraform-atlas-omni-cluster/blob/main/variables.tf">`storage_provider`</Anchor> variable overrides the storage class used for edge clusters. Defaults to `premium-rwo` for GKE and `gp3` for EKS when empty.

```toml
module "omni" {
  # ...
  storage_provider = "my-custom-sc"
}
```

### LoadBalancer provider

This applies to **EKS only**. Provisioning load balancers requires either the <Anchor target="_blank" href="https://kubernetes-sigs.github.io/aws-load-balancer-controller/">AWS Load Balancer Controller</Anchor> (recommended) or the legacy in-tree cloud provider.

> ⚠️ Warning
>
> The legacy in-tree provider (`nlb`) is not recommended for production clusters. Install the AWS Load Balancer Controller for better performance and features.

#### Onboarding script

The onboarding script auto-detects the controller by checking existence of `elbv2.k8s.aws` CRDs. If found, AWS Load Balancer Controller is used. If not, it falls back to the legacy provider.

#### Terraform

The <Anchor target="_blank" href="https://github.com/atlas/terraform-atlas-omni-cluster/blob/main/variables.tf">`loadbalancer_provider`</Anchor> variable sets the provider for EKS edge clusters. Accepts `nlb` or `external`. Defaults to `external` when empty.

```toml
module "omni" {
  # ...
  loadbalancer_provider = "external"
}
```

The module assumes you have the AWS Load Balancer Controller installed in production clusters. Install it before applying the module, or set `loadbalancer_provider = "nlb"` to use the legacy provider.

# Step 1: Onboard your cluster to OMNI

Onboarding deploys the OMNI components to your cluster. You can onboard in two ways.

## Option A: During Phase 2 onboarding (recommended for new clusters)

<Columns layout="auto">
  <Column>
    If you're onboarding a cluster from Phase 1 (Read-only) to Phase 2 (Automation), you can enable OMNI at the same time.

    1. In the Atlas console, navigate to your cluster

    2. Follow the standard Phase 2 onboarding flow

    3. Select **Extend cluster to other regions and cloud providers** under **Advanced** settings

    The script will be updated to include `INSTALL_OMNI=true`.

    4. Copy and run the script in your terminal

    5. Wait for the script to complete (typically 1-2 minutes)
  </Column>

  <Column>
    <Image align="center" src="https://files.readme.io/2905a7071754123a33099bd17b4fb5c96a53999d748fe184be81e0a5f9c3db98-image.png" width="100% " />
  </Column>
</Columns>

## Option B: Enable OMNI on an existing Phase 2 cluster

If you already have a Phase 2 cluster with cluster optimization enabled:

1. Navigate to your cluster in the **Cluster list**
2. Click on the **ellipsis** and choose **Atlas features**

![](https://files.readme.io/768642751529091a7e2b2d2c00fb44d025a345383cba1cf8c125bdf3c3d58bad-image.png)

<Columns layout="auto">
  <Column>
    3. Under **Other features**, check the box for **Extend cluster to other regions and cloud providers**

    4. Copy the updated script and run it in your terminal

    5. Wait for the script to complete (typically 1-2 minutes)
  </Column>

  <Column>
    <Image align="center" src="https://files.readme.io/2e839e3f5c30f2f07e25479d2b3e4c2b0d38e25447ab7baa7f4787701956e944-image.png" width="100% " />
  </Column>
</Columns>

<Accordion title="Verify onboarding (optional)" icon="fa-info-circle">
  ### Verify onboarding (optional)

  After the script completes, verify OMNI is enabled:

  ```shell
  kubectl get pods -n atlas-omni
  ```

  You should see OMNI components running, including:

* `liqo-*` pods (controller manager, CRD replicator, fabric, IPAM, proxy, webhook)
* `atlas-omni-agent` pod
* `k0smotron-controller-manager` pod

  Example output:

  ```shell
  NAME                                          READY   STATUS    RESTARTS   AGE
  atlas-omni-agent-595c4b97d9-xxxxx            1/1     Running   0          2m
  k0smotron-controller-manager-5c75bbd5b8-xxxxx 1/1     Running   0          2m
  liqo-controller-manager-7cf59bcc64-xxxxx      1/1     Running   0          2m
  liqo-crd-replicator-687bdc6f66-xxxxx          1/1     Running   0          2m
  liqo-fabric-xxxxx                             1/1     Running   0          2m
  liqo-ipam-8667dbccbb-xxxxx                    1/1     Running   0          2m
  liqo-proxy-77c66dfb88-xxxxx                   1/1     Running   0          2m
  liqo-webhook-6f648484cc-xxxxx                 1/1     Running   0          2m
  ```

  All pods should be in `Running` status.
</Accordion>

# Step 2: Create and onboard an edge location

Edge locations define the regions where edge nodes can be provisioned. Each edge location is cluster-specific and requires its own setup.

1. In the Atlas console, navigate to **Autoscaler** → **Edge locations**

![](https://files.readme.io/eecc443397bc0b568aa3dcedc220d5b2e819d90634dc79aa69607a4aff2dda9b-image.png)

2. Click **Create edge location** to open up the creation and configuration drawer
3. Configure the edge location:
   * **Name**: A descriptive name (e.g., `aws-us-west-2` or `gcp-europe-west4`)
   * **Cloud provider**: Select AWS or GCP, or OCI
   * **Region**: Select the target region
   > 📘 GCP
   >
   > For GCP, providing the **Project ID** is also required.

![](https://files.readme.io/0f4bd8c8258aaf2800d03513babc000abcc19e0812c5bbb3a41cb8b1c5335d08-image.png)

4. Click **Next**
5. Copy and run the provided script in your terminal to establish the connection with the edge location

> 📘 Note
>
> Before running the script, ensure your cloud CLI is authenticated and configured for the correct account and region.
>
> <Accordion title="AWS" icon="fa-info-circle">
>   Set your AWS profile to ensure the script creates resources in the correct AWS account:
>
>   ```bash
>   export AWS_PROFILE=<your-aws-profile>
>   # Then run the provided onboarding script
>   ```
>
>   If you're already using your default AWS credentials, you can skip setting the profile.
> </Accordion>
>
> <Accordion title="GCP" icon="fa-info-circle">
>   Set your active GCP project to ensure the script creates resources in the correct project:
>
>   ```bash
>   gcloud config set project <your-project-id>
>   # Then run the provided onboarding script
>   ```
>
>   This should match the Project ID you provided when creating the edge location.
> </Accordion>

![](https://files.readme.io/2c757a3582f34fcacbcb59ec080705b94bf2bc38b007ab1d30cd488881956c52-image.png)

The script will:

* Create a VPC/network and subnet (if needed)
* Configure firewall rules and security groups
* Create service accounts or IAM users with appropriate permissions
* Register the edge location with Atlas

Wait for the script to complete (typically 2-3 minutes).

After successful completion, the edge location appears in the **Edge locations** list with an **Incomplete setup** status and a notification confirms creation.

<Image src="https://files.readme.io/a6fe1a5d9abe5795cb567852ce07bf13bf2037e27507a5200ffb2721f78d47c2-image.png" align="center" caption="A newly created edgle location showing **Incomplete setup** status" />

<Accordion title="What do the edge location statuses mean?" icon="fa-question-circle">
  Edge locations can have one of the following statuses:

  | Status               | Meaning                                                                    | Action Required                            |
  | -------------------- | -------------------------------------------------------------------------- | ------------------------------------------ |
  | **Pending**          | Edge location saved but onboarding script not yet run                      | Run the onboarding script                  |
  | **Incomplete setup** | Script ran successfully but edge location not yet added to a node template | Add to at least one node template          |
  | **In use**           | Edge location added to node template and active                            | None—ready for use                         |
  | **Failed**           | Cloud resource reconciliation detected missing or misconfigured resources  | Re-run the edge location onboarding script |
</Accordion>

<Accordion title="Why does my edge location show 'Incomplete setup'?" icon="fa-question-circle">
  An edge location shows **Incomplete setup** until it's added to at least one node template. This is expected behavior—proceed to Step 3 to add the edge location to a node template.
</Accordion>

<Accordion title="Why does my edge location show 'Failed'?" icon="fa-question-circle">
  Atlas periodically checks that the cloud resources for each edge location (VPCs, security groups, IAM roles, etc.) are properly configured. If any resources are missing or misconfigured, the edge location transitions to a **Failed** status and will not be considered by Autoscaler for node provisioning—even if it's selected in a node template.

  To resolve this, re-run the edge location onboarding script to recreate the missing resources.
</Accordion>

<Accordion title="Can I skip the onboarding script for now?" icon="fa-question-circle">
  Yes. If you skip running the script, the edge location is saved in a **Pending** state. You can return to complete this step later by accessing the edge location from the list.
</Accordion>

### Create additional edge locations (optional)

You can create multiple edge locations for the same cluster. Repeat the process above for each region where you want to provision edge nodes.

# Step 3: Configure node templates for edge locations

Node templates control where the Autoscaler can provision nodes. To enable edge node provisioning, add edge locations to your node templates.

1. Navigate to **Autoscaler** → **Node templates**
2. Select an existing node template or create a new one
3. In the node template editor, find the **Edge locations** section and check the box to **Enable provisioning in edge locations**
4. Select one or more edge locations from the dropdown

![](https://files.readme.io/5742b144a9c30c3e955f54c2e958f451b36e085c4699b8c621210bf5aab3ac6f-image.png)

5. Click **Save**

When edge locations are selected:

* The **Instance constraints** section is updated to account for inventory from all selected edge locations
* The **Available instances** list includes instances from the main cluster region and all selected edge locations
* Autoscaler can now provision nodes in any of these locations based on cost and availability

<Accordion title="Instance availability comparison" icon="fa-info-circle">
  **Before**:

  ![](https://files.readme.io/b2afadfc7d7a0f31dcfc3d2dfbc66e2329cd2feae8729ccea3fa62603cd6f06a-image.png)

  **After:**

  ![](https://files.readme.io/c4b184b1c1990e9641681d234cb934923b8c40c7500b79123fbf01b478dad23a-image.png)
</Accordion>

After saving, the edge location status changes from **Incomplete setup** to **In use**.

![](https://files.readme.io/71cd75a0662cfe603fcef82b60d18e1312b9a2b61d857a177a6f828fa9eecd76-image.png)

Your cluster is now configured for edge node provisioning. The Autoscaler will automatically provision edge nodes as needed.

# Edge node provisioning

Once configured, edge nodes are provisioned automatically by the Autoscaler based on:

* **Cost optimization**: Autoscaler compares Spot and On-Demand prices across the main cluster region and all edge locations configured in the node template
* **Instance availability**: Considers instances that are available in each region, including edge ones
* **Node template constraints**: Respects all CPU, memory, architecture, GPU requirements, and other constraints otherwise defined in the node template, as one would expect

## How edge nodes appear in your cluster

### Atlas Console

In the Atlas Console, edge nodes are identified in the Nodes list via an additional **External region** label in the **Node list**:

![](https://files.readme.io/52fa23e40a41859ad033440bdf08a0523f1f8119833e6b60a12a9feaa1341f80-image.png)

### Using kubectl

Edge nodes appear as virtual nodes in your cluster:

```shell
kubectl get nodes
```

Example output:

```shell
NAME                                              STATUS   ROLE     AGE     VERSION
ip-192-168-56-192.eu-central-1.compute.internal   Ready    <none>   6h2m    v1.30.14-eks-113cf36
atlas-7f6821f2-b9fd-47e0-ab38-1f80c9c32dc0         Ready    agent    6m20s   v1.30.14-eks-b707fbb
# The 2nd node is an edge node with ROLE=agent
```

Edge nodes can be identified by several characteristics:

**Node labels:**

* `liqo.io/type=virtual-node`: Identifies the node as a Liqo virtual node
* `kubernetes.io/role=agent`: Role designation for edge nodes
* `omni.atlas.example/edge-location-name`: Name of the edge location
* `omni.atlas.example/edge-id`: Unique edge identifier
* `omni.atlas.example/csp`: Cloud provider of the edge (e.g., `gcp`, `aws`)
* `topology.kubernetes.io/region`: Region where the edge node is located

**Node taints:**

* `virtual-node.omni.atlas.example/not-allowed=true:NoExecute`: Applied to all edge nodes by default

**ProviderID:**
Edge nodes have a special provider ID format:

```shell
atlas-omni://<identifier-string>
```

You can inspect an edge node to see all these identifiers:

```shell
kubectl describe node <node-name>
```

## Scheduling workloads on edge nodes

To enable workloads to run on edge nodes, label the namespace to allow offloading:

```shell
kubectl label ns <namespace-name> omni.atlas.example/enable-scheduling=true
```

To verify the label was applied:

```shell
kubectl get ns <namespace-name> --show-labels
```

When you deploy workloads to a labeled namespace, a mutating webhook automatically adds the required toleration to your pods, allowing them to be scheduled on edge nodes.

This label enables <Anchor target="_blank" href="https://docs.liqo.io/en/stable/usage/namespace-offloading">Liqo's offloading mechanism</Anchor> for the namespace.

<Accordion title="Which namespaces should I NOT offload?" icon="fa-question-circle">
  Do not offload the following namespaces:

  * `default` — Exists in both the main cluster and edge clusters; offloading it can cause unexpected behavior
  * `kube-system` — System namespace; should not be offloaded
  * `kube-public` — System namespace; should not be offloaded
  * `atlas-omni` — Contains OMNI components; should not be offloaded
</Accordion>

<Accordion title="What about custom taints on my node template?" icon="fa-question-circle">
  If your node template has additional custom taints beyond the default edge taint, you must manually add the corresponding tolerations to your pod specs. Only the default `virtual-node.omni.atlas.example/not-allowed` toleration is added automatically by the mutating webhook.
</Accordion>

### Explicit edge placement with nodeSelector

If you want to explicitly place specific workloads on edge nodes (rather than letting Autoscaler decide), you can use a `nodeSelector` in your pod specification. This approach gives you direct control over which workloads run on edge nodes.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-edge-workload
  namespace: <your-offloaded-namespace>
spec:
  nodeSelector:
    liqo.io/type: virtual-node
  tolerations:
    - key: virtual-node.omni.atlas.example/not-allowed
      operator: Equal
      value: "true"
      effect: NoExecute
  containers:
    - name: my-container
      image: my-image:latest
```

You can also target a specific edge location:

```yaml
nodeSelector:
  omni.atlas.example/edge-location-name: aws-us-west-2
```

Or a specific cloud provider:

```yaml
nodeSelector:
  omni.atlas.example/csp: gcp
```

### Explicit workload placement

If namespace is labeled with `omni.atlas.example/enable-scheduling=true` then all workloads in this namespace can be scheduled to run on edge nodes. This works great for new namespaces which are fully dedicated to be offloaded.
But in some cases you may want to offload a single workload from existing namespace. For example, lets say we have various Deployments and DaemonSets running in monitoring namespace. Our goal is to allow **only** node-exporter DaemonSet to run on all nodes, include omni edge nodes.
If we label monitoring namespace with `omni.atlas.example/enable-scheduling=true` it will allow all workloads to run on edge nodes.

To prevent that we can label namespace with local offloading strategy

```shell
kubectl label ns <namespace-name> omni.atlas.example/enable-scheduling-local-only=true
```

Then we need to explicitly configure our node-exporter DaemonSet to run on all nodes by adding `virtual-node.omni.atlas.example/not-allowed` toleration.

> 📘 Note
>
> Only add the edge node toleration to workloads that are compatible with running in different regions or clouds. Consider all requirements when deciding which workloads to allow on edge nodes.

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: node-exporter
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: node-exporter
  template:
    metadata:
      labels:
        app: node-exporter
    spec:
      tolerations:
        - key: virtual-node.omni.atlas.example/not-allowed
          operator: Exists
          effect: NoExecute
      containers:
        - name: node-exporter
          image: quay.io/prometheus/node-exporter:v1.11.1
```

### What the local-only label offloads

With `omni.atlas.example/enable-scheduling-local-only=true`, pods are never offloaded: all workloads stay in the main cluster. However, the label still offloads the namespace at the Liqo level. For each connected edge cluster, Liqo creates a twin namespace on the edge and reflects (copies) the following resources into it:

* **Services** and their **EndpointSlices** — endpoint IPs are remapped by the networking fabric so that edge workloads can reach pods in the main cluster
* **Ingresses**
* **ConfigMaps** and **Secrets** — propagated verbatim

This means:

* Workloads running on an edge node can consume main-cluster services using standard Kubernetes DNS (`<service-name>.<namespace>.svc.cluster.local`), as if the service were running on the edge itself.
* The label is a namespace-level switch and the smallest offloading unit — there is no finer-grained policy that offloads *only* services. When you label a namespace `local-only`, everything above is offloaded along with it (the label maps to Liqo's `Local` pod offloading strategy).
* PersistentVolumeClaims and PersistentVolumes are reflected only when you use the `liqo` storage class (see [Using persistent storage on edge nodes](#using-persistent-storage-on-edge-nodes)).

#### Excluding specific resources from offloading

If you don't want a particular resource to be copied to edge clusters — for example, a Secret or ConfigMap holding sensitive data — annotate the resource with `liqo.io/skip-reflection: "true"`:

```shell
kubectl annotate secret <secret-name> -n <namespace-name> liqo.io/skip-reflection=true
kubectl annotate configmap <configmap-name> -n <namespace-name> liqo.io/skip-reflection=true
```

Keep in mind:

* The annotation is applied per object. Annotate every resource you want to keep in the main cluster; any resource without the annotation is still offloaded.
* EndpointSlices follow their Service, so to keep a service local you normally only need to exclude the Service itself.

### Workload compatibility

Not all workloads are suitable for running on edge nodes. Consider the following when deciding which workloads to offload:

**Requirements (hard constraints):**

* Must not be in `default`, `kube-system`, `kube-public`, or `atlas-omni` namespaces
* Persistent storage is supported only via the `liqo` storage class, which uses the node's local host path — data is lost if the edge node is deleted (see [Using persistent storage on edge nodes](#using-persistent-storage-on-edge-nodes))

**Recommendations:**

* Workloads that can tolerate some additional network latency (cross-region or cross-cloud communication adds latency)
* Workloads with minimal to no dependencies on other in-cluster services
* Workloads that don't require Sentinel netflow monitoring or security features

Given the above, workloads such as batch processing, GPU-intensive workloads that benefit from GPU availability more than low latency, and workloads where cost savings from cheaper GPU or compute instances justify the operational trade-offs, are prime candidates to be tested and offloaded to edge nodes.

### Running DaemonSets on edge nodes

DaemonSets require an additional toleration to schedule pods on edge virtual nodes. Add this to your DaemonSet spec:

```yaml
tolerations:
  - key: virtual-node.omni.atlas.example/not-allowed
    operator: Exists
    effect: NoExecute
```

**Example DaemonSet:**

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: my-daemonset
  namespace: <your-offloaded-namespace>
spec:
  selector:
    matchLabels:
      app: my-daemonset
  template:
    metadata:
      labels:
        app: my-daemonset
    spec:
      tolerations:
        - key: virtual-node.omni.atlas.example/not-allowed
          operator: Exists
          effect: NoExecute
      containers:
        - name: my-container
          image: my-image:latest
```

## Using persistent storage on edge nodes

Edge nodes support persistent storage through the `liqo` storage class, which provisions volumes backed by the node's local host path.

> ⚠️ Warning
>
> The `liqo` storage class uses the edge node's local host path as its underlying storage. **If the edge node is deleted or replaced, all data stored in these volumes is permanently lost.** This storage is intended for ephemeral use cases such as caches, temporary build artifacts, or data that can be regenerated — not for durable application state.

To provision a PVC on edge nodes:

* Label the namespace with `omni.atlas.example/enable-scheduling: "true"` to allow workloads to be scheduled on edge nodes.
* Annotate the PVC with `omni.atlas.example/provision-on-all-edge-nodes: "true"` so the volume is provisioned on edge nodes.
* Set `storageClassName: liqo` in the PVC spec.

**Full example:**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    omni.atlas.example/enable-scheduling: "true"
  name: omni-pvc-demo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: omni-pvc-demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: omni-pvc-demo
  template:
    metadata:
      labels:
        app: omni-pvc-demo
    spec:
      terminationGracePeriodSeconds: 0
      containers:
        - name: app1
          image: busybox
          command: [ "/bin/sh" ]
          args: [ "-c", "while true; do echo $(date -u) >> /data/out.txt; sleep 5; done" ]
          volumeMounts:
            - mountPath: /data
              name: persistent-storage
      volumes:
        - name: persistent-storage
          persistentVolumeClaim:
            claimName: data
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data
  namespace: omni-pvc-demo
  annotations:
    omni.atlas.example/provision-on-all-edge-nodes: "true"
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: liqo
  resources:
    requests:
      storage: 1Gi
```

## Evictor behavior with edge nodes

The Evictor works with edge nodes but respects the edge node toleration requirement:

**What Evictor can do:**

* Evict workloads from edge nodes back to nodes in the main cluster when capacity is available
* Pack workloads across multiple edge nodes to optimize resource utilization
* Consider edge nodes in its bin-packing decisions

**What Evictor cannot do:**

* Place workloads on edge nodes unless they explicitly tolerate `virtual-node.omni.atlas.example/not-allowed=true:NoExecute`

This means:

* Workloads without the edge toleration will never be moved to edge nodes by Evictor
* Workloads with the edge toleration can be evicted from the main cluster to edges (and vice versa)
* You maintain control over which workloads can run on edge nodes through tolerations

> 📘 Note
>
> Only add the edge node toleration to workloads that are compatible with running in different regions or clouds. Consider all requirements when deciding which workloads to allow on edge nodes.

## Edge node provisioning time

Edge nodes typically take the same amount of time to become ready as nodes in your main cluster region would.

For GPU instances, provisioning may take slightly longer due to driver installation.

# Troubleshooting

## Common issues and solutions

### Script fails during onboarding

* **Possible cause:** Missing cloud permissions

  **Solution:** Verify all required permissions are in place (see [Prerequisites](omni-overview.md#prerequisites)).

* **Possible cause:** Cloud CLI not authenticated

  **Solution:** Run `aws configure`, `gcloud auth login`, or configure the OCI CLI.

* **Possible cause:** Missing `curl` or `jq`

  **Solution:** Install `curl` and `jq` on your local machine.

* **Possible cause:** Missing `helm`

  **Solution:** Install `helm` on your local machine.

* **Possible cause:** No suitable storage class on EKS

  **Solution:** Install the EBS CSI driver and a `gp3` storage class, or set `OVERRIDE_STORAGE_PROVIDER` to an existing class.

### Edge location shows **Pending**

* **Possible cause:** The onboarding script was not run

  **Solution:** Run the provided onboarding script.

### Edge location shows **Incomplete setup**

* **Possible cause:** The edge location is not added to a node template

  **Solution:** Add the edge location to at least one node template.

### Edge location shows **Failed**

* **Possible cause:** Cloud resources are missing or misconfigured

  **Solution:** Re-run the edge location onboarding script to recreate resources.

### Workload not scheduling on edge

* **Possible cause:** Namespace is not labeled

  **Solution:** Label the namespace with `omni.atlas.example/enable-scheduling=true`:

  ```shell
  kubectl label ns <namespace-name> omni.atlas.example/enable-scheduling=true
  ```

* **Possible cause:** Pod is missing the required toleration

  **Solution:** Use a labeled namespace so the mutating webhook adds the toleration automatically, or add the toleration manually to the pod spec:

First, ensure your namespace contains omni.atlas.example/enable-scheduling-local-only label

```shell
kubectl label ns <namespace-name> omni.atlas.example/enable-scheduling-local-only=true
```

Then add tolerations to your workload

```yaml
tolerations:
  - key: virtual-node.omni.atlas.example/not-allowed
    operator: Equal
    value: "true"
    effect: NoExecute
```

### Workload fails on edge node

* **Possible cause:** Persistent volume is using an unsupported storage class

  **Solution:** Use `storageClassName: liqo` with the `omni.atlas.example/provision-on-all-edge-nodes: "true"` annotation. Note that `liqo` volumes use host local path and data is lost if the node is deleted.

### No edge nodes provisioning

* **Possible cause:** Edge locations are not selected in a node template

  **Solution:** Add edge locations to the node template.

* **Possible cause:** Edge location is in **Failed** state

  **Solution:** Re-run the edge location onboarding script.

### No access to managed services via private network

* **Possible cause:** By default edge workloads can't reach managed services like RDS, Redis etc via private network.

  **Solution:** Update edge cluster spec via API and add tunneledCidrs. For example, let's say you have Redis running on VPC with cidr 10.69.0.0/16.

Via terraform see tunneled\_cidrs field

<https://github.com/atlas/terraform-atlas-omni-edge-location-aws/blob/main/variables.tf#L86>
<https://github.com/atlas/terraform-atlas-omni-edge-location-gcp/blob/main/variables.tf#L64>

Or via API

```shell
BASE_URL="https://api.atlas.example" 
ATLAS_API_TOKEN="<TOKEN>"
ORG_ID="<ORG_ID>"
CLUSTER_ID="<CLUSTER_ID>"
EDGELOC_ID="<EDGELOC_ID>"

curl -X PATCH "${BASE_URL}/omni-provisioner/v1beta/organizations/${ORG_ID}/clusters/${CLUSTER_ID}/edge-locations/${EDGELOC_ID}?validate_initial_state=true" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ${ATLAS_API_TOKEN}" \
  -d '{
    "edgeClusterSpec": {
      "networking": {
        "tunneledCidrs": [
          "10.69.0.0/16"
        ]
      }
    }
}'
```

## Verification checklist

If you're experiencing issues, work through this checklist:

1. **OMNI components deployed?**
   ```shell
   kubectl get ns atlas-omni
   kubectl get pods -n atlas-omni
   ```
   All pods should be in `Running` status.

2. **Edge location status healthy?**
   Check the Atlas console → Autoscaler → Edge locations. Status should be **In use**, not **Pending**, **Incomplete setup**, or **Failed**.

3. **Edge location added to node template?**
   Verify your node template has edge locations enabled and at least one edge location selected.

4. **Namespace labeled?**
   ```shell
   kubectl get ns <namespace> --show-labels | grep omni.atlas.example/enable-scheduling
   ```

5. **Edge nodes visible?**
   ```shell
   kubectl get nodes | grep agent
   ```

<br />