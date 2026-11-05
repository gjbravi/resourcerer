# Webhook configuration

> Configure the webhook used by the Applier to apply recommendations at admission time


This topic covers the configuration parameters available for the Orion webhook.

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


**Webhook parameters:**

- [Enable or disable the webhook](#webhook-enabled)
- [Namespace allow and deny lists](#namespace-lists)
- [TLS configuration](#tls)
- [Failure policy](#failure-policy)

---

## Enable or disable the webhook {#webhook-enabled}

Controls whether the Applier uses a [mutating admission webhook][mutating-webhook] to apply recommendations.

| Parameter          | Default value |
| ------------------ | ------------- |
| `webhook.enabled`  | `true`        |

### Description {#webhook-enabled-description}

By default, the Applier can use a mutating admission webhook to apply recommendations. This setting also grants the additional RBAC permissions required for webhook operation to the Applier.

When disabled, recommendations are still generated and can be applied as patches, but webhook-based apply methods (`WebhookImmediateRollout`, `WebhookHybridRollout`, `WebhookDeferredRollout`) will not work.

Teams can set `live.orion.example/apply.method` on individual [namespaces] or [workloads] to use a webhook-based method. For annotation syntax details, see [Apply method][apply-method] in the Apply Settings topic.

[Continuous reconciliation][cont-rec-applier], if enabled, doesn't run for workloads using a webhook-based apply method. Webhook methods rely on admission-time mutation rather than continuous patching.

### Valid values {#webhook-enabled-valid-values}

- `true` (default)
- `false`

---

## Namespace allow and deny lists {#namespace-lists}

Restrict webhook use to specific namespaces.

| Parameter                   | Default value |
| --------------------------- | ------------- |
| `webhook.allowNamespaces`   | `[]`          |
| `webhook.denyNamespaces`    | `["kube-system"]` |

### Description {#namespace-lists-description}

You can restrict webhook use to specific namespaces:
- To allow webhook use on specific namespaces, configure the `webhook.allowNamespaces` list.
- To deny webhook use on specific namespaces, configure the `webhook.denyNamespaces` list.

If both `webhook.allowNamespaces` and `webhook.denyNamespaces` are configured, the `denyNamespaces` list is ignored and only the allow list is enforced.

If a namespace or workload uses a webhook-based apply method but the webhook is disabled or a namespace list restricts access, recommendations are still generated, but can't be applied via webhook. In this case, UI and CLI values may differ from the values on Pods.

### Valid values {#namespace-lists-valid-values}

An array of strings or regular expressions to match against namespace names.

To explicitly unset either of these values, use `null`.

### Examples {#namespace-lists-examples}

```yaml
webhook:
  denyNamespaces:
  - ".*-system"
```

or

```yaml
webhook:
  denyNamespaces: null
  allowNamespaces:
  - "monitoring"
  - "frontend"
  - ".ecommerce-*"
```

---

## TLS configuration {#tls}

Configure TLS for the webhook.

| Parameter                 | Default value |
| ------------------------- | ------------- |
| `webhook.tlsSecretName`   | `""`          |
| `webhook.caBundle`        | `""`          |

### Description {#tls-description}

By default, Orion uses an integrated cert-controller to provision and rotate TLS certificates for the webhook automatically.

- Set `webhook.tlsSecretName` to the name of an existing TLS Secret to disable the integrated cert-controller and use your own certificate.
- Set `webhook.caBundle` to provide the CA bundle directly when managing your own TLS certificates without the integrated cert-controller.

---

## Failure policy {#failure-policy}

Controls what happens if the webhook is unavailable or returns an error.

| Parameter                  | Default value |
| -------------------------- | ------------- |
| `webhook.failurePolicy`    | `Ignore`      |

### Description {#failure-policy-description}

`Ignore` (default): If the webhook fails, the admission request proceeds — workloads are created and updated normally, but recommendations won't be applied at admission time.

`Fail`: If the webhook fails, the admission request is rejected. Use with caution — a webhook outage will block Pod creation in namespaces covered by the webhook.

### Valid values {#failure-policy-valid-values}

- `Ignore` (default)
- `Fail`

## Related topics

- [Applier configuration](installation__applier-config.md)
- [Apply settings](configure__settings__apply-settings.md)

[apply-method]: /docs/configure/settings/apply-settings/#apply-method
[namespaces]: /docs/configure/annotate-namespaces/
[workloads]: /docs/configure/annotate-workloads/
[cont-rec-applier]: /docs/installation/applier-config/#continuous-rec
[mutating-webhook]: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/
