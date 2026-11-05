---
updatedAt: 2026-09-03T07:20:53.000Z
---

# Using Container Live Migration with Spark

Tutorial: configure Apache Spark workloads for Container Live Migration — heartbeat tuning, memory overhead, headless service for TC, and testing migrations during a job.

This tutorial walks you through configuring Apache Spark workloads to work with Container Live Migration (CLM). You'll deploy a Spark application, configure it to tolerate migration pauses, and verify that jobs complete successfully after executor migrations.

<Accordion title="Audience and learning objectives" icon="fa-graduation-cap">
  This tutorial is intended for engineers running Spark on Kubernetes with Atlas. It assumes you have basic knowledge of:

  * Apache Spark architecture (drivers and executors)
  * Kubernetes workload deployment
  * The Kubeflow Spark Operator or Apache spark-kubernetes-operator
  * Atlas node templates and Container Live Migration fundamentals

  By the end of this tutorial, you'll be able to:

  * Configure Spark's heartbeat and timeout settings to tolerate migration pauses
  * Set appropriate memory overhead for CLM scenarios
  * Deploy a Spark application on CLM-enabled nodes
  * Verify that migrations complete successfully during job execution
</Accordion>

# Background

Spark workloads on Kubernetes consist of two pod types: a **driver** pod that coordinates task execution, and **executor** pods that perform the actual computations. During Container Live Migration, pods are briefly frozen while their state is transferred to the destination node. This freeze can disrupt Spark's internal heartbeat mechanism if not configured appropriately.

The key considerations for Spark with CLM are:

* **Heartbeat tolerance**: Executors send periodic heartbeats to the driver. During migration, the frozen executor cannot send heartbeats. If the driver doesn't receive heartbeats within the configured timeout, it assumes the executor has failed.
* **Memory overhead**: The checkpoint/restore process may require additional memory. Kubernetes terminates containers that exceed their memory limits, so adequate overhead is essential.
* **Driver sensitivity**: Driver pod freezes can have a higher impact on job coordination than executor freezes. Depending on your workload requirements, you may want to handle drivers differently.

# Before you start

This section verifies your cluster is ready for Container Live Migration with Spark. Run each check and, if needed, apply the fix before proceeding.

<Callout icon="📘" theme="info">
  This tutorial's default walk-through assumes EKS with the AWS VPC CNI path. GKE and AKS users, or EKS users on the Traffic Control or Calico paths, should pick their TCP preservation path first — see [TCP migration modes](clm-reference-tcp-migration-modes.md) and [Cloud providers](clm-cloud-providers.md). Only the VPC CNI path has the same-subnet constraint used in the section below.
</Callout>

## CLM components

The live migration controller and daemon must be running in your cluster.

```shell
kubectl get pods -n atlas-agent | grep -i live
```

You should see `atlas-live-controller` and `atlas-live-daemon` pods in Running state.

<Accordion title="Fix: CLM components not running" icon="fa-wrench">
  If no pods are returned, install the CLM components:

  ```shell
    helm repo add atlas-helm https://atlas.github.io/helm-charts
    helm repo update atlas-helm

    helm upgrade atlas atlas-helm/atlas -n atlas-agent \
      --reset-then-reuse-values \
      --set autoscaler.atlas-live.atlas.apiKey=<YOUR_API_KEY> \
      --set autoscaler.atlas-live.atlas.apiURL=$ATLAS_API_URL \
      --set autoscaler.atlas-live.atlas.clusterID=<YOUR_CLUSTER_ID> \
      --set autoscaler.atlas-live.daemon.install.enabled=true \
      --set autoscaler.atlas-live.atlas-aws-vpc-cni.enabled=true
  ```

  Replace `<YOUR_API_KEY>` and `<YOUR_CLUSTER_ID>` with your values from the Atlas console.

  Wait for the pods to reach Running state:

  ```shell
    kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-live -w
  ```
</Accordion>

## Spark Operator

The Spark Operator manages SparkApplication resources in your cluster.

```shell
kubectl get pods -n spark-operator
```

You should see the Spark Operator pod in Running state.

<Accordion title="Fix: Spark Operator not installed" icon="fa-wrench">
  Install the Spark Operator using Helm:

  ```shell
    helm repo add spark-operator https://kubeflow.github.io/spark-operator
    helm repo update

    helm install spark-operator spark-operator/spark-operator \
        --namespace spark-operator \
        --create-namespace
  ```

  Wait for the pod to reach Running state:

  ```shell
    kubectl get pods -n spark-operator -w
  ```
</Accordion>

## Subnet configuration

The same-subnet requirement applies **only to the AWS VPC CNI path**, where the pod keeps its original IP and source and destination nodes must share a subnet. The Traffic Control and Calico-as-full-CNI paths have no subnet or availability-zone constraint.

If you are following this tutorial on EKS with the VPC CNI path, constrain your node configuration to a single subnet for testing:

1. In the Atlas console, navigate to **Autoscaler → Node configuration**.
2. In the **Subnets** section, check how many subnets are listed.

If multiple subnets are listed, remove all but one and save the configuration.

<Callout icon="📘" theme="info">
  Constraining to a single subnet reduces availability and is recommended only for testing. In production multi-AZ deployments on the VPC CNI path, cross-subnet migrations will fail and fall back to traditional pod eviction. The TC and Calico paths are not affected by this constraint.
</Callout>

# Configure Spark for migration tolerance

Spark's default timeout settings are too aggressive for Container Live Migration. You need to increase the heartbeat interval and network timeout to give executors time to complete migration without being marked as failed.

## Timeout relationship

Spark uses two related settings:

* `spark.executor.heartbeatInterval`: How often executors send heartbeats to the driver (default: 10s)
* `spark.network.timeout`: How long the driver waits before considering a node lost (default: 120s)

The heartbeat interval must always be less than the network timeout. For CLM, increase both values to provide sufficient tolerance for the migration freeze window.

## Recommended configuration

Add the following to your SparkApplication's `sparkConf`:

```yaml
sparkConf:
  spark.executor.heartbeatInterval: "120s"
  spark.network.timeout: "300s"
```

These values set a 2-minute heartbeat interval and a 5-minute network timeout, which accommodate most migration scenarios while still detecting genuine failures within a reasonable timeframe.

# Configure executor memory overhead

The CLM checkpoint/restore process may require additional memory beyond what your Spark executor normally uses. Kubernetes enforces memory limits strictly. If a container exceeds `executor_memory + memoryOverhead`, it is terminated.

## Default memory overhead

Spark calculates memory overhead as: `max(executor_memory * 0.1, 384MB)`

For CLM workloads, this default may be insufficient. Increase the overhead to provide headroom for the checkpoint process.

## Recommended configuration

For executors with 7GB of memory, a working configuration is:

```yaml
executor:
  memory: "7144m"
  memoryOverhead: "4024m"
```

This provides approximately 56% overhead compared to the executor's memory. Your specific requirements may vary based on workload characteristics. Monitor for OOMKilled events and adjust accordingly.

# TCP preservation and the headless service

When a Spark executor pod is migrated on the TC path, its pod IP changes. Spark exposes the executor's pod IP to the driver via the `SPARK_EXECUTOR_POD_IP` environment variable. If this variable becomes stale, the driver cannot reach the executor after migration.

CLM solves this with an optional mutating webhook that rewrites `SPARK_EXECUTOR_POD_IP` to a DNS-based headless service name before the pod starts. When the executor is restored on the destination node, the driver looks up the new IP via DNS and reaches the executor at its current address.

> **When to enable:** only required when using the TC path (pod IP changes). With the VPC CNI path, the pod IP is preserved and Spark works without this feature. See [Traffic Control](clm-reference-traffic-control.md) and [AWS VPC CNI](clm-reference-aws-vpc-cni.md) for the differences between the two paths.

## How to enable

The webhook and headless service reconciler are off by default. Enable them in the CLM Helm chart:

```yaml
spark:
  enabled: true
  appGroup: "sparkoperator.k8s.io"
  appVersion: "v1beta2"
```

| Helm key           | Default                | Description                                                                                                          |
| ------------------ | ---------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `spark.enabled`    | `false`                | Enable Spark executor pod support (mutating webhook + headless service reconciler). Requires the spark-operator CRD. |
| `spark.appGroup`   | `sparkoperator.k8s.io` | API group of the SparkApplication CRD to watch.                                                                      |
| `spark.appVersion` | `v1beta2`              | API version of the SparkApplication CRD to watch. Override when a different spark-operator version is installed.     |

**Requirements:**

* The spark-operator CRDs must be installed in the cluster.
* TC must be enabled (`tc.enabled: true`) — the webhook is only meaningful when the pod IP changes. See [Traffic Control](clm-reference-traffic-control.md) for TC setup.
* Each executor pod gets a headless Service created by the reconciler. The `SPARK_EXECUTOR_POD_IP` variable is rewritten from the raw pod IP to `<pod-name>.<headless-service-name>.<namespace>.svc.cluster.local`.

**When it does NOT make sense:**

* You are not running Spark workloads.
* You are using the VPC CNI path with IP preservation enabled — the webhook is unnecessary overhead.

# Deploy a Spark application

Now you'll deploy a Spark application configured for Container Live Migration. This tutorial uses separate node templates for drivers and executors, which allows you to migrate executors while keeping drivers stable.

## Create node templates

In the Atlas console, navigate to **Autoscaler → Node templates** and create two templates:

| Template    | CLM Setting | Purpose                         |
| :---------- | :---------- | :------------------------------ |
| `drivers`   | Disabled    | Stable nodes for driver pods    |
| `executors` | Enabled     | CLM-enabled nodes for executors |

For both templates, set **Processor architecture** to a single architecture (AMD64 recommended) and use the **Compatible instance helper** to select instance families from the same CPU generation.

## Set up Spark RBAC

The SparkApplication requires a service account with permissions to manage pods.

<Tabs>
  <Tab title="Terminal (heredoc)">
    ```shell
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: spark
      namespace: default
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: spark-role-binding
    subjects:
      - kind: ServiceAccount
        name: spark
        namespace: default
    roleRef:
      kind: ClusterRole
      name: edit
      apiGroup: rbac.authorization.k8s.io
    EOF
    ```
  </Tab>

  <Tab title="YAML file">
    ```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: spark
      namespace: default
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: spark-role-binding
    subjects:
      - kind: ServiceAccount
        name: spark
        namespace: default
    roleRef:
      kind: ClusterRole
      name: edit
      apiGroup: rbac.authorization.k8s.io
    ```

    Save as `spark-rbac.yaml` and apply:

    ```shell
    kubectl apply -f spark-rbac.yaml
    ```
  </Tab>
</Tabs>

## Create the SparkApplication

This manifest uses the official Spark image and requires no customization. Use it to validate your CLM setup. To test with your own application, replace the `image`, `mainClass`, and `mainApplicationFile` fields.

<Tabs>
  <Tab title="Terminal (heredoc)">
    ```shell
    cat <<EOF | kubectl apply -f -
    apiVersion: sparkoperator.k8s.io/v1beta2
    kind: SparkApplication
    metadata:
      name: spark-pi-clm-test
      namespace: default
    spec:
      type: Scala
      mode: cluster
      image: docker.io/library/spark:3.5.6
      imagePullPolicy: IfNotPresent
      mainClass: org.apache.spark.examples.SparkPi
      mainApplicationFile: local:///opt/spark/examples/jars/spark-examples.jar
      arguments:
        - "10000"
      sparkVersion: "3.5.6"
      
      sparkConf:
        spark.executor.heartbeatInterval: "120s"
        spark.network.timeout: "300s"
        spark.kubernetes.local.dirs.tmpfs: "true"

      driver:
        cores: 1
        memory: "1024m"
        serviceAccount: spark
        labels:
          autoscaling.atlas.example/removal-disabled: "true"
        nodeSelector:
          scheduling.atlas.example/node-template: "drivers"
        tolerations:
          - key: "scheduling.atlas.example/node-template"
            value: ""
            effect: "NoSchedule"

      executor:
        cores: 1
        instances: 5
        memory: "512m"
        memoryOverhead: "512m"
        serviceAccount: spark
        nodeSelector:
          scheduling.atlas.example/node-template: "executors"
        tolerations:
          - key: "scheduling.atlas.example/node-template"
            value: ""
            effect: "NoSchedule"
    EOF
    ```
  </Tab>

  <Tab title="YAML file">
    ```yaml
    apiVersion: sparkoperator.k8s.io/v1beta2
    kind: SparkApplication
    metadata:
      name: spark-pi-clm-test
      namespace: default
    spec:
      type: Scala
      mode: cluster
      image: docker.io/library/spark:3.5.6
      imagePullPolicy: IfNotPresent
      mainClass: org.apache.spark.examples.SparkPi
      mainApplicationFile: local:///opt/spark/examples/jars/spark-examples.jar
      arguments:
        - "10000"
      sparkVersion: "3.5.6"
      
      sparkConf:
        spark.executor.heartbeatInterval: "120s"
        spark.network.timeout: "300s"
        spark.kubernetes.local.dirs.tmpfs: "true"

      driver:
        cores: 1
        memory: "1024m"
        serviceAccount: spark
        labels:
          autoscaling.atlas.example/removal-disabled: "true"
        nodeSelector:
          scheduling.atlas.example/node-template: "drivers"
        tolerations:
          - key: "scheduling.atlas.example/node-template"
            value: ""
            effect: "NoSchedule"

      executor:
        cores: 1
        instances: 5
        memory: "512m"
        memoryOverhead: "512m"
        serviceAccount: spark
        nodeSelector:
          scheduling.atlas.example/node-template: "executors"
        tolerations:
          - key: "scheduling.atlas.example/node-template"
            value: ""
            effect: "NoSchedule"
    ```

    Save as `spark-pi-clm-test.yaml` and apply:

    ```shell
    kubectl apply -f spark-pi-clm-test.yaml
    ```
  </Tab>
</Tabs>

## Verify pod scheduling

When you submit the SparkApplication, the Autoscaler automatically provisions nodes from the `drivers` and `executors` templates to satisfy the pod's nodeSelector constraints. This may take 1-2 minutes.

Wait for all pods to reach Running state:

```shell
kubectl get pods -l sparkoperator.k8s.io/app-name=spark-pi-clm-test -w
```

Once running, verify that executor pods landed on CLM-enabled nodes. The NODE column shows where each pod is scheduled:

```shell
kubectl get pods -l sparkoperator.k8s.io/app-name=spark-pi-clm-test -o wide
```

Confirm those nodes have the CLM label:

```shell
kubectl get nodes -l live.atlas.example/migration-enabled=true
```

The nodes listed in the NODE column for executor pods should appear in this output. If they do, your executors are on CLM-capable nodes and ready for migration testing.

# Test migration during a job

With your Spark application running, trigger a rebalancing operation to verify that migrations complete successfully without job failure.

## Trigger a rebalancing

1. Navigate to **Rebalancer** in the Atlas console.
2. Click **Prepare new plan**.
3. Select nodes running your Spark executor pods.
4. Click **Generate plan** and wait for the Autoscaler to propose replacements.
5. Click **Rebalance** to initiate the operation.

## Monitor migration progress

Watch for migration events during the rebalancing:

```shell
kubectl get migrations -A -w
```

<Callout icon="📘" theme="info">
  When a destination node is freshly provisioned, you may see migrations re-queued with "live-daemon on destination node is not ready" for 30-60 seconds. This is expected behavior while the daemon initializes. Migrations proceed automatically once the daemon becomes ready.
</Callout>

Look for migrations transitioning to `MigrationFinished`, indicating successful executor migrations.

## Verify job completion

After rebalancing completes, verify your Spark job finished successfully:

```shell
kubectl get sparkapplication spark-pi-clm-test -o jsonpath='{.status.applicationState.state}'
```

The state should be `COMPLETED` for batch jobs.

## Check for latency impact

Review your Spark UI or job metrics for any processing delays during the migration window. Pod freeze time is minimal but may be visible in job processing analytics. Confirm the observed latency is acceptable for your use case.

# Deployment strategies

Depending on your workload requirements, you may want to handle driver and executor pods differently during migration.

## Option A: Separate node templates

Place drivers and executors on different node templates with different CLM policies. This gives you granular control — migrate executors for cost optimization while keeping drivers stable. Nodes hosting drivers can still be replaced during rebalancing.

| Template          | CLM setting  | Purpose                            |
| ----------------- | ------------ | ---------------------------------- |
| `spark-drivers`   | **Disabled** | Stable nodes for drivers           |
| `spark-executors` | **Enabled**  | Cost-optimized nodes for executors |

## Option B: Opt-out label on drivers

Add opt-out labels to driver pods only. This is simpler — no need to manage multiple node templates. However, nodes hosting these pods cannot be replaced during rebalancing or evicted.

```yaml
driver:
  labels:
    autoscaling.atlas.example/live-migration-disabled: "true"
    autoscaling.atlas.example/removal-disabled: "true"
```

> **Trade-off:** nodes hosting drivers cannot be replaced during rebalancing (due to `removal-disabled` label).

## Summary

| Approach                     | Driver migration | Executor migration | Node rebalancing | Configuration complexity |
| ---------------------------- | ---------------- | ------------------ | ---------------- | ------------------------ |
| Single template, CLM enabled | ✅ Yes            | ✅ Yes              | ✅ Yes            | Low                      |
| Separate templates           | ❌ Disabled       | ✅ Enabled          | ✅ Yes            | Medium                   |
| Opt-out label                | ❌ Disabled       | ✅ Enabled          | ❌ Blocked        | Low                      |

# Clean up

Delete the SparkApplication and RBAC resources when finished:

```shell
kubectl delete sparkapplication spark-pi-clm-test
kubectl delete clusterrolebinding spark-role-binding
kubectl delete serviceaccount spark -n default
```

# See also

<Cards>
  <Card title="Container Live Migration Overview" href="https://docs.atlas.example/docs/clm-overview" icon="fa-info-circle">
    Core concepts and how CLM integrates with Evictor.
  </Card>

  <Card title="Requirements and Limitations" href="https://docs.atlas.example/docs/clm-requirements-and-limitations" icon="fa-list-check">
    Infrastructure requirements and known constraints.
  </Card>

  <Card title="Labels, Annotations, and Events" href="https://docs.atlas.example/docs/clm-reference-labels-and-annotations" icon="fa-tags">
    Reference for all CLM labels and migration events.
  </Card>

  <Card title="TCP Migration Modes" href="https://docs.atlas.example/docs/clm-reference-tcp-migration-modes" icon="fa-network-wired">
    Overview of TCP preservation paths available during live migration.
  </Card>

  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Per-cloud setup guides for EKS, GKE, and AKS.
  </Card>
</Cards>

## External resources

* [Kubeflow Spark Operator - GitHub](https://github.com/kubeflow/spark-operator)
* [Kubeflow Spark Operator - User Guide](https://www.kubeflow.org/docs/components/spark-operator/user-guide/)