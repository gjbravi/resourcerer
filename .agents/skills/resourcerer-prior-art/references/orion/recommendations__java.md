# Java heap size

> Max heap size recommendations for Java workloads


Orion recommendations can optionally include the Java max heap size that the machine learning has determined to be optimal for each Java container in a workload.

When configured, Orion automatically analyzes critical Java metrics—such as heap and non-heap usage, as well as garbage collection data—and provides tailored recommendations for heap size adjustments alongside its recommendations for the container's requests and limits.

## Enabling heap size recommendations

To create Java Heap size recommendations, the Orion Agent must be configured to collect JMX metrics. Use the `jvmWorkloadConfigs` Helm parameter to accomplish this.

Typically, JMX metrics are exposed by using either [Prometheus Java Metrics][java-library] (Java library) or the [Prometheus JMX Exporter][jmx-exporter] (Java agent). However, as long as the method you use exposes the metrics in OpenMetrics format, you can use the `jvmWorkloadConfigs` Helm parameter described below.

To identify the Java pods to optimize, Orion uses a label selector comprised of the following:

* A Pod label selector that identifies the Java pods to optimize
* The port name or port number for each container that exports metrics in OpenMetrics format
* The URL path to scrape from that port (for example, `/metrics`)

Prerequisites:

* Helm version 3.14.0 or later (required for the `--reset-then-reuse parameter`)
* Orion Agent version 2.20.2 or later

### Steps

1. To enable the Agent to collect metrics and generate Java Heap size recommendations, use the  `jvmWorkloadConfigs` Helm parameter.

    You can copy the following template into a file named, for example, jvm-config.yaml, and then update it to suit your needs:

    ```yaml
    ---
    jvmWorkloadConfigs:
    - labelSelector: "example.com/jvm-runtime=true,example.com/environment in (dev, prod)"
       metricsPort:
         scheme: ""      # Optional string. Scheme port uses. Default is "http"
         matchNumber: 0  # Optional number. Matches port number to scrape metrics from
         matchName: ""   # Optional string. Matches port name to scrape metrics from
       metricsPath: ""   # Default is "/metrics"
    ```

2. Run `helm upgrade` to apply the setting:

    ```sh
    helm upgrade orion oci://registry.orion.example/library/orion \
      -n orion-system \
      --reset-then-reuse-values \
      -f jvm-config.yaml
    ```

### Example

This example assumes that your Java pods are:

* Labeled with `example.com/jvm-runtime: "true"`
* Exporting their metrics on a port named `web` to the `/metrics` URL path

The `jvmWorkloadConfigs` parameter in your `jvm-config.yaml` file would look like this:

```yaml
# Helm values to enable Orion Java Heap Size recommendations (beta)
jvmWorkloadConfigs:
- labelSelector: "example.com/jvm-runtime=true"
  metricsPort:
    matchName: "web"
  metricsPath: "/jvm-metrics"
```

## Applying heap size recommendations

Orion can apply heap size recommendations either by:

1. **Using the container memory limit to adjust max heap size** – works if your app is configured with `-XX:MaxRamPercentage`
2. **Setting an environment variable** – requires knowledge about how your app is configured

### Using the container memory limit to adjust max heap size

<div class="alert alert-info" role="alert">


**Note:** This is Orion's default mode, but it is still advised that you configure the behavior directly for internal visibility.
</div>


If the JVM MaxRamPercentage setting used by the app is taken as a constant, changing the memory limit will effectively change the max heap size too.

Configure cluster-defaults according to this example.

```yaml
live.orion.example/containers.java.max-heap.patch-path: "-" # Set heap using memory limit
live.orion.example/containers.memory.optimization-policy: runtime:java=RequestsAndLimits
```

Orion will calculate a recommended value for max heap. Rather than including the max heap value directly in the patch, the container's memory limit will be set such that given the observed MaxRamPercentage, the desired max heap will be achieved.

In this case, no additional knowledge about the app's configuration is needed to automatically manage the heap size.

While it is not strictly necessary to change the memory optimization policy, it is recommended to use RequestsAndLimits for memory when setting heap size via memory limits. Other memory optimization policies may have constraints that prevent Orion from being able to fully implement the desired heap size using this method.

### Setting an environment variable

Configure cluster-defaults according to this example.

```yaml
live.orion.example/containers.java.max-heap.patch-path: '{{ .EnvVarPath "ORION_JAVA_ARGS" }}'
live.orion.example/containers.java.max-heap.patch-format: '{{- jvmOption "-XX:MaxHeapSize" .Value }}'
```

If you do nothing more, Orion will include the recommended max heap value in a patch thusly:

```yaml
spec:
  containers:
  - name: my-jvm-container
    env:
    - name: ORION_JAVA_ARGS
      value: -XX:MaxHeapSize=741m
```

Note that however you configure this environment variable, you must ensure that your application is configured to use it in order to ensure the value is actually consumed and takes effect.

## Optimization policies and limit-to-request ratio (LRR)

When `-` is configured as the Java max heap patch-path, this changes Orion's default calculation for the base (pre-policy and pre-constraints) recommended memory limit.

Normally, Orion calculates a base recommended memory limit based on the configured limit-to-request ratio (LRR).

```
limit = recommended-request ⨉ limit-request-ratio
```

For Java containers that have their Java max-heap patch-path set to `-`, Orion will change its base memory limit calculation to be:

```
limit = recommended-max-heap ÷ java-max-ram-percentage
```

All other characteristics and constraints of the selected memory optimization policy will still apply, including minimum and maximum bounds. This may mean that implementing the complete Java max heap recommendation is not possible, if the constraints or the optimization policy prevent setting the memory limit to exactly the desired value.

For the RequestsRaiseLimitsIfNeeded policy in particular, note that Orion will still never lower memory limits for Java containers.

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span>Settings<span class="related-topics-divider">/</span><a href="/docs/configure/settings/runtime/">Runtime</a>
                </h5>
                <p>Max heap size settings for Java workloads</p>
            </div>
        </div>


[java-library]: https://prometheus.github.io/client_java/
[jmx-exporter]: https://prometheus.github.io/jmx_exporter/
[java-settings]: /docs/configure/settings/runtime/#java
