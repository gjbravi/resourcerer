# Troubleshooting

> Get help with common questions, errors, and warnings


**[Common troubleshooting steps](#key-points)**
- [Agent troubleshooting](#trbl-agent)
  - [Agent status](#trbl-agent-stat)
  - [Agent logs](#trbl-agent-depl-logs)
  - [ConfigMap settings](#trbl-agent-edit-configmap)
  - [Restart the Agent](#trbl-agent-restart)
- [Applier troubleshooting](#trbl-applier)
  - [Applier logs](#trbl-applier-depl-logs)
- [Node Agent troubleshooting](#trbl-node-agent)
  - [Node Agent logs](#trbl-node-agent-logs)
  - [Restart the Node Agent](#trbl-node-agent-restart)
  - [No pods on GPU nodes](#trbl-node-agent-not-scheduled)
  - [GPU data missing](#trbl-node-agent-no-gpu-data)

**[Installation](#installation)**

- [Health check failed with a `Failed to create Orion client` message: What should I do?](#fail-create-client)
- [The installation failed. What can I check?](#install-failed)

**[Workloads and recommendations](#workloads-recs)**

- [I don't see the workloads (or namespaces) that I expect on the dashboard. What can I check?](#check-namespaces)
- [How do I grant RBAC permissions on a patch target?](#additional-rbac)
- [I didn't get a recommendation for a workload. What can I check?](#no-rec)
- [Recommendations don't seem right based on the CPU/memory utilization. What can I do?](#workload-profile-change)
- [My cluster is disconnected. What can I check?](#cluster-disconnected)

**[Errors and warnings](#errors-warnings)**

- [`Failed to send batch, retrying` message: What should I do?](#failed-to-send-batch-retry)
- [Recommender failure](#err-rec-fail)
- [Scaled to zero](#err-scale-zero)
- [Missing data](#err-missing-data)
- [Expired](#err-exp)
- [Applier failure](#err-app-fail)
- [Missing patch path for a Job](#err-missing-patch-path)
- [Job doesn't run long enough to make a recommendation](#unsup-jobs)
- [Corrupt data detected for container](#unsup-corrupt-data)
- [Oscillating replicas](#unsup-osc-rep)
- [Target utilization max](#unsup-targ-ut-max)

For status details, see [Workload recommendation statuses](reference__statuses.md).

## Common troubleshooting steps {#key-points}

Most troubleshooting involves the checks listed below. 

### Agent troubleshooting {#trbl-agent}

#### View high-level information such as Agent status and cluster connection status {#trbl-agent-stat}

  Check the Agent pod status:
  ```sh
  kubectl get pods -n orion-system -l app.kubernetes.io/component=cluster-agent
  ```

  View the agent pod logs:
  ```sh
  kubectl logs -n orion-system -l app.kubernetes.io/component=cluster-agent
  ```

  Use the optional CLI image to check the cluster connection status:
  ```sh
  kubectl run orion-check-connect --attach --rm --restart=Never --quiet \
    -n orion-system --image registry.orion.example/library/orion-cli \
    --override-type=strategic \
    --overrides='{"spec": {
                    "serviceAccount": "orion-cluster-agent",
                    "containers": [{
                      "name": "orion-check-connect",
                      "envFrom": [
                        {"secretRef": {"name": "orion-auth"}},
                        {"secretRef": {"name": "orion-env"}}]}]}}' \
     -- check connect
  ```
  
  Examine the output to determine:

  - Whether the agent pod is running
  - Whether the agent pod is reporting any errors
  - Whether the application can connect to the `api.orion.example` endpoint

#### Check whether the Agent is running {#trbl-agent-run}

  ```sh
  kubectl get pods -n orion-system -l app.kubernetes.io/component=cluster-agent
  ```

#### View the Agent logs {#trbl-agent-depl-logs}

These logs show output and errors related to the Agent Deployments. 
- If the Agent won't start (for example, connectivity issues when using a proxy), check the init container status: 
  ```shell
  kubectl logs -n orion-system -l app.kubernetes.io/component=cluster-agent -c init --tail=-1
  ```

- If the Agent is running:
  ```shell
  kubectl logs -n orion-system -l app.kubernetes.io/component=cluster-agent --tail=-1
  ```

#### Describe the Agent Pod and events {#trbl-agent-descr-pod}

```shell
kubectl describe pod -n orion-system -l app.kubernetes.io/component=cluster-agent
```

Examine the output:

- Check the pod status, state and reason, conditions, and events.
- In the `.containers` spec, check the arguments, state, liveness, and readiness.

#### View the ConfigMaps settings {#trbl-agent-edit-configmap}

```shell
kubectl get configmaps -n orion-system -l app.kubernetes.io/name=orion -o yaml
```

Scan the relevant ConfigMaps for any unusual or unexpected settings. To make changes, identify the correct Helm values needed to change the ConfigMap and set them by running `helm upgrade`. If you make changes by editing the ConfigMaps directly, those manual edits will be overwritten on upgrade. 

#### Restart the Agent {#trbl-agent-restart}

In some scenarios, restarting the Agent solves the problem. Remember that this will result in a few moments of downtime while new pods scale up: 

```shell
kubectl rollout restart -n orion-system deployment.apps/orion-cluster-agent
```
To check the status of the restart, run:
```sh
kubectl rollout status deployment -n orion-system orion-cluster-agent -w
```

A successful restart generates this message:\
```deployment "orion-cluster-agent" successfully rolled out```

### Applier troubleshooting {#trbl-applier}

#### View the Applier logs {#trbl-applier-depl-logs}

These logs contain output and errors related to the Applier Deployment. 
  ```shell
  kubectl logs -n orion-system -l component=applier --tail=-1
  ```

#### Describe the Applier Pod and events {#trbl-applier-descr-pod}

  ```shell
  kubectl describe pod -n orion-system -l component=applier
  ```
Examine the output:

- Check the pod status, state and reason, conditions, and events.
- In the `.containers` spec, check the arguments, state, liveness, and readiness.
 
### Node Agent troubleshooting {#trbl-node-agent}

The Node Agent is optional and disabled by default. These steps apply only if you enabled it to collect [GPU metrics][gpu].

#### Check whether the Node Agent is running {#trbl-node-agent-run}

  ```sh
  kubectl get pods -n orion-system -l app.kubernetes.io/component=node-agent -o wide
  ```

Expect one pod per GPU node. The `-o wide` output shows which node each pod is on.

#### View the Node Agent logs {#trbl-node-agent-logs}

  ```shell
  kubectl logs -n orion-system -l app.kubernetes.io/component=node-agent --tail=-1
  ```

#### Describe the Node Agent Pod and events {#trbl-node-agent-descr-pod}

  ```shell
  kubectl describe pod -n orion-system -l app.kubernetes.io/component=node-agent
  ```

Examine the output:

- Check the pod status, state and reason, conditions, and events.
- In the `.containers` spec, check the arguments, state, liveness, and readiness.

#### Restart the Node Agent {#trbl-node-agent-restart}

  ```shell
  kubectl rollout restart daemonset/orion-node-agent -n orion-system
  ```

#### No Node Agent pods on the GPU nodes {#trbl-node-agent-not-scheduled}

If the DaemonSet reports 0 desired pods, or pods are missing from nodes that have GPUs, the node selector or the tolerations don't match those nodes.

1. Check what the DaemonSet expects:

    ```shell
    kubectl get daemonset orion-node-agent -n orion-system \
      -o jsonpath='{.spec.template.spec.nodeSelector}{"\n"}{.spec.template.spec.tolerations}{"\n"}'
    ```

2. Check the labels and taints on a GPU node, replacing NODE_NAME with the node name:

    ```shell
    kubectl get node NODE_NAME -o jsonpath='{.metadata.labels}{"\n"}{.spec.taints}{"\n"}'
    ```

3. If they don't match, set `nodeAgent.nodeSelector` and `nodeAgent.tolerations` to match your GPU nodes. See [Node selection][node-selection].

#### Node Agent is running but GPU data is missing {#trbl-node-agent-no-gpu-data}

If Node Agent pods are running on your GPU nodes but GPU workloads show no GPU data:

1. Confirm that Orion identifies the workload as a GPU workload:

    ```shell
    orion view workloads --gpu
    ```

    If the workload isn't listed, see [How GPU workloads are identified][gpu-identified].

2. Check the Node Agent logs for GPU initialization warnings:

    ```shell
    kubectl logs -n orion-system -l app.kubernetes.io/component=node-agent --tail=-1 --prefix \
      | grep -E "gpu provider init failed|dcgm profiling unavailable"
    ```

    - `gpu provider init failed`: the Node Agent couldn't initialize the NVIDIA driver library on that node, so it isn't collecting GPU metrics. If the line includes `ERROR_LIBRARY_NOT_FOUND`, confirm that the node runs the NVIDIA container runtime and that your cluster policies admit the Node Agent's `SYS_ADMIN` capability. See [Permissions and host access][node-agent-permissions]. If the node has no GPU, adjust [Node selection][node-selection] so that the Node Agent doesn't run there.

    - `dcgm profiling unavailable`: engine activity metrics, such as tensor core activity, are missing. Memory, power, and utilization metrics are unaffected. See [How GPU metrics are collected][gpu-collection].

3. Check the Node Agent's collection health metrics. Forward its metrics port to your machine, replacing POD_NAME with a Node Agent pod name:

    ```shell
    kubectl port-forward -n orion-system pod/POD_NAME 9401:9401
    ```

    In a second terminal, query the metrics:

    ```shell
    curl -s localhost:9401/metrics \
      | grep -E "^orion_gpu_(scrape_errors|allocation_scrape_errors|allocated_unresolved_devices)"
    ```

    If you changed `nodeAgent.metricsPort`, use that port instead of `9401`. If the command prints nothing and the port-forward shows no errors, the Node Agent isn't collecting GPU metrics on that node. Check its logs as described in the previous step.

    | Metric | A value above 0 means | What to do |
    | ------ | --------------------- | ---------- |
    | `orion_gpu_scrape_errors` | The Node Agent's last read from the NVIDIA driver failed. | Check the Node Agent logs for `gpu scrape`. |
    | `orion_gpu_allocation_scrape_errors` | The Node Agent's last read of GPU allocations from the kubelet failed, so the GPU tab is missing data. | Check the Node Agent logs for `pod-resources scrape`. If the value stays above 0, confirm that the kubelet serves the PodResources API socket under `/var/lib/kubelet/pod-resources` on the node. |
    | `orion_gpu_allocated_unresolved_devices` | Some allocated GPUs couldn't be matched to a physical GPU, so they're missing from the affected workloads' GPU tab. The `reason` label is `dra_slice` for a GPU allocated through Dynamic Resource Allocation, or `mig_geometry` for a MIG device. | For `dra_slice`, a value can appear briefly after the Node Agent starts. If it stays above 0, confirm that the cluster runs Kubernetes 1.34 or later. For `mig_geometry`, [restart the Node Agent](#trbl-node-agent-restart). If either value stays above 0, [contact Orion support](https://www.orioncorp.io/company/contact-us/). |

#### Related topics

- [Monitoring the health of Orion][monitoring-health-optlive]

## Installation {#installation}

### Health check failed with a `Failed to create Orion client` message {#fail-create-client}


This message can occur when installing either the Agent or Applier, and it indicates a connectivity problem between your environment and the Orion environment. For example, a proxy or a firewall that isn't configured correctly can cause timeout, proxy, and TLS issues. 

Start by checking the connectivity to the Orion endpoints: 
```sh
kubectl run orion-check-connect --attach --rm --restart=Never --quiet \
  -n orion-system --image registry.orion.example/library/orion-cli \
  --override-type=strategic \
  --overrides='{"spec": {
                  "serviceAccount": "orion-cluster-agent",
                  "containers": [{
                    "name": "orion-check-connect",
                    "envFrom": [
                      {"secretRef": {"name": "orion-auth"}},
                      {"secretRef": {"name": "orion-env"}}]}]}}' \
    -- check connect
```

If any endpoint status in the output is not `OK`, try the following steps:

1. Run the following `busybox` and `wget` commands to check for a network issue such as a problematic network policy, firewall, or egress controller:

    ```sh
    kubectl -n orion-system run -i --tty busybox --image=busybox --restart=Never -- sh
    ```
    At the busybox command prompt, enter:
    ```sh
    wget https://api.orion.example
    ```
    In this example, although it is counterintuitive, the `404` error indicates connectivity to Orion servers:
    ```sh
    % kubectl -n orion-system run -i --tty busybox --image=busybox --restart=Never -- sh
    If you don't see a command prompt, try pressing enter.
    / # 
    / # wget https://api.orion.example
    Connecting to api.orion.example (XX.XXX.XXX.XXX:443)
    wget: note: TLS certificate validation not implemented
    wget: server returned error: HTTP/1.1 404 Not Found
    ```
    Type `exit` to exit busybox and then delete the pod:
    ```sh
    kubectl delete pod busybox -n orion-system
    ```
1. If the `busybox` and `wget` commands above do not return a `404` error, run the following command, which uses a `netshoot` image to determine which network hop is dropping the connection.\
  **Note:** You must be connected to the cluster with admin rights and the pod must run in privileged mode (`.spec.privileged: true`): 

    ```sh
    kubectl -n orion-system run tmp-shell --rm -i --tty --image nicolaka/netshoot
    ```
    At the `[tmp-shell]` prompt, enter: 
    ```sh
    tcptraceroute api.orion.example 443
    ```
    The following example illustrates end-to-end connectivity. If you are troubleshooting, you might see `* * *` indefinitely at a particular hop.  
    ```sh
    % kubectl -n orion-system run tmp-shell --rm -i --tty --image nicolaka/netshoot
    tmp-shell  ~  tcptraceroute api.orion.example 443
    Selected device dev0, address 192.168.0.3, port 33305 for outgoing packets
    Tracing the path to api.orion.example (198.51.100.1) on TCP port 443 (HTTPS), 30 hops max
    1  192.168.0.1  0.056 ms  0.052 ms  0.048 ms
    2  192.168.0.2  0.274 ms  0.215 ms  0.205 ms
    3  * * *
    4  ec2-198-51-100-1.compute-1.amazonaws.com (198.51.100.1)  0.787 ms  0.678 ms  0.542 ms
    5  ec2-198-51-100-1.compute-1.amazonaws.com (198.51.100.1)  1.520 ms  0.832 ms  0.817 ms
    6  ec2-198-51-100-1.compute-1.amazonaws.com (198.51.100.1) [open]  1.288 ms  2.021 ms  2.333 ms
    ```

### The installation failed. What can I check? {#install-failed}

#### Agent

If the Agent didn't install correctly, you might not be able to:
- Log in to [app.orion.example](https://app.orion.example/)
- View your workloads after logging in to [app.orion.example](https://app.orion.example/)

Here are a few things you can check:
- [Check the Agent status and cluster connection status](#trbl-agent-stat).
- [View the Agent logs](#trbl-agent-depl-logs).

#### Applier

The Applier uses the secret provided by the Agent — check for the secret's existence by running:

```shell
kubectl get deployment,secret -n orion-system
```

Check the output for the following item:

```
NAME                           TYPE     DATA   AGE
secret/orion-auth         Opaque   5      31d
```

If the secret is missing, reinstall Orion using `helm install orion oci://registry.orion.example/library/orion`.

## Workloads and recommendations {#workloads-recs}

### I don't see the workloads (or namespaces) that I expect on the dashboard. What can I check? {#check-namespaces}

Reasons why you might not see some workloads or namespaces include: 
- The namespace exists but it is empty - no pods, no built-in types.
- The namespace contains only batch jobs (which are not supported) or pods without an owner reference (this is rare).
- The namespace has no workloads in a healthy state: for example, a DaemonSet isn't fully operational (which occurs when not all are healthy).
- The namespace allowNamespaces or denyNamespaces Helm chart values constrain which namespaces to optimize. No namespaces are excluded by default, so this applies only if one of these lists has been set. Check with `helm get values`.

There are a few things you can check: pods on the cluster, [the Agent logs](#trbl-agent-depl-logs), and the allow/deny namespaces lists.

**Get a list of all running pods on the cluster**
```shell
kubectl get all -A
```

You might choose to modify this command or add other arguments, based on your familiarity with Kubernetes. 

In the output: 
 - In the NAME column, look at the pod list (`pod/POD_NAME` rows) and make sure the `orion-cluster-agent-POD_SUFFIX` pod and your application pods in other namespaces are running.
 - In the NAME column, look for `batch.apps`: If it exists, make sure that the corresponding namespace contains other Kubernetes types (deployment.apps, replicaset.apps, daemonset.apps).
 - For each pod whose STATUS column that has a status of RUNNING, check the READY column: Make sure all the nodes are ready. For example, 7/7 means the workload is healthy (all 7 pods are running); 3/7 is unhealthy.

**Check whether there are any workloads on the cluster**\
Run `kubectl get deployments` (include the `-n NAMESPACE` argument if you're not checking the `default` namespace). If there are no deployments listed in the output, Orion won't collect any metrics.

  - **If you need a sample workload**, you can download the Orion sample application, which installs one workload on your cluster. From a command line, run:
    ```shell
    helm install optlive-showcase-app oci://registry.orion.example/examples/optlive-showcase-app
    ```
    Learn more about the sample application in [this guide](/guides/showcase-app).

**Check the Agent logs**
1. Get the Agent logs as described [above](#trbl-agent-depl-logs). 
1. Scan the output for errors and any unusual or unexpected settings. 

**Check whether the allowList or denyList constraints are too restrictive**
  1. Get the Orion ConfigMap as described [above](#trbl-agent-edit-configmap).
  1. Find `cluster-agent.yml` in the output, and look for the allowNamespaces or denyNamespaces argument.
      - If either list is present and doesn't match what you expect, adjust it as needed. If both are present, the allowList has precedence and the denyList is ignored.
      - If neither list is included in this spec, then neither list has been set, and something else is causing the problem.

### How do I grant RBAC permissions on a patch target? {#additional-rbac}

To apply or export the recommended workload settings, Orion must:
- Support the patch target type.
- Have RBAC access permissions on the patch target. 

When these permissions are missing, you'll see a message similar to this:  
```
please grant rbac access permissions on name/namespace: %s/%s, kind: %s, apiversion: %s
```

To grant RBAC access permissions:  
1. Find the `values.yaml` file from the Helm chart that was used to install Orion.

    - If you don't have the file:
       1. Download the chart:
          ```sh
          helm show values oci://registry.orion.example/library/orion
          ```
          If you use a proxy server or private container registry, follow the steps in [Advanced installation](installation__install-adv.md#advanced-installation-scenarios).

       1. Copy the `clusterAgent.rbac` section into a new file named `rbac-update.yaml`. Proceed to the next step.

1. In the `clusterAgent.rbac.additionalRules` section, add a `.apiGroups` section that includes the resources listed in the message.

    - Remember to replace API_VERSION and KIND with the values from the `please grant rbac access permissions ...` message.

      ```yaml
      clusterAgent:
        rbac:
          additionalRules: |
            - apiGroups:
              - API_VERSION
              resources:
              - KIND
              verbs:
              - get
              - list
              - watch
      ```

1. Apply the changes by running one of the following commands: 
    - If you edited the `values.yaml` file, run: 

      ```shell
      helm upgrade --values.yaml 
      ```

    - If you created a separate file named `rbac-update.yaml`, run: 
      ```shell
      helm upgrade --values.yaml -f rbac-update.yaml 
      ```

### I didn't get a recommendation for a workload. What can I check? {#no-rec}

Check the basics: 
- Look for more information in a message in the Orion banner.
- Check that the recommendation schedule is set to the frequency that you expect: 
  1. Click **Workloads** in left navigation and drill down to the workload. 
  1. Click **Settings** > **Schedule** and check the **Recommendation Schedule**.
  
      To change the schedule, set the `live.orion.example/schedule` annotation as described in the [Schedule] topic.

Check for scale-down events: You won't get a recommendation if the workload has been scaled to zero replicas for the entire 7-day period.

### Recommendations don't seem right based on the CPU/memory utilization. What can I do? {#workload-profile-change}

Orion responds to changes in a workload's profile (CPU/memory utilization) and generates recommendations based on those changes and recent data.

If a planned change to a workload dramatically alters the profile, [Orion Support][support] can help you restart the workload's _learning period_ (initial observation period). Orion will collect fresh metrics and generate new preliminary recommendations, which you can apply on demand, for the duration of the learning period.

When the learning period is complete, Orion generates recommendations that can be applied automatically if auto-deploy is enabled, or you can apply them on demand.

Related information: 
- [Learning period][learning-period]
- [Deploying recommendations automatically][auto-deploy] 

### My cluster is disconnected. What can I check? {#cluster-disconnected}

If Orion doesn't get a response from a cluster after about one hour, it marks the cluster as disconnected (![](/img/docs/ol-cluster-disc.png)). The data shown is the last view that Orion has of the data.

If you believe the cluster should be connected, [check the Agent status and cluster connection status](reference__troubleshooting.md#trbl-agent-stat). [Restart the Agent](reference__troubleshooting.md#restart-the-agent) if needed.

## Errors and warnings {#errors-warnings}

### Failed to send batch, retrying {#failed-to-send-batch-retry}

`<span style="color:#C44F39;">Post &#34;https://in.orion.example/prometheus/write&#34;: oauth2: &#34;too_many_attempts&#34; We have detected suspicious login behavior and further attempts will be blocked. Please contact the administrator.</span>
`

If enough incorrect client_ID/client_secret combinations originate from one IP address, any further attempts to connect to Orion are blocked.

Troubleshooting steps:
1. Determine which authentication setup you're using in your estate. An administrator typically makes this decision, usually before installing Orion on an initial cluster.
    - One clientID/clientSecret pair used by all clusters
    - One clientID/clientSecret pair per cluster

1. On each cluster, locate your access credentials file (the `.yaml` file that contains your clientID and clientSecret). This file was created when you either:
    - Added a cluster by clicking **Add Cluster** in the Orion UI
    - Ran `orion auth create CREDENTIAL_NAME > CREDENTIAL_FILE` from the command line before you installed the Orion Agent

Errors and warnings are grouped by type when you hover over the status icon in a workload list.

### Recommender failure {#err-rec-fail}

`<span style="color:#C44F39;">Failed to generate recommendation (1000)</span>
`

This error can occur for different reasons. Typically, [contacting support][support] is the best action to take, but you can also try one of the following:
- Try to generate a new recommendation by: 
    - Clicking **Regenerate** on the workload details page and then applying the recommendation.  
    - Running [`orion create recommendations [flags]`](orion-cli__reference.md#orion-create-recommendations). Append the `--autodeploy` flag to apply it immediately, or run [`orion apply --recommendation RECOMMENDATION_NAME [flags]`](orion-cli__reference.md#orion-apply) to apply it manually.

        **Tip:** To get the 10-digit recommendation name, run [`orion get recommendations [flags]`](orion-cli__reference.md#orion-get-recommendations).\
        Example (replace placeholders with your cluster, namespace, and workload names):
        ```sh
        orion get recommendations --cluster CLUSTER_NAME --namespace NAMESPACE_NAME --name WORKLOAD_NAME –state Generated -o yaml
        ```
- Wait for the next scheduled recommendation.

### Scaled to zero {#err-scale-zero}

No recommendation is generated when a workload scales to zero for too long. 

`<span style="color:#C44F39;">This workload has been scaled to zero for the entire last 7 days. No non-zero replica data is available to generate a recommendation.</span>
`

If the scaling down is unexpected, investigate the workload using `kubectl` commands.  

### Missing data {#err-missing-data}

No recommendation was generated because data is missing or is stale. 
When this happens, you'll see one of the following errors:
- `<span style="color:#C44F39;">We don&#39;t have any recent data for this workload - the available data is more than 20 minutes old.</span>
`
- `<span style="color:#C44F39;">Data not found for cpu_usage.</span>
`
- `<span style="color:#C44F39;">Data not found for requests.</span>
`
- `<span style="color:#C44F39;">No containers metrics observed for this workload in the last 10 minutes.</span>
`

  Try these standard Orion [troubleshooting steps](reference__troubleshooting.md#key-points):
  - If the cluster is disconnected (![](/img/docs/ol-cluster-disc.png)), [check the Agent status and cluster connection status](reference__troubleshooting.md#trbl-agent-stat).
  - [Check whether the Agent is running](reference__troubleshooting.md#trbl-agent-run).
  - [View the Agent logs](reference__troubleshooting.md#trbl-agent-depl-logs).
  - [Describe the Pod and Events for the Agent](reference__troubleshooting.md#trbl-agent-logs).


### Expired {#err-exp}

`<span style="color:#C44F39;">The recommendation was not computed in a timely manner.</span>
`\
You can either wait for the next recommendation to be generated or generate a new one by:
    - Clicking **Regenerate** on the workload details page and then applying the recommendation.  
    - Running [`orion create recommendations [flags]`](orion-cli__reference.md#orion-create-recommendations). Append the `--autodeploy` flag to apply it immediately, or run [`orion apply --recommendation RECOMMENDATION_NAME [flags]`](orion-cli__reference.md#orion-apply) to apply it manually.

### Applier failure {#err-app-fail}

If a workload doesn't restart or function as expected after a recommendation is applied, the Orion Applier rolls back the workload to its last known good state and logs one of the following messages in the Applier logs.
- `<span style="color:#C44F39;">applied patch but timed out waiting for workload to become ready</span>
`\
`<span style="color:#C44F39;">rolled back to previous state not waiting for workload to become ready</span>
`\
`<span style="color:#C44F39;">workload was unhealthy prior to patch</span>
`

  The workload wasn't healthy even before applying any changes, and you should investigate the workload directly.

  Start by describing the pods relating to the workload resource and examine the Events section for clues.
    ```sh
    kubectl describe pod NAME
    ```
    You can learn more about [troubleshooting common issues][troubleshooting] with Kubernetes resources in the Kubernetes docs.

- `<span style="color:#C44F39;">applied patch but timed out waiting for workload to become ready</span>
`\
`<span style="color:#C44F39;">rolled back to previous state workload returned to ready</span>
`

  _or_

  `<span style="color:#C44F39;">applied patch but timed out waiting for workload to become ready rolled back to previous state timed out waiting for pods to be ready</span>
`

  While rollback is sometimes a desirable safety feature to have enabled, at scale and in large environments these errors often occur when the Applier's monitoring period of 5min isn't long enough for workloads to become Ready. The Applier determines that the change is a failure and rolls the workload back to the previous state. 

  If this error happens often, disable the rollback feature.

  #### Disable rollback {#disable-rollback}
  
    1. Add the following to your `values.yaml`:
          ```yaml
          additionalEnvVars:
          - name: ORION_SKIP_ROLLBACK
            value: "true"
          ```
    1. Upgrade Orion:
          ```sh
          helm upgrade orion oci://registry.orion.example/library/orion \
            -n orion-system \
            --reset-then-reuse-values \
            -f values.yaml
          ```

### Missing patch path for a Job {#err-missing-patch-path}

Orion marks a recommendation for a standalone Job as not patchable with the reason `MissingPatchPath` and a message such as:

- `The workload is missing podtemplatemetadata patchPath configuration for its Job patchTarget.`
- `The workload is missing CPU requests patchPath configuration for its Job patchTarget.`

Kubernetes doesn't allow changes to a Job's Pod template after creation, so Orion can't patch the Job. Use a webhook-based apply method to set recommended values on Pods at admission:

1. Enable the [Applier webhook][webhook].
1. Add the following annotation to the Job:
    ```yaml
    live.orion.example/apply.method: "WebhookHybridRollout"
    ```

If the Job uses `WebhookImmediateRollout`, setting `live.orion.example/pod-template.metadata.patch-path: "-"` on the Job has the same effect. See [Pod template metadata][pod-template-metadata].

### Unsupported workloads {#unsup}

Workloads whose status is **Unsupported** (![](/img/docs/ol-status-wl-unsup-sm.png)) are configured in such a way that Orion does not generate recommendations for them. The following messages shown for unsupported workloads are informational only, and typically, no action is required.

#### Job doesn't run long enough to make a recommendation {#unsup-jobs}

A CronJob (or any short-duration Job) must typically run for at least 10 continuous minutes in order for Orion to generate a recommendation. In some scenarios, 5-10 minutes is sufficient to generate a recommendation, depending on the frequency of the Job.

If the Job does not provide enough metrics within 5-10 minutes, Orion labels the workload as Unsupported.

Orion does not generate recommendations for Jobs that run for less than 5 minutes.

#### Corrupt data detected for container {#unsup-corrupt-data}

No recommendation was generated because some of the metrics data Orion has collected exhibit signs of corruption that make the metrics data unusable, such as metrics showing impossibly high CPU or memory usage. If you see workloads with this Unsupported reason, please contact Support.

#### Oscillating replicas {#unsup-osc-rep}

No recommendation was generated because the HPA oscillates between min and max replicas. If this is unexpected, investigate the workload. Increasing the max replicas can sometimes eliminate the oscillation, enabling Orion to generate recommendations.

#### Target utilization max {#unsup-targ-ut-max}

No recommendation was generated because the HPA `target.averageUtilization` is set to greater than or equal to 100.

[support]: https://www.orioncorp.io/company/contact-us/
[troubleshooting]: https://kubernetes.io/docs/tasks/debug/debug-application/
[Schedule]:  /docs/configure/settings/schedule/
[learning-period]: /docs/configure/settings/apply-settings/#learning-period
[auto-deploy]: /docs/configure/settings/apply-settings/#auto-deploy
[monitoring-health-optlive]: /docs/guides/monitoring-health-optlive/
[gpu]: /docs/recommendations/gpu/
[node-selection]: /docs/installation/node-agent-config/#node-selection
[node-agent-permissions]: /docs/installation/node-agent-config/#permissions
[gpu-identified]: /docs/recommendations/gpu/#how-gpu-workloads-are-identified
[gpu-collection]: /docs/recommendations/gpu/#how-gpu-metrics-are-collected
[webhook]: /docs/installation/webhook-config/#webhook-enabled
[pod-template-metadata]: /docs/configure/settings/apply-settings/#pod-template-metadata
