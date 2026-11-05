---
updatedAt: 2026-10-23T08:36:00.000Z
---

# JVM workload optimization

Java workloads manage memory through the JVM heap, which is invisible to standard Kubernetes container memory metrics. The JVM preallocates heap memory at startup, and Kubernetes reports this preallocated memory as "used" even when only a fraction is consumed internally. This creates two problems: over-provisioning when relying on container-level metrics, and an inability to detect internal JVM heap OOM errors (`java.lang.OutOfMemoryError`) that are invisible to Kubernetes.

The Workload Autoscaler addresses this with two complementary capabilities:

* **JVM memory optimization** — Uses actual JVM runtime data (heap usage, GC behavior, thread count) to generate memory recommendations instead of relying on container-level working set memory.
* **Auto-instrumentation** — Automatically injects a JMX metrics agent into Java containers at pod admission time, so no application changes are required to collect JVM metrics.

<Image src="https://files.readme.io/2384f435c7b32675d377e481c088e22f8fb0b83d34824d83f901549c8a092f78-image.png" align="center" />

## How it works

### Choosing a collection method

The Workload Autoscaler collects JVM metrics three ways. Choose the one that fits your workloads before reading further — the prerequisites and enablement steps below apply to whichever method you pick.

| Method                   | When to use                                                                                        | Pod changes                                       | Requires                                                          |
| ------------------------ | -------------------------------------------------------------------------------------------------- | ------------------------------------------------- | ----------------------------------------------------------------- |
| **Auto-instrumentation** | Default. Your application exposes no metrics of its own.                                           | Injects an init container and a `-javaagent` flag | Java 9+, inline Java options                                      |
| **Endpoint scraping**    | Your application already exposes Prometheus-format JVM metrics (for example, Spring Boot Actuator) | None — a label and annotations only               | `prometheus.io/*` pod annotations                                 |
| **Your own Prometheus**  | You already run a JMX or JVM metrics pipeline                                                      | None                                              | A connected [custom metrics data source](woop-custom-metrics.md) |

### JVM workload detection

The Workload Autoscaler automatically identifies JVM containers using multiple signals from the cluster snapshot:

* **Container image** — Images matching known JVM runtimes and frameworks (OpenJDK, Kafka, and others)
* **Environment variables** — Presence of `JAVA_OPTS`, `JAVA_TOOL_OPTIONS`, `JDK_JAVA_OPTIONS`, `JVM_OPTS`, `_JAVA_OPTIONS`, or `CATALINA_OPTS`
* **Pod annotations** — OpenTelemetry Java instrumentation (`instrumentation.opentelemetry.io/inject-java`) and Datadog Java agent annotations
* **Explicit Atlas annotation** — Adding `workloads.atlas.example/runtime: jvm` to the pod template's annotations marks the container as a JVM workload. The value is case-insensitive, other values have no effect, and the annotation applies to single-container pods only. See the [annotations reference](workload-autoscaler-annotations-reference.md) for the full entry.
* **Custom metrics** — JVM heap metrics already flowing into the Workload Autoscaler from an existing Prometheus setup

Detection is automatic — no labels or annotations are required on your part. Detected JVM workloads are tagged with a **JVM** badge in the Atlas console. To find them, navigate to **Workload Autoscaler** > **Optimization** and use the **JVM** filter chip in the quick search bar.

![](https://files.readme.io/41ebb0feb97cf03540b6b30cb236e43c7ab8f4c17243aba794e1c616eeb78a2a-image.png)

### JVM memory optimization

Standard container-level autoscaling cannot distinguish between heap memory that the JVM has reserved and heap memory that the application is actively using. This leads to two failure modes:

* **Container OOM kill** — When the total process RSS exceeds the Kubernetes memory limit. Kubernetes emits an `OOMKilled` event, and the Workload Autoscaler handles this automatically by increasing memory.
* **Internal JVM heap OOM** — When the JVM exceeds its own heap limit (`-Xmx`, `-XX:MaxRAMPercentage`). The JVM throws `java.lang.OutOfMemoryError` internally, but Kubernetes sees nothing wrong unless a liveness probe fails. The container may keep running in a degraded state.

When JVM memory optimization is enabled in a scaling policy, the Workload Autoscaler uses actual JVM runtime data — heap usage, non-heap usage, garbage collection behavior, and thread count — to compute memory request recommendations. This lets it detect and prevent internal JVM heap OOM scenarios that are invisible to standard Kubernetes monitoring.

Recommendations include `-Xms` and `-Xmx` flags injected into the container's Java options environment variable. The flags are appended to any existing value, preserving your existing JVM configuration. The Workload Autoscaler automatically identifies the right environment variable to use based on what is already present in the container, falling back to `JDK_JAVA_OPTIONS` if none is detected. To pin the flags to a specific variable, set `optionsEnvVar` in the [configuration annotation](workload-autoscaler-annotations-reference.md#verticaljvmoptionsenvvar).

If JVM metrics are not yet available or there is insufficient history, the Workload Autoscaler falls back to the standard memory recommendation flow — the workload continues to receive recommendations, and no disruption occurs.

> 📘 Note
>
> After enabling JVM heap scaling, the Workload Autoscaler needs time to collect enough data points before generating heap-based recommendations. During this warm-up period, the workload detail page may show a message like "JVM recommendation skipped: insufficient JVM metrics." This is normal — the Workload Autoscaler uses standard recommendations in the meantime.

When JVM metrics are available, the workload detail page in the Atlas console displays additional charts showing heap size and heap recommendation metrics alongside standard memory metrics.

### Auto-instrumentation

Auto-instrumentation enables the Workload Autoscaler to collect JVM heap metrics from any Java container without code changes or involvement from the application team. When enabled, the workload-autoscaler webhook injects the following at pod admission time:

1. An **init container** that copies the JMX Prometheus Java agent JAR into a shared `emptyDir` volume
2. A `-javaagent`**&#x20;flag** appended to the container's Java options environment variable, pointing the agent at the shared volume
3. A **Prometheus scrape port** (port 9404 by default) declared on the container for metrics collection
4. A **pod label** (`workloads.atlas.example/jmx-exporter-injected=true`) indicating the agent was successfully injected

Nothing in the application image changes.

**Injection behavior:**

* **New pods and restarts** — Injection occurs at admission time. Rolling updates and pod restarts are safe — the webhook skips containers that are already instrumented.
* **Running pods** — Not affected until they restart. The Workload Autoscaler does not restart pods when auto-instrumentation is toggled. You must restart affected pods manually for the injection to take effect.
* **Multiple JVM containers** — If a pod contains multiple JVM containers, each gets its own port and Prometheus scrape target.
* **Port conflicts** — If port 9404 is already in use, the Workload Autoscaler automatically picks the next available port (9405, 9406, and so on).
* **Prometheus** — If Prometheus is not already present in the cluster, the Workload Autoscaler installs it automatically when auto-instrumentation is enabled.

### Scrape existing JVM metrics endpoints

If your Java workloads already expose JVM metrics in Prometheus format — Spring Boot applications with [Micrometer](https://micrometer.io) expose them at the [Actuator `/actuator/prometheus` endpoint](https://docs.spring.io/spring-boot/reference/actuator/metrics.html), for example — the Workload Autoscaler can scrape these endpoints directly with the Prometheus instance bundled in `atlas-workload-autoscaler-exporter`:

* No agent is injected and nothing changes in the pod
* No external data source needs to be connected — metrics flow through the automatically created [Default Exporter](woop-custom-metrics.md#default-exporter)
* JVM heap scaling works the same as with the other collection methods

This method is available in Workload Autoscaler 1.11.0 and later and is enabled by default on new installs. On clusters upgraded from an earlier chart version with `--reuse-values` or `--reset-then-reuse-values`, the `scrapeJVMMetrics` key is absent from your stored values and the scrape job stays disabled — set it explicitly as described in [Enable or disable endpoint scraping](#enable-or-disable-endpoint-scraping).

To opt a workload in, add a label to the pod template that tells the Workload Autoscaler to start scraping, and describe the metrics endpoint with the standard Prometheus pod annotations — the same annotations used by the [Prometheus community Helm chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus#scraping-pod-metrics-via-annotations):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-spring-boot-app
spec:
  selector:
    matchLabels:
      app: my-spring-boot-app
  template:
    metadata:
      labels:
        app: my-spring-boot-app
        # Tells the Workload Autoscaler to start scraping this pod
        workloads.atlas.example/jvm-metrics-scrape: "true"
      annotations:
        # Standard Prometheus annotations describing the metrics endpoint
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/actuator/prometheus"
    spec:
      containers:
        - name: app
          image: my-spring-boot-app:1.0
```

| Annotation             | Required | Description                                                                                                                                   |
| ---------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `prometheus.io/scrape` | Yes      | Must be set to `"true"` for the pod to be scraped                                                                                             |
| `prometheus.io/port`   | Yes      | Port the metrics endpoint listens on, for example `8080`. Without it the scraped target address is malformed and Prometheus drops the target. |
| `prometheus.io/path`   | No       | Path of the metrics endpoint. Defaults to `/metrics`. For Spring Boot Actuator, use `/actuator/prometheus`.                                   |
| `prometheus.io/scheme` | No       | `http` or `https`. Defaults to `http`.                                                                                                        |

**How scraping behaves:**

* **Only JVM metrics are collected** — metrics with names starting in `jvm_`, `java_`, or `jmx_scrape_` (Micrometer's JVM metrics, for example); any other metrics exposed by the endpoint are dropped.
* **Auto-instrumented pods are excluded** — pods carrying the `workloads.atlas.example/jmx-exporter-injected=true` label are skipped, so a workload is never scraped twice.
* **Scrape interval** — each endpoint is scraped every 15 seconds.

> 📘 Note
>
> The label and annotations only provide metrics. Heap-based recommendations still require **JVM heap scaling** to be enabled in the scaling policy assigned to the workload — auto-instrumentation can stay disabled. See [Enable JVM optimization](#enable-jvm-optimization).

Because no agent is injected, this method also fits workloads where injection is not an option — for example, Java 8 applications (see [Known limitations](#known-limitations)) or containers whose Java options are loaded from a ConfigMap or Secret.

#### Verify endpoint scraping

After labeling pods and adding the annotations, verify that JVM metrics are being collected:

1. Confirm the label and annotations are present on the pod:

   ```shell
   kubectl get pod <pod-name> -n <namespace> \
     -o jsonpath='{.metadata.labels.workloads\.atlas\.example/jvm-metrics-scrape}'
   ```

   Expected output: `true`

2. Confirm the endpoint exposes JVM metrics by forwarding the application port and querying it. Substitute your `prometheus.io/port` and `prometheus.io/path` values for `8080` and `/actuator/prometheus`:

   ```shell
   kubectl port-forward pod/<pod-name> 8080:8080 -n <namespace>
   ```

   ```shell
   curl -s http://localhost:8080/actuator/prometheus | grep "^jvm_"
   ```

3. Check that the Workload Autoscaler's Prometheus has the target up. Forward the exporter's Prometheus service, then open `http://localhost:9090/targets` in a browser:

   ```shell
   kubectl port-forward -n atlas-agent svc/atlas-workload-autoscaler-exporter-prometheus 9090:9090
   ```

   The pod should appear under the `jvm-customer-metrics` job with state `UP`.

4. In the Atlas console, navigate to **Workload Autoscaler** > **Custom metrics** and use the [Metrics explorer](woop-custom-metrics.md#metrics-explorer) to confirm that JVM metrics (heap usage, GC behavior, thread count) are flowing for the workload. Because these metrics flow through the automatically created [Default Exporter](woop-custom-metrics.md#default-exporter), no external data source connection is required.

#### Enable or disable endpoint scraping

The scrape job is controlled cluster-wide by the `prometheus.scrapeJVMMetrics` Helm value of the `atlas-workload-autoscaler-exporter` chart (default `true`). Set it explicitly to enable the job on clusters upgraded from an earlier chart version, or set it to `false` to disable the job.

> ⚠️ Configuration-only changes
>
> This command upgrades all components in the umbrella chart, not just Workload Autoscaler. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler-exporter.prometheus.scrapeJVMMetrics=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler-exporter.prometheus.scrapeJVMMetrics=true
```

  </Tab>
</Tabs>

Set the value to `false` to disable the scrape job cluster-wide.

## Prerequisites

Before JVM optimization produces metrics and heap-based recommendations, the following must be in place:

| Component                             | Minimum version | Description                                                    |
| ------------------------------------- | --------------- | -------------------------------------------------------------- |
| `atlas-workload-autoscaler`          | —               | Main component that generates and applies recommendations      |
| `atlas-workload-autoscaler-exporter` | `0.85.0`        | Collects custom metrics from the cluster                       |
| `atlas-agent`                        | `0.113.0`       | Cluster agent that communicates with the Atlas control plane |

For upgrade instructions, see [Workload Autoscaler configuration](workload-autoscaling-configuration.md).

<Accordion title="Verify installed component versions" icon="fa-circle-check">

Run the following commands to check that the required components are installed and running:

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler
```

```shell
kubectl get pods -n atlas-agent -l app.kubernetes.io/name=atlas-workload-autoscaler-exporter
```

```shell
kubectl get deploy -n atlas-agent atlas-agent \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

All pods should be in `Running` status.

</Accordion>

Additionally:

1. A **[custom metrics data source](woop-custom-metrics.md)** feeding JVM metrics. If you bring your own Prometheus, connect it in the Atlas console. With auto-instrumentation or [endpoint scraping](#scrape-existing-jvm-metrics-endpoints), the Workload Autoscaler routes metrics through the automatically created [Default Exporter](woop-custom-metrics.md#default-exporter) instead — no connection step is needed.
2. A scaling policy with **JVM heap scaling** enabled
3. **At least one JVM metrics source** feeding the Workload Autoscaler — auto-instrumentation enabled in the same scaling policy, [endpoint scraping](#scrape-existing-jvm-metrics-endpoints) (label plus `prometheus.io/*` annotations), or your own Prometheus metrics pipeline connected as a data source
4. If you use auto-instrumentation, affected pods **restarted** so the JMX agent is injected (not required for endpoint scraping or an external Prometheus pipeline)

> ⚠️ Warning
>
> Enabling JVM heap scaling with **no** JVM metrics source — no auto-instrumentation, no [endpoint scraping](#scrape-existing-jvm-metrics-endpoints), and no external Prometheus pipeline — results in missing metrics errors for JVM. At least one collection method must be active for the feature to work.

## Enable JVM optimization

### Enable via the Atlas console

1. Ensure a Prometheus data source is connected. Navigate to **Workload Autoscaler** > **Custom metrics** and verify that a data source is connected (or click **Add data source** to connect one). See [Custom metrics](woop-custom-metrics.md) for details.

2. Navigate to **Workload Autoscaler** > **Scaling policies**

3. Select an existing policy or [create a new one](woop-scaling-policies-create.md)

4. Click **Custom metrics** in the policy settings sidebar

5. Under **JVM heap scaling**, enable **Scale JVM memory**

6. Under **Auto-instrumentation**, enable **Auto-instrument JVM workloads**

   <Image src="https://files.readme.io/2384f435c7b32675d377e481c088e22f8fb0b83d34824d83f901549c8a092f78-image.png" align="left" />

7. Configure the remaining policy settings and save your policy. See [Create a scaling policy](woop-scaling-policies-create.md) for details.

8. Assign JVM workloads to this policy using [assignment rules](woop-scaling-policies-create.md#configure-assignment-rules) or manual assignment

9. Restart affected pods to trigger injection:

```shell
kubectl rollout restart deployment/<name> -n <namespace>
```

### Helm configuration

<Callout icon="📘" theme="info">
  ### Which upgrade method to use

  The Helm commands on this page use the **umbrella chart** (`atlas-helm/atlas`) by default. If you need to use a different method:

  - **[atlasctl](connect-with-atlasctl.md)**: To upgrade all Atlas components at once without managing Helm flags:
    ```shell
    atlasctl atlasware upgrade
    ```
    This preserves your existing configuration. See the [atlasctl documentation](connect-with-atlasctl.md) for installation and authentication instructions.
  - **Individual charts**: If you installed each component as a separate Helm release (e.g., for ArgoCD or custom GitOps), replace the release name and chart reference with the component-specific ones (e.g., `atlas-workload-autoscaler` and `atlas-helm/atlas-workload-autoscaler`) and remove the `autoscaler.atlas-workload-autoscaler.` value prefix.

  Not sure which method you used? Run `helm list -n atlas-agent`. A single release named `atlas` means umbrella chart; separate releases like `atlas-workload-autoscaler` mean individual charts.
</Callout>

<br />

Auto-instrumentation is enabled per scaling policy in the Atlas console. The following Helm value for the `atlas-workload-autoscaler` chart controls cluster-level behavior:

| Parameter             | Default | Description                                                                                                                      |
| --------------------- | ------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `jmxExporter.enabled` | `true`  | Master switch for auto-instrumentation. Set to `false` to disable injection cluster-wide, regardless of scaling policy settings. |

To disable auto-instrumentation at the cluster level:

> 📘 Configuration-only changes
>
> This command upgrades all components in the umbrella chart, not just Workload Autoscaler. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.jmxExporter.enabled=false
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.jmxExporter.enabled=false
```

  </Tab>
</Tabs>

### Enable via annotations

You can enable JVM optimization and auto-instrumentation on individual workloads using the `workloads.atlas.example/configuration` annotation. This overrides the scaling policy settings for the annotated workload.

```yaml
workloads.atlas.example/configuration: |
  vertical:
    jvm:
      enabled: true
      autoInstrument: true
```

To enable JVM optimization for a specific container in a multi-container pod:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    containers:
      my-container:
        jvm:
          enabled: true
          autoInstrument: true
```

For the full annotation schema and all available fields, see the [annotations reference](workload-autoscaler-annotations-reference.md).

### Verify auto-instrumentation

After enabling auto-instrumentation and restarting pods, verify that the injection was successful:

1. Check that the pod has the injection label:

   ```shell
   kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.metadata.labels.workloads\.atlas\.example/jmx-exporter-injected}'
   ```

   Expected output: `true`

2. Confirm that the JMX agent init container was injected:

   ```shell
   kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.spec.initContainers[*].name}'
   ```

   The output should include a `jmx-exporter-init` container alongside any other init containers in the pod.

3. Verify that JVM metrics are being exposed by forwarding the metrics port to your local machine and querying it:

   ```shell
   kubectl port-forward pod/<pod-name> 9404:9404 -n <namespace>
   ```

   Then, in a separate terminal:

   ```shell
   curl -s http://localhost:9404/metrics | head -20
   ```

4. In the Atlas console, navigate to **Workload Autoscaler** > **Optimization**, verify the **JVM** badge on the workload, and check for heap size and heap recommendation charts on the workload detail page

## When not to use auto-instrumentation

Disable auto-instrumentation (`jmxExporter.enabled: false`) in the following scenarios:

* **The application already runs a JMX agent.** The Workload Autoscaler does not detect existing JMX agents — it adds a second agent alongside the first, which can cause port conflicts or duplicate metrics. If the application already exposes JMX metrics in Prometheus format, feed them to the Workload Autoscaler through your existing Prometheus setup or use [endpoint scraping](#scrape-existing-jvm-metrics-endpoints) instead — no separate Prometheus instance required.
* `JDK_JAVA_OPTIONS`**&#x20;or&#x20;**`JAVA_TOOL_OPTIONS`**&#x20;is loaded from a ConfigMap or Secret.** The webhook appends the `-javaagent` flag to the environment variable value. If the value is sourced from a ConfigMap or Secret reference (`envFrom` or `valueFrom`), the webhook cannot modify it. Injection may appear to succeed (the pod label is applied), but the agent will not actually load.
* **A strict security policy blocks init containers.** The injection relies on an init container. If your cluster enforces a Pod Security Policy, OPA/Gatekeeper policy, or Kyverno policy that rejects pods with init containers added by the Workload Autoscaler, the pod will not start.
* **You already have Prometheus scraping JVM metrics, or use endpoint scraping.** If your monitoring stack already collects JVM heap metrics and they are flowing to the Workload Autoscaler — or you're using [endpoint scraping](#scrape-existing-jvm-metrics-endpoints) — auto-instrumentation is unnecessary.

## Known limitations

**Elasticsearch and OpenSearch are not auto-instrumented.** Workload Autoscaler detects Elasticsearch and OpenSearch workloads as JVM containers and applies heap scaling, but does not inject the JMX exporter into them. To scale Elasticsearch or OpenSearch memory with Workload Autoscaler, configure your own Prometheus metrics pipeline and connect it as a [custom metrics data source](woop-custom-metrics.md).

**Java 9+ required.** Auto-instrumentation defaults to the `JDK_JAVA_OPTIONS` environment variable, which is only available in Java 9 and later. Java 8 workloads may show the injection label (`workloads.atlas.example/jmx-exporter-injected=true`), but the agent will not load and no metrics will be produced because Java 8 does not recognize `JDK_JAVA_OPTIONS`.

Java 8 uses `JAVA_TOOL_OPTIONS` instead. The ability to configure the target environment variable per scaling policy is planned. Until then, Java 8 workloads should use an external JMX metrics pipeline instead of auto-instrumentation.

> 📘 Note
>
> If you see the injection label on a pod but no metrics are being produced, check the Java version of the workload. Java 8 workloads require manual JMX configuration rather than auto-instrumentation.

## Troubleshooting

### "Metric is missing" errors for JVM metrics

**Symptom:** The Atlas console shows "metric is missing" for metrics like `jvm.gc.pause_count`, `jvm.gc.pause_seconds`, `jvm.memory.heap.allocated`, `jvm.memory.heap.committed`, `jvm.memory.heap.used`, `jvm.memory.non_heap.used`, or `jvm.threads.live_threads`.

**Cause:** JVM heap scaling is enabled in the scaling policy, but one or more prerequisites are not met — auto-instrumentation is not enabled, pods have not been restarted, or the custom metrics data source is not configured.

**Solution:**

1. Verify that both **Scale JVM memory** and **Auto-instrument JVM workloads** are enabled in the scaling policy assigned to the workload
2. Verify that a [custom metrics data source](woop-custom-metrics.md) is connected and healthy
3. Restart affected pods:
   ```shell
   kubectl rollout restart deployment/<name> -n <namespace>
   ```
4. Wait 2-3 minutes for metrics to begin flowing

### Injection label present but no metrics

**Symptom:** The pod has the label `workloads.atlas.example/jmx-exporter-injected=true`, but port 9404 is not listening and no JVM metrics are produced.

**Cause 1: Java 8 or earlier.** The default environment variable (`JDK_JAVA_OPTIONS`) is not recognized by Java 8. The agent JAR is copied to the shared volume, but it never loads.

**Solution:** Upgrade the workload to Java 9 or later. If upgrading is not possible, set up an external JMX metrics pipeline compatible with Java 8 and feed the metrics to the Workload Autoscaler through a Prometheus data source.

**Cause 2: Java options environment variable sourced from a ConfigMap or Secret.** The webhook could not append the `-javaagent` flag because the environment variable value is a reference, not an inline string.

**Solution:** Move the environment variable value inline into the pod spec, or manually add the `-javaagent` flag to the ConfigMap or Secret value.

### Auto-instrumentation blocked by security policy

**Symptom:** Pods fail admission with errors referencing the init container injected by the Workload Autoscaler.

**Cause:** Cluster security policies (Pod Security Policy, OPA/Gatekeeper, Kyverno) reject the init container that the Workload Autoscaler adds during admission.

**Solution:** Add an exception for the Workload Autoscaler JMX agent init container in your security policy configuration, or disable auto-instrumentation and use an external JMX metrics pipeline instead.

## Next steps

<Cards>
  <Card title="Custom metrics" href="https://docs.atlas.example/docs/woop-custom-metrics" icon="fa-chart-line">
    Configure Prometheus data sources for collecting JVM and other custom metrics.
  </Card>

  <Card title="Scaling policies" href="https://docs.atlas.example/docs/woop-scaling-policies" icon="fa-sliders">
    Create and manage scaling policies that control JVM optimization settings.
  </Card>

  <Card title="Annotations reference" href="https://docs.atlas.example/docs/workload-autoscaler-annotations-reference" icon="fa-code">
    Full annotation schema for per-workload JVM settings and all other configuration overrides.
  </Card>
</Cards>

<br />