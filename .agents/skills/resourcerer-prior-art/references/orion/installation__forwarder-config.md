# Forwarder configuration

> Configure the Orion Forwarder component


The Forwarder collects workload metrics and forwards them to Orion. This topic covers the additional configuration parameters available.

## Applying these settings

These are Helm values for the `orion` chart. Add them to a YAML values file (for example, `values.yaml`) and pass it to your Helm command.

At install time:

```shell
helm install orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  -f values.yaml
```

To apply changes after installation:

```shell
helm upgrade orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  --reset-then-reuse-values \
  -f values.yaml
```

---


**Forwarder parameters:**

- [Private container registry](#private-registry)
- [Additional scrape configs](#additional-scrape-configs)
- [Scrape interval and timeout](#scrape-interval)
- [Storage volume size](#storage-volume)
- [Pod priority and scheduling](#pod-priority)

---

## Private container registry {#private-registry}

Override the Forwarder image repository for environments that use a private container registry.

| Helm parameter              | Default value                          |
| --------------------------- | -------------------------------------- |
| `forwarder.image.repository` | `quay.io/prometheus/prometheus`        |
| `forwarder.image.tag`        | (pinned to a tested Prometheus version)|

### Description {#private-registry-description}

The Forwarder uses the Prometheus container image, separate from the Orion Agent image. When using a private registry, you must push the Prometheus image to your registry and set `forwarder.image.repository` accordingly.

For the full setup procedure, see [If you have a private container registry](installation__install-adv.md#if-you-have-a-private-container-registry).

### Example {#private-registry-example}

```yaml
forwarder:
  image:
    repository: my-registry.example.com/prometheus/prometheus
```

---

## Additional scrape configs {#additional-scrape-configs}

Extend the Forwarder's Prometheus scrape configuration with custom scrape jobs.

| Helm parameter                       | Default value |
| ------------------------------------ | ------------- |
| `forwarder.additionalScrapeConfigs`   | `[]`          |

### Description {#additional-scrape-configs-description}

`forwarder.additionalScrapeConfigs` accepts a list of Prometheus scrape config objects appended to the default scrape configuration. Use this to collect metrics from custom endpoints or JVM exporters.

### Valid values {#additional-scrape-configs-valid-values}

An array of Prometheus [scrape config][prom-scrape-config] objects.

### Example {#additional-scrape-configs-example}

```yaml
forwarder:
  additionalScrapeConfigs:
    - job_name: custom-metrics
      static_configs:
        - targets:
          - myservice.mynamespace.svc.cluster.local:9090
```

---

## Scrape interval and timeout {#scrape-interval}

Controls how frequently the Forwarder scrapes metrics and when a scrape times out.

| Helm parameter              | Default value |
| --------------------------- | ------------- |
| `forwarder.scrapeInterval`   | `30s`         |
| `forwarder.scrapeTimeout`    | `16s`         |

### Description {#scrape-interval-description}

The default values are appropriate for most environments. Increase `scrapeInterval` only if the Forwarder pod is resource-constrained; decrease it only if you need more frequent metrics.

### Valid values {#scrape-interval-valid-values}

A [Prometheus duration string](https://prometheus.io/docs/prometheus/latest/querying/basics/#float-literals-and-time-durations) (for example, `30s`, `1m`). `scrapeTimeout` must be less than `scrapeInterval`.

---

## Storage volume size {#storage-volume}

Controls the size of the Forwarder's local metrics storage volume.

| Helm parameter                 | Default value |
| ------------------------------ | ------------- |
| `forwarder.storageVolumeSize`   | `2G`          |

### Description {#storage-volume-description}

The Forwarder buffers scraped metrics locally before forwarding them to Orion. Increase this value if you observe dropped metrics due to write-ahead log overflow in high-cardinality environments.

### Valid values {#storage-volume-valid-values}

A Kubernetes [resource quantity](https://kubernetes.io/docs/reference/glossary/?all=true#term-quantity) string (for example, `2G`, `4Gi`).

---

## Pod priority and scheduling {#pod-priority}

Enables scheduling and priority configurations on Orion Forwarder pods.

| Helm parameter                | Default value | Scope                    |
| ----------------------------- | ------------- | ------------------------ |
| `priorityClassName`           | `""`          | Top-level                |
| `forwarder.priorityClassName`  | `""`          | Forwarder only (override) |

### Description {#pod-priority-description}

The top-level `priorityClassName` applies to all Orion components. Set `forwarder.priorityClassName` to override it for the Forwarder only. A non-empty per-component value **replaces** the top-level default.

`forwarder.priorityClassName` assigns an existing Pod `PriorityClass` to the `orion-forwarder` pod.

### Valid values {#pod-priority-valid-values}

A string that matches the `.metadata.name` of a `PriorityClass` defined in your cluster.

### Example {#pod-priority-example}

To apply a priority class to all Orion components:

```yaml
priorityClassName: "high-priority"
```

To override for the Forwarder only:

```yaml
forwarder:
  priorityClassName: "high-priority"
```

---

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Deprecated exporter parameters</div>


The `exporter` parameter group is deprecated and will be removed in a future
major version; use `forwarder` instead. Until then, `exporter.*` values
still work and override equivalent `forwarder.*` values.
</div>


## Related topics

- Kubernetes official documentation:
  - [PriorityClass](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#priorityclass-v1-scheduling-k8s-io)
  - [Pod Priority and Preemption](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)

[prom-scrape-config]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config
