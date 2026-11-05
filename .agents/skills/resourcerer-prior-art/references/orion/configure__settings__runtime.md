# Runtime

> Max heap size settings for Java workloads


## Java runtime {#java}

You can define how Orion applies JVM max heap size recommendations to JVM containers.

The Java max heap recommendation settings apply to workloads identified as JVM workloads that Orion is able to collect JVM metrics for. For non-Java workloads, these settings will have no effect.

Orion by default assumes that Java applications use the `-XX:MaxRamPercentage` heap size management mechanism, and that changing the container's memory limit will therefore change the effective max heap size. Orion will attempt to use this relationship to optimize heap size without changing any configuration for your app.

If an app does not use `-XX:MaxRamPercentage`, you will need to adjust the [max heap patch path](#maxheap-patchpath) setting to define another way of automating the heap size.

For more information about Java recommendations, see [Java heap size][java-recs].

### Max heap patch path {#maxheap-patchpath}

Specifies where the value for Java Max Heap should be in the Orion patch.

| Annotation                                                 | Default value                |
| ---------------------------------------------------------- | ---------------------------- |
| `live.orion.example/containers.java.max-heap.patch-path`   | `-` |

#### Description

This setting controls how Orion applies its recommended max heap value to workloads.

If you don't configure a patch-path, Orion won't directly update a specific value for max heap anywhere in the workload. Instead, Orion will take the JVM's observed `MaxRamPercentage` as a constant and adjust the container's memory limit in order to change the heap size.

If you configure a patch-path, Orion will instead inject the recommended max heap value into the workload at the location you specify.

| In order to: | Set patch-path to: |
| --- | --- |
| Use the memory limit to effect changes to heap size | `-` (or None) |
| Set heap values using an environment variable | `{{ .EnvVarPath "ORION_JAVA_ARGS" }}` |
| Set heap value in a custom field (JSON pointer) | `/spec/{{ .ContainerName }}/jvm-max-heap` |

### Max heap patch format {#maxheap-patchformat}

Specifies how to format the Java Max Heap recommended value in Orion patches.

| Annotation                                                 | Default value |
| ---------------------------------------------------------- | ------------- |
| `live.orion.example/containers.java.max-heap.patch-format` | None |

#### Description

The format Orion uses when it injects a recommended value into a workload can be customized by setting a patch-format.

If you don't configure a specific patch-format, Orion's default behavior is to produce a `MaxRamPercentage` setting (`-XX:MaxRamPercentage=VALUE`) if the workload has a memory limit. If the workload does not have a memory limit, it will produce a `MaxHeapSize` setting (`-XX:MaxHeapSize=VALUE`) instead.

This setting has no effect when [patch-path](#maxheap-patchpath) is set to `-`.

| In order to: | Set patch-format to: |
| --- | --- |
| Let Orion pick how to output max heap setting | None |
| Produce a `-XX:MaxRamPercentage` setting | `{{ jvmOption "-XX:MaxRamPercentage" (round (.MemoryLimit.AsApproximateFloat64 \| divf .Value.AsApproximateFloat64 \| mulf 100) 2 1) }}` |
| Produce a `-XX:MaxHeapSize` setting | `{{- jvmOption "-XX:MaxHeapSize" .Value }}` |

### Minimum and maximum bounds for max heap {#maxheap-bounds}

Values indicate the lower bound and upper bound of the range for Java max heap.

| Annotation                                        | Default value |
| ------------------------------------------------  | ------------- |
| `live.orion.example/containers.java.max-heap.min` | `"512Mi"` |
| `live.orion.example/containers.java.max-heap.max` | None          |

#### Description

The minimum and maximum bounds for Java max heap acts a constraint on what max heap values Orion will recommend, regardless of the observed heap usage. These bounds can be used to set a floor below which the max heap should never drop and a ceiling above which Orion will never raise it.

When used as defaults for the cluster, max heap bounds can be used to implement default automation policies such as “never set max heap of less than 128Mi” or “never set max heap higher than 8192Gi”.

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/java/">Java heap size</a>
                </h5>
                <p>Max heap size recommendations for Java workloads</p>
            </div>
        </div>


[java-recs]: /docs/recommendations/java/
