---
updatedAt: 2026-09-10T08:10:32.000Z
---

# Enable Workload Autoscaler with pod mutations

## Overview

In this tutorial, you'll learn how to use pod mutations to automatically enable Atlas Workload Autoscaler on workloads without modifying deployment manifests. This approach is useful when you want to roll out vertical autoscaling across namespaces or workload types without requiring application teams to update their configurations.

This tutorial assumes you have basic knowledge of:

* Kubernetes workloads (Deployments, pods, labels)
* Atlas pod mutations concepts
* Atlas Workload Autoscaler functionality

By the end of this tutorial, you'll be able to:

* Create a pod mutation that enables Workload Autoscaler using the custom workloads feature
* Target specific workloads using namespace and label filters
* Verify that the mutation is correctly enabling autoscaling on new pods

## Background

[Atlas Workload Autoscaler](workload-autoscaling-overview.md) optimizes container resource requests based on actual usage patterns. To enable it on standard workloads, you typically add the `workloads.atlas.example/configuration` annotation to your Deployment's metadata. However, pod mutations operate at the pod level, not the controller level, so they cannot modify Deployment resources directly.

To work around this, we use the [custom workloads](workload-autoscaling-configuration.md#custom-workload-support) feature. Custom workloads are identified by a label on the pod itself, and their configuration annotation is also read from the pod. This makes them ideal for pod mutations.

The mutation adds two items to each matching pod:

* **`workloads.atlas.example/custom-workload` label** — Registers the pod as a custom workload with Workload Autoscaler
* **`workloads.atlas.example/configuration` annotation** — Contains the autoscaling configuration

This approach lets you enable Workload Autoscaler across your cluster declaratively, without touching application manifests or requiring coordination with development teams.

<Accordion title="Custom workloads vs. standard workloads" icon="fa-info-circle">
  **Standard workloads** read their Workload Autoscaler configuration from the parent controller (Deployment, StatefulSet, etc.). This is the typical approach when you control the application manifests.

  **Custom workloads** read their configuration from pod-level labels and annotations. This is useful for:

  * Workloads managed by custom controllers
  * Bare pods without parent controllers
  * Scenarios where you want to inject configuration via pod mutations

  Both approaches provide the same optimization capabilities. The main difference is where the configuration lives and how workloads are identified in the Workload Autoscaler UI.

  For more details, see [Custom workload support](workload-autoscaling-configuration.md#custom-workload-support).
</Accordion>

## Before you start

Before starting this tutorial, ensure you have:

* A Kubernetes cluster connected to Atlas
* The pod mutator installed and running (see [Pod mutations quickstart](pod-mutations-quickstart.md))
* Workload Autoscaler enabled on your cluster (see [Workload Autoscaler overview](workload-autoscaling-overview.md))
* kubectl configured to access your cluster
* A scaling policy created (or use the `default` policy)

<Accordion title="How to verify the pod mutator is running" icon="fa-info-circle">
  Run the following command to check that the pod mutator pods are running:

  ```shell
  kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-pod-mutator
  ```

  You should see output similar to:

  ```shell
  NAME                                  READY   STATUS    RESTARTS   AGE
  atlas-pod-mutator-767d48f477-lshm2   1/1     Running   0          10m
  atlas-pod-mutator-767d48f477-wvprs   1/1     Running   0          10m
  ```

  If no pods are listed or the status isn't `Running`, see the installation accordion below.
</Accordion>

<Accordion title="How to install the pod mutator" icon="fa-download">
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
       --set autoscaler.atlas-pod-mutator.atlas.apiKey="${ATLAS_API_KEY}" \
       --set autoscaler.atlas-pod-mutator.atlas.clusterID="${CLUSTER_ID}"
     ```

       </Tab>
       <Tab title="Anywhere clusters">

     On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

     ```shell
     helm upgrade atlas atlas-helm/atlas -n atlas-agent \
       --reset-then-reuse-values \
       --set autoscaler-anywhere.atlas-pod-mutator.atlas.apiKey="${ATLAS_API_KEY}" \
       --set autoscaler-anywhere.atlas-pod-mutator.atlas.clusterID="${CLUSTER_ID}"
     ```

       </Tab>
     </Tabs>

     Replace `${ATLAS_API_KEY}` and `${CLUSTER_ID}` with your actual values. You can find these in the Atlas console under **User** → **API keys** and in your cluster's settings.

  3. Verify the installation:

     ```shell
     kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-pod-mutator
     ```

  For detailed installation options, see the [Pod mutations quickstart](pod-mutations-quickstart.md).
</Accordion>

<Accordion title="How to verify Workload Autoscaler is enabled" icon="fa-info-circle">
  Run the following command to check that the Workload Autoscaler component is running:

  ```shell
  kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler
  ```

  You should see output similar to:

  ```shell
  NAME                                         READY   STATUS    RESTARTS   AGE
  atlas-workload-autoscaler-5f8b7c9d4-xxxxx   1/1     Running   0          10m
  ```

  You can also verify in the Atlas console by navigating to **Workload Autoscaler** in the sidebar. If the component is not running, see the installation accordion below.
</Accordion>

<Accordion title="How to install Workload Autoscaler" icon="fa-download">
  1. Add the Atlas Helm repository (if not already added):

     ```shell
     helm repo add atlas-helm https://atlas.github.io/helm-charts
     helm repo update
     ```

  2. Install the Workload Autoscaler:

     <Tabs>
       <Tab title="Standard (EKS, AKS, GKE)">

     ```shell
     helm upgrade atlas atlas-helm/atlas -n atlas-agent \
       --reset-then-reuse-values \
       --set autoscaler.atlas-workload-autoscaler.atlas.apiKey="${ATLAS_API_KEY}" \
       --set autoscaler.atlas-workload-autoscaler.atlas.clusterID="${CLUSTER_ID}"
     ```

       </Tab>
       <Tab title="Anywhere clusters">

     On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

     ```shell
     helm upgrade atlas atlas-helm/atlas -n atlas-agent \
       --reset-then-reuse-values \
       --set autoscaler-anywhere.atlas-workload-autoscaler.atlas.apiKey="${ATLAS_API_KEY}" \
       --set autoscaler-anywhere.atlas-workload-autoscaler.atlas.clusterID="${CLUSTER_ID}"
     ```

       </Tab>
     </Tabs>

     Replace `${ATLAS_API_KEY}` and `${CLUSTER_ID}` with your actual values.

  3. Verify the installation:

     ```shell
     kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler
     ```

  For detailed configuration options and scaling policies, see [Workload Autoscaler overview](workload-autoscaling-overview.md).
</Accordion>

## Deploy a test workload

First, create a test namespace and deployment that we'll use for this tutorial to apply and verify the mutation.

1. Create a namespace for testing:

   ```shell
   kubectl create namespace woop-mutation-demo
   ```

2. Deploy a sample workload without any Workload Autoscaler annotations:

   ```shell
   kubectl apply -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: demo-app
     namespace: woop-mutation-demo
   spec:
     replicas: 2
     selector:
       matchLabels:
         app: demo-app
     template:
       metadata:
         labels:
           app: demo-app
           enable-autoscaling: "true"
       spec:
         containers:
           - name: nginx
             image: nginx:1.25
             resources:
               requests:
                 cpu: 100m
                 memory: 128Mi
               limits:
                 cpu: 200m
                 memory: 256Mi
   EOF
   ```

   <Accordion title="About the test workload configuration" icon="fa-info-circle">
     This deployment includes:

     * **Two replicas**: Allows you to verify the mutation applies consistently across pods
     * **`enable-autoscaling: "true"`label**: Provides a targeting mechanism for the mutation filter, demonstrating label-based filtering
     * **Resource requests and limits**: Gives Workload Autoscaler baseline values to optimize

     The label `enable-autoscaling` is arbitrary—you can use any label that fits your organization's conventions. In production, you might target workloads by team, environment, or application tier.
   </Accordion>

3. Verify the deployment is running:

   ```shell
   kubectl get pods -n woop-mutation-demo
   ```

Expected output:

```shell
NAME                       READY   STATUS    RESTARTS   AGE
demo-app-869684576-6tzm9   1/1     Running   0          11s
demo-app-869684576-cwhjn   1/1     Running   0          11s
```

## Create the pod mutation

Now create a pod mutation that registers matching pods as custom workloads and enables Workload Autoscaler. The mutation uses a JSON patch to add both the required label and configuration annotation.

1. In the Atlas console, navigate to **Autoscaler** → **Pod mutations**.

2. Click **Add mutation**.

   <Image border={false} src="https://files.readme.io/e587721527fbd5b8a0c35bf5c35d742bb7d2d0338f32277c41bdef8ab93478c1-image.png" />

3. Enter a name for your mutation:

   ```
   enable-workload-autoscaler
   ```

4. Under **Default configuration**, click the **Configuration** dropdown and select **Custom JSON patch**.

5. Enter the following JSON patch operation:

   ```json
   [
     {
       "op": "add",
       "path": "/metadata/labels/workloads.atlas.example~1custom-workload",
       "value": "demo-app"
     },
     {
       "op": "add",
       "path": "/metadata/annotations/workloads.atlas.example~1configuration",
       "value": "vertical:\n  optimization: on"
     }
   ]
   ```

   <Accordion title="Understanding the JSON patch" icon="fa-info-circle">
     This JSON patch performs two operations:

     **1. Add the custom workload label:**

     ```json
     {"op":"add","path":"/metadata/labels/workloads.atlas.example~1custom-workload","value":"demo-app"}
     ```

     This registers the pod as a custom workload. The label value (`demo-app`) serves as an identifier—pods with the same value are treated as a single workload for optimization purposes.

     **2. Add the configuration annotation:**

     ```json
     {"op":"add","path":"/metadata/annotations/workloads.atlas.example~1configuration","value":"vertical:\n  optimization: on"}
     ```

     This enables vertical autoscaling with default settings. The `\n` represents an actual newline character in the YAML configuration.

     Note that `/` in annotation and label keys must be escaped as `~1` per RFC 6901 (JSON Pointer).
   </Accordion>

   <Accordion title="Customizing the workload identifier" icon="fa-info-circle">
     The `workloads.atlas.example/custom-workload` label value serves as the workload identifier. In this tutorial, we use a static value (`demo-app`), but in production you might want to:

     * Use namespace-specific values to separate workloads across namespaces
     * Create different mutations for different application groups
     * Match the value to your existing workload naming conventions

     Pods with the same label value in the same namespace are grouped together as a single workload in the Workload Autoscaler UI.
   </Accordion>

   <Image border={false} src="https://files.readme.io/34add61a12a89f5df12dfabd59e7a3ff3f288de89a5da3fd923cf2419a630e8a-image.png" />

   <br />

6. Click **Go to target workloads** to proceed to filter configuration.

## Configure target filters

Configure filters to target the workloads that should receive the Workload Autoscaler configuration.

1. Under **Filter target workloads**, click the **Filter type** dropdown and select **Include namespaces**.

2. Select `woop-mutation-demo` from the namespace list.

3. (Optional) Add a label filter to target only pods with the `enable-autoscaling: "true"` label:

   a. Click **Add filter type** and select **Include pod labels**.

   b. Add the label condition:

   * **Key:** `enable-autoscaling`
   * **Value:** `true`

   <Image border={false} src="https://files.readme.io/20bb473fbea24a113b4d6f9c76771826d97cd3bc8e0fc2b910c595d19ad7a0a3-image.png" />

4. Click **Go to review** to see the configuration summary.

5. Review the **Affected workloads** panel to confirm your `demo-app` deployment appears in the affected list.

   <Image border={false} src="https://files.readme.io/1e9cfa05c5615ac9fd69ce0a326ac0f7524381d60e431c540093e99c1c17ac94-image.png" />

6. Click **Create** and confirm enablement.

   At this point, the mutation will be enabled and active, but it will show that 0 pods are mutated. This is because we need to restart the pods, since mutations apply at creation time, while our deployment is already running.

   <Image border={false} src="https://files.readme.io/dfa15169271a6788ec97a4943b89d17a8f50903452ff279d9c0764f5eed17146-image.png" />

## Verify the mutation

Test that the mutation is correctly adding the custom workload label and Workload Autoscaler annotation to new pods.

1. Delete the existing pods to trigger recreation with the mutation applied:

   ```shell
   kubectl delete pods -n woop-mutation-demo -l app=demo-app
   ```

2. Wait for the new pods to be created:

   ```shell
   kubectl get pods -n woop-mutation-demo -w
   ```

3. Check that the new pods have the custom workload label:

   ```shell
   kubectl get pods -n woop-mutation-demo -l app=demo-app -o jsonpath='{.items[0].metadata.labels.workloads\.atlas\.example/custom-workload}'
   ```

   Expected output:

   ```
   demo-app
   ```

4. Check that the pods have the Workload Autoscaler configuration annotation:

   ```shell
   kubectl get pod -n woop-mutation-demo -l app=demo-app -o json | jq '.items[0].metadata.annotations["workloads.atlas.example/configuration"]'
   ```

   Expected output:

   ```
   "vertical:\n  optimization: on"
   ```

5. Verify in the Atlas console that the workload appears in **Workload Autoscaler** as a custom workload:

   a. Navigate to **Workload Autoscaler** → **Optimization** in the sidebar.

   b. Find `demo-app` in the workloads list. It will appear as a custom workload.

   c. Confirm the workload shows up as being optimized.

   <Image border={false} src="https://files.readme.io/26b1b18d84bc4e239973cc571072b981dc4a4edcae0f054dc1dcb216737a2918-image.png" />

   <br />

## Clean up

Remove the test resources when you're done:

1. Delete the test namespace:

   ```shell
   kubectl delete namespace woop-mutation-demo
   ```

2. (Optional) Delete the mutation in the Atlas console:

   a. Navigate to **Autoscaler** → **Pod mutations**.

   b. Click on `enable-workload-autoscaler`.

   c. Click **Delete mutation** and confirm.

## Summary

In this tutorial, you learned how to:

* Create a pod mutation using JSON patch to register pods as custom workloads and enable Workload Autoscaler
* Use namespace and label filters to target specific workloads
* Verify that new pods receive the configuration and appear in Workload Autoscaler

This approach enables you to roll out Workload Autoscaler across your cluster without modifying application manifests, using the custom workloads feature to work around the controller-level configuration requirement.

## Mutation reference

After creating the mutation, it exists as a `PodMutation` custom resource in your cluster. Below is the YAML representation for reference.

### PodMutation YAML

```yaml
apiVersion: pod-mutations.atlas.example/v1
kind: PodMutation
metadata:
  annotations:
    pod-mutations.atlas.example/pod-mutation-id: 5c9e70fb-a9bc-4237-b2e1-6ca6467dd61d
    pod-mutations.atlas.example/pod-mutation-name: enable-workload-autoscaler
    pod-mutations.atlas.example/pod-mutation-source: api
  creationTimestamp: "2025-12-22T16:44:34Z"
  generation: 2
  name: api-mutation-5c9e70fb-a9bc-4237-b2e1-6ca6467dd61d
  resourceVersion: "19682"
  uid: 4d605ae7-400f-4c61-9361-101858dfb89a
spec:
  filter:
    pod: {}
    workload: {}
  filterV2:
    pod:
      excludeLabels: {}
      labels:
        matchers:
        - keyMatcher:
            type: exact
            value: enable-autoscaling
          valueMatcher:
            type: exact
            value: "true"
        operator: or
    workload:
      namespaces:
      - type: exact
        value: woop-mutation-demo
  patchesV2:
  - operations:
    - op: add
      path: /metadata/labels/workloads.atlas.example~1custom-workload
      value: demo-app
    - op: add
      path: /metadata/annotations/workloads.atlas.example~1configuration
      value: |-
        vertical:
          optimization: on
  restartPolicy: deferred
  spotConfig: {}
```

## Next steps

* [Custom workload support](workload-autoscaling-configuration.md#custom-workload-support) — Learn more about custom workloads and their configuration options
* [Workload Autoscaler annotations reference](workload-autoscaler-annotations-reference.md) — Complete reference for all annotation configuration options
* [Pod mutations reference](pod-mutations-reference.md) — Complete CRD specification and configuration options