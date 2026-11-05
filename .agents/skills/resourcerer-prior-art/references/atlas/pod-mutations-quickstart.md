---
updatedAt: 2026-09-10T08:10:32.000Z
---

# Quickstart

This quickstart guide walks you through installing the Pod Mutator Controller and creating your first pod mutation. By the end, you'll have a working mutation that automatically adds a label to pods in a target namespace.

The quickstart covers:

* [Part 1: Install the pod mutator](#part-1-install-the-pod-mutator)
* [Part 2: Create a pod mutation](#part-2-create-a-pod-mutation)
* [Part 3: Configure target filters](#part-3-configure-target-filters)
* [Part 4: Verify the mutation](#part-4-verify-the-mutation)

This guide assumes you have a basic familiarity with Kubernetes concepts, such as namespaces, pods, and labels, and that you can run kubectl commands against your cluster.

# Before you start

Before starting this quickstart, ensure you have:

* A Kubernetes cluster connected to Atlas
* kubectl configured to access your cluster
* The following Atlas components at minimum versions:
  * `atlas-agent`: 0.123.0 or higher
  * `atlas-cluster-controller`: 0.85.0 or higher

You can verify your component versions with:

```shell
helm list -n atlas-agent --filter 'atlas-agent|cluster-controller'
```

Example output:

```shell
robertas@Robertass-MacBook-Pro ~ % helm list -n atlas-agent --filter 'atlas-agent|cluster-controller'
NAME              	NAMESPACE   	REVISION	UPDATED                                	STATUS  	CHART                           	APP VERSION
atlas-agent      	atlas-agent	1       	2025-12-17 10:33:15.856274144 +0000 UTC	deployed	atlas-agent-0.129.0            	v0.108.0   
cluster-controller	atlas-agent	1       	2025-12-17 10:36:27.258252445 +0000 UTC	deployed	atlas-cluster-controller-0.86.0	v0.60.0    
```

# Part 1: Install the pod mutator

The pod mutator is the controller that applies mutations to pods at creation time. You can install it via the Atlas console or Helm.

## Option A: Install via console

1. In the Atlas console, select your cluster from the cluster list.

2. In the Classic console, navigate to **Autoscaler** → **Pod mutations** in the sidebar.

   <Accordion title="I'm using Enhanced console" icon="fa-info-circle">
     In the Enhanced console, navigate to **Settings** in the sidebar, then **Pod mutations**.

     ![](https://files.readme.io/575f5c7c9131fe7b3ab262ee66e99745805a1ce2d51a27773399c5a4f2f72652-image.png)
   </Accordion>

   If the pod mutator is not installed, you'll see an installation prompt with a script. Copy it.

   <Image border={false} src="https://files.readme.io/6e436ef4481838ddd37d6ea23cca03a08c580038e3870e3f7129c779f2a97b32-image.png" />

3. Open your cloud shell or terminal and run the script to install the pod mutator in your cluster.

## Option B: Install via Helm

1. Add the Atlas Helm repository:

   ```shell
   helm repo add atlas-helm https://atlas.github.io/helm-charts
   helm repo update
   ```

2. Install the pod mutator:

   <Tabs>
     <Tab title="Standard (EKS, AKS, GKE)">

   ```shell
   helm upgrade atlas atlas-helm/atlas -n atlas-agent \
     --reset-then-reuse-values \
     --set autoscaler.atlas-pod-mutator.atlas.apiUrl="$ATLAS_API_URL" \
     --set autoscaler.atlas-pod-mutator.atlas.apiKey="${ATLAS_API_KEY}" \
     --set autoscaler.atlas-pod-mutator.atlas.clusterID="${CLUSTER_ID}"
   ```

     </Tab>
     <Tab title="Anywhere clusters">

   On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

   ```shell
   helm upgrade atlas atlas-helm/atlas -n atlas-agent \
     --reset-then-reuse-values \
     --set autoscaler-anywhere.atlas-pod-mutator.atlas.apiUrl="$ATLAS_API_URL" \
     --set autoscaler-anywhere.atlas-pod-mutator.atlas.apiKey="${ATLAS_API_KEY}" \
     --set autoscaler-anywhere.atlas-pod-mutator.atlas.clusterID="${CLUSTER_ID}"
   ```

     </Tab>
   </Tabs>

   Replace `${ATLAS_API_KEY}` and `${CLUSTER_ID}` with your actual values. You can find these in the Atlas console under **User** → **API keys** and in your cluster's settings.

<Accordion title="Where to find your API key and Cluster ID" icon="fa-key">
  **API Key:**

  1. In the Atlas console, click your profile in the top right corner.
  2. Select **API keys** from the dropdown menu.
  3. Use an existing key, or click **Create access key** to generate a new one.

  ![](https://files.readme.io/43aef2c88cf151e6b8d1d6b14f7f32f503c235eb25a820694c9cedcde68798cc-image.png)

  **Cluster ID:**

  1. Navigate to your cluster list in the Atlas console.
  2. Select the cluster you want to install the pod mutator on. It will take you to **Cluster overview** > **Dashboard** in the sidebar navigation.
  3. Copy the **Cluster ID** value from the cluster details section.

  ![](https://files.readme.io/5a9c163e516d62a8954439c97fb3ba6db8649d0fa0f5e64bd5fca1a100acbdd0-image.png)
</Accordion>

## Verify installation

Confirm the pod mutator is running:

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-pod-mutator
```

Expected output:

```shell
NAME                                  READY   STATUS    RESTARTS      AGE
atlas-pod-mutator-767d48f477-4tdrw   1/1     Running   0             172m
atlas-pod-mutator-767d48f477-d7s45   1/1     Running   1 (97s ago)   172m
```

# Part 2: Create a pod mutation

With the pod mutator installed, you can create your first mutation. This example creates a mutation that adds a label to pods.

1. In the Atlas console, navigate to **Autoscaler** → **Pod mutations**.
   <Accordion title="I'm using Enhanced console" icon="fa-info-circle">
     In the Enhanced console, navigate to **Settings** in the sidebar, then **Pod mutations**.

     ![](https://files.readme.io/575f5c7c9131fe7b3ab262ee66e99745805a1ce2d51a27773399c5a4f2f72652-image.png)
   </Accordion>

2. Click **Add mutation**.

3. Enter a name for your mutation in the **Mutation name** field:

   ```
   demo-label-mutation
   ```

   > 📘 Note
   >
   > Mutation names cannot be changed after creation. Always choose a descriptive name that reflects what the mutation does.

4. Under **Default configuration**, click the **Configuration** dropdown and select **Labels**.

5. In the Labels configuration that appears, add a label:

   * **Key:** `mutated-by`
   * **Value:** `atlas`

6. Click **Go to target workloads** to proceed to filter configuration.

<Image border={false} src="https://files.readme.io/56cc0342dfbe0ee1a15b61e12cb7cf54af6f8cfa9540bba109d9b864773e9992-image.png" />

# Part 3: Configure target filters

Filters determine which pods receive the mutation. For this quickstart, we'll target pods in a specific namespace.

1. Under **Filter target workloads**, click the **Filter type** dropdown and select the **Include namespace** filter.

2. Select a namespace where you want to test the mutation. If you don't have a test namespace, you can create one:

   ```shell
   kubectl create namespace pod-mutation-demo
   ```

   Then select `pod-mutation-demo` in the console.

3. Click **Go to review** to review the mutation configuration.

   <Image border={false} src="https://files.readme.io/0fd773b62c872c26975d64cf0edacc25b669233fd719a4378090043c8d326afb-image.png" />

4. (Optional) Review the **Configuration summary** of your mutation configuration:

   <Image border={false} src="https://files.readme.io/2e9aa1d0c8aae992471389f3214eb6ba199d2a221ea601f230c5b82a2f82e7ff-image.png" />

5. (Optional) Review the **Affected/Unaffected workloads**. This shows existing workloads match your filter criteria, split into:

* **Affected:** Workloads that will receive the mutation
* **Unaffected:** Workloads that will not receive it

For the purpose of this quickstart, we've deployed a simple workload with 4 replicas, called `mutation-demo`:

<Image border={false} src="https://files.readme.io/d683668be4d67a3b5cbf9843274a80d8b9b30a0bd81f1187058ee6d3ff4242d8-image.png" />

6. Click **Create**  and confirm enablement to proceed.

The mutation is now active. Any new pods created in the target namespace will automatically receive the `mutated-by: atlas` label.

# Part 4: Verify the mutation

Test that the mutation is working by creating a pod in the target namespace.

1. Create a test pod:

   ```shell
   kubectl run test-pod --image=nginx -n pod-mutation-demo
   ```

2. Wait for the pod to be created, then check its labels:

   ```shell
   kubectl get pod test-pod -n pod-mutation-demo -o jsonpath='{.metadata.labels}' | jq
   ```

   Expected output (formatted):

   ```json
   {
     "mutated-by": "atlas",
     "run": "test-pod"
   }
   ```

   The `mutated-by: atlas` label confirms the mutation was applied.

3. You can also verify by describing the pod:

   ```shell
   kubectl describe pod test-pod -n pod-mutation-demo | grep -A5 "Labels:"
   ```

4. Clean up the test resources:

   ```shell
   kubectl delete namespace pod-mutation-demo
   ```

5. (Optional) Delete the test mutation in the Atlas console by navigating to **Autoscaler** → **Pod mutations**, clicking on `demo-label-mutation`, and selecting **Delete mutation**:

   <Image border={false} src="https://files.readme.io/54d2772cc5406d65f85cf335be5368918ee9742fb0ddd6939aba281f5a903a35-image.png" />

# Troubleshooting

If the mutation isn't applied to your test pod, check the following:

**Pod mutator not running:**

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-pod-mutator
```

If no pods are listed or the status isn't `Running`, review the installation steps.

**Check pod mutator logs:**

```shell
kubectl logs -n atlas-agent -l app.kubernetes.io/name=atlas-pod-mutator --tail=50
```

Look for errors related to your mutation or webhook configuration.

**Mutation not synced to cluster:**

Mutations created in the console are synced to your cluster as custom resources. Verify the mutation exists:

```shell
kubectl get podmutations.pod-mutations.atlas.example 
```

**Namespace filter mismatch:**

Ensure the namespace you're creating pods in matches the filter configured in the mutation.

# Next steps

Now that you've created your first mutation, explore more advanced use cases:

* [Pod mutations reference](pod-mutations-reference.md) — Complete reference for all configuration options and CRD specification

<br />